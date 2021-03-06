package controllers

import (
	"fmt"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	monitoringAnnotationKey         = "monitor.k8s.webcenter.fr"
	centreonMonitoringAnnotationKey = "centreon.monitor.k8s.webcenter.fr"
	waitDurationWhenError           = 1 * time.Minute
)

type Reconciler struct {
	recorder   record.EventRecorder
	log        *logrus.Entry
	reconciler controller.Reconciler
	platforms  map[string]*ComputedPlatform
}

func (r *Reconciler) SetLogger(log *logrus.Entry) {
	r.log = log
}

func (r *Reconciler) SetRecorder(recorder record.EventRecorder) {
	r.recorder = recorder
}

func (r *Reconciler) SetReconsiler(reconciler controller.Reconciler) {
	r.reconciler = reconciler
}

// SetPlatforms permit to set the platform list
func (r *Reconciler) SetPlatforms(p map[string]*ComputedPlatform) {
	r.platforms = p
}

// Handle only resources that have the monitoring annotation or TemplateCentreonService type
func viewResourceWithMonitoringTemplate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isMonitoringTemplateAnnotation(e.ObjectOld.GetAnnotations()) || isMonitoringTemplateAnnotation(e.ObjectNew.GetAnnotations()) || isTemplateCentreonService(e.ObjectNew)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isMonitoringTemplateAnnotation(e.Object.GetAnnotations()) || isTemplateCentreonService(e.Object)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return isMonitoringTemplateAnnotation(e.Object.GetAnnotations()) || isTemplateCentreonService(e.Object)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isMonitoringAnnotation(e.Object.GetAnnotations()) || isTemplateCentreonService(e.Object)
		},
	}
}

// Return true if monitoring annotation is present
func isMonitoringTemplateAnnotation(annotations map[string]string) bool {
	if annotations == nil {
		return false
	}
	watchKey := fmt.Sprintf("%s/templates", monitoringAnnotationKey)
	for key, value := range annotations {
		if key == watchKey && value != "" {
			return true
		}
	}
	return false
}

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

func getClient(platformRef string, platforms map[string]*ComputedPlatform) (meta any, platform *v1alpha1.Platform, err error) {
	if platformRef == "" {
		if p, ok := platforms["default"]; ok {
			return p.client, p.platform, nil
		}

		return nil, nil, errors.New("No default platform")
	}

	if p, ok := platforms[platformRef]; ok {
		return p.client, p.platform, nil
	}

	return nil, nil, errors.Errorf("Platform %s not found", platformRef)
}
