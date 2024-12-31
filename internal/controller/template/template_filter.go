package template

import (
	"fmt"
	"reflect"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// Handle only resources that have the monitoring annotation or TemplateCentreonService type
func ViewResourceWithMonitoringTemplate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isMonitoringTemplateAnnotation(e.ObjectOld) || isMonitoringTemplateAnnotation(e.ObjectNew) || isTemplate(e.ObjectNew) || isGeneratedFromTemplate(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isMonitoringTemplateAnnotation(e.Object) || isTemplate(e.Object) || isGeneratedFromTemplate(e.Object)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isMonitoringTemplateAnnotation(e.Object) || isTemplate(e.Object) || isGeneratedFromTemplate(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isMonitoringTemplateAnnotation(e.Object) || isTemplate(e.Object) || isGeneratedFromTemplate(e.Object)
		},
	}
}

// Return true if monitoring annotation is present
func isMonitoringTemplateAnnotation(o client.Object) bool {
	if o.GetAnnotations() == nil {
		return false
	}
	watchKey := fmt.Sprintf("%s/templates", centreoncrd.MonitoringAnnotationKey)
	for key, value := range o.GetAnnotations() {
		if key == watchKey && value != "" {
			return true
		}
	}
	return false
}

// Return true if object type is Template
func isTemplate(o client.Object) bool {
	return reflect.TypeOf(o).Elem().Name() == "Template"
}

func isGeneratedFromTemplate(o client.Object) bool {
	if o.GetLabels() == nil {
		return false
	}
	watchKey := fmt.Sprintf("%s/template", centreoncrd.MonitoringAnnotationKey)
	for key, value := range o.GetLabels() {
		if key == watchKey && value != "" {
			return true
		}
	}
	return false
}
