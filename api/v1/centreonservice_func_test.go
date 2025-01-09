package v1

import (
	"testing"

	"github.com/disaster37/operator-sdk-extra/pkg/apis"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func TestCentreonServiceGetStatus(t *testing.T) {
	status := CentreonServiceStatus{
		BasicRemoteObjectStatus: apis.BasicRemoteObjectStatus{
			LastAppliedConfiguration: "test",
		},
	}
	o := &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Status: status,
	}

	assert.Equal(t, &status, o.GetStatus())
}

func TestCentreonServiceGetExternalName(t *testing.T) {
	var o *CentreonService

	// When name is set
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceSpec{
			Name: "test2",
		},
	}

	assert.Equal(t, "test2", o.GetExternalName())

	// When name isn't set
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceSpec{},
	}

	assert.Equal(t, "test", o.GetExternalName())
}

func TestCentreonServiceGetPlatform(t *testing.T) {
	var o *CentreonService

	// When platform is set
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceSpec{
			PlatformRef: "test2",
		},
	}

	assert.Equal(t, "test2", o.GetPlatform())

	// When platform isn't set
	o = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceSpec{},
	}

	assert.Equal(t, "default", o.GetPlatform())
}
