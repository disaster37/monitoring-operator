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
	routev1 "github.com/openshift/api/route/v1"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteReconciler reconciles a Route object
type RouteCentreonReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons,verbs=get;list;watch
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
func (r *RouteCentreonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, "", r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	route := &routev1.Route{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, route, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouteCentreonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&routev1.Route{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewResourceWithMonitoringAnnotationPredicate()).
		Complete(r)
}

// Configure do nothink here
func (r *RouteCentreonReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected Centreon service that reflect route
func (r *RouteCentreonReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	route := resource.(*routev1.Route)

	// Get global Centreon Spec
	centreonSpec, err := getCentreonSpec(ctx, r.Client)
	if err != nil {
		return res, errors.Wrap(err, "Error when get Centreon Spec")
	}
	if centreonSpec == nil {
		r.log.Warning("It's recommanded to set some default values on custom resource called `Centreon` on the same operator namespace. It avoid to set on route all Centreon service properties as annotations")
	}

	// Get if current CentreonService object already exist
	currentCS := &v1alpha1.CentreonService{}
	err = r.Get(ctx, types.NamespacedName{Name: route.Name, Namespace: route.Namespace}, currentCS)
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
			Name:        route.Name,
			Namespace:   route.Namespace,
			Labels:      route.GetLabels(),
			Annotations: route.GetAnnotations(),
		},
		Spec: monitorv1alpha1.CentreonServiceSpec{},
	}
	placeholders := generatePlaceholdersRouteCentreonService(route)
	initCentreonServiceDefaultValue(centreonSpec, expectedCS, placeholders)
	if err = initCentreonServiceFromAnnotations(route.GetAnnotations(), expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when init CentreonService from route annotations")
	}

	// Check CentreonService is valide
	if !expectedCS.IsValid() {
		return res, fmt.Errorf("Generated CentreonService is not valid: %+v", expectedCS.Spec)
	}

	// Set route instance as the owner
	ctrl.SetControllerReference(route, expectedCS, r.Scheme)

	data["expectedCentreonService"] = expectedCS

	return res, nil

}

// Create add new CentreonService object
func (r *RouteCentreonReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
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
func (r *RouteCentreonReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
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
func (r *RouteCentreonReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *RouteCentreonReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
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
func (r *RouteCentreonReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *RouteCentreonReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

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
func generatePlaceholdersRouteCentreonService(r *routev1.Route) (placeholders map[string]string) {
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
