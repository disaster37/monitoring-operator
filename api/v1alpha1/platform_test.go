package v1alpha1

import (
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestPlatformCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *Platform
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: PlatformSpec{
			IsDefault:    true,
			PlatformType: "centreon",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &Platform{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)

}
