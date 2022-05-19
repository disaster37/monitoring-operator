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
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// IngressReconciler reconciles a Ingress object
type IngressCentreonReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
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
func (r *IngressCentreonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, "", r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	ingress := &networkv1.Ingress{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, ingress, data)

}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressCentreonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkv1.Ingress{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewResourceWithMonitoringAnnotationPredicate()).
		Complete(r)
}

// Configure do nothink here
func (r *IngressCentreonReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected Centreon service that reflect ingress
func (r *IngressCentreonReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	ingress := resource.(*networkv1.Ingress)

	// Get global Centreon Spec
	centreonSpec, err := getCentreonSpec(ctx, r.Client)
	if err != nil {
		return res, errors.Wrap(err, "Error when get Centreon Spec")
	}
	if centreonSpec == nil {
		r.log.Warning("It's recommanded to set some default values on custom resource called `Centreon` on the same operator namespace. It avoid to set on each ingress all Centreon service properties as annotations")
	}

	// Get if current CentreonService object already exist
	currentCS := &v1alpha1.CentreonService{}
	err = r.Get(ctx, types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, currentCS)
	if err != nil && k8serrors.IsNotFound(err) {
		data["currentCentreonService"] = nil
	} else if err != nil {
		return res, errors.Wrap(err, "Error when get current CentreonService object")
	} else {
		data["currentCentreonService"] = currentCS
	}

	// Compute expected Centreon service
	expectedCS := &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        ingress.Name,
			Namespace:   ingress.Namespace,
			Labels:      ingress.GetLabels(),
			Annotations: ingress.GetAnnotations(),
		},
		Spec: monitorv1alpha1.CentreonServiceSpec{},
	}
	placeholders := generatePlaceholdersIngressCentreonService(ingress)
	initCentreonServiceDefaultValue(centreonSpec, expectedCS, placeholders)
	if err = initCentreonServiceFromAnnotations(ingress.GetAnnotations(), expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when init CentreonService from ingress annotations")
	}

	// Check CentreonService is valide
	if !expectedCS.IsValid() {
		return res, fmt.Errorf("Generated CentreonService is not valid: %+v", expectedCS.Spec)
	}

	// Set ingress instance as the owner
	ctrl.SetControllerReference(ingress, expectedCS, r.Scheme)

	data["expectedCentreonService"] = expectedCS

	return res, nil

}

// Create add new CentreonService object
func (r *IngressCentreonReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return res, err
	}
	expectedCS := d.(*v1alpha1.CentreonService)

	if err = r.Client.Create(ctx, expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when create CentreonService object")
	}

	return res, nil
}

// Update permit to update CentreonService object
func (r *IngressCentreonReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	var d any

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return res, err
	}
	expectedCS := d.(*v1alpha1.CentreonService)

	if err = r.Client.Update(ctx, expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when update CentreonService object")
	}

	return res, nil
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *IngressCentreonReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *IngressCentreonReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	var (
		d          any
		currentCS  *v1alpha1.CentreonService
		expectedCS *v1alpha1.CentreonService
	)

	d, err = helper.Get(data, "currentCentreonService")
	if err != nil {
		return diff, err
	}
	if d == nil {
		currentCS = nil
	} else {
		currentCS = d.(*v1alpha1.CentreonService)
	}

	d, err = helper.Get(data, "expectedCentreonService")
	if err != nil {
		return diff, err
	}
	expectedCS = d.(*v1alpha1.CentreonService)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	if currentCS == nil {
		diff.NeedCreate = true
		diff.Diff = "CentreonService object not exist"
		return diff, nil
	}

	diffSpec := cmp.Diff(currentCS.Spec, expectedCS.Spec)
	diffLabels := cmp.Diff(currentCS.GetLabels(), expectedCS.GetLabels())
	diffAnnotations := cmp.Diff(currentCS.GetAnnotations(), expectedCS.GetAnnotations())
	if diffSpec != "" || diffLabels != "" || diffAnnotations != "" {
		diff.NeedUpdate = true
		diff.Diff = fmt.Sprintf("%s\n%s\n%s", diffLabels, diffAnnotations, diffSpec)

		currentCS.SetLabels(expectedCS.GetLabels())
		currentCS.SetAnnotations(expectedCS.GetAnnotations())
		currentCS.Spec = expectedCS.Spec
		data["expectedCentreonService"] = currentCS
	}

	return
}

// OnError permit to set status condition on the right state and record error
func (r *IngressCentreonReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *IngressCentreonReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

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
func generatePlaceholdersIngressCentreonService(i *networkv1.Ingress) (placeholders map[string]string) {
	placeholders = map[string]string{}
	if i == nil {
		return placeholders
	}

	//Main properties
	placeholders["name"] = i.Name
	placeholders["namespace"] = i.Namespace

	// Labels properties
	for key, value := range i.GetLabels() {
		placeholders[fmt.Sprintf("label.%s", key)] = value
	}

	// Annotations properties
	for key, value := range i.GetAnnotations() {
		placeholders[fmt.Sprintf("annotation.%s", key)] = value
	}

	// Ingress properties
	for j, rule := range i.Spec.Rules {
		placeholders[fmt.Sprintf("rule.%d.host", j)] = rule.Host

		// Check if scheme is http or https
		placeholders[fmt.Sprintf("rule.%d.scheme", j)] = "http"
		for _, tls := range i.Spec.TLS {
			for _, host := range tls.Hosts {
				if host == rule.Host {
					placeholders[fmt.Sprintf("rule.%d.scheme", j)] = "https"
				}
			}
		}

		// Add path
		for k, path := range rule.HTTP.Paths {
			placeholders[fmt.Sprintf("rule.%d.path.%d", j, k)] = path.Path
		}
	}

	return placeholders

}
