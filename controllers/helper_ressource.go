package controllers

import (
	"errors"
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetSpec permit to get the Spec field from Object interface
func GetSpec(o client.Object) (spec any, err error) {
	if o == nil || reflect.ValueOf(o).IsNil() {
		return nil, errors.New("Ressource can't be nil")
	}

	val := reflect.ValueOf(o).Elem()
	valueField := val.FieldByName("Spec")

	return valueField.Interface(), nil
}

// SetSpec permit to set the spec contend on Spec field from Object interface
func SetSpec(o client.Object, spec any) (err error) {
	if o == nil || reflect.ValueOf(o).IsNil() {
		return errors.New("Ressource can't be nil")
	}

	val := reflect.ValueOf(o).Elem()
	valSpec := reflect.ValueOf(spec)
	valueField := val.FieldByName("Spec")

	if !valueField.CanSet() {
		return errors.New("The field Struct can't be setted")
	}
	valueField.Set(valSpec)

	return nil
}

// GetItems permit to get items contend from ObjectList interface
func GetItems(o client.ObjectList) (items []client.Object, err error) {
	if o == nil || reflect.ValueOf(o).IsNil() {
		return nil, errors.New("Ressource can't be nil")
	}

	val := reflect.ValueOf(o).Elem()
	valueField := val.FieldByName("Items")

	items = make([]client.Object, valueField.Len())
	for i := range items {
		items[i] = valueField.Index(i).Addr().Interface().(client.Object)
	}

	return items, nil
}
