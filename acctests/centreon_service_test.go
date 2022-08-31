package acctests

import (
	"context"
	"fmt"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (t *AccTestSuite) TestCentreonService() {

	var (
		cs        *monitorapi.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		err       error
	)

	centreonServiceGVR := monitorapi.GroupVersion.WithResource("centreonservices")

	/***
	 * Create new centreon service resource
	 */
	cs = &monitorapi.CentreonService{
		TypeMeta: v1.TypeMeta{
			Kind:       "CentreonService",
			APIVersion: fmt.Sprintf("%s/%s", monitorapi.GroupVersion.Group, monitorapi.GroupVersion.Version),
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: monitorapi.CentreonServiceSpec{
			Host:         "localhost",
			Name:         "test-ping",
			CheckCommand: "ping",
			Template:     "template-test",
			Activated:    true,
		},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ping",
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

	ucs, err = structuredToUntructured(cs)
	if err != nil {
		t.T().Fatal(err)
	}

	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Create(context.Background(), ucs, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that status is updated
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update Centreon resource
	 */
	time.Sleep(30 * time.Second)
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	cs.Spec.Groups = []string{"sg1"}
	cs.Spec.Categories = []string{"Ping"}
	cs.Spec.Arguments = []string{"arg1"}
	cs.Spec.NormalCheckInterval = "60"
	cs.Spec.RetryCheckInterval = "10"
	cs.Spec.MaxCheckAttempts = "2"
	cs.Spec.Macros = map[string]string{"MAC1": "value"}
	ucs, err = structuredToUntructured(cs)
	if err != nil {
		t.T().Fatal(err)
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ping",
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
	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Update(context.Background(), ucs, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete service
	 */
	time.Sleep(20 * time.Second)
	err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Delete(context.Background(), "test", v1.DeleteOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.Nil(t.T(), s)
}
