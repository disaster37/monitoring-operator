package v1

import (
	"context"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *APITestSuite) TestSetupPlatformIndexer() {
	platform := &Platform{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
		},
		Spec: PlatformSpec{
			IsDefault:    false,
			PlatformType: "centreon",
			CentreonSettings: &PlatformSpecCentreonSettings{
				URL:                   "http://localhost",
				SelfSignedCertificate: true,
				Secret:                "my-secret",
			},
		},
	}

	err := t.k8sClient.Create(context.Background(), platform)
	assert.NoError(t.T(), err)
}
