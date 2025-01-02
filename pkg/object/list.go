package object

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ObjectList interface {
	metav1.ListInterface
	runtime.Object
	GetItems() []client.Object
}
