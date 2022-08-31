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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	PlatformFinalizer = "platform.monitor.k8s.webcenter.fr/finalizer"
	PlatformCondition = "LoadConfig"
)

// PlatformReconciler reconciles a Platform object
type PlatformReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	name   string
}

type ComputedPlatform struct {
	client   any
	platform *monitorapi.Platform
	hash     string
}

func NewPlatformReconciler(client client.Client, scheme *runtime.Scheme) *PlatformReconciler {

	r := &PlatformReconciler{
		Client: client,
		Scheme: scheme,
		name:   "platform",
	}

	controllerMetrics.WithLabelValues(r.name).Add(0)

	return r
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Platform object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, PlatformFinalizer, r.reconciler, r.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	platform := &monitorapi.Platform{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, platform, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(r.name).
		For(&monitorapi.Platform{}).
		Watches(&source.Kind{Type: &core.Secret{}}, handler.EnqueueRequestsFromMapFunc(watchCentreonPlatformSecret(r.Client))).
		WithEventFilter(viewOperatorNamespacePredicate()).
		Complete(r)
}

func viewOperatorNamespacePredicate() predicate.Predicate {
	ns, err := helpers.GetOperatorNamespace()
	if err != nil {
		panic(err)
	}
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			return e.ObjectNew.GetNamespace() == ns || e.ObjectOld.GetNamespace() == ns
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return e.Object.GetNamespace() == ns
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return e.Object.GetNamespace() == ns
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return e.Object.GetNamespace() == ns
		},
	}
}

// watchPlatformSecret permit to update client if platform secret change
func watchCentreonPlatformSecret(c client.Client) handler.MapFunc {
	return func(a client.Object) []reconcile.Request {

		reconcileRequests := make([]reconcile.Request, 0)
		listPlatforms := &monitorapi.PlatformList{}

		fs := fields.ParseSelectorOrDie(fmt.Sprintf("spec.centreonSettings.secret=%s", a.GetName()))

		// Get all platforms that use the current secret
		if err := c.List(context.Background(), listPlatforms, &client.ListOptions{Namespace: a.GetNamespace(), FieldSelector: fs}); err != nil {
			panic(err)
		}

		for _, p := range listPlatforms.Items {
			reconcileRequests = append(reconcileRequests, reconcile.Request{NamespacedName: types.NamespacedName{Name: p.Name, Namespace: p.Namespace}})
		}

		return reconcileRequests
	}
}

// Configure permit to init condition
func (r *PlatformReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	platform := resource.(*monitorapi.Platform)

	// Init condition status if not exist
	if condition.FindStatusCondition(platform.Status.Conditions, PlatformCondition) == nil {
		condition.SetStatusCondition(&platform.Status.Conditions, v1.Condition{
			Type:   PlatformCondition,
			Status: v1.ConditionFalse,
			Reason: "Initialize",
		})
	}

	return nil, nil
}

// Read
func (r *PlatformReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	p := resource.(*monitorapi.Platform)

	// Skip compute platform if on delete step because off secret can be already deleted
	// It avoid dead lock
	if !p.ObjectMeta.DeletionTimestamp.IsZero() {
		return res, nil
	}

	switch p.Spec.PlatformType {
	case "centreon":
		// Get secret
		s := &core.Secret{}
		k := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      p.Spec.CentreonSettings.Secret,
		}
		if err = r.Get(ctx, k, s); err != nil {
			if k8serrors.IsNotFound(err) {
				r.log.Warnf("Secret %s not yet exist, try later", p.Spec.CentreonSettings.Secret)
				return res, errors.Errorf("Secret %s not yet exist", p.Spec.CentreonSettings.Secret)
			}
		}

		computedPlatform, err := getComputedCentreonPlatform(p, s, r.log)
		if err != nil {
			return res, errors.Wrapf(err, "Error when compute platform %s", p.Name)
		}
		data["platform"] = computedPlatform

	default:
		return res, errors.Errorf("Plaform %s is not supported", p.Spec.PlatformType)
	}

	return res, nil
}

// Create add new Service on Centreon
func (r *PlatformReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	p := resource.(*monitorapi.Platform)
	var (
		d any
	)

	d, err = helper.Get(data, "platform")
	if err != nil {
		return res, err
	}
	if p.Spec.IsDefault {
		r.platforms["default"] = d.(*ComputedPlatform)
	}
	r.platforms[p.Name] = d.(*ComputedPlatform)

	// Update prometheus mectric
	controllerMetrics.WithLabelValues(r.name).Inc()

	return res, nil
}

