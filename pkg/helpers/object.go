package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetItems permit to get items contend from ObjectList interface
func GetItems(o client.ObjectList) (items []client.Object) {
	if o == nil || reflect.ValueOf(o).IsNil() {
		panic("ressource can't be nil")
	}

	val := reflect.ValueOf(o).Elem()
	valueField := val.FieldByName("Items")

	items = make([]client.Object, valueField.Len())
	for i := range items {
		items[i] = valueField.Index(i).Addr().Interface().(client.Object)
	}

	return items
}

// GetObjectWithMeta return current object with TypeMeta to kwons the object type
func GetObjectWithMeta(o client.Object, s runtime.ObjectTyper) client.Object {
	if o == nil {
		panic("Object can't be nil")
	}

	y := printers.NewTypeSetter(s).ToPrinter(&printers.JSONPrinter{})
	buf := new(bytes.Buffer)
	if err := y.PrintObj(o, buf); err != nil {
		panic(err)
	}

	if err := json.Unmarshal(buf.Bytes(), o); err != nil {
		panic(err)
	}

	return o
}

// GetObjectType print the current object type
func GetObjectType(o schema.ObjectKind) string {
	if o == nil {
		panic("Object can't be nil")
	}
	return fmt.Sprintf("%s/%s/%s", o.GroupVersionKind().Group, o.GroupVersionKind().Version, o.GroupVersionKind().Kind)
}

// CloneObject permit to clone current object type
func CloneObject[objectType comparable](o objectType) objectType {
	if reflect.TypeOf(o).Kind() != reflect.Pointer {
		panic("CloneObject work only with pointer")
	}

	if reflect.ValueOf(o).IsNil() {
		panic("Object can't be nill")
	}

	return reflect.New(reflect.TypeOf(o).Elem()).Interface().(objectType)
}
