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
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	monitorv1alpha1 "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	CertificateFinalizer = "certificate.monitor.k8s.webcenter.fr/finalizer"
)

// CertificateReconciler reconciles a Secret of type TLS object
type CertificateReconciler struct {
	Reconciler
	client.Client
	Scheme *runtime.Scheme
	TemplateController
	name string
}

func NewCertificateReconciler(client client.Client, scheme *runtime.Scheme, templateController TemplateController) *CertificateReconciler {

	r := &CertificateReconciler{
		Client:             client,
		Scheme:             scheme,
		TemplateController: templateController,
		name:               "certificate",
	}

	controllerMetrics.WithLabelValues(r.name).Add(0)

	return r
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonServices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="monitor.k8s.webcenter.fr",resources=templates,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Node object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	reconciler, err := controller.NewStdReconciler(r.Client, CertificateFinalizer, r.reconciler, r.Reconciler.log, r.recorder, waitDurationWhenError)
	if err != nil {
		return ctrl.Result{}, err
	}

	secret := &core.Secret{}
	data := map[string]any{}

	return reconciler.Reconcile(ctx, req, secret, data)

}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&core.Secret{}).
		Owns(&monitorv1alpha1.CentreonService{}).
		WithEventFilter(predicate.And(viewResourceWithMonitoringTemplate(), viewCertificate())).
		Watches(&source.Kind{Type: &v1alpha1.Template{}}, handler.EnqueueRequestsFromMapFunc(watchTemplate(r.Client))).
		Complete(r)
}

// Configure do nothink here
func (r *CertificateReconciler) Configure(ctx context.Context, req ctrl.Request, resource client.Object) (meta any, err error) {

	return nil, nil
}

// Read permit to compute expected monitoring service that reflect node
func (r *CertificateReconciler) Read(ctx context.Context, resource client.Object, data map[string]any, meta any) (res ctrl.Result, err error) {
	secret := resource.(*core.Secret)
	placeholders, err := generatePlaceholdersCertificate(secret)
	if err != nil {
		return res, err
	}
	return r.TemplateController.readTemplating(ctx, secret, data, meta, placeholders)
}

// Create add new service object
func (r *CertificateReconciler) Create(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.TemplateController.createOrUpdateRessourcesFromTemplate(ctx, resource, data, meta)
}

// Update permit to update service object
func (r *CertificateReconciler) Update(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (res ctrl.Result, err error) {
	return r.Create(ctx, resource, data, meta)
}

// Delete do nothink here
// We add parent link, so k8s auto delete children
func (r *CertificateReconciler) Delete(ctx context.Context, resource client.Object, data map[string]interface{}, meta interface{}) (err error) {

	// Update metrics
	controllerMetrics.WithLabelValues(r.name).Dec()

	return nil
}

// Diff permit to check if diff between actual and expected CentreonService exist
func (r *CertificateReconciler) Diff(resource client.Object, data map[string]interface{}, meta interface{}) (diff controller.Diff, err error) {
	return r.TemplateController.diffRessourcesFromTemplate(resource, data, meta)
}

// OnError permit to set status condition on the right state and record error
func (r *CertificateReconciler) OnError(ctx context.Context, resource client.Object, data map[string]any, meta any, err error) {

	r.Reconciler.log.Error(err)
	r.recorder.Event(resource, core.EventTypeWarning, "Failed", fmt.Sprintf("Error when generate CentreonService: %s", err.Error()))

	// Update metrics
	totalErrors.Inc()

}

// OnSuccess permit to set status condition on the right state is everithink is good
func (r *CertificateReconciler) OnSuccess(ctx context.Context, resource client.Object, data map[string]any, meta any, diff controller.Diff) (err error) {

	if diff.NeedCreate {
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Create CentreonService successfully")
		return nil
	}

	if diff.NeedUpdate {
		r.recorder.Event(resource, core.EventTypeNormal, "Completed", "Update CentreonService successfully")
		return nil
	}

	return nil
}

// It generate map of placeholders from secret of type TLS
func generatePlaceholdersCertificate(s *core.Secret) (placeholders map[string]any, err error) {
	placeholders = map[string]any{}
	if s == nil {
		return placeholders, nil
	}

	//Main properties
	placeholders["name"] = s.Name
	placeholders["namespace"] = s.Namespace
	placeholders["labels"] = s.GetLabels()
	placeholders["annotations"] = s.GetAnnotations()

	// Read certificates
	var (
		blocks []byte
		rest   []byte
		block  *pem.Block
	)
	rest = s.Data["tls.crt"]
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			return placeholders, nil
		}
		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}

	certs, err := x509.ParseCertificates(blocks)
	if err != nil {
		return nil, errors.Wrap(err, "Error whe read TLS certificate")
	}

	placeholders["certificates"] = certs

	return placeholders, nil

}

func viewCertificate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			switch e.ObjectOld.(type) {
			case *core.Secret:
				return e.ObjectOld.(*core.Secret).Type == core.SecretTypeTLS
			default:
				return true
			}

		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			switch e.Object.(type) {
			case *core.Secret:
				return e.Object.(*core.Secret).Type == core.SecretTypeTLS
			default:
				return true
			}
		},
		CreateFunc: func(e event.CreateEvent) bool {
			switch e.Object.(type) {
			case *core.Secret:
				return e.Object.(*core.Secret).Type == core.SecretTypeTLS
			default:
				return true
			}
		},
		GenericFunc: func(e event.GenericEvent) bool {
			switch e.Object.(type) {
			case *core.Secret:
				return e.Object.(*core.Secret).Type == core.SecretTypeTLS
			default:
				return true
			}
		},
	}
}
