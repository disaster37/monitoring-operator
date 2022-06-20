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
	CentreonServicedFinalizer = "service.monitor.k8s.webcenter.fr/finalizer"
	CentreonServiceCondition  = "UpdateCentreonService"
)

// CentreonServiceReconciler reconciles a CentreonService object
type CentreonServiceReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the CentreonService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CentreonServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	reconciler, err := controller.NewStdReconciler(r.Client, CentreonServicedFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	cs := &v1alpha1.CentreonService{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, cs, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CentreonService{}).
		Complete(r)
}

// Configure permit to init condition and choose the right client
func (r *CentreonServiceReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	cs := resource.(*v1alpha1.CentreonService)

	// Init condition status if not exist
	if condition.FindStatusCondition(cs.Status.Conditions, CentreonServiceCondition) == nil {
		condition.SetStatusCondition(&cs.Status.Conditions, v1.Condition{
			Type:   CentreonServiceCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	// Get Centreon client
	meta, _, err = getClient(cs.Spec.PlatformRef, r.platforms)
	return meta, err

}

// Read permit to get current service from Centreon
func (r *CentreonServiceReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	cs := resource.(*v1alpha1.CentreonService)
	cHandler := meta.(centreonhandler.CentreonHandler)

	// Check if the current service name and host is right before to search on Centreon
	var (
		host        string
		serviceName string
	)
	if cs.Status.Host != "" && cs.Status.ServiceName != "" {
		host = cs.Status.Host
		serviceName = cs.Status.ServiceName
	} else {
		host = cs.Spec.Host
		serviceName = cs.Spec.Name
	}

	actualCS, err := cHandler.GetService(host, serviceName)
	if err != nil {
		return res, errors.Wrap(err, "Unable to get Service from Centreon")
	}

	data["currentService"] = actualCS
	return res, nil
}

// Create add new Service on Centreon
func (r *CentreonServiceReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	cs := resource.(*v1alpha1.CentreonService)

	// Create service on Centreon
	expectedCS, err := cs.ToCentreoonService()
	if err != nil {
		return res, errors.Wrap(err, "Error when convert to Centreon Service")
	}
	if err = cHandler.CreateService(expectedCS); err != nil {
		return res, errors.Wrap(err, "Error when create service on Centreoon")
	}

	cs.Status.Host = cs.Spec.Host
	cs.Status.ServiceName = cs.Spec.Name

	return res, nil
}

// Update permit to update service on Centreon
func (r *CentreonServiceReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	cs := resource.(*v1alpha1.CentreonService)
	var d any

	d, err = helper.Get(data, "expectedService")
	if err != nil {
		return res, err
	}
	expectedService := d.(*centreonhandler.CentreonServiceDiff)

	if err = cHandler.UpdateService(expectedService); err != nil {
		return res, errors.Wrap(err, "Error when create service on Centreoon")
	}

	cs.Status.Host = cs.Spec.Host
	cs.Status.ServiceName = cs.Spec.Name

	return res, nil
}

// Delete permit to delete service from Centreon
func (r *CentreonServiceReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	cs := resource.(*v1alpha1.CentreonService)

	actualCS, err := cHandler.GetService(cs.Spec.Host, cs.Spec.Name)
	if err != nil {
		return err
	}

	if actualCS == nil {
		r.log.Info("Service already deleted on Centreon by external process, skip it")
		return nil
	}

	if err = cHandler.DeleteService(cs.Spec.Host, cs.Spec.Name); err != nil {
		return errors.Wrap(err, "Error when delete service from Centreon")
	}

	return nil

}

// Diff permit to check if diff between actual and expected Centreon service exist
func (r *CentreonServiceReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	cHandler := meta.(centreonhandler.CentreonHandler)
	cs := resource.(*v1alpha1.CentreonService)
	var d any

	expectedCS, err := cs.ToCentreoonService()
	if err != nil {
		return diff, errors.Wrap(err, "Error when convert to Centreon Service")
	}

	d, err = helper.Get(data, "currentService")
	if err != nil {
		return diff, err
	}
	currentService := d.(*centreonhandler.CentreonService)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}
	if currentService == nil {
		diff.NeedCreate = true
		diff.Diff = "Service not exist on Centreon"
		return diff, nil
	}

	currentDiff, err := cHandler.DiffService(currentService, expectedCS)
	if err != nil {
		return diff, errors.Wrap(err, "Error when diff Centreon service")
	}
	if currentDiff.IsDiff {
		diff.NeedUpdate = true
		diff.Diff = currentDiff.String()
	}

	data["expectedService"] = currentDiff

	return
}

// OnError permit to set status condition on the right state and record error
func (r *CentreonServiceReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	cs := resource.(*v1alpha1.CentreonService)

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&cs.Status.Conditions, v1.Condition{
		Type:    CentreonServiceCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *CentreonServiceReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	cs := resource.(*v1alpha1.CentreonService)

	if diff.NeedCreate {
		condition.SetStatusCondition(&cs.Status.Conditions, v1.Condition{
			Type:    CentreonServiceCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: fmt.Sprintf("Service %s/%s successfully created on Centreon", cs.Spec.Host, cs.Spec.Name),
		})
		r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "Service %s/%s successfully created on Centreon", cs.Spec.Host, cs.Spec.Name)

		return nil
	}

	if diff.NeedUpdate {
		condition.SetStatusCondition(&cs.Status.Conditions, v1.Condition{
			Type:    CentreonServiceCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: fmt.Sprintf("Service %s/%s successfully updated on Centreon", cs.Spec.Host, cs.Spec.Name),
		})

		r.recorder.Eventf(resource, core.EventTypeNormal, "Completed", "Service %s/%s successfully updated on Centreon", cs.Spec.Host, cs.Spec.Name)

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, CentreonServiceCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&cs.Status.Conditions, v1.Condition{
			Type:    CentreonServiceCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: fmt.Sprintf("Service %s/%s already exit on Centreon", cs.Spec.Host, cs.Spec.Name),
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Service already exit on Centreon")
	}

	return nil
}
