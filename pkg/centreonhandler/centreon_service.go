package centreonhandler

import (
	"strings"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/pkg/errors"
	"github.com/thoas/go-funk"
)

// CreateService permit to create new service on Centreon from spec
func (h *CentreonHandlerImpl) CreateService(service *CentreonService) (err error) {

	if service == nil {
		return errors.New("Service must be provided")
	}
	if service.Host == "" {
		return errors.New("Host must be provided")
	}
	if service.Name == "" {
		return errors.New("Service name must be provided")
	}

	// Create main object
	if err = h.client.API.Service().Add(service.Host, service.Name, service.Template); err != nil {
		return err
	}
	h.log.Debug("Create service core from Centreon")

	// Set extra params
	params := map[string]string{
		"check_command":           service.CheckCommand,
		"normal_check_interval":   service.NormalCheckInterval,
		"retry_check_interval":    service.RetryCheckInterval,
		"max_check_attempts":      service.MaxCheckAttempts,
		"check_command_arguments": service.CheckCommandArgs,
		"activate":                service.Activated,
		"active_checks_enabled":   service.ActiveCheckEnabled,
		"passive_checks_enabled":  service.PassiveCheckEnabled,
		"comment":                 service.Comment,
	}
	for param, value := range params {
		if value != "" {
			if err = h.client.API.Service().SetParam(service.Host, service.Name, param, value); err != nil {
				return err
			}
			h.log.Debugf("Set param %s on service from Centreon", param)
		}
	}

	// Set service groups
	if service.Groups != nil && len(service.Groups) > 0 {
		if err = h.client.API.Service().SetServiceGroups(service.Host, service.Name, service.Groups); err != nil {
			return err
		}
		h.log.Debugf("Set service groups %s from Centreon", strings.Join(service.Groups, "|"))
	}

	// Set categories
	if service.Categories != nil && len(service.Categories) > 0 {
		if err = h.client.API.Service().SetCategories(service.Host, service.Name, service.Categories); err != nil {
			return err
		}
		h.log.Debugf("Set categories %s from Centreon", strings.Join(service.Categories, "|"))
	}

	// Set macros
	if service.Macros != nil && len(service.Macros) > 0 {
		for _, macro := range service.Macros {
			if err = h.client.API.Service().SetMacro(service.Host, service.Name, macro); err != nil {
				return err
			}
			h.log.Debugf("Set macro %s from Centreon", macro.Name)
		}
	}

	h.log.Debug("Create service successfully on Centreon")

	return nil

}

