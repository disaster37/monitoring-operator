package controllers

/*
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
func initCentreonServiceDefaultValue(endpointSpec *v1alpha1.CentreonSpecEndpoint, cs *v1alpha1.CentreonService, placesholders map[string]string) {
	if endpointSpec == nil || cs == nil {
		return
	}

	cs.Spec.Activated = endpointSpec.ActivateService
	cs.Spec.Categories = endpointSpec.Categories
	cs.Spec.Groups = endpointSpec.ServiceGroups
	cs.Spec.Host = endpointSpec.DefaultHost
	cs.Spec.Template = endpointSpec.Template

	// Need placeholders
	if endpointSpec.Arguments != nil && len(endpointSpec.Arguments) > 0 {
		arguments := make([]string, len(endpointSpec.Arguments))
		for i, arg := range endpointSpec.Arguments {
			arguments[i] = helpers.PlaceholdersInString(arg, placesholders)
		}
		cs.Spec.Arguments = arguments
	}

	if endpointSpec.Macros != nil && len(endpointSpec.Macros) > 0 {
		macros := map[string]string{}
		for key, value := range endpointSpec.Macros {
			macros[key] = helpers.PlaceholdersInString(value, placesholders)
		}
		cs.Spec.Macros = macros
	}

	cs.Spec.Name = helpers.PlaceholdersInString(endpointSpec.NameTemplate, placesholders)
}
*/
