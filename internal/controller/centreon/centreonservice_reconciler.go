package centreon

import (
	"context"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/common"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/object"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type centreonServiceReconciler struct {
	controller.RemoteReconcilerAction[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler]
	name      string
	platforms map[string]*ComputedPlatform
}

func newCentreonServiceReconciler(name string, client client.Client, recorder record.EventRecorder) controller.RemoteReconcilerAction[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler] {
	return &centreonServiceReconciler{
		RemoteReconcilerAction: controller.NewRemoteReconcilerAction[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler](
			client,
			recorder,
		),
		name: name,
	}
}

func (h *centreonServiceReconciler) GetRemoteHandler(ctx context.Context, req ctrl.Request, o object.RemoteObject, logger *logrus.Entry) (handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], res ctrl.Result, err error) {
	cs := o.(*centreoncrd.CentreonService)

	meta, _, err := getClient(cs.Spec.PlatformRef, h.platforms)
	if err != nil {
		return nil, res, err
	}

	handler = newCentreonServiceApiClient(meta.(centreonhandler.CentreonHandler), logger)

	return handler, res, nil
}

func (h *centreonServiceReconciler) Configure(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], logger *logrus.Entry) (res ctrl.Result, err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(1)

	return h.RemoteReconcilerAction.Configure(ctx, o, data, handler, logger)
}

func (h *centreonServiceReconciler) Delete(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], logger *logrus.Entry) (err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	return h.RemoteReconcilerAction.Delete(ctx, o, data, handler, logger)
}

func (h *centreonServiceReconciler) OnError(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], currentErr error, logger *logrus.Entry) (res ctrl.Result, err error) {
	common.TotalErrors.Inc()
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Inc()

	return h.RemoteReconcilerAction.OnError(ctx, o, data, handler, currentErr, logger)
}

func (h *centreonServiceReconciler) OnSuccess(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler], diff controller.RemoteDiff[*CentreonService], logger *logrus.Entry) (res ctrl.Result, err error) {
	sg := o.(*centreoncrd.CentreonService)

	// Reset the current cluster errors
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	if diff.NeedCreate() || diff.NeedUpdate() {
		sg.Status.ServiceName = sg.GetExternalName()
		sg.Status.Host = sg.Spec.Host
	}

	return h.RemoteReconcilerAction.OnSuccess(ctx, o, data, handler, diff, logger)
}
