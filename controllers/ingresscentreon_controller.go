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
	"encoding/json"
	errorspkg "errors"
	"fmt"
	"regexp"
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

const (
	monitoringAnnotationKey         = "monitor.k8s.webcenter.fr"
	centreonMonitoringAnnotationKey = "centreon.monitor.k8s.webcenter.fr"
)

// IngressReconciler reconciles a Ingress object
type IngressCentreonReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Log            *logrus.Entry
	Recorder       record.EventRecorder
	CentreonConfig *atomic.Value
}

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update

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

	diffSpec := cmp.Diff(expectedCs.Spec, cs.Spec)
	diffLabels := cmp.Diff(expectedCs.GetLabels(), cs.GetLabels())
	diffAnnotations := cmp.Diff(expectedCs.GetAnnotations(), cs.GetAnnotations())
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
func (r *IngressCentreonReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		For(&networkv1.Ingress{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(viewIngressWithMonitoringAnnotationPredicate()).
		Complete(r)
}

func centreonServiceFromIngress(i *networkv1.Ingress, centreonSpec *v1alpha1.CentreonSpec, scheme *runtime.Scheme) (cs *v1alpha1.CentreonService, err error) {

	if i == nil {
		return nil, errorspkg.New("Ingress must be provided")
	}
	if scheme == nil {
		return nil, errorspkg.New("Scheme must be provided")
	}
	cs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:        i.Name,
			Namespace:   i.Namespace,
			Labels:      i.GetLabels(),
			Annotations: i.GetAnnotations(),
		},
		Spec: monitorv1alpha1.CentreonServiceSpec{},
	}

	// Generate placeholders
	placeholders := generatePlaceholdersIngressCentreonService(i)
	initIngressCentreonServiceDefaultValue(centreonSpec, cs, placeholders)

	// Then, init with annotations
	if err = initIngressCentreonServiceFromAnnotations(i.GetAnnotations(), cs); err != nil {
		return nil, err
	}

	// Check CentreonService is valide
	if !cs.IsValid() {
		return nil, fmt.Errorf("Generated CentreonService is not valid: %+v", cs.Spec)
	}

	// Set ingress instance as the owner
	ctrl.SetControllerReference(i, cs, scheme)

	return cs, nil
}

// Handle only ingress that have the monitoring annotation
func viewIngressWithMonitoringAnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isMonitoringAnnotation(e.ObjectOld.GetAnnotations()) || isMonitoringAnnotation(e.ObjectNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
	}
}

func isMonitoringAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	watchKey := fmt.Sprintf("%s/discover", monitoringAnnotationKey)
	for key, value := range annotations {
		if key == watchKey && value == "true" {
			return true
		}
	}
	return false
}

func initIngressCentreonServiceDefaultValue(centreon *v1alpha1.CentreonSpec, cs *v1alpha1.CentreonService, placesholders map[string]string) {
	if centreon == nil || centreon.Endpoints == nil || cs == nil {
		return
	}

	cs.Spec.Activated = centreon.Endpoints.ActivateService
	cs.Spec.Categories = centreon.Endpoints.Categories
	cs.Spec.Groups = centreon.Endpoints.ServiceGroups
	cs.Spec.Host = centreon.Endpoints.DefaultHost
	cs.Spec.Template = centreon.Endpoints.Template

	// Need placeholders
	if centreon.Endpoints.Arguments != nil && len(centreon.Endpoints.Arguments) > 0 {
		arguments := make([]string, len(centreon.Endpoints.Arguments))
		for i, arg := range centreon.Endpoints.Arguments {
			arguments[i] = helpers.PlaceholdersInString(arg, placesholders)
		}
		cs.Spec.Arguments = arguments
	}

	if centreon.Endpoints.Macros != nil && len(centreon.Endpoints.Macros) > 0 {
		macros := map[string]string{}
		for key, value := range centreon.Endpoints.Macros {
			macros[key] = helpers.PlaceholdersInString(value, placesholders)
		}
		cs.Spec.Macros = macros
	}

	cs.Spec.Name = helpers.PlaceholdersInString(centreon.Endpoints.NameTemplate, placesholders)
}

func initIngressCentreonServiceFromAnnotations(ingressAnnotations map[string]string, cs *v1alpha1.CentreonService) (err error) {
	if ingressAnnotations == nil || cs == nil {
		return
	}

	// Init values from annotations
	re := regexp.MustCompile(fmt.Sprintf("^%s/(.+)$", centreonMonitoringAnnotationKey))
	for key, value := range ingressAnnotations {
		if match := re.FindStringSubmatch(key); len(match) > 0 {
			switch match[1] {
			case "name":
				cs.Spec.Name = value
				break
			case "host":
				cs.Spec.Host = value
				break
			case "template":
				cs.Spec.Template = value
				break
			case "activated":
				t := helpers.StringToBool(value)
				if t != nil {
					cs.Spec.Activated = *t
				}
				break
			case "normal-check-interval":
				cs.Spec.NormalCheckInterval = value
				break
			case "retry-check-interval":
				cs.Spec.RetryCheckInterval = value
				break
			case "max-check-attempts":
				cs.Spec.MaxCheckAttempts = value
				break
			case "active-check-enabled":
				cs.Spec.ActiveCheckEnabled = helpers.StringToBool(value)
				break
			case "passive-check-enabled":
				cs.Spec.PassiveCheckEnabled = helpers.StringToBool(value)
				break
			case "arguments":
				cs.Spec.Arguments = helpers.StringToSlice(value, ",")
				break
			case "groups":
				cs.Spec.Groups = helpers.StringToSlice(value, ",")
				break
			case "categories":
				cs.Spec.Categories = helpers.StringToSlice(value, ",")
				break
			case "macros":
				t := map[string]string{}
				if err = json.Unmarshal([]byte(value), &t); err != nil {
					return err
				}
				cs.Spec.Macros = t
				break
			}
		}
	}

	return nil

}

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
