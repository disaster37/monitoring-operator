package centreonhandler

import (
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

// CreateServiceGroup permit to create new serviceGroup on Centreon from spec
func (h *CentreonHandlerImpl) CreateServiceGroup(sg *CentreonServiceGroup) (err error) {
	if sg == nil {
		return errors.New("ServiceGroup must be provided")
	}
	if sg.Name == "" {
		return errors.New("ServiceGroup name must be provided")
	}
	if sg.Description == "" {
		return errors.New("ServiceGroup description must be provided")
	}

	// Create main object
	if err = h.client.API.ServiceGroup().Add(sg.Name, sg.Description); err != nil {
		return err
	}
	h.log.Debug("Create serviceGroup core from Centreon")

	// Set extra params
	params := map[string]string{
		"activate": sg.Activated,
		"comment":  sg.Comment,
	}
	for param, value := range params {
		if value != "" {
			if err = h.client.API.ServiceGroup().SetParam(sg.Name, param, value); err != nil {
				return err
			}
			h.log.Debugf("Set param %s on service from Centreon", param)
		}
	}

	h.log.Debug("Create serviceGroup successfully on Centreon")

	return nil
}

// UpdateServiceGroup permit to update existing serviceGroup on Centreon from spec
func (h *CentreonHandlerImpl) UpdateServiceGroup(serviceGroupDiff *CentreonServiceGroupDiff) (err error) {
	if serviceGroupDiff == nil {
		return errors.New("ServiceGroupDiff must be provided")
	}

	if serviceGroupDiff.Name == "" {
		return errors.New("ServiceGroup name must be provided")
	}

	if !serviceGroupDiff.IsDiff {
		h.log.Debug("No update needed, skip it")
		return nil
	}

	// Update properties
	if len(serviceGroupDiff.ParamsToSet) > 0 {
		for param, value := range serviceGroupDiff.ParamsToSet {
			err = h.client.API.ServiceGroup().SetParam(serviceGroupDiff.Name, param, value)
			if err != nil {
				return err
			}
			h.log.Debugf("Update param %s from Centreon", param)
			// Handle special param name "name". It change service name
			if param == "name" {
				serviceGroupDiff.Name = value
			}
		}
	}

	return nil
}

// DeleteServiceGroup permit to delete an existing serviceGroup on Centreon
func (h *CentreonHandlerImpl) DeleteServiceGroup(name string) (err error) {
	return h.client.API.ServiceGroup().Delete(name)
}

// DiffServiceGroup permit to diff actual and expected serviceGroup to know what it need to modify
func (h *CentreonHandlerImpl) DiffServiceGroup(actual, expected *CentreonServiceGroup, ignoreFields []string) (diff *CentreonServiceGroupDiff, err error) {
	diff = &CentreonServiceGroupDiff{
		Name:        actual.Name,
		IsDiff:      false,
		ParamsToSet: map[string]string{},
	}

	// Check params
	if !funk.Contains(ignoreFields, "name") && actual.Name != expected.Name {
		diff.ParamsToSet["name"] = expected.Name
	}
	if !funk.Contains(ignoreFields, "activate") && actual.Activated != expected.Activated {
		diff.ParamsToSet["activate"] = expected.Activated
	}
	if !funk.Contains(ignoreFields, "description") && actual.Description != expected.Description {
		diff.ParamsToSet["alias"] = expected.Description
	}
	if !funk.Contains(ignoreFields, "comment") && actual.Comment != expected.Comment {
		diff.ParamsToSet["comment"] = expected.Comment
	}

	// Compute IsDiff
	if len(diff.ParamsToSet) > 0 {
		diff.IsDiff = true
		h.log.Debugf("Some diff founds :%s", diff)
	} else {
		h.log.Debug("No diff found")
	}

	return diff, nil
}

// GetServiceGroup permit to get serviceGroup by it name
func (h *CentreonHandlerImpl) GetServiceGroup(name string) (sg *CentreonServiceGroup, err error) {
	if name == "" {
		return nil, errors.New("ServiceGroup name must be provided")
	}

	// Get serviceGroup from Centreon
	baseSG, err := h.client.API.ServiceGroup().Get(name)
	if err != nil {
		return nil, err
	}
	if baseSG == nil {
		return nil, nil
	}

	// Get extras params
	extras, err := h.client.API.ServiceGroup().GetParam(name, []string{"activate", "comment"})
	if err != nil {
		return nil, err
	}

	sg = &CentreonServiceGroup{
		Name:        name,
		Description: baseSG.Description,
		Comment:     extras["comment"],
		Activated:   extras["activate"],
	}

	h.log.Debugf("Actual serviceGroup: %s", sg)

	return sg, nil
}
