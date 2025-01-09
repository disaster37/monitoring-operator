package v1

import (
	"testing"

	"github.com/disaster37/operator-sdk-extra/pkg/apis"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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

func TestCentreonServiceGroupGetStatus(t *testing.T) {
	status := CentreonServiceGroupStatus{
		BasicRemoteObjectStatus: apis.BasicRemoteObjectStatus{
			LastAppliedConfiguration: "test",
		},
	}
	o := &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Status: status,
	}

	assert.Equal(t, &status, o.GetStatus())
}

func TestCentreonServiceGroupGetExternalName(t *testing.T) {
	var o *CentreonServiceGroup

	// When name is set
	o = &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceGroupSpec{
			Name: "test2",
		},
	}

	assert.Equal(t, "test2", o.GetExternalName())

	// When name isn't set
	o = &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceGroupSpec{},
	}

	assert.Equal(t, "test", o.GetExternalName())
}

func TestCentreonServiceGroupGetPlatform(t *testing.T) {
	var o *CentreonServiceGroup

	// When platform is set
	o = &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceGroupSpec{
			PlatformRef: "test2",
		},
	}

	assert.Equal(t, "test2", o.GetPlatform())

	// When platform isn't set
	o = &CentreonServiceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: CentreonServiceGroupSpec{},
	}

	assert.Equal(t, "default", o.GetPlatform())
}
