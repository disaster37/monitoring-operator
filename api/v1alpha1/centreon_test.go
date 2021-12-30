package v1alpha1

import (
	"time"

	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestCentreonCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *Centreon
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &Centreon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: CentreonSpec{
			Endpoints: &CentreonSpecEndpoint{},
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &Centreon{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)

}

func (t *V1alpha1TestSuite) TestCentreonIsSubmitted() {
	centreon := &Centreon{}
	assert.False(t.T(), centreon.IsSubmitted())

	centreon.Status.CreatedAt = "test"
	assert.True(t.T(), centreon.IsSubmitted())
}

func (t *V1alpha1TestSuite) TestCentreonIsBeingDeleted() {
	centreon := &Centreon{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
		},
	}
	assert.True(t.T(), centreon.IsBeingDeleted())
}

func (t *V1alpha1TestSuite) TestCentreonFinalizer() {
	centreon := &Centreon{}

	centreon.AddFinalizer()
	assert.Equal(t.T(), 1, len(centreon.GetFinalizers()))
	assert.True(t.T(), centreon.HasFinalizer())

	centreon.RemoveFinalizer()
	assert.Equal(t.T(), 0, len(centreon.GetFinalizers()))
	assert.False(t.T(), centreon.HasFinalizer())
}
