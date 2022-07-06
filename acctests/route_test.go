package acctests

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	api "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func (t *AccTestSuite) TestRoute() {

	var (
		cs        *api.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		route     *routev1.Route
		uRoute    *unstructured.Unstructured
		err       error
	)

	isRouteCRD, err := controllers.IsRouteCRD(t.config)
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if !isRouteCRD {
		t.T().Skip("Not Openshift cluster, skit it")
	}

	centreonServiceGVR := api.GroupVersion.WithResource("centreonservices")
	templateCentreonServiceGVR := api.GroupVersion.WithResource("templatecentreonservices")

	routeGVR := schema.GroupVersionResource{
		Group:    "route.openshift.io",
		Version:  "v1",
		Resource: "routes",
	}

	/***
	 * Create new template dedicated for route test
	 */
	 tcs := &api.TemplateCentreonService{
		TypeMeta: v1.TypeMeta{
			Kind:       "TemplateCentreonService",
			APIVersion: fmt.Sprintf("%s/%s", api.GroupVersion.Group, api.GroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "check-route",
		},
		Spec: api.TemplateCentreonServiceSpec{
			Template: `
{{ $rule := index .rules 0}}
{{ $path := index $rule.paths 0}}
host: "localhost"
name: "test-route-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
  SCHEME: "{{ $rule.scheme }}"
  HOST: "{{ $rule.host }}"
  PATH: "{{ $path }}"
activate: true`,
		},
	}
	tcsu, err := structuredToUntructured(tcs)
	if err != nil {
		t.T().Fatal(err)
	}
	if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Create(context.Background(), tcsu, v1.CreateOptions{}); err != nil {
		t.T().Fatal(err)
	}

	/***
	 * Create new route
	 */
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-route",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-route"}]`,
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "test",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("8080"),
			},
		},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-route-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros: []*models.Macro{
			{
				Name:   "LABEL",
				Value:  "bar",
				Source: "direct",
			},
			{
				Name:   "SCHEME",
				Value:  "http",
				Source: "direct",
			},
			{
				Name:   "HOST",
				Value:  "front.local.local",
				Source: "direct",
			},
			{
				Name:   "PATH",
				Value:  "/",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	uRoute, err = structuredToUntructured(route)
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	_, err = t.k8sclient.Resource(routeGVR).Namespace("default").Create(context.Background(), uRoute, v1.CreateOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &api.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-route", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-route-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-route-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)

	// Sort macro to fix test
	sort.Slice(expectedS.Macros, func(i, j int) bool {
		return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	})
	sort.Slice(s.Macros, func(i, j int) bool {
		return s.Macros[i].Name < s.Macros[j].Name
	})
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update Route
	 */
	time.Sleep(30 * time.Second)
	uRoute, err = t.k8sclient.Resource(routeGVR).Namespace("default").Get(context.Background(), "test-route", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(uRoute, route); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	route.Labels = map[string]string{"foo": "bar2"}
	uRoute, err = structuredToUntructured(route)
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-route-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros: []*models.Macro{
			{
				Name:   "LABEL",
				Value:  "bar2",
				Source: "direct",
			},
			{
				Name:   "SCHEME",
				Value:  "http",
				Source: "direct",
			},
			{
				Name:   "HOST",
				Value:  "front.local.local",
				Source: "direct",
			},
			{
				Name:   "PATH",
				Value:  "/",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	_, err = t.k8sclient.Resource(routeGVR).Namespace("default").Update(context.Background(), uRoute, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-route", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "bar2", cs.Spec.Macros["LABEL"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-route-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)

	// Sort macro to fix test
	sort.Slice(expectedS.Macros, func(i, j int) bool {
		return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	})
	sort.Slice(s.Macros, func(i, j int) bool {
		return s.Macros[i].Name < s.Macros[j].Name
	})
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update route template
	 */
	 time.Sleep(30 * time.Second)
	 tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Get(context.Background(), "check-route", v1.GetOptions{})
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 if err = unstructuredToStructured(tcsu, tcs); err != nil {
		 t.T().Fatal(err)
	 }
	tcs.Spec.Template = `
{{ $rule := index .rules 0}}
{{ $path := index $rule.paths 0}}
host: "localhost"
name: "test-route-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
  SCHEME: "{{ $rule.scheme }}"
  HOST: "{{ $rule.host }}"
  PATH: "{{ $path }}"
  TEST: "plop"
activate: true`
 
	 tcsu, err = structuredToUntructured(tcs)
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Update(context.Background(), tcsu, v1.UpdateOptions{}); err != nil {
		 t.T().Fatal(err)
	 }
 
	 expectedS = &centreonhandler.CentreonService{
		 Host:                "localhost",
		 Name:                "test-route-ping",
		 CheckCommand:        "ping",
		 Template:            "template-test",
		 PassiveCheckEnabled: "2",
		 ActiveCheckEnabled:  "2",
		 Comment:             "Managed by monitoring-operator",
		 Groups:              []string{},
		 Categories:          []string{},
		 Macros: []*models.Macro{
			 {
				 Name:   "LABEL",
				 Value:  "bar2",
				 Source: "direct",
			 },
			 {
				 Name:   "SCHEME",
				 Value:  "http",
				 Source: "direct",
			 },
			 {
				 Name:   "HOST",
				 Value:  "front.local.local",
				 Source: "direct",
			 },
			 {
				 Name:   "PATH",
				 Value:  "/",
				 Source: "direct",
			 },
			 {
				 Name:   "TEST",
				 Value:  "plop",
				 Source: "direct",
			 },
		 },
		 Activated: "1",
	 }
	 time.Sleep(20 * time.Second)
 
	 ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-route", v1.GetOptions{})
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 if err = unstructuredToStructured(ucs, cs); err != nil {
		 t.T().Fatal(err)
	 }
	 assert.Equal(t.T(), "plop", cs.Spec.Macros["TEST"])
 
	 // Check service updated on Centreon
	 s, err = t.centreon.GetService("localhost", "test-route-ping")
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 assert.NotNil(t.T(), s)
	 // Sort macro to fix test
	 sort.Slice(expectedS.Macros, func(i, j int) bool {
		 return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	 })
	 sort.Slice(s.Macros, func(i, j int) bool {
		 return s.Macros[i].Name < s.Macros[j].Name
	 })
	 assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete route
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclient.Resource(routeGVR).Namespace("default").Delete(context.Background(), "check-route", *metav1.NewDeleteOptions(0)); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService deleted
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-route", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not deleted after delete route")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-route-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
