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
	"sync/atomic"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// CentreonReconciler reconciles a Ingress object
type CentreonReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Log            *logrus.Entry
	Recorder       record.EventRecorder
	CentreonConfig *atomic.Value
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Ingress object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CentreonReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.Log.Infof("Starting reconcile loop for %v", req.NamespacedName)
	defer r.Log.Infof("Finish reconcile loop for %v", req.NamespacedName)

	// Get instance
	instance := &monitorv1alpha1.Centreon{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Errorf("Error when get resource: %s", err.Error())
		return ctrl.Result{}, err
	}
	r.Log = r.Log.WithFields(logrus.Fields{
		"name":      instance.Name,
		"namespace": instance.Namespace,
	})

	// Delete
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		r.Log.Info("Centreon is being deleted, unshare Centreon")
		r.CentreonConfig.Store(nil)
		return ctrl.Result{}, nil
	}

	// Load current Centreon share
	var centreonSpec *v1alpha1.CentreonSpec
	if r.CentreonConfig != nil {
		data := r.CentreonConfig.Load()
		if data != nil {
			centreonSpec = data.(*v1alpha1.CentreonSpec)
		}
	}

	// Share new CrentreonSpec with other controllers
	if centreonSpec == nil {
		r.CentreonConfig.Store(&instance.Spec)
		r.Log.Info("Share Centreon successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon shared successfully")
		return ctrl.Result{}, nil
	}

	// Rconcile
	diff := cmp.Diff(centreonSpec, &instance.Spec)
	if diff != "" {
		r.Log.Infof("Diff detected:\n%s", diff)
		r.CentreonConfig.Store(&instance.Spec)
		r.Log.Info("Update share Centreon successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon shared successfully updated")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&monitorv1alpha1.Centreon{}).
		WithEventFilter(viewCentreonNamespacePredicate()).
		Complete(r)
}

// Handle only Centreon on the same controller namespace, or default namespace if not found
func viewCentreonNamespacePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isOnControllerNamespace(e.ObjectOld.GetNamespace()) || isOnControllerNamespace(e.ObjectNew.GetNamespace())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isOnControllerNamespace(e.Object.GetNamespace())
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isOnControllerNamespace(e.Object.GetNamespace())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isOnControllerNamespace(e.Object.GetNamespace())
		},
	}
}

func isOnControllerNamespace(ns string) bool {
	expectedNs, err := helpers.GetCentreonNamespace()
	if err != nil {
		expectedNs = "default"
	}

	return ns == expectedNs
}
