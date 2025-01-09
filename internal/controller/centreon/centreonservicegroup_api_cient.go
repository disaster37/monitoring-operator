package centreon

import (
	"github.com/disaster37/generic-objectmatcher/patch"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type centreonServiceGroupApiClient struct {
	*controller.BasicRemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler]
	logger *logrus.Entry
}

func newCentreonServiceGroupApiClient(client centreonhandler.CentreonHandler, logger *logrus.Entry) controller.RemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler] {
	return &centreonServiceGroupApiClient{
		BasicRemoteExternalReconciler: controller.NewBasicRemoteExternalReconciler[*centreoncrd.CentreonServiceGroup, *CentreonServiceGroup, centreonhandler.CentreonHandler](client),
		logger:                        logger,
	}
}

func (h *centreonServiceGroupApiClient) Build(o *centreoncrd.CentreonServiceGroup) (csg *CentreonServiceGroup, err error) {
	csg = &CentreonServiceGroup{
		CentreonServiceGroup: &centreonhandler.CentreonServiceGroup{
			Name:        o.GetExternalName(),
			Activated:   helpers.BoolToString(&o.Spec.Activated),
			Comment:     "Managed by monitoring-operator",
			Description: o.Spec.Description,
		},
	}

	return csg, nil
}

func (h *centreonServiceGroupApiClient) Get(o *centreoncrd.CentreonServiceGroup) (object *CentreonServiceGroup, err error) {
	var serviceGroupName string

	// Check if the current serviceGroup name is right before to search on Centreon
	// Maybee we should to change it name
	if o.Status.ServiceGroupName != "" {
		serviceGroupName = o.Status.ServiceGroupName
	} else {
		serviceGroupName = o.GetExternalName()
	}

	csg, err := h.Client().GetServiceGroup(serviceGroupName)
	if err != nil {
		return nil, err
	}

	if csg == nil {
		return nil, nil
	}

	object = &CentreonServiceGroup{
		CentreonServiceGroup: csg,
	}

	return object, nil
}

func (h *centreonServiceGroupApiClient) Create(object *CentreonServiceGroup, o *centreoncrd.CentreonServiceGroup) (err error) {
	// Check policy
	if o.Spec.Policy.NoCreate {
		h.logger.Info("Skip create service (policy NoCreate)")
		return nil
	}

	// Create service on Centreon
	return h.Client().CreateServiceGroup(object.CentreonServiceGroup)
}

func (h *centreonServiceGroupApiClient) Update(object *CentreonServiceGroup, o *centreoncrd.CentreonServiceGroup) (err error) {
	// Check policy
	if o.Spec.Policy.NoUpdate {
		h.logger.Info("Skip update service (policy NoUpdate)")
		return nil
	}

	// Update service on Centreon
	return h.Client().UpdateServiceGroup(object.CentreonServiceGroupDiff)
}

func (h *centreonServiceGroupApiClient) Delete(o *centreoncrd.CentreonServiceGroup) (err error) {
	// Check policy
	if o.Spec.Policy.NoDelete {
		h.logger.Info("Skip delete service (policy NoDelete)")
		return nil
	}

	return h.Client().DeleteServiceGroup(o.GetExternalName())
}

func (h *centreonServiceGroupApiClient) Diff(currentOject *CentreonServiceGroup, expectedObject *CentreonServiceGroup, originalObject *CentreonServiceGroup, o *centreoncrd.CentreonServiceGroup, ignoresDiff ...patch.CalculateOption) (patchResult *patch.PatchResult, err error) {
	patchResult = &patch.PatchResult{}

	csDiff, err := h.Client().DiffServiceGroup(currentOject.CentreonServiceGroup, expectedObject.CentreonServiceGroup, o.Spec.Policy.ExcludeFieldsOnDiff)
	if err != nil {
		return nil, errors.Wrap(err, "Error when diff CentreonServiceGroup")
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
