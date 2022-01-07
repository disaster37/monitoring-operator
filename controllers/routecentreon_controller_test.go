package controllers

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestGeneratePlaceholdersRouteCentreonService(t *testing.T) {

	var (
		route      *routev1.Route
		ph         map[string]string
		expectedPh map[string]string
	)

	// When route is nil
	ph = generatePlaceholdersRouteCentreonService(nil)
	assert.Empty(t, ph)

	// When all no path
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
		},
	}

	expectedPh = map[string]string{
		"name":             "test",
		"namespace":        "default",
		"rule.host":        "front.local.local",
		"rule.scheme":      "http",
		"rule.path":        "/",
		"label.app":        "appTest",
		"label.env":        "dev",
		"annotation.anno1": "value1",
		"annotation.anno2": "value2",
	}

	// When all properties without tls
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
		},
	}

	expectedPh = map[string]string{
		"name":             "test",
		"namespace":        "default",
		"rule.host":        "front.local.local",
		"rule.scheme":      "http",
		"rule.path":        "/",
		"label.app":        "appTest",
		"label.env":        "dev",
		"annotation.anno1": "value1",
		"annotation.anno2": "value2",
	}

	ph = generatePlaceholdersRouteCentreonService(route)
	assert.Equal(t, expectedPh, ph)

	// When all properties with tls
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
			},
		},
	}

	expectedPh = map[string]string{
		"name":             "test",
		"namespace":        "default",
		"rule.host":        "front.local.local",
		"rule.scheme":      "https",
		"rule.path":        "/",
		"label.app":        "appTest",
		"label.env":        "dev",
		"annotation.anno1": "value1",
		"annotation.anno2": "value2",
	}

	ph = generatePlaceholdersRouteCentreonService(route)
	assert.Equal(t, expectedPh, ph)

}

func TestCentreonServiceFromRoute(t *testing.T) {

	var (
		route        *routev1.Route
		cs           *v1alpha1.CentreonService
		expectedCs   *v1alpha1.CentreonService
		centreonSpec *v1alpha1.CentreonSpec
		err          error
	)

	// When ingress is nil
	_, err = centreonServiceFromIngress(nil, nil, nil)
	assert.Error(t, err)

	// When no centreonSpec and not all annotations
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
		},
	}
	_, err = centreonServiceFromRoute(route, nil, runtime.NewScheme())
	assert.Error(t, err)

	// When no centreonSpec and all annotations
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromRoute(route, nil, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec and all annotations, priority to annotations
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
		},
	}
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
			Template:        "template",
			NameTemplate:    "name",
			DefaultHost:     "localhost",
			ActivateService: true,
			Arguments:       []string{"arg1"},
			ServiceGroups:   []string{"sg1"},
			Categories:      []string{"cat1"},
			Macros: map[string]string{
				"MACRO1": "value1",
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromRoute(route, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec without annotations
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
		},
	}
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
			Template:        "template",
			NameTemplate:    "name-<label.app>-<label.env>-<namespace>",
			DefaultHost:     "localhost",
			ActivateService: true,
			ServiceGroups:   []string{"sg1"},
			Macros: map[string]string{
				"SCHEME": "<rule.scheme>",
				"URL":    "<rule.host><rule.path>",
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "name-appTest-dev-default",
			Template: "template",
			Macros: map[string]string{
				"SCHEME": "http",
				"URL":    "front.local.local/",
			},
			Activated: true,
			Groups:    []string{"sg1"},
		},
	}
	cs, err = centreonServiceFromRoute(route, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

}

func (t *ControllerTestSuite) TestRouteCentreonControllerWhenNoCentreonSpec() {

	var (
		err                     error
		fetched                 *routev1.Route
		cs                      *v1alpha1.CentreonService
		expectedCentreonService *v1alpha1.CentreonService
		isTimeout               bool
	)
	routeName := "t-route-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      routeName,
		Namespace: "default",
	}

	//Create new route
	toCreate := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover":                       "true",
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "fake",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("8080"),
			},
		},
	}
	expectedCentreonService = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover":                       "true",
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}
	isTimeout, err = RunWithTimeout(func() error {
		cs = &v1alpha1.CentreonService{}
		if err := t.k8sClient.Get(context.Background(), key, cs); err != nil {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isTimeout)
	assert.Equal(t.T(), expectedCentreonService.Spec, cs.Spec)
	assert.Equal(t.T(), expectedCentreonService.GetLabels(), cs.GetLabels())
	assert.Equal(t.T(), expectedCentreonService.GetAnnotations(), cs.GetAnnotations())
	time.Sleep(10 * time.Second)

	// Update route
	fetched = &routev1.Route{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}

	fetched.ObjectMeta.Annotations["centreon.monitor.k8s.webcenter.fr/max-check-attempts"] = "6"

	expectedCentreonService = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover":                       "true",
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "6",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "6",
		},
	}
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(30 * time.Second)
	cs = &v1alpha1.CentreonService{}
	if err := t.k8sClient.Get(context.Background(), key, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), expectedCentreonService.Spec, cs.Spec)
	assert.Equal(t.T(), expectedCentreonService.GetLabels(), cs.GetLabels())
	assert.Equal(t.T(), expectedCentreonService.GetAnnotations(), cs.GetAnnotations())
	time.Sleep(10 * time.Second)

}

