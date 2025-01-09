package v1

import (
	"testing"

	"github.com/disaster37/operator-sdk-extra/pkg/apis"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPlaformGetStatus(t *testing.T) {
	status := PlatformStatus{
		BasicRemoteObjectStatus: apis.BasicRemoteObjectStatus{
			LastAppliedConfiguration: "test",
		},
	}
	o := &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Status: status,
	}

	assert.Equal(t, &status, o.GetStatus())
}

func TestPlatformGetExternalName(t *testing.T) {
	// When name isn't set
	o := &Platform{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "test",
		},
		Spec: PlatformSpec{},
	}

	assert.Equal(t, "test", o.GetExternalName())
}
