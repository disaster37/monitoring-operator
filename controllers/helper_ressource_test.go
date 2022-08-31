package controllers

import (
	"testing"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/stretchr/testify/assert"
	networkv1 "k8s.io/api/networking/v1"
)

func TestGetSpec(t *testing.T) {
	// When OK
	name := "test"
	spec := networkv1.IngressSpec{
		IngressClassName: &name,
	}
	i := &networkv1.Ingress{
		Spec: spec,
	}
	currentSpec, err := GetSpec(i)
	assert.NoError(t, err)
	assert.Equal(t, spec, currentSpec)

	// When KO
	var i2 *networkv1.Ingress
	_, err = GetSpec(i2)
	assert.Error(t, err)

	_, err = GetSpec(nil)
	assert.Error(t, err)
}

func TestSetSpec(t *testing.T) {
	// When OK
	name := "test"
	spec := networkv1.IngressSpec{
		IngressClassName: &name,
	}
	i := &networkv1.Ingress{}
	err := SetSpec(i, spec)
	assert.NoError(t, err)
	assert.Equal(t, spec, i.Spec)

	// When KO
	var i2 *networkv1.Ingress
	err = SetSpec(i2, spec)
	assert.Error(t, err)

	err = SetSpec(nil, spec)
	assert.Error(t, err)
}

func TestGetItems(t *testing.T) {
	// When OK
	i := &monitorapi.CentreonServiceGroup{}
	iList := &monitorapi.CentreonServiceGroupList{
		Items: []monitorapi.CentreonServiceGroup{
			*i,
		},
	}
	currentItems, err := GetItems(iList)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(currentItems))
	assert.Equal(t, i, currentItems[0])

	// When KO
	var iList2 *monitorapi.CentreonServiceGroupList
	_, err = GetItems(iList2)
	assert.Error(t, err)

	_, err = GetItems(nil)
	assert.Error(t, err)
}
