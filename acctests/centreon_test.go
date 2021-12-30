package acctests

import (
	"context"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (t *AccTestSuite) TestCentreon() {

	var (
		c   *v1alpha1.Centreon
		uc  *unstructured.Unstructured
		err error
	)

	centreonGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreons",
	}

	/***
	 * Create new centreon resource
	 */
	c = &v1alpha1.Centreon{
		TypeMeta: v1.TypeMeta{
			Kind:       "Centreon",
			APIVersion: "monitor.k8s.webcenter.fr/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.CentreonSpec{
			Endpoints: &v1alpha1.CentreonSpecEndpoint{
				Template:     "my-template",
				NameTemplate: "ping",
				DefaultHost:  "localhost",
				Macros: map[string]string{
					"macro1": "value",
				},
				Arguments:       []string{"arg1"},
				ServiceGroups:   []string{"sg1"},
				Categories:      []string{"cat1"},
				ActivateService: true,
			},
		},
	}

	uc, err = structuredToUntructured(c)
	if err != nil {
		t.T().Fatal(err)
	}

	_, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Create(context.Background(), uc, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that status is updated
	uc, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(uc, c); err != nil {
		t.T().Fatal(err)
	}
	assert.NotEmpty(t.T(), c.Status.CreatedAt)

	/***
	 * Update Centreon resource
	 */
	time.Sleep(30 * time.Second)
	uc, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(uc, c); err != nil {
		t.T().Fatal(err)
	}
	c.Spec.Endpoints.Template = "template2"

	uc, err = structuredToUntructured(c)
	if err != nil {
		t.T().Fatal(err)
	}
	_, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Update(context.Background(), uc, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that status is updated
	uc, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(uc, c); err != nil {
		t.T().Fatal(err)
	}
	assert.NotEmpty(t.T(), c.Status.UpdatedAt)

	/***
	 * Delete service
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclient.Resource(centreonGVR).Namespace("default").Delete(context.Background(), "test", *&v1.DeleteOptions{}); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService delete on k8s
	uc, err = t.k8sclient.Resource(centreonGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "Centreon not delete")
	}

}
