/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	IngressFinalizer = "ingress.monitor.k8s.webcenter.fr/finalizer"
)

// IngressReconciler reconciles a Ingress object
type IngressReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	CentreonController
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="monitor.k8s.webcenter.fr",resources=templatecentreonservices,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, IngressFinalizer, r.reconciler, r.Reconciler.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	ingress := &networkv1.Ingress{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, ingress, data)

}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkv1.Ingress{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewResourceWithMonitoringTemplate()).
		Watches(&source.Kind{Type: &v1alpha1.TemplateCentreonService{}}, handler.EnqueueRequestsFromMapFunc(watchCentreonTemplate(r.Client))).
		Complete(r)
}

// Configure do nothink here
func (r *IngressReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected monitoring service that reflect ingress
func (r *IngressReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	ingress := resource.(*networkv1.Ingress)

	_, platform, err := getClient(ingress.Annotations[fmt.Sprintf("%s/platform-ref", centreonMonitoringAnnotationKey)], r.platforms)
	if err != nil {
		return res, errors.Wrap(err, "Error when get platform")
	}
	data["platform"] = platform

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.CentreonController.readTemplatingCentreonService(ctx, ingress, data, meta, generatePlaceholdersIngress(ingress))
	default:
		return res, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}

}

// Create add new service object
func (r *IngressReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return res, err
	}
	platform := d.(*v1alpha1.Platform)

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.CentreonController.createOrUpdateCentreonServiceFromTemplate(ctx, resource, data, meta)
	default:
		return res, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}
}

// Update permit to update service object
func (r *IngressReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *IngressReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *IngressReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return diff, err
	}
	platform := d.(*v1alpha1.Platform)

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.CentreonController.diffCentreonService(resource, data, meta)
	default:
		return diff, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}
}

// OnError permit to set status condition on the right state and record error
func (r *IngressReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.Reconciler.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *IngressReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

	if diff.NeedCreate {
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Create CentreonService successfully")
		return nil
	}

	if diff.NeedUpdate {
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Update CentreonService successfully")
		return nil
	}

	return nil
}

// It generate map of placeholders from ingress spec
func generatePlaceholdersIngress(i *networkv1.Ingress) (placeholders map[string]any) {
	placeholders = map[string]any{}
	if i == nil {
		return placeholders
	}

	//Main properties
	placeholders["name"] = i.Name
	placeholders["namespace"] = i.Namespace
	placeholders["labels"] = i.GetLabels()
	placeholders["annotations"] = i.GetAnnotations()

	// Ingress properties
	rules := make([]map[string]any, 0, len(i.Spec.Rules))
	for _, rule := range i.Spec.Rules {
		r := map[string]any{
			"host":   rule.Host,
			"scheme": "http",
		}

		// Check if scheme is https
		for _, tls := range i.Spec.TLS {
			for _, host := range tls.Hosts {
				if host == rule.Host {
					r["scheme"] = "https"
				}
			}
		}

		// Add path
		paths := make([]string, 0, len(rule.HTTP.Paths))
		for _, path := range rule.HTTP.Paths {
			paths = append(paths, path.Path)
		}
		r["paths"] = paths
		rules = append(rules, r)
	}
	placeholders["rules"] = rules

	return placeholders

}
