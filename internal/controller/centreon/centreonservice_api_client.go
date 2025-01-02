package centreon

import (
	"strings"

	"github.com/disaster37/generic-objectmatcher/patch"
	"github.com/disaster37/go-centreon-rest/v21/models"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type centreonServiceApiClient struct {
	*controller.BasicRemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler]
	logger *logrus.Entry
}

func newCentreonServiceApiClient(client centreonhandler.CentreonHandler, logger *logrus.Entry) controller.RemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler] {
	return &centreonServiceApiClient{
		BasicRemoteExternalReconciler: controller.NewBasicRemoteExternalReconciler[*centreoncrd.CentreonService, *CentreonService, centreonhandler.CentreonHandler](client),
		logger:                        logger,
	}
}

func (h *centreonServiceApiClient) Build(o *centreoncrd.CentreonService) (cs *CentreonService, err error) {
	cs = &CentreonService{
		CentreonService: &centreonhandler.CentreonService{
			Host:                o.Spec.Host,
			Name:                o.GetExternalName(),
			CheckCommand:        o.Spec.CheckCommand,
			CheckCommandArgs:    helpers.CheckArgumentsToString(o.Spec.Arguments),
			NormalCheckInterval: o.Spec.NormalCheckInterval,
			RetryCheckInterval:  o.Spec.RetryCheckInterval,
			MaxCheckAttempts:    o.Spec.MaxCheckAttempts,
			ActiveCheckEnabled:  helpers.BoolToString(o.Spec.ActiveCheckEnabled),
			PassiveCheckEnabled: helpers.BoolToString(o.Spec.PassiveCheckEnabled),
			Activated:           helpers.BoolToString(&o.Spec.Activated),
			Template:            o.Spec.Template,
			Comment:             "Managed by monitoring-operator",
			Groups:              o.Spec.Groups,
			Categories:          o.Spec.Categories,
			Macros:              make([]*models.Macro, 0, len(o.Spec.Macros)),
		},
	}

	for name, value := range o.Spec.Macros {
		macro := &models.Macro{
			Name:       strings.ToUpper(name),
			Value:      value,
			IsPassword: "0",
		}
		cs.CentreonService.Macros = append(cs.CentreonService.Macros, macro)
	}

	return cs, nil
}

func (h *centreonServiceApiClient) Get(o *centreoncrd.CentreonService) (object *CentreonService, err error) {
	var (
		host        string
		serviceName string
	)

	// Check if the current service name and host is right before to search on Centreon
	if o.Status.Host != "" && o.Status.ServiceName != "" {
		host = o.Status.Host
		serviceName = o.Status.ServiceName
	} else {
		host = o.GetExternalName()
		serviceName = o.Spec.Name
	}

	cs, err := h.Client().GetService(host, serviceName)
	if err != nil {
		return nil, err
	}

	if cs == nil {
		return nil, nil
	}

	object = &CentreonService{
		CentreonService: cs,
	}

	return object, nil
}

func (h *centreonServiceApiClient) Create(object *CentreonService, o *centreoncrd.CentreonService) (err error) {
	// Check policy
	if o.Spec.Policy.NoCreate {
		h.logger.Info("Skip create service (policy NoCreate)")
		return nil
	}

	// Create service on Centreon
	return h.Client().CreateService(object.CentreonService)
}

func (h *centreonServiceApiClient) Update(object *CentreonService, o *centreoncrd.CentreonService) (err error) {
	// Check policy
	if o.Spec.Policy.NoUpdate {
		h.logger.Info("Skip update service (policy NoUpdate)")
		return nil
	}

	// Update service on Centreon
	return h.Client().UpdateService(object.CentreonServiceDiff)
}

func (h *centreonServiceApiClient) Delete(o *centreoncrd.CentreonService) (err error) {
	// Check policy
	if o.Spec.Policy.NoDelete {
		h.logger.Info("Skip delete service (policy NoDelete)")
		return nil
	}

	return h.Client().DeleteService(o.Spec.Host, o.GetExternalName())

}

func (h *centreonServiceApiClient) Diff(currentOject *CentreonService, expectedObject *CentreonService, originalObject *CentreonService, o *centreoncrd.CentreonService, ignoresDiff ...patch.CalculateOption) (patchResult *patch.PatchResult, err error) {

	patchResult = &patch.PatchResult{}

	csDiff, err := h.Client().DiffService(currentOject.CentreonService, expectedObject.CentreonService, o.Spec.Policy.ExcludeFieldsOnDiff)
	if err != nil {
		return nil, errors.Wrap(err, "Error when diff CentreonService")
	}

	if csDiff.IsDiff {
		patchDiff, err := json.ConfigCompatibleWithStandardLibrary.Marshal(csDiff)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to convert patched object to byte sequence")
		}

		patchResult.Patch = patchDiff

	}

	return patchResult, nil
}
