package acctests

import (
	"context"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	api "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	core "k8s.io/api/core/v1"
)

func (t *AccTestSuite) TestNamespace() {

	var (
		cs        *api.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		namespace   *core.Namespace
		err       error
	)

	centreonServiceGVR := api.GroupVersion.WithResource("centreonServices")
	templateCentreonServiceGVR := api.GroupVersion.WithResource("templateCentreonServices")

	/***
	 * Create new template dedicated for namespace test
	 */
	tcs := &api.TemplateCentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "check-namespace",
			Namespace: "default",
		},
		Spec: api.TemplateCentreonServiceSpec{
			Template: `
host: "localhost"
name: "test-namespace-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
activate: true`,
		},
	}
	tcsu, err := structuredToUntructured(tcs)
	if err != nil {
		t.T().Fatal(err)
	}
	if tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Create(context.Background(), tcsu, v1.CreateOptions{}); err != nil {
		t.T().Fatal(err)
	}


	/***
	 * Create new namespace
	 */
	namespace = &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-namespace"}]`,
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-namespace-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros:              []*models.Macro{
			{
				Name: "LABEL",
				Value: "bar",
			},
		},
		Activated:           "1",
	}
	namespace, err = t.k8sclientStd.CoreV1().Namespaces().Create(context.Background(), namespace, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &api.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("test-namespace").Get(context.Background(), "check-namespace", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-namespace-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-namespace-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update namespace
	 */
	time.Sleep(30 * time.Second)
	namespace, err = t.k8sclientStd.CoreV1().Namespaces().Get(context.Background(), "test-namespace", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	namespace.Labels = map[string]string{"foo": "bar2"}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-namespace-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros:              []*models.Macro{
			{
				Name: "LABEL",
				Value: "bar2",
			},
		},
		Activated:           "1",
	}
	_, err = t.k8sclientStd.CoreV1().Namespaces().Update(context.Background(), namespace, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("test-namespace").Get(context.Background(), "check-namespace", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "bar2", cs.Spec.Macros["LABEL"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-namespace-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update namespace template
	 */
	 time.Sleep(30 * time.Second)
	 tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Get(context.Background(), "check-namespace", v1.GetOptions{})
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 if err = unstructuredToStructured(tcsu, tcs); err != nil {
		t.T().Fatal(err)
	 }
	 tcs.Spec.Template = `
host: "localhost"
name: "test-namespace-ping"
template: "template-test"
checkCommand: "ping"
macros:
	LABEL: "{{ .labels.foo }}"
	TEST: "plop"
activate: true`
	 tcsu, err = structuredToUntructured(tcs)
	 if err != nil {
		t.T().Fatal(err)
	 }
	 if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Update(context.Background(), tcsu, v1.UpdateOptions{}); err != nil {
		t.T().Fatal(err)
	 }
 
	 expectedS = &centreonhandler.CentreonService{
		 Host:                "localhost",
		 Name:                "test-namespace-ping",
		 CheckCommand:        "ping",
		 Template:            "template-test",
		 PassiveCheckEnabled: "2",
		 ActiveCheckEnabled:  "2",
		 Comment:             "Managed by monitoring-operator",
		 Groups:              []string{},
		 Categories:          []string{},
		 Macros:              []*models.Macro{
			 {
				 Name: "LABEL",
				 Value: "bar2",
			 },
			 {
				 Name: "TEST",
				 Value: "plop",
			 },
		 },
		 Activated:           "1",
	 }
	 time.Sleep(20 * time.Second)
 
	 ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("test-namespace").Get(context.Background(), "check-namespace", v1.GetOptions{})
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 if err = unstructuredToStructured(ucs, cs); err != nil {
		 t.T().Fatal(err)
	 }
	 assert.Equal(t.T(), "plop", cs.Spec.Macros["TEST"])
 
	 // Check service updated on Centreon
	 s, err = t.centreon.GetService("localhost", "test-namespace-ping")
	 if err != nil {
		 t.T().Fatal(err)
	 }
	 assert.NotNil(t.T(), s)
	 assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete namespace
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclientStd.CoreV1().Namespaces().Delete(context.Background(), "test-namespace", *metav1.NewDeleteOptions(0)); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService delete on k8s
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("test-namespace").Get(context.Background(), "template1", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not delete on k8s after delete namespace")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-namespace-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
