package controllers

import (
	"fmt"
	"time"

	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	monitoringAnnotationKey         = "monitor.k8s.webcenter.fr"
	centreonMonitoringAnnotationKey = "centreon.monitor.k8s.webcenter.fr"
)

var (
	waitDurationWhenError time.Duration = 1 * time.Minute
)

// Handle only resources that have the monitoring annotation
func viewResourceWithMonitoringAnnotationPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			return isMonitoringAnnotation(e.ObjectOld.GetAnnotations()) || isMonitoringAnnotation(e.ObjectNew.GetAnnotations())
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations())
		},
	}
}
func isMonitoringAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	watchKey := fmt.Sprintf("%s/discover", monitoringAnnotationKey)
	for key, value := range annotations {
		if key == watchKey && value == "true" {
			return true
		}
	}
	return false
}

// Only spec update and finalizer predicate
func centreonServicePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.ObjectOld.GetGeneration() != e.ObjectNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return !e.DeleteStateUnknown
		},
	}
}

// IsRouteCRD check if apiGroup called "route.openshift.io" exist on cluster.
// It usefull to start controller that manage this ressource only if exist on cluster
func IsRouteCRD(cfg *rest.Config) (bool, error) {
	// Create the discoveryClient
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}
	apiGroups, _, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return false, err
	}

	for _, apiGroup := range apiGroups {
		if apiGroup.Name == "route.openshift.io" {
			return true, nil
		}
	}

	return false, nil
}