// Update permit to update service on Centreon
func (r *PlatformReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete service from Centreon
func (r *PlatformReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {
	p := resource.(*monitorapi.Platform)

	if p.Spec.IsDefault {
		delete(r.platforms, "default")
	}
	delete(r.platforms, p.Name)

	// Update prometheus mectric
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if diff between actual and expected Centreon service exist
func (r *PlatformReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	p := resource.(*monitorapi.Platform)
	var d any

	d, err = helper.Get(data, "platform")
	if err != nil {
		return diff, err
	}
	pTarget := d.(*ComputedPlatform)

	diff = controller.Diff{
		NeedCreate: false,
		NeedUpdate: false,
	}

	// New platform
	if r.platforms[p.Name] == nil {
		diff.NeedCreate = true
		diff.Diff = "New plaform"

		return diff, nil
	}

	// Client change
	if r.platforms[p.Name].hash != pTarget.hash {
		diff.NeedUpdate = true
		diff.Diff = "Secret change on platform"
		return diff, nil
	}

	// Platform change
	diffStr := cmp.Diff(r.platforms[p.Name].platform.Spec, p.Spec)
	if diffStr != "" {
		diff.NeedUpdate = true
		diff.Diff = diffStr
		return diff, nil
	}

	return diff, nil
}

// OnError permit to set status condition on the right state and record error
func (r *PlatformReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	platform := resource.(*monitorapi.Platform)

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&platform.Status.Conditions, v1.Condition{
		Type:    PlatformCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})

	// Update prometheus mectric
	totalErrors.Inc()
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *PlatformReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	platform := resource.(*monitorapi.Platform)

	if diff.NeedCreate || diff.NeedUpdate {
		condition.SetStatusCondition(&platform.Status.Conditions, v1.Condition{
			Type:    PlatformCondition,
			Status:  v1.ConditionTrue,
			Reason:  "Success",
			Message: "Load platform successfully",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Load platform successfully")

		return nil
	}

	// Update condition status if needed
	if condition.IsStatusConditionPresentAndEqual(platform.Status.Conditions, PlatformCondition, v1.ConditionFalse) {
		condition.SetStatusCondition(&platform.Status.Conditions, v1.Condition{
			Type:    PlatformCondition,
			Reason:  "Success",
			Status:  v1.ConditionTrue,
			Message: "Load platform successfully",
		})

		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Load platform successfully")
	}

	return nil
}

// ComputedPlatformList permit to get the list of coomputed platform object
// It usefull to init controller with client to access on external monitoring resources
func ComputedPlatformList(ctx context.Context, cd dynamic.Interface, c kubernetes.Interface, log *logrus.Entry) (platforms map[string]*ComputedPlatform, err error) {
	platforms = map[string]*ComputedPlatform{}
	platformList := &monitorapi.PlatformList{}
	platformGVR := monitorapi.GroupVersion.WithResource("platforms")
	ns, err := helpers.GetOperatorNamespace()
	if err != nil {
		return nil, err
	}

	// Get list of current platform
	pluc, err := cd.Resource(platformGVR).Namespace(ns).List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(pluc.UnstructuredContent(), platformList); err != nil {
		return nil, err
	}

	// Create computed platform
	for _, p := range platformList.Items {
		log.Debugf("Start to Compute platform %s", p.Name)

		switch p.Spec.PlatformType {
		case "centreon":
			// Get centreon secret
			s, err := c.CoreV1().Secrets(p.Namespace).Get(ctx, p.Spec.CentreonSettings.Secret, v1.GetOptions{})
			if err != nil {
				if k8serrors.IsNotFound(err) {
					log.Warnf("Secret %s not yet exist, skip platform %s", p.Spec.CentreonSettings.Secret, p.Name)
					continue
				}
				return nil, err
			}
			cp, err := getComputedCentreonPlatform(&p, s, log)
			if err != nil {
				return nil, errors.Wrapf(err, "Error when compute platform %s", p.Name)
			}
			platforms[p.Name] = cp
			if p.Spec.IsDefault {
				platforms["default"] = cp
			}

		default:
			return nil, errors.Errorf("Platform %s of type %s is not supported", p.Name, p.Spec.PlatformType)

		}
	}

	return platforms, nil
}

func getComputedCentreonPlatform(p *monitorapi.Platform, s *core.Secret, log *logrus.Entry) (cp *ComputedPlatform, err error) {

	if p == nil {
		return nil, errors.New("Platform can't be null")
	}
	if s == nil {
		return nil, errors.New("Secret can't be null")
	}

	username := string(s.Data["username"])
	password := string(s.Data["password"])
	if username == "" || password == "" {
		return nil, errors.Errorf("You need to set username and password on secret %s", s.Name)
	}

	// Create client
	cfg := &models.Config{
		Address:          p.Spec.CentreonSettings.URL,
		Username:         username,
		Password:         password,
		DisableVerifySSL: p.Spec.CentreonSettings.SelfSignedCertificate,
	}
	if log.Level == logrus.DebugLevel {
		cfg.Debug = true
	}
	client, err := centreon.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create Centreon client")
	}
	shaByte, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	sha := sha256.New()
	if _, err := sha.Write([]byte(shaByte)); err != nil {
		return nil, err
	}

	return &ComputedPlatform{
		client:   centreonhandler.NewCentreonHandler(client, log),
		platform: p,
		hash:     hex.EncodeToString(sha.Sum(nil)),
	}, nil

}
