package acctests

import (
	"context"
	"time"

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

	// Create new centreon service resource
	centreonServiceGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreonservices",
	}
	/*
		centreonServiceCR := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "monitor.k8s.webcenter.fr/v1alpha1",
				"kind":       "CentreonService",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"host":         "localhost",
					"name":         "test-ping",
					"checkCommand": "ping",
					"activate":     true,
				},
			},
		}
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
			Activated:    true,
		},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:         "localhost",
		Name:         "test-ping",
		CheckCommand: "ping",
		Activated:    "1",
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

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-ping")
	assert.NoError(t.T(), err)
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

}
