package template

import (
	"fmt"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/thoas/go-funk"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// getLabels permit to return global label must be set on all resources
func getLabels(o client.Object, customLabels ...map[string]string) (labels map[string]string) {
	labels = map[string]string{
		centreoncrd.MonitoringAnnotationKey:                           "true",
		fmt.Sprintf("%s/parent", centreoncrd.MonitoringAnnotationKey): fmt.Sprintf("%s/%s", o.GetNamespace(), o.GetName()),
	}
	for _, label := range customLabels {
		for key, val := range label {
			labels[key] = val
		}
	}

	labels = funk.UnionStringMap(labels, o.GetLabels())

	return labels
}

// getAnnotations permit to return global annotations must be set on all resources
func getAnnotations(o client.Object, customAnnotation ...map[string]string) (annotations map[string]string) {
	annotations = map[string]string{
		centreoncrd.MonitoringAnnotationKey: "true",
	}
	for _, annotation := range customAnnotation {
		for key, val := range annotation {
			annotations[key] = val
		}
	}

	return annotations
}