// UpdateService permit to update existing service on Centreon from spec
func (h *CentreonHandlerImpl) UpdateService(serviceDiff *CentreonServiceDiff) (err error) {

	if serviceDiff == nil {
		return errors.New("ServiceDiff must be provided")
	}

	if serviceDiff.Host == "" {
		return errors.New("Host must be provided")
	}
	if serviceDiff.Name == "" {
		return errors.New("Service name must be provided")
	}

	if !serviceDiff.IsDiff {
		h.log.Debug("No update needed, skip it")
		return nil
	}

	// Update properties
	if len(serviceDiff.ParamsToSet) > 0 {
		for param, value := range serviceDiff.ParamsToSet {
			err = h.client.API.Service().SetParam(serviceDiff.Host, serviceDiff.Name, param, value)
			if err != nil {
				return err
			}
			h.log.Debugf("Update param %s from Centreon", param)
			// Handle special param name "description". It change service name
			if param == "description" {
				serviceDiff.Name = value
			}
		}
	}

	// Update service groups
	if len(serviceDiff.GroupsToSet) > 0 {
		err = h.client.API.Service().SetServiceGroups(serviceDiff.Host, serviceDiff.Name, serviceDiff.GroupsToSet)
		if err != nil {
			return err
		}
		h.log.Debugf("Set service groups %s from Centreon", strings.Join(serviceDiff.GroupsToSet, "|"))
	}
	if len(serviceDiff.GroupsToDelete) > 0 {
		err = h.client.API.Service().DeleteServiceGroups(serviceDiff.Host, serviceDiff.Name, serviceDiff.GroupsToDelete)
		if err != nil {
			return err
		}
		h.log.Debugf("Delete service groups %s from Centreon", strings.Join(serviceDiff.GroupsToDelete, "|"))
	}

	// Update categories
	if len(serviceDiff.CategoriesToSet) > 0 {
		err = h.client.API.Service().SetCategories(serviceDiff.Host, serviceDiff.Name, serviceDiff.CategoriesToSet)
		if err != nil {
			return err
		}
		h.log.Debugf("Set categories %s from Centreon", strings.Join(serviceDiff.CategoriesToSet, "|"))
	}
	if len(serviceDiff.CategoriesToDelete) > 0 {
		err = h.client.API.Service().DeleteCategories(serviceDiff.Host, serviceDiff.Name, serviceDiff.CategoriesToDelete)
		if err != nil {
			return err
		}
		h.log.Debugf("Delete categories %s from Centreon", strings.Join(serviceDiff.CategoriesToDelete, "|"))
	}

	// Update macros
	if len(serviceDiff.MacrosToSet) > 0 {
		for _, macro := range serviceDiff.MacrosToSet {
			err = h.client.API.Service().SetMacro(serviceDiff.Host, serviceDiff.Name, macro)
			if err != nil {
				return err
			}
			h.log.Debugf("Set macro %s from Centreon", macro.Name)
		}
	}
	if len(serviceDiff.MacrosToDelete) > 0 {
		for _, macro := range serviceDiff.MacrosToDelete {
			err = h.client.API.Service().DeleteMacro(serviceDiff.Host, serviceDiff.Name, macro.Name)
			if err != nil {
				return err
			}
			h.log.Debugf("Delete macro %s from Centreon", macro.Name)
		}
	}

	// Finnaly update host
	if serviceDiff.HostToSet != "" {
		if err = h.client.API.Service().SetHost(serviceDiff.Host, serviceDiff.Name, serviceDiff.HostToSet); err != nil {
			return err
		}
		h.log.Debugf("Set host %s on service %s from Centreon", serviceDiff.HostToSet, serviceDiff.Name)

		serviceDiff.Host = serviceDiff.HostToSet
	}

	return nil
}

// DeleteService permit to delete an existing service on Centreon
func (h *CentreonHandlerImpl) DeleteService(host, name string) (err error) {
	return h.client.API.Service().Delete(host, name)
}

