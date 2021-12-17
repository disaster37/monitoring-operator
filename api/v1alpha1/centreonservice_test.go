package v1alpha1

import (
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestCentreonServiceCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *CentreonService
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: CentreonServiceSpec{
			Host: "central",
			Name: "ping",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &CentreonService{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func (t *V1alpha1TestSuite) TestCentreonServiceIsSubmitted() {
	centreonService := &CentreonService{}
	assert.False(t.T(), centreonService.IsSubmitted())

	centreonService.Status.ID = "test"
	assert.True(t.T(), centreonService.IsSubmitted())
}

func (t *V1alpha1TestSuite) TestCentreonServiceIsBeingDeleted() {
	centreonService := &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
		},
	}
	assert.True(t.T(), centreonService.IsBeingDeleted())
}

func (t *V1alpha1TestSuite) TestCentreonServiceFinalizer() {
	centreonService := &CentreonService{}

	centreonService.AddFinalizer()
	assert.Equal(t.T(), 1, len(centreonService.GetFinalizers()))
	assert.True(t.T(), centreonService.HasFinalizer())

	centreonService.RemoveFinalizer()
	assert.Equal(t.T(), 0, len(centreonService.GetFinalizers()))
	assert.False(t.T(), centreonService.HasFinalizer())
}

func TestCentreonServiceIsValid(t *testing.T) {
	var centreonService *CentreonService

	// When is valid
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
		},
	}
	assert.True(t, centreonService.IsValid())

	// When invalid
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "",
			Name:     "ping",
			Template: "template",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "",
			Template: "template",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{}
	assert.False(t, centreonService.IsValid())
}
