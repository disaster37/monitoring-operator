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
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	// Check default value is specified
	if r.CentreonConfig.Load() == nil {

	}

	// Get instance
	instance := &networkv1.Ingress{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Errorf("Error when get resource: %s", err.Error())
		return ctrl.Result{}, err
	}
	r.Log = r.Log.WithFields(logrus.Fields{
		"Ingress": instance.Name,
	})

	// Delete
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		r.Log.Info("Ingress is being deleted, auto delete CentreonService if exist")
		return ctrl.Result{}, nil
	}

	// Reconcile
	var centreonSpec *v1alpha1.CentreonSpec
	if r.CentreonConfig != nil {
		data := r.CentreonConfig.Load()
		if data != nil {
			centreonSpec = data.(*v1alpha1.CentreonSpec)
		}
	}
	if centreonSpec == nil {
		r.Log.Warning("It's recommanded to set some default values on custom resource called `Centreon` on the same operator namespace. It avoid to set on each ingress all Centreon service properties as annotations")
	}

	cs := &v1alpha1.CentreonService{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cs)
	if err != nil && errors.IsNotFound(err) {
		//Create
		cs, err = centreonServiceFromIngress(instance, centreonSpec, r.Scheme)
		if err != nil {
			r.Log.Errorf("Error when generate CentreonService from Ingress: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
			return ctrl.Result{}, err
		}
		if err = r.Create(ctx, cs); err != nil {
			r.Log.Errorf("Error when create CentreonService: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
			return ctrl.Result{}, err
		}

		r.Log.Info("Create CentreonService successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon Service created successfully")
		return ctrl.Result{}, nil
	} else if err != nil {
		r.Log.Errorf("Failed to get CentreonService: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{}, err
	}

	// Update if needed
	expectedCs, err := centreonServiceFromIngress(instance, centreonSpec, r.Scheme)
	if err != nil {
		r.Log.Errorf("Error when generate CentreonService from Ingress: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{}, err
	}

	diffSpec := cmp.Diff(cs.Spec, expectedCs.Spec)
	diffLabels := cmp.Diff(cs.GetLabels(), expectedCs.GetLabels())
	diffAnnotations := cmp.Diff(cs.GetAnnotations(), expectedCs.GetAnnotations())
	if diffSpec != "" || diffLabels != "" || diffAnnotations != "" {
		r.Log.Infof("Diff detected:\n%s\n%s\n%s", diffSpec, diffLabels, diffAnnotations)
		//Update
		cs.SetLabels(expectedCs.GetLabels())
		cs.SetAnnotations(expectedCs.GetAnnotations())
		cs.Spec = expectedCs.Spec
		if err = r.Update(ctx, cs); err != nil {
			r.Log.Errorf("Error when update CentreonService: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
			return ctrl.Result{}, err
		}
		r.Log.Info("Update CentreonService successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon Service updated successfully")
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkv1.Ingress{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewIngressWithMonitoringAnnotationPredicate()).
		Complete(r)
}
