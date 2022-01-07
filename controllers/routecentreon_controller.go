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
	errorspkg "errors"
	"fmt"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/google/go-cmp/cmp"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// RouteReconciler reconciles a Route object
type RouteCentreonReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Log      *logrus.Entry
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreons,verbs=get;list;watch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete

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
	r.Log.Infof("Starting reconcile loop for %v", req.NamespacedName)
	defer r.Log.Infof("Finish reconcile loop for %v", req.NamespacedName)

	// Get instance
	instance := &routev1.Route{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Errorf("Error when get resource: %s", err.Error())
		return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
	}
	r.Log = r.Log.WithFields(logrus.Fields{
		"name":      instance.Name,
		"namespace": instance.Namespace,
	})

	// Delete
	if !instance.ObjectMeta.DeletionTimestamp.IsZero() {
		r.Log.Info("Route is being deleted, auto delete CentreonService if exist")
		return ctrl.Result{}, nil
	}

	// Reconcile
	centreonSpec, err := getCentreonSpec(ctx, r.Client)
	if err != nil {
		r.Log.Errorf("Error when get CentreonSpec: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
	}
	if centreonSpec == nil {
		r.Log.Warning("It's recommanded to set some default values on custom resource called `Centreon` on the same operator namespace. It avoid to set on each route all Centreon service properties as annotations")
	}

	cs := &v1alpha1.CentreonService{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cs)
	if err != nil && errors.IsNotFound(err) {
		//Create
		cs, err = centreonServiceFromRoute(instance, centreonSpec, r.Scheme)
		if err != nil {
			r.Log.Errorf("Error when generate CentreonService from route: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
			return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
		}
		if err = r.Create(ctx, cs); err != nil {
			r.Log.Errorf("Error when create CentreonService from route: %s", err.Error())
			r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
			return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
		}

		r.Log.Info("Create CentreonService successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon Service created successfully")
		return ctrl.Result{}, nil
	} else if err != nil {
		r.Log.Errorf("Failed to get CentreonService: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
	}

	// Update if needed
	expectedCs, err := centreonServiceFromRoute(instance, centreonSpec, r.Scheme)
	if err != nil {
		r.Log.Errorf("Error when generate CentreonService from route: %s", err.Error())
		r.Recorder.Eventf(instance, corev1.EventTypeWarning, "Failed", "Error when reconcile: %s", err.Error())
		return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
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
			return ctrl.Result{RequeueAfter: waitDurationWhenError}, err
		}
		r.Log.Info("Update CentreonService successfully")
		r.Recorder.Event(instance, corev1.EventTypeNormal, "Completed", "Centreon Service updated successfully")
	}

	return ctrl.Result{}, nil
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

// It compute the expected CentreonServiceSpec from route and CentreonSpec
func centreonServiceFromRoute(r *routev1.Route, centreonSpec *v1alpha1.CentreonSpec, scheme *runtime.Scheme) (cs *v1alpha1.CentreonService, err error) {

	if r == nil {
		return nil, errorspkg.New("Route must be provided")
	}
	if scheme == nil {
		return nil, errorspkg.New("Scheme must be provided")
	}
	cs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        r.Name,
			Namespace:   r.Namespace,
			Labels:      r.GetLabels(),
			Annotations: r.GetAnnotations(),
		},
		Spec: monitorv1alpha1.CentreonServiceSpec{},
	}

	// Generate placeholders
	placeholders := generatePlaceholdersRouteCentreonService(r)
	initCentreonServiceDefaultValue(centreonSpec, cs, placeholders)

	// Then, init with annotations
	if err = initCentreonServiceFromAnnotations(r.GetAnnotations(), cs); err != nil {
		return nil, err
	}

	// Check CentreonService is valide
	if !cs.IsValid() {
		return nil, fmt.Errorf("Generated CentreonService is not valid: %+v", cs.Spec)
	}

	// Set route instance as the owner
	ctrl.SetControllerReference(r, cs, scheme)

	return cs, nil
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
	placeholders["rule.host"] = r.Spec.Host
	if r.Spec.Path != "" {
		placeholders["rule.path"] = r.Spec.Path
	} else {
		placeholders["rule.path"] = "/"
	}
	if r.Spec.TLS != nil && r.Spec.TLS.Termination != "" {
		placeholders["rule.scheme"] = "https"
	} else {
		placeholders["rule.scheme"] = "http"
	}

	return placeholders

}
