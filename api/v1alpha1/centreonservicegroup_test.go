package v1alpha1

import (
	"testing"

	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestCentreonServiceGroupCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *CentreonServiceGroup
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: CentreonServiceGroupSpec{
			Name:        "sg1",
			Description: "my sg",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &CentreonServiceGroup{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func TestToCentreonServiceGroup(t *testing.T) {

	sg := &CentreonServiceGroup{
		Spec: CentreonServiceGroupSpec{
			Name:        "sg1",
			Description: "my sg",
			Activated:   true,
		},
	}

	expectedCsg := &centreonhandler.CentreonServiceGroup{
		Name:        "sg1",
		Description: "my sg",
		Activated:   "1",
		Comment:     "Managed by monitoring-operator",
	}

	currentCsg, err := sg.ToCentreonServiceGroup()
	assert.NoError(t, err)
	assert.Equal(t, expectedCsg, currentCsg)
}

func TestCentreonServiceGroupIsValid(t *testing.T) {
	var centreonServiceGroup *CentreonServiceGroup

	// When is valid
	centreonServiceGroup = &CentreonServiceGroup{
		Spec: CentreonServiceGroupSpec{
			Name:        "sg1",
			Description: "my sg",
		},
	}
	assert.True(t, centreonServiceGroup.IsValid())

	// When invalid
	centreonServiceGroup = &CentreonServiceGroup{
		Spec: CentreonServiceGroupSpec{
			Name:        "",
			Description: "my sg",
		},
	}
	assert.False(t, centreonServiceGroup.IsValid())

	centreonServiceGroup = &CentreonServiceGroup{
		Spec: CentreonServiceGroupSpec{
			Name:        "sg1",
			Description: "",
		},
	}
	assert.False(t, centreonServiceGroup.IsValid())

	centreonServiceGroup = &CentreonServiceGroup{}
	assert.False(t, centreonServiceGroup.IsValid())
}
