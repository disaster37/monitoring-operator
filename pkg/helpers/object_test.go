package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestGetObjectWithMeta(t *testing.T) {
	err := scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	o := &corev1.Secret{}

	o = GetObjectWithMeta(o, scheme.Scheme).(*corev1.Secret)
	assert.Equal(t, "Secret", o.GetObjectKind().GroupVersionKind().Kind)
}

func TestGetItems(t *testing.T) {
	list := &corev1.SecretList{
		Items: []corev1.Secret{
			{},
		},
	}

	assert.Len(t, GetItems(list), 1)
}

func TestGetObjectType(t *testing.T) {
	err := scheme.AddToScheme(scheme.Scheme)
	if err != nil {
		panic(err)
	}
	o := &corev1.Secret{}
	o = GetObjectWithMeta(o, scheme.Scheme).(*corev1.Secret)

	assert.Equal(t, "/v1/Secret", GetObjectType(o))
}

func TestCloneObject(t *testing.T) {
	// Normal
	o := &corev1.Secret{}
	clone := CloneObject(o)
	assert.Equal(t, o, clone)

	// When nil
	var s *corev1.Secret
	assert.Panics(t, func() {
		CloneObject(s)
	})
}
