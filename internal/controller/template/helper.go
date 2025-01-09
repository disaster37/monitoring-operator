package template

import (
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/thoas/go-funk"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getLabels permit to return global label must be set on all resources
func getLabels(o client.Object, customLabels ...map[string]string) (labels map[string]string) {
	labels = map[string]string{
		centreoncrd.MonitoringAnnotationKey: "true",
	}
	for _, label := range customLabels {
		for key, val := range label {
			labels[key] = val
		}
	}

	labels = funk.UnionStringMap(labels, o.GetLabels())

	return labels
}
