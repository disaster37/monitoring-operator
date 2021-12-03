package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/disaster37/go-centreon-rest/v21"
	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
)

func isServicealreadyExist(client *centreon.Client, log *logrus.Entry, host, service string) (isExist bool, err error) {

	s, err := client.API.Service().Get(host, service)
	if err != nil {
		return false, err
	}

	if s == nil {
		return false, nil
	}
	return true, nil
}

func createService(client *centreon.Client, log *logrus.Entry, serviceSpec *v1alpha1.CentreonServiceSpec) (err error) {
	// Require before
	//   - Compute default value from centreon config
	//   - Replace tags by value on macros, arguments, service name

	// Create main object
	if err = client.API.Service().Add(serviceSpec.Host, serviceSpec.Name, serviceSpec.Template); err != nil {
		return err
	}
	log.Debugf("Create service core %s/%s", serviceSpec.Host, serviceSpec.Name)

	// Set extra params
	params := map[string]string{
		"check_command":           serviceSpec.CheckCommand,
		"normal_check_interval":   serviceSpec.NormalCheckInterval,
		"retry_check_interval":    serviceSpec.RetryCheckInterval,
		"max_check_attempts":      serviceSpec.MaxCheckAttempts,
		"check_command_arguments": checkArgumentsToString(serviceSpec.Arguments),
		"activate":                boolToString(serviceSpec.Activated),
		"active_checks_enabled":   boolToString(serviceSpec.ActiveCheckEnabled),
		"passive_checks_enabled":  boolToString(serviceSpec.PassiveCheckEnabled),
		"comment":                 fmt.Sprintf("Created by k8s opeator at %s", time.Now()),
	}
	for param, value := range params {
		if value != "" {
			if err = client.API.Service().SetParam(serviceSpec.Host, serviceSpec.Name, param, value); err != nil {
				return err
			}
			log.Debugf("Set param %s on service %s/%s", serviceSpec.Host, serviceSpec.Name, param)
		}
	}

	// Set service groups
	if serviceSpec.Groups != nil {
		if err = client.API.Service().SetServiceGroups(serviceSpec.Host, serviceSpec.Name, serviceSpec.Groups); err != nil {
			return err
		}
		log.Debugf("Set service groups %s on service %s/%s", strings.Join(serviceSpec.Groups, "|"), serviceSpec.Host, serviceSpec.Name)
	}

	// Set categories
	if serviceSpec.Categories != nil {
		if err = client.API.Service().SetCategories(serviceSpec.Host, serviceSpec.Name, serviceSpec.Categories); err != nil {
			return err
		}
		log.Debugf("Set categories %s on service %s/%s", strings.Join(serviceSpec.Categories, "|"), serviceSpec.Host, serviceSpec.Name)
	}

	// Set macros
	if serviceSpec.Macros != nil && len(serviceSpec.Macros) > 0 {
		for key, value := range serviceSpec.Macros {
			macro := &models.Macro{
				Name:       key,
				Value:      value,
				IsPassword: "0",
			}
			if err = client.API.Service().SetMacro(serviceSpec.Host, serviceSpec.Name, macro); err != nil {
				return err
			}
			log.Debugf("Set macro %s on service %s/%s", key, serviceSpec.Host, serviceSpec.Name)
		}
	}

	return nil
}

