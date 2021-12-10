package centreonhandler

import (
	"fmt"
	"strings"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

// CreateService permit to create new service on Centreon from spec
func (h *CentreonHandlerImpl) CreateService(spec *v1alpha1.CentreonServiceSpec) (err error) {

	// Create main object
	if err = h.client.API.Service().Add(spec.Host, spec.Name, spec.Template); err != nil {
		return err
	}
	h.log.Debug("Create service core from Centreon")

	// Set extra params
	params := map[string]string{
		"check_command":           spec.CheckCommand,
		"normal_check_interval":   spec.NormalCheckInterval,
		"retry_check_interval":    spec.RetryCheckInterval,
		"max_check_attempts":      spec.MaxCheckAttempts,
		"check_command_arguments": helpers.CheckArgumentsToString(spec.Arguments),
		"activate":                helpers.BoolToString(&spec.Activated),
		"active_checks_enabled":   helpers.BoolToString(spec.ActiveCheckEnabled),
		"passive_checks_enabled":  helpers.BoolToString(spec.PassiveCheckEnabled),
		"comment":                 fmt.Sprintf("Created by monitoring-opeator at %s", time.Now()),
	}
	for param, value := range params {
		if value != "" {
			if err = h.client.API.Service().SetParam(spec.Host, spec.Name, param, value); err != nil {
				return err
			}
			h.log.Debugf("Set param %s on service from Centreon", param)
		}
	}

	// Set service groups
	if spec.Groups != nil && len(spec.Groups) > 0 {
		if err = h.client.API.Service().SetServiceGroups(spec.Host, spec.Name, spec.Groups); err != nil {
			return err
		}
		h.log.Debugf("Set service groups %s from Centreon", strings.Join(spec.Groups, "|"))
	}

	// Set categories
	if spec.Categories != nil && len(spec.Categories) > 0 {
		if err = h.client.API.Service().SetCategories(spec.Host, spec.Name, spec.Categories); err != nil {
			return err
		}
		h.log.Debugf("Set categories %s from Centreon", strings.Join(spec.Categories, "|"))
	}

	// Set macros
	if spec.Macros != nil && len(spec.Macros) > 0 {
		for key, value := range spec.Macros {
			macro := &models.Macro{
				Name:       key,
				Value:      value,
				IsPassword: "0",
			}
			if err = h.client.API.Service().SetMacro(spec.Host, spec.Name, macro); err != nil {
				return err
			}
			h.log.Debugf("Set macro %s from Centreon", key)
		}
	}

	h.log.Debug("Create service successfully on Centreon")

	return nil

}

// UpdateService permit to update existing service on Centreon from spec
func (h *CentreonHandlerImpl) UpdateService(spec *v1alpha1.CentreonServiceSpec) (err error) {
	// Get service from Centreon
	service, err := h.client.API.Service().Get(spec.Host, spec.Name)
	if err != nil {
		return err
	}
	if service == nil {
		return errors.Errorf("Service %s/%s not found on Centreon", spec.Host, spec.Name)
	}

	// Get extras params
	extras, err := h.client.API.Service().GetParam(spec.Host, spec.Name, []string{"template"})
	if err != nil {
		return err
	}

	// Get service groups
	sgs, err := h.client.API.Service().GetServiceGroups(spec.Host, spec.Name)
	if err != nil {
		return err
	}

	// Get catgeories
	cats, err := h.client.API.Service().GetCategories(spec.Host, spec.Name)
	if err != nil {
		return err
	}

	// Get macros
	macros, err := h.client.API.Service().GetMacros(spec.Host, spec.Name)
	if err != nil {
		return err
	}

	propertiesChange := map[string]string{}
	// Check the main properties
	if helpers.BoolToString(&spec.Activated) != service.Activated {
		propertiesChange["activate"] = helpers.BoolToString(&spec.Activated)
	}
	if helpers.BoolToString(spec.ActiveCheckEnabled) != service.ActiveCheckEnabled {
		propertiesChange["active_checks_enabled"] = helpers.BoolToString(spec.ActiveCheckEnabled)
	}
	if spec.CheckCommand != service.CheckCommand {
		propertiesChange["check_command"] = spec.CheckCommand
	}
	if helpers.CheckArgumentsToString(spec.Arguments) != service.CheckCommandArgs {
		propertiesChange["check_command_arguments"] = helpers.CheckArgumentsToString(spec.Arguments)
	}
	if spec.MaxCheckAttempts != service.MaxCheckAttempts {
		propertiesChange["max_check_attempts"] = spec.MaxCheckAttempts
	}
	if spec.NormalCheckInterval != service.NormalCheckInterval {
		propertiesChange["normal_check_interval"] = spec.NormalCheckInterval
	}
	if helpers.BoolToString(spec.PassiveCheckEnabled) != service.PassiveCheckEnabled {
		propertiesChange["passive_checks_enabled"] = helpers.BoolToString(spec.PassiveCheckEnabled)
	}
	if spec.RetryCheckInterval != service.RetryCheckInterval {
		propertiesChange["retry_check_interval"] = spec.RetryCheckInterval
	}
	if spec.Template != extras["template"] {
		propertiesChange["template"] = spec.Template
	}

	// Check the service groups
	sgNeededTmp, sgDeleteTmp := funk.Difference(spec.Groups, sgs)
	sgNeeded := sgNeededTmp.([]string)
	sgDelete := sgDeleteTmp.([]string)

	// Check the categories
	catNeededTmp, catDeleteTmp := funk.Difference(spec.Categories, cats)
	catNeeded := catNeededTmp.([]string)
	catDelete := catDeleteTmp.([]string)

	// Check macros
	macrosNeeded := make([]*models.Macro, 0)
	for key, value := range spec.Macros {
		isFound := false
		for i, m := range macros {
			if key == m.Name {
				if value == m.Value {
					isFound = true
				}
				macros = append(macros[:i], macros[i+1:]...)
				break
			}
		}

		if !isFound {
			macro := &models.Macro{
				Name:       key,
				Value:      value,
				IsPassword: "0",
			}
			macrosNeeded = append(macrosNeeded, macro)
		}
	}

	// Operate changes
	if len(propertiesChange) > 0 {
		for param, value := range propertiesChange {
			err = h.client.API.Service().SetParam(spec.Host, spec.Name, param, value)
			if err != nil {
				return err
			}
			h.log.Debugf("Update param %s from Centreon", param)
		}
	}

	if len(sgNeeded) > 0 {
		err = h.client.API.Service().SetServiceGroups(spec.Host, spec.Name, sgNeeded)
		if err != nil {
			return err
		}
		h.log.Debugf("Set service groups %s from Centreon", strings.Join(sgNeeded, "|"))
	}
	if len(sgDelete) > 0 {
		err = h.client.API.Service().DeleteServiceGroups(spec.Host, spec.Name, sgDelete)
		if err != nil {
			return err
		}
		h.log.Debugf("Delete service groups %s from Centreon", strings.Join(sgDelete, "|"))
	}
	if len(catNeeded) > 0 {
		err = h.client.API.Service().SetCategories(spec.Host, spec.Name, catNeeded)
		if err != nil {
			return err
		}
		h.log.Debugf("Set categories %s from Centreon", strings.Join(catNeeded, "|"))
	}
	if len(catDelete) > 0 {
		err = h.client.API.Service().DeleteCategories(spec.Host, spec.Name, catDelete)
		if err != nil {
			return err
		}
		h.log.Debugf("Delete categories %s from Centreon", strings.Join(catDelete, "|"))
	}
	if len(macrosNeeded) > 0 {
		for _, macro := range macrosNeeded {
			err = h.client.API.Service().SetMacro(spec.Host, spec.Name, macro)
			if err != nil {
				return err
			}
			h.log.Debugf("Set macro %s from Centreon", macro.Name)
		}
	}
	if len(macros) > 0 {
		for _, macro := range macros {
			err = h.client.API.Service().DeleteMacro(spec.Host, spec.Name, macro.Name)
			if err != nil {
				return err
			}
			h.log.Debugf("Delete macro %s from Centreon", macro.Name)
		}
	}

	h.log.Debug("Update service successfully on Centreon")

	return nil
}

// DeleteService permit to delete an existing service on Centreon
func (h *CentreonHandlerImpl) DeleteService(spec *v1alpha1.CentreonServiceSpec) (err error) {
	return h.client.API.Service().Delete(spec.Host, spec.Name)
}
