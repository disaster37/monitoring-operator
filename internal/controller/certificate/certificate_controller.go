package certificate

import (
	"context"
	"crypto/x509"
	"encoding/pem"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/template"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
)

const (
	name string = "certificate"
)

// CertificateReconciler reconciles a Secret (Certificate type)
type CertificateReconciler struct {
	controller.Controller
	controller.SentinelReconciler
	controller.SentinelReconcilerAction
	name string
}

func NewCertificateReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder) (sentienelReconciler controller.Controller) {
	return &CertificateReconciler{
		Controller: controller.NewBasicController(),
		SentinelReconciler: controller.NewBasicSentinelReconciler(
			client,
			name,
			logger,
			recorder,
		),
		SentinelReconcilerAction: template.NewTemplateReconciler(client, recorder),
		name:                     name,
	}
}

//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=secrets/finalizers,verbs=update
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=monitor.k8s.webcenter.fr,resources=centreonservicegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=patch;get;create

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cerebro object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.0/pkg/reconcile
func (r *CertificateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	s := &corev1.Secret{}
	data := map[string]any{}

	return r.SentinelReconciler.Reconcile(
		ctx,
		req,
		s,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertificateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&corev1.Secret{}).
		Owns(&centreoncrd.CentreonService{}).
		Owns(&centreoncrd.CentreonServiceGroup{}).
		WithOptions(k8scontroller.Options{
			RateLimiter: helper.DefaultControllerRateLimiter[reconcile.Request](),
		}).
		WithEventFilter(predicate.And(template.ViewResourceWithMonitoringTemplate(), viewCertificate())).
		Watches(&centreoncrd.Template{}, handler.EnqueueRequestsFromMapFunc(template.WatchTemplate(r.Client(), &corev1.SecretList{}))).
		Complete(r)
}

func (r *CertificateReconciler) Read(ctx context.Context, o client.Object, data map[string]any, logger *logrus.Entry) (read controller.SentinelRead, res ctrl.Result, err error) {
	s := o.(*corev1.Secret)

	placeholders := map[string]any{}

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
			break
		}
		blocks = append(blocks, block.Bytes...)
		if len(rest) == 0 {
			break
		}
	}

	if len(blocks) > 0 {
		certs, err := x509.ParseCertificates(blocks)
		if err != nil {
			logger.Errorf("Error when read TLS certificate: %s", err.Error())
		}
		placeholders["certificates"] = certs
	}

	data["placeholders"] = placeholders

	return r.SentinelReconcilerAction.Read(ctx, o, data, logger)
}

func viewCertificate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			switch e.ObjectOld.(type) {
			case *corev1.Secret:
				return e.ObjectOld.(*corev1.Secret).Type == corev1.SecretTypeTLS
			default:
				return true
			}
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			switch e.Object.(type) {
			case *corev1.Secret:
				return e.Object.(*corev1.Secret).Type == corev1.SecretTypeTLS
			default:
				return true
			}
		},
		CreateFunc: func(e event.CreateEvent) bool {
			switch e.Object.(type) {
			case *corev1.Secret:
				return e.Object.(*corev1.Secret).Type == corev1.SecretTypeTLS
			default:
				return true
			}
		},
		GenericFunc: func(e event.GenericEvent) bool {
			switch e.Object.(type) {
			case *corev1.Secret:
				return e.Object.(*corev1.Secret).Type == corev1.SecretTypeTLS
			default:
				return true
			}
		},
	}
}
