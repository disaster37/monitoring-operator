package centreon

import (
	"context"
	"encoding/json"
	"fmt"

	"emperror.dev/errors"
	"github.com/disaster37/generic-objectmatcher/patch"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/common"
	"github.com/disaster37/monitoring-operator/internal/controller/platform"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/object"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type centreonServiceGroupReconciler struct {
	controller.RemoteReconcilerAction[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler]
	name      string
	platforms map[string]*platform.ComputedPlatform
}

func newCentreonServiceGroupReconciler(name string, client client.Client, recorder record.EventRecorder, platforms map[string]*platform.ComputedPlatform) controller.RemoteReconcilerAction[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler] {
	return &centreonServiceGroupReconciler{
		RemoteReconcilerAction: controller.NewRemoteReconcilerAction[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler](
			client,
			recorder,
		),
		name:      name,
		platforms: platforms,
	}
}

func (h *centreonServiceGroupReconciler) GetRemoteHandler(ctx context.Context, req ctrl.Request, o object.RemoteObject, logger *logrus.Entry) (handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], res ctrl.Result, err error) {
	cs := o.(*centreoncrd.CentreonServiceGroup)

	meta, _, err := platform.GetClient(cs.Spec.PlatformRef, h.platforms)
	if err != nil {
		return nil, res, err
	}

	handler = newCentreonServiceGroupApiClient(meta.(centreonhandler.CentreonHandler), logger)

	return handler, res, nil
}

func (h *centreonServiceGroupReconciler) Configure(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], logger *logrus.Entry) (res ctrl.Result, err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(1)

	return h.RemoteReconcilerAction.Configure(ctx, o, data, handler, logger)
}

func (h *centreonServiceGroupReconciler) Delete(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], logger *logrus.Entry) (err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	return h.RemoteReconcilerAction.Delete(ctx, o, data, handler, logger)
}

func (h *centreonServiceGroupReconciler) OnError(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], currentErr error, logger *logrus.Entry) (res ctrl.Result, err error) {
	common.TotalErrors.Inc()
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Inc()

	return h.RemoteReconcilerAction.OnError(ctx, o, data, handler, currentErr, logger)
}

func (h *centreonServiceGroupReconciler) OnSuccess(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], diff controller.RemoteDiff[*CentreonServiceGroup], logger *logrus.Entry) (res ctrl.Result, err error) {
	sg := o.(*centreoncrd.CentreonServiceGroup)

	// Reset the current cluster errors
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	if diff.NeedCreate() || diff.NeedUpdate() {
		sg.Status.ServiceGroupName = sg.GetExternalName()
	}

	return h.RemoteReconcilerAction.OnSuccess(ctx, o, data, handler, diff, logger)
}

func (h *centreonServiceGroupReconciler) Diff(ctx context.Context, o object.RemoteObject, read controller.RemoteRead[*CentreonServiceGroup], data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler], logger *logrus.Entry, ignoreDiff ...patch.CalculateOption) (diff controller.RemoteDiff[*CentreonServiceGroup], res ctrl.Result, err error) {
	// Get the original object from status to use 3-way diff

	originalObject := new(CentreonServiceGroup)
	if o.GetStatus().GetLastAppliedConfiguration() != "" {
		if err = helper.UnZipBase64Decode(o.GetStatus().GetLastAppliedConfiguration(), originalObject); err != nil {
			return diff, res, errors.Wrap(err, "Error when create object from 'lastAppliedConfiguration'")
		}
	}

	diff = controller.NewBasicRemoteDiff[*CentreonServiceGroup]()

	// Check if need to create object on remote
	if read.GetCurrentObject() == nil {
		diff.SetObjectToCreate(read.GetExpectedObject())
		diff.AddDiff(fmt.Sprintf("Need to create new object %s on remote target", o.GetName()))

		return diff, res, nil
	}

	differ, err := handler.Diff(read.GetCurrentObject(), read.GetExpectedObject(), originalObject, o.(*centreoncrd.CentreonServiceGroup), ignoreDiff...)
	if err != nil {
		return diff, res, errors.Wrapf(err, "Error when diffing %s for remote target", o.GetName())
	}

	if !differ.IsEmpty() {
		csgDiff := &centreonhandler.CentreonServiceGroupDiff{}
		if err = json.Unmarshal(differ.Patch, csgDiff); err != nil {
			return diff, res, errors.Wrap(err, "Error when unmarshall the CentreonServiceGroup patch")
		}
		diff.AddDiff(string(differ.Patch))
		csg := read.GetExpectedObject()
		csg.CentreonServiceGroupDiff = csgDiff
		diff.SetObjectToUpdate(csg)
	}

	return diff, res, nil
}
