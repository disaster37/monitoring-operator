package controllers

/*

import (
	"testing"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestInitCentreonServiceFromAnnotations(t *testing.T) {
	var (
		cs          *v1alpha1.CentreonService
		expectedCS  *v1alpha1.CentreonService
		annotations map[string]string
		err         error
	)

	// When nil value
	err = initCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Nil(t, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	err = initCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	annotations = map[string]string{}
	err = initCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

	// When annotations
	annotations = map[string]string{
		"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
		"centreon.monitor.k8s.webcenter.fr/template":              "template",
		"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
		"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
		"centreon.monitor.k8s.webcenter.fr/check-command":         "command",
		"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
		"centreon.monitor.k8s.webcenter.fr/activated":             "1",
		"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
		"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
		"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
		"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
		"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
		"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
		"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
	}

	expectedCS = &v1alpha1.CentreonService{
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			CheckCommand:        "command",
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}

	err = initCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

}

func TestInitCentreonServiceDefaultValue(t *testing.T) {
	var (
		endpointSpec *v1alpha1.CentreonSpecEndpoint
		expectedCS   *v1alpha1.CentreonService
		cs           *v1alpha1.CentreonService
		placeholders map[string]string
	)

	placeholders = map[string]string{}

	// When no value
	cs = &v1alpha1.CentreonService{}
	endpointSpec = &v1alpha1.CentreonSpecEndpoint{}
	expectedCS = &v1alpha1.CentreonService{}
	initCentreonServiceDefaultValue(endpointSpec, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	// When nil value
	cs = &v1alpha1.CentreonService{}
	endpointSpec = nil
	expectedCS = &v1alpha1.CentreonService{}
	initCentreonServiceDefaultValue(endpointSpec, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	endpointSpec = &v1alpha1.CentreonSpecEndpoint{}
	expectedCS = &v1alpha1.CentreonService{}
	initCentreonServiceDefaultValue(endpointSpec, nil, placeholders)
	assert.Equal(t, expectedCS, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	initCentreonServiceDefaultValue(nil, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	// Whan values
	cs = &v1alpha1.CentreonService{}
	endpointSpec = &v1alpha1.CentreonSpecEndpoint{
		Template:        "template",
		NameTemplate:    "name",
		DefaultHost:     "localhost",
		ActivateService: true,
		Arguments:       []string{"arg1"},
		ServiceGroups:   []string{"sg1"},
		Categories:      []string{"cat1"},
		Macros: map[string]string{
			"MACRO1": "value1",
		},
	}

	expectedCS = &v1alpha1.CentreonService{
		Spec: v1alpha1.CentreonServiceSpec{
			Template:   "template",
			Name:       "name",
			Host:       "localhost",
			Activated:  true,
			Arguments:  []string{"arg1"},
			Groups:     []string{"sg1"},
			Categories: []string{"cat1"},
			Macros: map[string]string{
				"MACRO1": "value1",
			},
		},
	}
	initCentreonServiceDefaultValue(endpointSpec, cs, placeholders)
	assert.Equal(t, expectedCS, cs)
}
*/
