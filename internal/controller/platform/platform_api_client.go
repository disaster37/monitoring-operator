package platform

import (
	"os"

	"emperror.dev/errors"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/sirupsen/logrus"
)

type platformApiClient struct {
	*controller.BasicRemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler]
	logger    *logrus.Entry
	platforms map[string]*ComputedPlatform
}

func newPlaformApiClient(client centreonhandler.CentreonHandler, logger *logrus.Entry, platforms map[string]*ComputedPlatform) controller.RemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler] {
	return &platformApiClient{
		BasicRemoteExternalReconciler: controller.NewBasicRemoteExternalReconciler[*centreoncrd.Platform, *ComputedPlatform, centreonhandler.CentreonHandler](client),
		logger:                        logger,
		platforms:                     platforms,
	}
}

func (h *platformApiClient) Build(o *centreoncrd.Platform) (p *ComputedPlatform, err error) {
	return &ComputedPlatform{
		Platform: o,
	}, nil
}

func (h *platformApiClient) Get(o *centreoncrd.Platform) (object *ComputedPlatform, err error) {
	object = h.platforms[o.Name]

	return object, nil
}

func (h *platformApiClient) Create(object *ComputedPlatform, o *centreoncrd.Platform) (err error) {
	if os.Getenv("TEST") != "true" {
		if err = object.Client.(centreonhandler.CentreonHandler).Auth(); err != nil {
			return errors.Wrapf(err, "Error when authentificate on platform %s", o.Name)
		}
	}
	if o.Spec.IsDefault {
		h.platforms["default"] = object
	}
	h.platforms[o.Name] = object

	h.logger.Infof("Add platform '%s'", o.Name)
	if o.Spec.IsDefault {
		h.logger.Infof("Platform '%s' is the default", o.Name)
	}

	return nil
}

func (h *platformApiClient) Update(object *ComputedPlatform, o *centreoncrd.Platform) (err error) {
	return h.Create(object, o)
}

func (h *platformApiClient) Delete(o *centreoncrd.Platform) (err error) {
	if o.Spec.IsDefault {
		delete(h.platforms, "default")
	}
	delete(h.platforms, o.Name)

	h.logger.Infof("Remove platform '%s'", o.Name)

	return nil
}
