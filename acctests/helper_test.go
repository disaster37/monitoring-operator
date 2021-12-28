package acctests

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func structuredToUntructured(s interface{}) (*unstructured.Unstructured, error) {
	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
	if err != nil {
		return nil, err
	}

	us := &unstructured.Unstructured{
		Object: data,
	}

	return us, nil
}

func unstructuredToStructured(us *unstructured.Unstructured, s interface{}) error {
	return runtime.DefaultUnstructuredConverter.FromUnstructured(us.Object, s)
}
