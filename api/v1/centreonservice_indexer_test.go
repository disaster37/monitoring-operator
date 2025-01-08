package v1

import (
	"context"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *APITestSuite) TestSetupCentreonServiceIndexer() {
	// Add CentreonService to force  indexer execution

	o := &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			PlatformRef: "test",
			Template:    "test",
		},
	}

	err := t.k8sClient.Create(context.Background(), o)
	assert.NoError(t.T(), err)
}
