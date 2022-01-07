package controllers

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
)

// It initiate CentreonService spec from annotations
func initCentreonServiceFromAnnotations(annotations map[string]string, cs *v1alpha1.CentreonService) (err error) {
	if annotations == nil || cs == nil {
		return
	}

	// Init values from annotations
	re := regexp.MustCompile(fmt.Sprintf("^%s/(.+)$", centreonMonitoringAnnotationKey))
	for key, value := range annotations {
		if match := re.FindStringSubmatch(key); len(match) > 0 {
			switch match[1] {
			case "name":
				cs.Spec.Name = value
				break
			case "host":
				cs.Spec.Host = value
				break
			case "template":
				cs.Spec.Template = value
				break
			case "activated":
				t := helpers.StringToBool(value)
				if t != nil {
					cs.Spec.Activated = *t
				}
				break
			case "normal-check-interval":
				cs.Spec.NormalCheckInterval = value
				break
			case "retry-check-interval":
				cs.Spec.RetryCheckInterval = value
				break
			case "max-check-attempts":
				cs.Spec.MaxCheckAttempts = value
				break
			case "active-check-enabled":
				cs.Spec.ActiveCheckEnabled = helpers.StringToBool(value)
				break
			case "passive-check-enabled":
				cs.Spec.PassiveCheckEnabled = helpers.StringToBool(value)
				break
			case "check-command":
				cs.Spec.CheckCommand = value
				break
			case "arguments":
				cs.Spec.Arguments = helpers.StringToSlice(value, ",")
				break
			case "groups":
				cs.Spec.Groups = helpers.StringToSlice(value, ",")
				break
			case "categories":
				cs.Spec.Categories = helpers.StringToSlice(value, ",")
				break
			case "macros":
				t := map[string]string{}
				if err = json.Unmarshal([]byte(value), &t); err != nil {
					return err
				}
				cs.Spec.Macros = t
				break
			}
		}
	}

	return nil

}

// It initiate CentreonService spec with default value provided by Centreon spec and some placeholders
func initCentreonServiceDefaultValue(centreon *v1alpha1.CentreonSpec, cs *v1alpha1.CentreonService, placesholders map[string]string) {
	if centreon == nil || centreon.Endpoints == nil || cs == nil {
		return
	}

	cs.Spec.Activated = centreon.Endpoints.ActivateService
	cs.Spec.Categories = centreon.Endpoints.Categories
	cs.Spec.Groups = centreon.Endpoints.ServiceGroups
	cs.Spec.Host = centreon.Endpoints.DefaultHost
	cs.Spec.Template = centreon.Endpoints.Template

	// Need placeholders
	if centreon.Endpoints.Arguments != nil && len(centreon.Endpoints.Arguments) > 0 {
		arguments := make([]string, len(centreon.Endpoints.Arguments))
		for i, arg := range centreon.Endpoints.Arguments {
			arguments[i] = helpers.PlaceholdersInString(arg, placesholders)
		}
		cs.Spec.Arguments = arguments
	}

	if centreon.Endpoints.Macros != nil && len(centreon.Endpoints.Macros) > 0 {
		macros := map[string]string{}
		for key, value := range centreon.Endpoints.Macros {
			macros[key] = helpers.PlaceholdersInString(value, placesholders)
		}
		cs.Spec.Macros = macros
	}

	cs.Spec.Name = helpers.PlaceholdersInString(centreon.Endpoints.NameTemplate, placesholders)
}