func (t *ControllerTestSuite) TestRouteCentreonControllerWhenCentreonSpec() {

	var (
		err                     error
		fetched                 *routev1.Route
		cs                      *v1alpha1.CentreonService
		expectedCentreonService *v1alpha1.CentreonService
		isTimeout               bool
		centreon                *v1alpha1.Centreon
	)
	routeName := "t-route-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      routeName,
		Namespace: "default",
	}
	keyCentreon := types.NamespacedName{
		Name:      centreonResourceName,
		Namespace: "default",
	}

	os.Setenv("OPERATOR_NAMESPACE", "default")

	//Create new route
	centreon = &v1alpha1.Centreon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      keyCentreon.Name,
			Namespace: keyCentreon.Namespace,
		},
		Spec: v1alpha1.CentreonSpec{
			Endpoints: &v1alpha1.CentreonSpecEndpoint{
				Template:     "template",
				DefaultHost:  "localhost",
				NameTemplate: "ping",
				Macros: map[string]string{
					"mac1": "value1",
					"mac2": "value2",
				},
				Arguments:       []string{"arg1", "arg2"},
				ActivateService: true,
				ServiceGroups:   []string{"sg1"},
				Categories:      []string{"cat1"},
			},
		},
	}
	if err = t.k8sClient.Create(context.Background(), centreon); err != nil {
		t.T().Fatal(err)
	}

	//Create new route
	toCreate := &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover": "true",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "fake",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("8080"),
			},
		},
	}
	expectedCentreonService = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover": "true",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:  []string{"arg1", "arg2"},
			Activated:  true,
			Groups:     []string{"sg1"},
			Categories: []string{"cat1"},
		},
	}
	err = t.k8sClient.Create(context.Background(), toCreate)
	assert.NoError(t.T(), err)
	isTimeout, err = RunWithTimeout(func() error {
		cs = &v1alpha1.CentreonService{}
		if err := t.k8sClient.Get(context.Background(), key, cs); err != nil {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isTimeout)
	assert.Equal(t.T(), expectedCentreonService.Spec, cs.Spec)
	assert.Equal(t.T(), expectedCentreonService.GetLabels(), cs.GetLabels())
	assert.Equal(t.T(), expectedCentreonService.GetAnnotations(), cs.GetAnnotations())
	time.Sleep(10 * time.Second)

	// Update route
	if err = t.k8sClient.Get(context.Background(), keyCentreon, centreon); err != nil {
		t.T().Fatal(err)
	}
	centreon.Spec.Endpoints.Arguments = []string{"arg1"}
	if err = t.k8sClient.Update(context.Background(), centreon); err != nil {
		t.T().Fatal(err)
	}

	fetched = &routev1.Route{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.ObjectMeta.Annotations["foo"] = "bar"

	expectedCentreonService = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover": "true",
				"foo":                               "bar",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:  []string{"arg1"},
			Activated:  true,
			Groups:     []string{"sg1"},
			Categories: []string{"cat1"},
		},
	}
	err = t.k8sClient.Update(context.Background(), fetched)
	assert.NoError(t.T(), err)
	time.Sleep(30 * time.Second)
	cs = &v1alpha1.CentreonService{}
	if err := t.k8sClient.Get(context.Background(), key, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), expectedCentreonService.Spec, cs.Spec)
	assert.Equal(t.T(), expectedCentreonService.GetLabels(), cs.GetLabels())
	assert.Equal(t.T(), expectedCentreonService.GetAnnotations(), cs.GetAnnotations())
	time.Sleep(10 * time.Second)

	// Clean centreon CR
	if err = t.k8sClient.Get(context.Background(), keyCentreon, centreon); err != nil {
		t.T().Fatal(err)
	}
	if err = t.k8sClient.Delete(context.Background(), centreon); err != nil {
		t.T().Fatal(err)
	}

}
