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
	"sync/atomic"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CentreonServiceReconciler reconciles a CentreonService object
type CentreonServiceReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Recorder       record.EventRecorder
	Log            *logrus.Entry
	Service        CentreonService
	CentreonConfig atomic.Value
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons,verbs=get;list;watch;create;update;patch;delete

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

	r.Log.Infof("Starting reconcile loop for %v", req.NamespacedName)
	defer r.Log.Infof("Finish reconcile loop for %v", req.NamespacedName)

	// Get instance
	instance := &v1alpha1.CentreonService{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Errorf("Error when get resource: %s", err.Error())
		return ctrl.Result{}, err
	}
	r.Log = r.Log.WithFields(logrus.Fields{
		"Host":    instance.Spec.Host,
		"Service": instance.Spec.Name,
	})

	// Add finalizer
	// Requeue if add finalizer to avoid lock resource
	if !instance.HasFinalizer() {
		instance.AddFinalizer()
		if err := r.Update(ctx, instance); err != nil {
			r.Log.Errorf("Error when add finalizer: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Adding finalizer", "Failed to add finalizer: %s", err)
			return ctrl.Result{}, err
		}
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Added", "Object finalizer is added")
		r.Log.Debug("Add finalizer successfully")
		return ctrl.Result{Requeue: true}, nil
	}

	r.Service.SetLogger(r.Log)

	// Delete
	if instance.IsBeingDeleted() {
		if instance.HasFinalizer() {
			if err := r.Service.Delete(instance); err != nil {
				r.Log.Errorf("Error when delete service on Centreon: %s", err.Error())
				r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when delete service on Centreon: %s", err.Error())
				return ctrl.Result{}, err
			}

			instance.RemoveFinalizer()
			if err := r.Update(ctx, instance); err != nil {
				r.Log.Errorf("Failed to remove finalizer: %s", err.Error())
				r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when remove finalizer: %s", err.Error())
				return ctrl.Result{}, err
			}
			r.Log.Debug("Remove finalizer successfully")
		}
		r.Log.Info("Delete Centreon service successfully")
		return ctrl.Result{}, nil
	}

	// Reconcile
	isCreated, isUpdated, err := r.Service.Reconcile(instance)
	if err != nil {
		r.Log.Errorf("Error when reconcile Centreon service: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{}, err
	}

	if isCreated || isUpdated {
		if isCreated {
			instance.Status.CreatedAt = time.Now().String()
			r.Log.Info("Create service on Centreon successfully")
			r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Service created on Centreon")
		} else {
			instance.Status.UpdatedAt = time.Now().String()
			r.Log.Info("Update service on Centreon successfully")
			r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Service updated on Centreon")
		}
		instance.Status.ID = fmt.Sprintf("%s/%s", instance.Spec.Host, instance.Spec.Name)
		if err := r.Status().Update(ctx, instance); err != nil {
			r.Log.Errorf("Failed to update status: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when update status: %s", err.Error())
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.CentreonService{}).
		Complete(r)
}
