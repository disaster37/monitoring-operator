package acctests

import (
	"context"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (t *AccTestSuite) TestCentreonService() {

	var (
		cs        *v1alpha1.CentreonService
		ucs       *unstructured.Unstructured
		ucsTmp    map[string]interface{}
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		err       error
	)

	centreonServiceGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreonservices",
	}

	/***
	 * Create new centreon service resource
	 */
	cs = &v1alpha1.CentreonService{
		TypeMeta: v1.TypeMeta{
			Kind:       "CentreonService",
			APIVersion: "monitor.k8s.webcenter.fr/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.CentreonServiceSpec{
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

	ucsTmp, err = runtime.DefaultUnstructuredConverter.ToUnstructured(cs)
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	ucs = &unstructured.Unstructured{
		Object: ucsTmp,
	}

	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Create(context.Background(), ucs, v1.CreateOptions{})
	assert.NoError(t.T(), err)
	time.Sleep(20 * time.Second)

	// Check that status is updated
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucs.Object, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.NotEmpty(t.T(), cs.Status.ID)
	assert.NotEmpty(t.T(), cs.Status.CreatedAt)

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	assert.NoError(t.T(), err)
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update Centreon resource
	 */
	time.Sleep(30 * time.Second)
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucs.Object, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	cs.Spec.Groups = []string{"sg1"}
	cs.Spec.Categories = []string{"Ping"}
	cs.Spec.Arguments = []string{"arg1"}
	cs.Spec.NormalCheckInterval = "60"
	cs.Spec.RetryCheckInterval = "10"
	cs.Spec.MaxCheckAttempts = "2"
	cs.Spec.Macros = map[string]string{"MAC1": "value"}
	ucsTmp, err = runtime.DefaultUnstructuredConverter.ToUnstructured(cs)
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	ucs = &unstructured.Unstructured{
		Object: ucsTmp,
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
	assert.NoError(t.T(), err)
	time.Sleep(20 * time.Second)

	// Check that status is updated
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(ucs.Object, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.NotEmpty(t.T(), cs.Status.UpdatedAt)

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	assert.NoError(t.T(), err)
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete service
	 */
	time.Sleep(20 * time.Second)
	err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Delete(context.Background(), "test", *&v1.DeleteOptions{})
	assert.NoError(t.T(), err)
	time.Sleep(20 * time.Second)

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
