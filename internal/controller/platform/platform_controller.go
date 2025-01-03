/*
Copyright 2022.

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

package platform

import (
	"context"
	"fmt"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	plaformName string = "platform"
)

// PlatformReconciler reconcile platform Object
type PlatformReconciler struct {
	controller.Controller
	controller.RemoteReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler]
	controller.RemoteReconcilerAction[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler]
	name string
}

func NewPlatformReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder, platforms map[string]*ComputedPlatform) controller.Controller {
	return &PlatformReconciler{
		Controller: controller.NewBasicController(),
		RemoteReconciler: controller.NewBasicRemoteReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler](
			client,
			plaformName,
			"platform.monitor.k8s.webcenter.fr/finalizer",
			logger,
			recorder,
		),
		RemoteReconcilerAction: newPlatformReconciler(
			plaformName,
			client,
			recorder,
			platforms,
		),
		name: plaformName,
	}
}

//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=platforms/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cerebro object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *PlatformReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	p := &centreoncrd.Platform{}
	data := map[string]any{}

	return r.RemoteReconciler.Reconcile(
		ctx,
		req,
		p,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (h *PlatformReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(h.name).
		For(&centreoncrd.Platform{}).
		Watches(&corev1.Secret{}, handler.EnqueueRequestsFromMapFunc(watchCentreonPlatformSecret(h.Client()))).
		WithEventFilter(viewOperatorNamespacePredicate()).
		WithOptions(k8scontroller.Options{
			RateLimiter: helper.DefaultControllerRateLimiter[reconcile.Request](),
		}).
		Complete(h)
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
	return func(ctx context.Context, a client.Object) []reconcile.Request {

		reconcileRequests := make([]reconcile.Request, 0)
		listPlatforms := &centreoncrd.PlatformList{}

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
