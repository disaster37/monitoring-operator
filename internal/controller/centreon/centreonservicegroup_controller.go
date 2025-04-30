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

package centreon

import (
	"context"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/platform"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	centreonServiceGroupName string = "centreonServiceGroup"
)

// CentreonServiceGroupReconciler reconciles a CentreonServiceGroup object
type CentreonServiceGroupReconciler struct {
	controller.Controller
	controller.RemoteReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler]
	controller.RemoteReconcilerAction[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler]
	name string
}

func NewCentreonServiceGroupReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder, platforms map[string]*platform.ComputedPlatform) controller.Controller {
	return &CentreonServiceGroupReconciler{
		Controller: controller.NewBasicController(),
		RemoteReconciler: controller.NewBasicRemoteReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler](
			client,
			centreonServiceGroupName,
			"servicegroup.monitor.k8s.webcenter.fr/finalizer",
			logger,
			recorder,
		),
		RemoteReconcilerAction: newCentreonServiceGroupReconciler(
			centreonServiceGroupName,
			client,
			recorder,
			platforms,
		),
		name: centreonServiceGroupName,
	}
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
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CentreonServiceGroupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	cs := &centreoncrd.CentreonServiceGroup{}
	data := map[string]any{}

	return r.RemoteReconciler.Reconcile(
		ctx,
		req,
		cs,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonServiceGroupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(r.name).
		For(&centreoncrd.CentreonServiceGroup{}).
		WithOptions(k8scontroller.Options{
			RateLimiter: helper.DefaultControllerRateLimiter[reconcile.Request](),
		}).
		Complete(r)
}
