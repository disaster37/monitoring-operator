package namespace

import (
	"context"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/template"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	name string = "namespace"
)

// NamespaceReconciler reconciles a namespace
type NamespaceReconciler struct {
	controller.Controller
	controller.SentinelReconciler
	controller.SentinelReconcilerAction
	name string
}

func NewNamespaceReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder) (sentienelReconciler controller.Controller) {
	return &NamespaceReconciler{
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

//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=namespaces/finalizers,verbs=update
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
func (r *NamespaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	n := &corev1.Namespace{}
	data := map[string]any{}

	return r.SentinelReconciler.Reconcile(
		ctx,
		req,
		n,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&corev1.Namespace{}).
		Owns(&centreoncrd.CentreonService{}).
		Owns(&centreoncrd.CentreonServiceGroup{}).
		WithEventFilter(template.ViewResourceWithMonitoringTemplate()).
		Watches(&centreoncrd.Template{}, handler.EnqueueRequestsFromMapFunc(template.WatchTemplate(r.Client(), &corev1.NamespaceList{}))).
		Complete(r)
}
