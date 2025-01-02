package platform

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"emperror.dev/errors"
	"github.com/disaster37/generic-objectmatcher/patch"
	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/internal/controller/common"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/disaster37/operator-sdk-extra/pkg/object"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type platformReconciler struct {
	controller.RemoteReconcilerAction[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler]
	name      string
	platforms map[string]*ComputedPlatform
}

func newPlatformReconciler(name string, client client.Client, recorder record.EventRecorder, platforms map[string]*ComputedPlatform) controller.RemoteReconcilerAction[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler] {
	return &platformReconciler{
		RemoteReconcilerAction: controller.NewRemoteReconcilerAction[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler](
			client,
			recorder,
		),
		name:      name,
		platforms: platforms,
	}
}

func (h *platformReconciler) GetRemoteHandler(ctx context.Context, req ctrl.Request, o object.RemoteObject, logger *logrus.Entry) (handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], res ctrl.Result, err error) {

	handler = newPlaformApiClient(nil, logger, h.platforms)

	return handler, res, nil
}

func (h *platformReconciler) Read(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], logger *logrus.Entry) (read controller.RemoteRead[*ComputedPlatform], res ctrl.Result, err error) {
	read, res, err = h.RemoteReconcilerAction.Read(ctx, o, data, handler, logger)
	if err != nil {
		return nil, res, err
	}

	p := o.(*centreoncrd.Platform)

	switch p.Spec.PlatformType {
	case "centreon":
		// Get secret
		s := &corev1.Secret{}
		k := types.NamespacedName{
			Namespace: p.Namespace,
			Name:      p.Spec.CentreonSettings.Secret,
		}
		if err = h.Client().Get(ctx, k, s); err != nil {
			if k8serrors.IsNotFound(err) {
				logger.Warnf("Secret %s not yet exist, try later", p.Spec.CentreonSettings.Secret)
				return nil, res, errors.Errorf("Secret %s not yet exist", p.Spec.CentreonSettings.Secret)
			}
		}

		computedPlatform, err := getComputedCentreonPlatform(p, s, logger)
		if err != nil {
			return nil, res, errors.Wrapf(err, "Error when compute platform %s", p.Name)
		}

		read.SetExpectedObject(computedPlatform)

	default:
		return nil, res, errors.Errorf("Plaform %s is not supported", p.Spec.PlatformType)
	}

	return read, res, nil
}

func (h *platformReconciler) Configure(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], logger *logrus.Entry) (res ctrl.Result, err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(1)

	return h.RemoteReconcilerAction.Configure(ctx, o, data, handler, logger)
}

func (h *platformReconciler) Delete(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], logger *logrus.Entry) (err error) {
	// Set prometheus Metrics
	common.ControllerInstances.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	return h.RemoteReconcilerAction.Delete(ctx, o, data, handler, logger)
}

func (h *platformReconciler) OnError(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], currentErr error, logger *logrus.Entry) (res ctrl.Result, err error) {
	common.TotalErrors.Inc()
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Inc()

	return h.RemoteReconcilerAction.OnError(ctx, o, data, handler, currentErr, logger)
}

func (h *platformReconciler) OnSuccess(ctx context.Context, o object.RemoteObject, data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], diff controller.RemoteDiff[*ComputedPlatform], logger *logrus.Entry) (res ctrl.Result, err error) {

	// Reset the current cluster errors
	common.ControllerErrors.WithLabelValues(h.name, o.GetNamespace(), o.GetName()).Set(0)

	return h.RemoteReconcilerAction.OnSuccess(ctx, o, data, handler, diff, logger)
}

func (h *platformReconciler) Diff(ctx context.Context, o object.RemoteObject, read controller.RemoteRead[*ComputedPlatform], data map[string]any, handler controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler], logger *logrus.Entry, ignoreDiff ...patch.CalculateOption) (diff controller.RemoteDiff[*ComputedPlatform], res ctrl.Result, err error) {

	diff = controller.NewBasicRemoteDiff[*ComputedPlatform]()

	// New platform
	if read.GetCurrentObject() == nil {
		diff.SetObjectToCreate(read.GetExpectedObject())
		diff.AddDiff("New plaform")

		return diff, res, nil
	}

	// Client change
	if read.GetCurrentObject().Hash != read.GetExpectedObject().Hash {
		diff.SetObjectToUpdate(read.GetExpectedObject())
		diff.AddDiff("Secret change on platform")
		return diff, res, nil
	}

	// Platform change
	diffStr := cmp.Diff(read.GetCurrentObject().Platform.Spec, read.GetExpectedObject().Platform.Spec)
	if diffStr != "" {
		diff.SetObjectToUpdate(read.GetExpectedObject())
		diff.AddDiff(diffStr)
		return diff, res, nil
	}

	return diff, res, nil
}

func getComputedCentreonPlatform(p *monitorapi.Platform, s *corev1.Secret, log *logrus.Entry) (cp *ComputedPlatform, err error) {

	if p == nil {
		return nil, errors.New("Platform can't be null")
	}
	if s == nil {
		return nil, errors.New("Secret can't be null")
	}

	username := string(s.Data["username"])
	password := string(s.Data["password"])
	if username == "" || password == "" {
		return nil, errors.Errorf("You need to set username and password on secret %s", s.Name)
	}

	// Create client
	cfg := &models.Config{
		Address:          p.Spec.CentreonSettings.URL,
		Username:         username,
		Password:         password,
		DisableVerifySSL: p.Spec.CentreonSettings.SelfSignedCertificate,
	}
	if log.Level == logrus.DebugLevel {
		cfg.Debug = true
	}
	client, err := centreon.NewClient(cfg)
	if err != nil {
		return nil, errors.Wrap(err, "Error when create Centreon client")
	}
	shaByte, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	sha := sha256.New()
	if _, err := sha.Write([]byte(shaByte)); err != nil {
		return nil, err
	}

	return &ComputedPlatform{
		Client:   centreonhandler.NewCentreonHandler(client, log),
		Platform: p,
		Hash:     hex.EncodeToString(sha.Sum(nil)),
	}, nil

}
