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
	routev1 "github.com/openshift/api/route/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	RouteFinalizer = "route.monitor.k8s.webcenter.fr/finalizer"
)

// RouteReconciler reconciles a Route object
type RouteReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	TemplateController
	name string
}

func NewRouteReconciler(client client.Client, scheme *runtime.Scheme, templateController TemplateController) *RouteReconciler {

	r := &RouteReconciler{
		Client:             client,
		Scheme:             scheme,
		TemplateController: templateController,
		name:               "route",
	}

	controllerMetrics.WithLabelValues(r.name).Add(0)

	return r
}

//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="route.openshift.io",resources=routes/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="monitor.k8s.webcenter.fr",resources=templatecentreonservices,verbs=get;list;watch;update;patch
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
	reconciler, err := controller.NewStdReconciler(r.Client, RouteFinalizer, r.reconciler, r.Reconciler.log, r.recorder, waitDurationWhenError)
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
		Named(r.name).
		For(&routev1.Route{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewResourceWithMonitoringTemplate()).
		Watches(&source.Kind{Type: &v1alpha1.Template{}}, handler.EnqueueRequestsFromMapFunc(watchTemplate(r.Client))).
		Complete(r)
}

// Configure do nothink here
func (r *RouteReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected monitoring service that reflect route
func (r *RouteReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	route := resource.(*routev1.Route)
	return r.TemplateController.readTemplating(ctx, route, data, meta, generatePlaceholdersRoute(route))
}

// Create add new monitoring service object
func (r *RouteReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.TemplateController.createOrUpdateRessourcesFromTemplate(ctx, resource, data, meta)
}

// Update permit to update monitoring service object
func (r *RouteReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *RouteReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	// Update prometheus mectric
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *RouteReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	return r.TemplateController.diffRessourcesFromTemplate(resource, data, meta)
}

// OnError permit to set status condition on the right state and record error
func (r *RouteReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.Reconciler.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

	// Update prometheus mectric
	totalErrors.Inc()

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
func generatePlaceholdersRoute(r *routev1.Route) (placeholders map[string]any) {
	placeholders = map[string]any{}
	if r == nil {
		return placeholders
	}

	//Main properties
	placeholders["name"] = r.Name
	placeholders["namespace"] = r.Namespace
	placeholders["labels"] = r.GetLabels()
	placeholders["annotations"] = r.GetAnnotations()

	// Set route placeholders on same format as ingress
	rules := make([]map[string]any, 0, 1)
	rule := map[string]any{
		"host": r.Spec.Host,
	}
	if r.Spec.Path != "" {
		rule["paths"] = []string{r.Spec.Path}
	} else {
		rule["paths"] = []string{"/"}
	}
	if r.Spec.TLS != nil && r.Spec.TLS.Termination != "" {
		rule["scheme"] = "https"
	} else {
		rule["scheme"] = "http"
	}

	rules = append(rules, rule)
	placeholders["rules"] = rules

	return placeholders

}
