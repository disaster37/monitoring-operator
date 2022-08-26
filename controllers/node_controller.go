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

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	NodeFinalizer = "node.monitor.k8s.webcenter.fr/finalizer"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	TemplateController
	name string
}

func NewNodeReconciler(client client.Client, scheme *runtime.Scheme, templateController TemplateController) *NodeReconciler {

	r := &NodeReconciler{
		Client:             client,
		Scheme:             scheme,
		TemplateController: templateController,
		name:               "node",
	}

	controllerMetrics.WithLabelValues(r.name).Add(0)

	return r
}

//+kubebuilder:rbac:groups="",resources=nodes,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=nodes/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="monitor.k8s.webcenter.fr",resources=templates,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, NodeFinalizer, r.reconciler, r.Reconciler.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	node := &core.Node{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, node, data)

}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&core.Node{}).
		Owns(&monitorapi.CentreonService{}).
		Owns(&monitorapi.CentreonServiceGroup{}).
		WithEventFilter(viewResourceWithMonitoringTemplate()).
		Watches(&source.Kind{Type: &monitorapi.Template{}}, handler.EnqueueRequestsFromMapFunc(watchTemplate(r.Client))).
		Complete(r)
}

// Configure do nothink here
func (r *NodeReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected monitoring service that reflect node
func (r *NodeReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	node := resource.(*core.Node)
	return r.TemplateController.readTemplating(ctx, node, data, meta, generatePlaceholdersNode(node))
}

// Create add new service object
func (r *NodeReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.TemplateController.createOrUpdateRessourcesFromTemplate(ctx, resource, data, meta)
}

// Update permit to update service object
func (r *NodeReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *NodeReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *NodeReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	return r.TemplateController.diffRessourcesFromTemplate(resource, data, meta)
}

// OnError permit to set status condition on the right state and record error
func (r *NodeReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.Reconciler.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

	// Update metrics
	totalErrors.Inc()

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *NodeReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

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

// It generate map of placeholders from node
func generatePlaceholdersNode(n *core.Node) (placeholders map[string]any) {
	placeholders = map[string]any{}
	if n == nil {
		return placeholders
	}

	//Main properties
	placeholders["name"] = n.Name
	placeholders["labels"] = n.GetLabels()
	placeholders["annotations"] = n.GetAnnotations()
	placeholders["nodeInfo"] = n.Status.NodeInfo
	placeholders["addresses"] = n.Status.Addresses
	placeholders["unschedulable"] = n.Spec.Unschedulable

	return placeholders

}
