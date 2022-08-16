package acctests

import (
	"context"
	"fmt"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	api "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (t *AccTestSuite) TestNode() {

	var (
		cs        *api.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		node      *core.Node
		err       error
	)

	centreonServiceGVR := api.GroupVersion.WithResource("centreonservices")
	templateCentreonServiceGVR := api.GroupVersion.WithResource("templates")

	/***
	 * Create new template dedicated for node test
	 */
	tcs := &api.Template{
		TypeMeta: v1.TypeMeta{
			Kind:       "Template",
			APIVersion: fmt.Sprintf("%s/%s", api.GroupVersion.Group, api.GroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "check-node",
		},
		Spec: api.TemplateSpec{
			Type: "CentreonService",
			Template: `
host: "localhost"
name: "test-node-ping"
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
	if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Create(context.Background(), tcsu, v1.CreateOptions{}); err != nil {
		t.T().Fatal(err)
	}

	/***
	 * Create new node
	 */
	node = &core.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-node"}]`,
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-node-ping",
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
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.CoreV1().Nodes().Create(context.Background(), node, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &api.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-node", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-node-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-node-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update node
	 */
	time.Sleep(30 * time.Second)
	node, err = t.k8sclientStd.CoreV1().Nodes().Get(context.Background(), "test-node", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	node.Labels = map[string]string{"foo": "bar2"}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-node-ping",
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
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.CoreV1().Nodes().Update(context.Background(), node, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-node", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "bar2", cs.Spec.Macros["LABEL"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-node-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update node template
	 */
	time.Sleep(30 * time.Second)
	tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Get(context.Background(), "check-node", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(tcsu, tcs); err != nil {
		t.T().Fatal(err)
	}
	tcs.Spec.Template = `
host: "localhost"
name: "test-node-ping"
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
	if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Update(context.Background(), tcsu, v1.UpdateOptions{}); err != nil {
		t.T().Fatal(err)
	}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-node-ping",
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
				Name:   "TEST",
				Value:  "plop",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-node", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "plop", cs.Spec.Macros["TEST"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-node-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete node
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclientStd.CoreV1().Nodes().Delete(context.Background(), "test-node", *metav1.NewDeleteOptions(0)); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService delete on k8s
	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-node", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not delete on k8s after delete node")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-node-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
