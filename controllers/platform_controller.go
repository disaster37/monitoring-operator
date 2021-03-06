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

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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
}

type ComputedPlatform struct {
	client   any
	platform *v1alpha1.Platform
	hash     string
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

	platform := &v1alpha1.Platform{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, platform, data)
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Platform{}).
		Owns(&core.Secret{}).
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

// Configure permit to init condition
func (r *PlatformReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {
	platform := resource.(*v1alpha1.Platform)

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
	p := resource.(*v1alpha1.Platform)

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
		data["secret"] = s

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
	p := resource.(*v1alpha1.Platform)
	var (
		d any
		s *core.Secret
	)

	// First, set owner on secret to track it
	d, err = helper.Get(data, "secret")
	if err != nil {
		return res, err
	}
	s = d.(*core.Secret)
	err = controllerutil.SetOwnerReference(p, s, r.Scheme)
	if err != nil {
		return res, err
	}
	if err = r.Client.Update(ctx, s); err != nil {
		return res, errors.Wrapf(err, "Error when update owner on secret %s", s.Name)
	}

	d, err = helper.Get(data, "platform")
	if err != nil {
		return res, err
	}
	if p.Spec.IsDefault {
		r.platforms["default"] = d.(*ComputedPlatform)
	}
	r.platforms[p.Spec.Name] = d.(*ComputedPlatform)

	return res, nil
}

// Update permit to update service on Centreon
func (r *PlatformReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete permit to delete service from Centreon
func (r *PlatformReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {
	p := resource.(*v1alpha1.Platform)

	if p.Spec.IsDefault {
		delete(r.platforms, "default")
	}
	delete(r.platforms, p.Spec.Name)

	return nil
}

// Diff permit to check if diff between actual and expected Centreon service exist
func (r *PlatformReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	p := resource.(*v1alpha1.Platform)
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
	if r.platforms[p.Spec.Name] == nil {
		diff.NeedCreate = true
		diff.Diff = "New plaform"

		return diff, nil
	}

	// Client change
	if r.platforms[p.Spec.Name].hash != pTarget.hash {
		diff.NeedUpdate = true
		diff.Diff = "Secret change on platform"
		return diff, nil
	}

	// Platform change
	diffStr := cmp.Diff(r.platforms[p.Spec.Name].platform.Spec, p.Spec)
	if diffStr != "" {
		diff.NeedUpdate = true
		diff.Diff = diffStr
		return diff, nil
	}

	return diff, nil
}

// OnError permit to set status condition on the right state and record error
func (r *PlatformReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {
	platform := resource.(*v1alpha1.Platform)

	r.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", err.Error())

	condition.SetStatusCondition(&platform.Status.Conditions, v1.Condition{
		Type:    PlatformCondition,
		Status:  v1.ConditionFalse,
		Reason:  "Failed",
		Message: err.Error(),
	})
}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *PlatformReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {
	platform := resource.(*v1alpha1.Platform)

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
	platformList := &v1alpha1.PlatformList{}
	platformGVR := v1alpha1.GroupVersion.WithResource("platforms")
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
			platforms[p.Spec.Name] = cp
			if p.Spec.IsDefault {
				platforms["default"] = cp
			}

		default:
			return nil, errors.Errorf("Platform %s of type %s is not supported", p.Name, p.Spec.PlatformType)

		}
	}

	return platforms, nil
}

func getComputedCentreonPlatform(p *v1alpha1.Platform, s *core.Secret, log *logrus.Entry) (cp *ComputedPlatform, err error) {

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
