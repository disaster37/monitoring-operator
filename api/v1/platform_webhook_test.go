package v1

import (
	"context"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *APITestSuite) TestSetupPlatformWebhook() {
	var (
		o   *Platform
		err error
	)

	// Need failed when create platform not on them namespace operator
	o = &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook",
			Namespace: "kube-system",
		},
		Spec: PlatformSpec{
			IsDefault:    false,
			PlatformType: "centreon",
			CentreonSettings: &PlatformSpecCentreonSettings{
				URL:    "http://localhost",
				Secret: "test",
			},
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when create multiple default platform
	o = &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook2",
			Namespace: "default",
		},
		Spec: PlatformSpec{
			IsDefault:    true,
			PlatformType: "centreon",
			CentreonSettings: &PlatformSpecCentreonSettings{
				URL:    "http://localhost",
				Secret: "test",
			},
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Update(context.Background(), o)
	assert.NoError(t.T(), err)

	o = &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook3",
			Namespace: "default",
		},
		Spec: PlatformSpec{
			IsDefault:    true,
			PlatformType: "centreon",
			CentreonSettings: &PlatformSpecCentreonSettings{
				URL:    "http://localhost",
				Secret: "test",
			},
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when not provide centronSetting
	o = &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook3",
			Namespace: "default",
		},
		Spec: PlatformSpec{
			IsDefault:    false,
			PlatformType: "centreon",
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

}
