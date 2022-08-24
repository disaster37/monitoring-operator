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
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	CentreonServiceGroupFinalizer = "servicegroup.monitor.k8s.webcenter.fr/finalizer"
	CentreonServiceGroupCondition = "UpdateCentreonServiceGroup"
)

// CentreonServiceGroupReconciler reconciles a CentreonServiceGroup object
type CentreonServiceGroupReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	name   string
}

func NewCentreonServiceGroupReconciler(client client.Client, scheme *runtime.Scheme) *CentreonServiceGroupReconciler {

	r := &CentreonServiceGroupReconciler{
		Client: client,
		Scheme: scheme,
		name:   "centreonservicegroup",
	}

	controllerMetrics.WithLabelValues(r.name).Add(0)

	return r
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservicegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservicegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservicegroups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CentreonServiceGroup object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.1/pkg/reconcile
func (r *CentreonServiceGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, CentreonServiceGroupFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	csg := &v1alpha1.CentreonServiceGroup{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, csg, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonServiceGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(r.name).
		For(&v1alpha1.CentreonServiceGroup{}).
		Complete(r)
}

// Configure permit to init condition and choose the right client
func (r *CentreonServiceGroupReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	csg := resource.(*v1alpha1.CentreonServiceGroup)

	// Init condition status if not exist
	if condition.FindStatusCondition(csg.Status.Conditions, CentreonServiceGroupCondition) == nil {
		condition.SetStatusCondition(&csg.Status.Conditions, v1.Condition{
			Type:   CentreonServiceGroupCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get Centreon client
	meta, _, err = getClient(csg.Spec.PlatformRef, r.platforms)
	return meta, err

}

// Read permit to get current serviceGroup from Centreon
func (r *CentreonServiceGroupReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	csg := resource.(*v1alpha1.CentreonServiceGroup)
	cHandler := meta.(centreonhandler.CentreonHandler)

	// Check if the current serviceGroup name is right before to search on Centreon
	// Maybee we should to change it name
	var (
		serviceGroupName string
	)
	if csg.Status.ServiceGroupName != "" {
		serviceGroupName = csg.Status.ServiceGroupName
	} else {
		serviceGroupName = csg.Spec.Name
	}

	// Update status in any case
	csg.Status.ServiceGroupName = csg.Spec.Name

	actualCSG, err := cHandler.GetServiceGroup(serviceGroupName)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get ServiceGroup from Centreon")
	}

	data["currentServiceGroup"] = actualCSG
	return res, nil
}

// Create add new ServiceGroup on Centreon
func (r *CentreonServiceGroupReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	csg := resource.(*v1alpha1.CentreonServiceGroup)

	// Check policy
	if csg.Spec.Policy.NoCreate {
		r.log.Info("Skip create serviceGroup (policy NoCreate)")
		return res, nil
	}

	// Create serviceGroup on Centreon
	expectedCSG, err := csg.ToCentreonServiceGroup()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to Centreon ServiceGroup")
	}
	if err = cHandler.CreateServiceGroup(expectedCSG); err != nil {
		return res, errors.Wrap(err, "Error when create serviceGroup on Centreoon")
	}

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Inc()

	return res, nil
}

// Update permit to update serviceGroup on Centreon
func (r *CentreonServiceGroupReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	csg := resource.(*v1alpha1.CentreonServiceGroup)
	var d any

	// Check policy
	if csg.Spec.Policy.NoUpdate {
		r.log.Info("Skip update serviceGroup (policy NoUpdate)")
		return res, nil
	}

	d, err = helper.Get(data, "expectedServiceGroup")
	if err != nil {
		return res, err
	}
	expectedServiceGroup := d.(*centreonhandler.CentreonServiceGroupDiff)

	if err = cHandler.UpdateServiceGroup(expectedServiceGroup); err != nil {
		return res, errors.Wrap(err, "Error when update serviceGroup on Centreoon")
	}

	return res, nil
}

// Delete permit to delete serviceGroup from Centreon
func (r *CentreonServiceGroupReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	csg := resource.(*v1alpha1.CentreonServiceGroup)

	// Check policy
	if csg.Spec.Policy.NoDelete {
		r.log.Info("Skip delete serviceGroup (policy NoDelete)")
		return nil
	}

	actualCSG, err := cHandler.GetServiceGroup(csg.Spec.Name)
	if err != nil {
		return err
	}

	if actualCSG == nil {
		r.log.Info("ServiceGroup already deleted on Centreon by external process, skip it")
		return nil
	}

	if err = cHandler.DeleteServiceGroup(csg.Spec.Name); err != nil {
		return errors.Wrap(err, "Error when delete serviceGroup from Centreon")
	}

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil

}

// Diff permit to check if diff between actual and expected Centreon serviceGroup exist
func (r *CentreonServiceGroupReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	csg := resource.(*v1alpha1.CentreonServiceGroup)
	var d any

	expectedCSG, err := csg.ToCentreonServiceGroup()
	if err != nil {
		return diff, errors.Wrap(err, "Error when convert to Centreon serviceGroup")
	}

	d, err = helper.Get(data, "currentServiceGroup")
	if err != nil {
		return diff, err
	}
	currentServiceGroup := d.(*centreonhandler.CentreonServiceGroup)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	if currentServiceGroup == nil {
		diff.NeedCreate = true
		diff.Diff = "ServiceGroup not exist on Centreon"
		return diff, nil
	}

	currentDiff, err := cHandler.DiffServiceGroup(currentServiceGroup, expectedCSG, csg.Spec.Policy.ExcludeFieldsOnDiff)
	if err != nil {
		return diff, errors.Wrap(err, "Error when diff Centreon serviceGroup")
	}
	if currentDiff.IsDiff {
		diff.NeedUpdate = true
		diff.Diff = currentDiff.String()
	}

	data["expectedServiceGroup"] = currentDiff

	return
}

// OnError permit to set status condition on the right state and record error
func (r *CentreonServiceGroupReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	csg := resource.(*v1alpha1.CentreonServiceGroup)

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&csg.Status.Conditions, v1.Condition{
		Type:    CentreonServiceGroupCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})

	// Update metrics
	totalErrors.Inc()
}

// OnSuccess permit to set status condition on the right state is everythink is good
func (r *CentreonServiceGroupReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	csg := resource.(*v1alpha1.CentreonServiceGroup)

	if diff.NeedCreate {
		r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "ServiceGroup %s successfully created on Centreon", csg.Spec.Name)
	}

	if diff.NeedUpdate {
		r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "ServiceGroup %s successfully updated on Centreon", csg.Spec.Name)
	}

	// Update condition status if needed
	if !condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, CentreonServiceGroupCondition, v1.ConditionTrue) {
		condition.SetStatusCondition(&csg.Status.Conditions, v1.Condition{
			Type:    CentreonServiceGroupCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: fmt.Sprintf("ServiceGroup %s up to date on Centreon", csg.Spec.Name),
		})

		if !diff.NeedCreate && !diff.NeedUpdate {
			r.recorder.Event(resource, core.EventTypeNormal, "Completed", "ServiceGroup already exit on Centreon")
		}
	}

	return nil
}
