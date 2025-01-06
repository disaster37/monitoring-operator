package route

import (
	"context"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/template"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
)

const (
	name string = "route"
)

// RouteReconciler reconciles a route
type RouteReconciler struct {
	controller.Controller
	controller.SentinelReconciler
	controller.SentinelReconcilerAction
	name string
}

func NewRouteReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder) (sentienelReconciler controller.Controller) {
	return &RouteReconciler{
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

//+kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="route.openshift.io",resources=routes/finalizers,verbs=update
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
func (r *RouteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	route := &routev1.Route{}
	data := map[string]any{}

	return r.SentinelReconciler.Reconcile(
		ctx,
		req,
		route,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *RouteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&routev1.Route{}).
		Owns(&centreoncrd.CentreonService{}).
		Owns(&centreoncrd.CentreonServiceGroup{}).
		WithEventFilter(template.ViewResourceWithMonitoringTemplate()).
		Watches(&centreoncrd.Template{}, handler.EnqueueRequestsFromMapFunc(template.WatchTemplate(r.Client(), &routev1.RouteList{}))).
		Complete(r)
}

func (h *RouteReconciler) Read(ctx context.Context, o client.Object, data map[string]any, logger *logrus.Entry) (read controller.SentinelRead, res ctrl.Result, err error) {
	r := o.(*routev1.Route)

	placeholders := map[string]any{}
	// Set route placeholders on same format as ingress
	rules := make([]map[string]any, 0, 1)
	rule := map[string]any{
		"host": r.Spec.Host,
	}
	if r.Spec.Path != "" {
		rule["paths"] = []string{r.Spec.Path}
	} else {
		rule["paths"] = []string{"/"}
	}
	if r.Spec.TLS != nil && r.Spec.TLS.Termination != "" {
		rule["scheme"] = "https"
	} else {
		rule["scheme"] = "http"
	}

	rules = append(rules, rule)
	placeholders["rules"] = rules

	data["placeholders"] = placeholders

	return h.SentinelReconcilerAction.Read(ctx, o, data, logger)
}
