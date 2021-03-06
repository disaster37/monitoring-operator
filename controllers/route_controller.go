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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Route object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *RouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, "", r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	route := &routev1.Route{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, route, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&routev1.Route{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewResourceWithMonitoringAnnotationPredicate()).
		Complete(r)
}

// Configure do nothink here
func (r *RouteReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected monitoring service that reflect route
func (r *RouteReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	route := resource.(*routev1.Route)

	_, platform, err := getClient(route.Annotations[fmt.Sprintf("%s/platform-ref", centreonMonitoringAnnotationKey)], r.platforms)
	if err != nil {
		return res, errors.Wrap(err, "Error when get platform")
	}
	data["platform"] = platform

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.readForCentreonPlatform(ctx, route, platform, data, meta)
	default:
		return res, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}

}

// Create add new monitoring service object
func (r *RouteReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return res, err
	}
	platform := d.(*v1alpha1.Platform)

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.createForCentreonPlatform(ctx, resource, data, meta)
	default:
		return res, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}
}

// Update permit to update monitoring service object
func (r *RouteReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return res, err
	}
	platform := d.(*v1alpha1.Platform)

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.updateForCentreonPlatform(ctx, resource, data, meta)
	default:
		return res, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *RouteReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *RouteReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return diff, err
	}
	platform := d.(*v1alpha1.Platform)

	switch platform.Spec.PlatformType {
	case "centreon":
		return r.diffForCentreonPlatform(resource, data, data)
	default:
		return diff, errors.Errorf("Platform of type %s is not supported", platform.Spec.PlatformType)
	}
}

// OnError permit to set status condition on the right state and record error
func (r *RouteReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *RouteReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

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

// It generate map of placeholders from route spec
func generatePlaceholdersRoute(r *routev1.Route) (placeholders map[string]string) {
	placeholders = map[string]string{}
	if r == nil {
		return placeholders
	}

	//Main properties
	placeholders["name"] = r.Name
	placeholders["namespace"] = r.Namespace

	// Labels properties
	for key, value := range r.GetLabels() {
		placeholders[fmt.Sprintf("label.%s", key)] = value
	}

	// Annotations properties
	for key, value := range r.GetAnnotations() {
		placeholders[fmt.Sprintf("annotation.%s", key)] = value
	}

	// Route properties
	placeholders["rule.0.host"] = r.Spec.Host
	if r.Spec.Path != "" {
		placeholders["rule.0.path"] = r.Spec.Path
	} else {
		placeholders["rule.0.path"] = "/"
	}
	if r.Spec.TLS != nil && r.Spec.TLS.Termination != "" {
		placeholders["rule.0.scheme"] = "https"
	} else {
		placeholders["rule.0.scheme"] = "http"
	}

	return placeholders

}
