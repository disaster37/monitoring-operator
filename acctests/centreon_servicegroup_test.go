package acctests

import (
	"context"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	condition "k8s.io/apimachinery/pkg/api/meta"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (t *AccTestSuite) TestCentreonServiceGroup() {

	var (
		csg        *v1alpha1.CentreonServiceGroup
		ucsg       *unstructured.Unstructured
		sg         *centreonhandler.CentreonServiceGroup
		expectedSG *centreonhandler.CentreonServiceGroup
		err       error
	)

	centreonServiceGroupGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreonservicegroups",
	}

	/***
	 * Create new centreon serviceGroup resource
	 */
	csg = &v1alpha1.CentreonServiceGroup{
		TypeMeta: v1.TypeMeta{
			Kind:       "CentreonServiceGroup",
			APIVersion: "monitor.k8s.webcenter.fr/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
		},
		Spec: v1alpha1.CentreonServiceGroupSpec{
			Name:         "sg1",
			Description: "my sg",
			Activated:    true,
		},
	}
	expectedSG = &centreonhandler.CentreonServiceGroup{
		Name:                "sg1",
		Comment:             "Managed by monitoring-operator",
		Description: "my sg",
		Activated:           "1",
	}

	ucsg, err = structuredToUntructured(csg)
	if err != nil {
		t.T().Fatal(err)
	}

	_, err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Create(context.Background(), ucsg, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that status is updated
	ucsg, err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucsg, csg); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "sg1", csg.Status.ServiceGroupName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, controllers.CentreonServiceGroupCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	sg, err = t.centreon.GetServiceGroup("sg1")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), sg)
	assert.Equal(t.T(), expectedSG, sg)

	/***
	 * Update Centreon resource
	 */
	time.Sleep(30 * time.Second)
	ucsg, err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucsg, csg); err != nil {
		t.T().Fatal(err)
	}
	csg.Spec.Description = "my sg2"
	ucsg, err = structuredToUntructured(csg)
	if err != nil {
		t.T().Fatal(err)
	}
	expectedSG = &centreonhandler.CentreonServiceGroup{
		Name:                "sg1",
		Comment:             "Managed by monitoring-operator",
		Description: "my sg2",
	}
	_, err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Update(context.Background(), ucsg, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucsg, err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Get(context.Background(), "test", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucsg, csg); err != nil {
		t.T().Fatal(err)
	}

	// Check service updated on Centreon
	sg, err = t.centreon.GetServiceGroup("sg1")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), sg)
	assert.Equal(t.T(), expectedSG, sg)

	/***
	 * Delete serviceGroup
	 */
	time.Sleep(20 * time.Second)
	err = t.k8sclient.Resource(centreonServiceGroupGVR).Namespace("default").Delete(context.Background(), "test", *&v1.DeleteOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check service is delete from centreon
	sg, err = t.centreon.GetServiceGroup("sg1")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.Nil(t.T(), sg)
}
