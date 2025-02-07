package ingress

import (
	"context"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/template"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/sirupsen/logrus"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	k8scontroller "sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	name string = "ingress"
)

// IngressReconciler reconciles a ingress
type IngressReconciler struct {
	controller.Controller
	controller.SentinelReconciler
	controller.SentinelReconcilerAction
	name string
}

func NewIngressReconciler(client client.Client, logger *logrus.Entry, recorder record.EventRecorder) (sentienelReconciler controller.Controller) {
	return &IngressReconciler{
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

//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;update
//+kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses/finalizers,verbs=update
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
func (r *IngressReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	i := &networkv1.Ingress{}
	data := map[string]any{}

	return r.SentinelReconciler.Reconcile(
		ctx,
		req,
		i,
		data,
		r,
	)
}

// SetupWithManager sets up the controller with the Manager.
func (r *IngressReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// Uncomment the following line adding a pointer to an instance of the controlled resource as an argument
		Named(r.name).
		For(&networkv1.Ingress{}).
		Owns(&centreoncrd.CentreonService{}).
		Owns(&centreoncrd.CentreonServiceGroup{}).
		WithOptions(k8scontroller.Options{
			RateLimiter: helper.DefaultControllerRateLimiter[reconcile.Request](),
		}).
		WithEventFilter(template.ViewResourceWithMonitoringTemplate()).
		Watches(&centreoncrd.Template{}, handler.EnqueueRequestsFromMapFunc(template.WatchTemplate(r.Client(), &networkv1.IngressList{}))).
		Complete(r)
}

func (r *IngressReconciler) Read(ctx context.Context, o client.Object, data map[string]any, logger *logrus.Entry) (read controller.SentinelRead, res ctrl.Result, err error) {
	i := o.(*networkv1.Ingress)

	placeholders := map[string]any{}
	rules := make([]map[string]any, 0, len(i.Spec.Rules))
	for _, rule := range i.Spec.Rules {
		r := map[string]any{
			"host":   rule.Host,
			"scheme": "http",
		}

		// Check if scheme is https
		for _, tls := range i.Spec.TLS {
			for _, host := range tls.Hosts {
				if host == rule.Host {
					r["scheme"] = "https"
				}
			}
		}

		// Add path
		paths := make([]string, 0, len(rule.HTTP.Paths))
		for _, path := range rule.HTTP.Paths {
			paths = append(paths, path.Path)
		}
		r["paths"] = paths
		rules = append(rules, r)
	}
	placeholders["rules"] = rules

	data["placeholders"] = placeholders

	return r.SentinelReconcilerAction.Read(ctx, o, data, logger)
}
