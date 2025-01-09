package v1

import (
	"context"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *APITestSuite) TestSetupCentreonServiceWebhook() {
	var (
		o   *CentreonService
		err error
	)

	// Need failed when create same resource by external name on same target platform
	// Check we can update it
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			PlatformRef: "webhook",
			Template:    "test",
			Host:        "localhost",
			Name:        "test",
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Update(context.Background(), o)
	assert.NoError(t.T(), err)

	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook2",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			PlatformRef: "webhook",
			Template:    "test",
			Host:        "localhost",
			Name:        "test",
		},
	}

	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when create same resource by external name on default platform
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook3",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			Template: "test",
			Host:     "localhost",
			Name:     "test",
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.NoError(t.T(), err)

	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook4",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			Template: "test",
			Host:     "localhost",
			Name:     "test",
		},
	}

	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when not specify template and checkCommand
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook5",
			Namespace: "default",
		},
		Spec: CentreonServiceSpec{
			Host: "localhost",
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when update platformRef (immutable)
	if err = t.k8sClient.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "test-webhook"}, o); err != nil {
		t.T().Fatal(err)
	}
	o.Spec.PlatformRef = "test2"
	err = t.k8sClient.Update(context.Background(), o)
	assert.Error(t.T(), err)
}
