package acctests

import (
	"context"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
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
		cs        *v1alpha1.CentreonService
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

	centreonServiceGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreonservices",
	}

	routeGVR := schema.GroupVersionResource{
		Group:    "route.openshift.io",
		Version:  "v1",
		Resource: "routes",
	}

	/***
	 * Create new route
	 */
	route = &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-route",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover":               "true",
				"centreon.monitor.k8s.webcenter.fr/name":          "test-route-ping",
				"centreon.monitor.k8s.webcenter.fr/template":      "template-test",
				"centreon.monitor.k8s.webcenter.fr/host":          "localhost",
				"centreon.monitor.k8s.webcenter.fr/activated":     "1",
				"centreon.monitor.k8s.webcenter.fr/check-command": "ping",
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
		Macros:              []*models.Macro{},
		Activated:           "1",
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
	cs = &v1alpha1.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-route", v1.GetOptions{})
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
	route.Annotations["centreon.monitor.k8s.webcenter.fr/groups"] = "sg1"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/categories"] = "Ping"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/arguments"] = "arg1"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/normal-check-interval"] = "60"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/retry-check-interval"] = "10"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/max-check-attempts"] = "2"
	route.Annotations["centreon.monitor.k8s.webcenter.fr/macros"] = `{"MAC1": "value"}`
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
		Groups:              []string{"sg1"},
		Categories:          []string{"Ping"},
		Macros: []*models.Macro{
			{
				Name:   "MAC1",
				Value:  "value",
				Source: "direct",
			},
		},
		Activated:           "1",
		NormalCheckInterval: "60",
		RetryCheckInterval:  "10",
		MaxCheckAttempts:    "2",
		CheckCommandArgs:    "!arg1",
	}
	_, err = t.k8sclient.Resource(routeGVR).Namespace("default").Update(context.Background(), uRoute, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-route", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-route-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete route
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclient.Resource(routeGVR).Namespace("default").Delete(context.Background(), "test-route", *metav1.NewDeleteOptions(0)); err != nil {
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
