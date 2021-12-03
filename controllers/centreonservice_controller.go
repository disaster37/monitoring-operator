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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
)

const centreonServicedFinalizer = "service.monitor.k8s.webcenter.fr/finalizer"

// CentreonServiceReconciler reconciles a CentreonService object
type CentreonServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
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
	_ = log.FromContext(ctx)
	log := logrus.WithFields(logrus.Fields{
		"namespace": req.Namespace,
		"name":      req.Name,
	})
	log.Info("Reconciling CentreonServiceReconciler")

	// Get the centreon service
	centreonService := &v1alpha1.CentreonService{}
	err := r.Get(ctx, req.NamespacedName, centreonService)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("CentreonService resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Errorf("Error when get CentreonService object with %s: %s", req.NamespacedName, err.Error())
		return ctrl.Result{}, err
	}
	log.Debugf("Found CentreonService: %+v", centreonService.Spec)

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(centreonService, centreonServicedFinalizer) {
		controllerutil.AddFinalizer(centreonService, centreonServicedFinalizer)
		if err := r.Update(ctx, centreonService); err != nil {
			log.Error("Error when add finalizer")
			return ctrl.Result{}, err
		}
		log.Info("Add finalizer successfully")
	}

	// Get client to access on Centreon
	client, err := r.newCentreonClient(ctx, log)
	if err != nil {
		switch err {
		case ErrorCentreonNotFound:
			log.Error("You need to create one Centreon object")
		case ErrorCentreonMultipleFound:
			log.Error("You need to have only one Centreon object, found multiple")
		default:
			log.Errorf("Error when get Centreon object: %s", err.Error())
		}
		return ctrl.Result{}, err
	}

	// Auth on client and reque if error. Maybee Centreon is temporary unavailable
	if err := client.Auth(); err != nil {
		log.Errorf("Error when authenticate with Centreon: %s", err.Error())
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	// Get current service on Centreon
	service, err := client.API.Service().Get(centreonService.Spec.Host, centreonService.Spec.Name)
	if err != nil {
		log.Errorf("Error get current service from Centreon: %s", err.Error())
		return ctrl.Result{}, err
	}

	// Create
	if service == nil {
		if err := createService(client, log, &centreonService.Spec); err != nil {
			log.Errorf("Error when create service from Centreon: %s", err.Error())
			return ctrl.Result{}, err
		}
		log.Infof("Create service %s/%s successfully on Centreon", centreonService.Spec.Host, centreonService.Spec.Name)
		centreonService.Status.ID = fmt.Sprintf("%s/%s", centreonService.Spec.Host, centreonService.Spec.Name)

		// Update status
		if err := r.Status().Update(ctx, centreonService); err != nil {
			log.Error("Failed to update CentreonService status")
			return ctrl.Result{}, err
		}
	} else if centreonService.GetDeletionTimestamp() != nil {
		// Delete
		if controllerutil.ContainsFinalizer(centreonService, centreonServicedFinalizer) {
			if err := deleteService(client, log, &centreonService.Spec); err != nil {
				return ctrl.Result{}, err
			}
			log.Infof("Delete service %s/%s from Centreon successfully", centreonService.Spec.Host, centreonService.Spec.Name)
			controllerutil.RemoveFinalizer(centreonService, centreonServicedFinalizer)
			if err := r.Update(ctx, centreonService); err != nil {
				return ctrl.Result{}, err
			}
			log.Infof("Remove finalizer successfully")
		}
		return ctrl.Result{}, nil

	} else {
		// Update
		if err := updateService(client, log, &centreonService.Spec, service); err != nil {
			log.Error("Error when update service from Centreon")
			return ctrl.Result{}, err
		}
		log.Infof("Update service %s/%s successfully on Centreon", centreonService.Spec.Host, centreonService.Spec.Name)
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *CentreonServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&monitorv1alpha1.CentreonService{}).
		Complete(r)
}