func (h *CentreonHandlerImpl) DiffService(actual, expected *CentreonService) (diff *CentreonServiceDiff, err error) {
	diff = &CentreonServiceDiff{
		Host:           actual.Host,
		Name:           actual.Name,
		IsDiff:         false,
		ParamsToSet:    map[string]string{},
		MacrosToSet:    make([]*models.Macro, 0),
		MacrosToDelete: make([]*models.Macro, 0),
	}

	// Check params
	if actual.Name != expected.Name {
		diff.ParamsToSet["description"] = expected.Name
	}
	if actual.Activated != expected.Activated {
		diff.ParamsToSet["activate"] = expected.Activated
	}
	if actual.ActiveCheckEnabled != expected.ActiveCheckEnabled {
		diff.ParamsToSet["active_checks_enabled"] = expected.ActiveCheckEnabled
	}
	if actual.CheckCommand != expected.CheckCommand {
		diff.ParamsToSet["check_command"] = expected.CheckCommand
	}
	if actual.CheckCommandArgs != expected.CheckCommandArgs {
		diff.ParamsToSet["check_command_arguments"] = expected.CheckCommandArgs
	}
	if actual.MaxCheckAttempts != expected.MaxCheckAttempts {
		diff.ParamsToSet["max_check_attempts"] = expected.MaxCheckAttempts
	}
	if actual.NormalCheckInterval != expected.NormalCheckInterval {
		diff.ParamsToSet["normal_check_interval"] = expected.NormalCheckInterval
	}
	if actual.PassiveCheckEnabled != expected.PassiveCheckEnabled {
		diff.ParamsToSet["passive_checks_enabled"] = expected.PassiveCheckEnabled
	}
	if actual.RetryCheckInterval != expected.RetryCheckInterval {
		diff.ParamsToSet["retry_check_interval"] = expected.RetryCheckInterval
	}
	if actual.Template != expected.Template {
		diff.ParamsToSet["template"] = expected.Template
	}
	if actual.Comment != expected.Comment {
		diff.ParamsToSet["comment"] = expected.Comment
	}

	// Check the host
	if actual.Host != expected.Host {
		diff.HostToSet = expected.Host
	}

	// Check the service groups
	sgNeed, sgDelete := funk.Difference(expected.Groups, actual.Groups)
	diff.GroupsToSet = sgNeed.([]string)
	diff.GroupsToDelete = sgDelete.([]string)

	// Check the categories
	catNeed, catDelete := funk.Difference(expected.Categories, actual.Categories)
	diff.CategoriesToSet = catNeed.([]string)
	diff.CategoriesToDelete = catDelete.([]string)

	// Check macros
	if actual.Macros == nil {
		actual.Macros = make([]*models.Macro, 0)
	}
	if expected.Macros == nil {
		expected.Macros = make([]*models.Macro, 0)
	}

	macros := make([]*models.Macro, len(actual.Macros))
	copy(macros, actual.Macros)
	// The IsPassword macro with value "" or "0" is the same
	// We mitigeate this behavior here
	for _, macro := range macros {
		if macro.IsPassword == "" {
			macro.IsPassword = "0"
		}
	}
	for _, macro := range expected.Macros {
		if macro.IsPassword == "" {
			macro.IsPassword = "0"
		}
	}
	for _, expectedMacro := range expected.Macros {
		isFound := false
		for i, actualMacro := range macros {
			if actualMacro.Name == expectedMacro.Name {
				if actualMacro.Value == expectedMacro.Value && actualMacro.IsPassword == expectedMacro.IsPassword {
					isFound = true
				}
				macros = append(macros[:i], macros[i+1:]...)
				break
			}
		}

		if !isFound {
			diff.MacrosToSet = append(diff.MacrosToSet, expectedMacro)
		}
	}
	// Remove indirect macro herited by templates or command (direct and null value)
	// There are no way to differentiate macro setted between service and command
	for _, macro := range macros {
		if macro.Source == "direct" && macro.Value != "" {
			diff.MacrosToDelete = append(diff.MacrosToDelete, macro)
		}
	}

	// Compute IsDiff
	if len(diff.ParamsToSet) > 0 || len(diff.CategoriesToDelete) > 0 || len(diff.CategoriesToSet) > 0 || len(diff.GroupsToDelete) > 0 || len(diff.GroupsToSet) > 0 || len(diff.MacrosToDelete) > 0 || len(diff.MacrosToSet) > 0 || diff.HostToSet != "" {
		diff.IsDiff = true
		h.log.Debugf("Some diff founds :%s", diff)
	} else {
		h.log.Debug("No diff found")
	}

	return diff, nil
}

func (h *CentreonHandlerImpl) GetService(host, name string) (service *CentreonService, err error) {
	if host == "" {
		return nil, errors.New("Host must be provided")
	}
	if name == "" {
		return nil, errors.New("Service name must be provided")
	}

	// Get service from Centreon
	baseService, err := h.client.API.Service().Get(host, name)
	if err != nil {
		return nil, err
	}
	if baseService == nil {
		return nil, nil
	}

	// Get extras params
	extras, err := h.client.API.Service().GetParam(host, name, []string{"template", "comment"})
	if err != nil {
		return nil, err
	}

	// Get service groups
	sgs, err := h.client.API.Service().GetServiceGroups(host, name)
	if err != nil {
		return nil, err
	}

	// Get catgeories
	cats, err := h.client.API.Service().GetCategories(host, name)
	if err != nil {
		return nil, err
	}

	// Get macros
	macros, err := h.client.API.Service().GetMacros(host, name)
	if err != nil {
		return nil, err
	}

	service = &CentreonService{
		Host:                host,
		Name:                name,
		Template:            extras["template"],
		Comment:             extras["comment"],
		CheckCommand:        baseService.CheckCommand,
		CheckCommandArgs:    baseService.CheckCommandArgs,
		NormalCheckInterval: baseService.NormalCheckInterval,
		RetryCheckInterval:  baseService.RetryCheckInterval,
		MaxCheckAttempts:    baseService.MaxCheckAttempts,
		ActiveCheckEnabled:  baseService.ActiveCheckEnabled,
		PassiveCheckEnabled: baseService.PassiveCheckEnabled,
		Activated:           baseService.Activated,
		Macros:              macros,
		Categories:          cats,
		Groups:              sgs,
	}

	h.log.Debugf("Actual service: %s", service)

	return service, nil
}