func updateService(client *centreon.Client, log *logrus.Entry, serviceSpec *v1alpha1.CentreonServiceSpec, service *models.ServiceGet) (err error) {

	// Get extras params
	extras, err := client.API.Service().GetParam(serviceSpec.Host, serviceSpec.Name, []string{"template"})
	if err != nil {
		return err
	}

	// Get service groups
	sgs, err := client.API.Service().GetServiceGroups(serviceSpec.Host, serviceSpec.Name)
	if err != nil {
		return err
	}

	// Get catgeories
	cats, err := client.API.Service().GetCategories(serviceSpec.Host, serviceSpec.Name)
	if err != nil {
		return err
	}

	// Get macros
	macros, err := client.API.Service().GetMacros(serviceSpec.Host, serviceSpec.Name)
	if err != nil {
		return err
	}

	propertiesChange := map[string]string{}
	// Check the main properties
	if boolToString(serviceSpec.Activated) != service.Activated {
		propertiesChange["activate"] = boolToString(serviceSpec.Activated)
	}
	if boolToString(serviceSpec.ActiveCheckEnabled) != service.ActiveCheckEnabled {
		propertiesChange["active_checks_enabled"] = boolToString(serviceSpec.ActiveCheckEnabled)
	}
	if serviceSpec.CheckCommand != service.CheckCommand {
		propertiesChange["check_command"] = serviceSpec.CheckCommand
	}
	if checkArgumentsToString(serviceSpec.Arguments) != service.CheckCommandArgs {
		propertiesChange["check_command_arguments"] = checkArgumentsToString(serviceSpec.Arguments)
	}
	if serviceSpec.MaxCheckAttempts != service.MaxCheckAttempts {
		propertiesChange["max_check_attempts"] = serviceSpec.MaxCheckAttempts
	}
	if serviceSpec.NormalCheckInterval != service.NormalCheckInterval {
		propertiesChange["normal_check_interval"] = serviceSpec.NormalCheckInterval
	}
	if boolToString(serviceSpec.PassiveCheckEnabled) != service.PassiveCheckEnabled {
		propertiesChange["passive_checks_enabled"] = boolToString(serviceSpec.PassiveCheckEnabled)
	}
	if serviceSpec.RetryCheckInterval != service.RetryCheckInterval {
		propertiesChange["retry_check_interval"] = serviceSpec.RetryCheckInterval
	}
	if serviceSpec.Template != extras["template"] {
		propertiesChange["retry_check_interval"] = serviceSpec.Template
	}

	// Check the service groups
	sgNeededTmp, sgOthersTmp := funk.Difference(serviceSpec.Groups, sgs)
	sgNeeded := sgNeededTmp.([]string)
	sgOthers := sgOthersTmp.([]string)

	// Check the categories
	catNeededTmp, catOthersTmp := funk.Difference(serviceSpec.Categories, cats)
	catNeeded := catNeededTmp.([]string)
	catOthers := catOthersTmp.([]string)

	// Check macros
	macrosNeeded := make([]*models.Macro, 0)
	for key, value := range serviceSpec.Macros {
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
	for param, value := range propertiesChange {
		err = client.API.Service().SetParam(serviceSpec.Host, serviceSpec.Name, param, value)
		if err != nil {
			return err
		}
		log.Debugf("Update param %s on service %s/%s", serviceSpec.Host, serviceSpec.Name, param)
	}
	if len(sgNeeded) > 0 {
		err = client.API.Service().SetServiceGroups(serviceSpec.Host, serviceSpec.Name, append(sgNeeded, sgOthers...))
		if err != nil {
			return err
		}
		log.Debugf("Update service groups %s on service %s/%s", serviceSpec.Host, serviceSpec.Name, strings.Join(append(sgNeeded, sgOthers...), "|"))
	}
	if len(catNeeded) > 0 {
		err = client.API.Service().SetCategories(serviceSpec.Host, serviceSpec.Name, append(catNeeded, catOthers...))
		if err != nil {
			return err
		}
		log.Debugf("Update categories %s on service %s/%s", serviceSpec.Host, serviceSpec.Name, strings.Join(append(catNeeded, catOthers...), "|"))
	}
	for _, macro := range macrosNeeded {
		err = client.API.Service().SetMacro(serviceSpec.Host, serviceSpec.Name, macro)
		if err != nil {
			return err
		}
		log.Debugf("Update macro %s on service %s/%s", serviceSpec.Host, serviceSpec.Name, macro.Name)
	}

	return nil
}

func deleteService(client *centreon.Client, log *logrus.Entry, serviceSpec *v1alpha1.CentreonServiceSpec) (err error) {
	return client.API.Service().Delete(serviceSpec.Host, serviceSpec.Name)
}
