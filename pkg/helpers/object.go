package helpers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetItems permit to get items contend from ObjectList interface
func GetItems(o client.ObjectList) (items []client.Object, err error) {
	if o == nil || reflect.ValueOf(o).IsNil() {
		return nil, errors.New("ressource can't be nil")
	}

	val := reflect.ValueOf(o).Elem()
	valueField := val.FieldByName("Items")

	items = make([]client.Object, valueField.Len())
	for i := range items {
		items[i] = valueField.Index(i).Addr().Interface().(client.Object)
	}

	return items, nil
}

// GetObjectWithMeta return current object with TypeMeta to kwons the object type
func GetObjectWithMeta(o client.Object, s runtime.ObjectTyper) client.Object {

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
func GetObjectType(o client.Object) string {
	return fmt.Sprintf("%s/%s/%s", o.GetObjectKind().GroupVersionKind().Kind, o.GetObjectKind().GroupVersionKind().Group, o.GetObjectKind().GroupVersionKind().Version)
}

// CloneObject permit to clone current object type
func CloneObject[objectType comparable](o objectType) objectType {
	if reflect.TypeOf(o).Kind() != reflect.Pointer {
		panic("CloneObject work only with pointer")
	}

	return reflect.New(reflect.TypeOf(o)).Interface().(objectType)
}
