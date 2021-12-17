package controllers

import (
	"testing"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestInitIngressCentreonServiceDefaultValue(t *testing.T) {
	var (
		centreon     *v1alpha1.CentreonSpec
		expectedCS   *v1alpha1.CentreonService
		cs           *v1alpha1.CentreonService
		placeholders map[string]string
	)

	placeholders = map[string]string{}

	// When no value
	cs = &v1alpha1.CentreonService{}
	centreon = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{},
	}
	expectedCS = &v1alpha1.CentreonService{}
	initIngressCentreonServiceDefaultValue(centreon, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	// When nil value
	cs = &v1alpha1.CentreonService{}
	centreon = &v1alpha1.CentreonSpec{}
	expectedCS = &v1alpha1.CentreonService{}
	initIngressCentreonServiceDefaultValue(centreon, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	centreon = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{},
	}
	expectedCS = &v1alpha1.CentreonService{}
	initIngressCentreonServiceDefaultValue(centreon, nil, placeholders)
	assert.Equal(t, expectedCS, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	initIngressCentreonServiceDefaultValue(nil, cs, placeholders)
	assert.Equal(t, expectedCS, cs)

	// Whan values
	cs = &v1alpha1.CentreonService{}
	centreon = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
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
	initIngressCentreonServiceDefaultValue(centreon, cs, placeholders)
	assert.Equal(t, expectedCS, cs)
}

func TestGeneratePlaceholdersIngressCentreonService(t *testing.T) {

	var (
		ingress    *networkv1.Ingress
		ph         map[string]string
		expectedPh map[string]string
	)

	// When ingress is nil
	ph = generatePlaceholdersIngressCentreonService(nil)
	assert.Empty(t, ph)

	// When all properties
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: "front.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
								{
									Path: "/api",
								},
							},
						},
					},
				},
				{
					Host: "back.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts: []string{"back.local.local"},
				},
			},
		},
	}

	expectedPh = map[string]string{
		"name":             "test",
		"namespace":        "default",
		"rule.0.host":      "front.local.local",
		"rule.0.scheme":    "http",
		"rule.0.path.0":    "/",
		"rule.0.path.1":    "/api",
		"rule.1.host":      "back.local.local",
		"rule.1.scheme":    "https",
		"rule.1.path.0":    "/",
		"label.app":        "appTest",
		"label.env":        "dev",
		"annotation.anno1": "value1",
		"annotation.anno2": "value2",
	}

	ph = generatePlaceholdersIngressCentreonService(ingress)
	assert.Equal(t, expectedPh, ph)

}

func TestInitIngressCentreonServiceFromAnnotations(t *testing.T) {
	var (
		cs          *v1alpha1.CentreonService
		expectedCS  *v1alpha1.CentreonService
		annotations map[string]string
		err         error
	)

	// When nil value
	err = initIngressCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Nil(t, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	err = initIngressCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

	cs = &v1alpha1.CentreonService{}
	expectedCS = &v1alpha1.CentreonService{}
	annotations = map[string]string{}
	err = initIngressCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

	// When annotations
	annotations = map[string]string{
		"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
		"centreon.monitor.k8s.webcenter.fr/template":              "template",
		"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
		"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
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
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}

	err = initIngressCentreonServiceFromAnnotations(annotations, cs)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs)

}

func TestVentreonServiceFromIngress(t *testing.T) {

	var (
		ingress      *networkv1.Ingress
		cs           *v1alpha1.CentreonService
		expectedCs   *v1alpha1.CentreonService
		centreonSpec *v1alpha1.CentreonSpec
		err          error
	)

	// When ingress is nil
	_, err = centreonServiceFromIngress(nil, nil, nil)
	assert.Error(t, err)

	// When no centreonSpec and not all annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"anno1": "value1",
				"anno2": "value2",
			},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: "front.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
								{
									Path: "/api",
								},
							},
						},
					},
				},
				{
					Host: "back.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts: []string{"back.local.local"},
				},
			},
		},
	}
	_, err = centreonServiceFromIngress(ingress, nil, runtime.NewScheme())
	assert.Error(t, err)

	// When no centreonSpec and all annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: "front.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
								{
									Path: "/api",
								},
							},
						},
					},
				},
				{
					Host: "back.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts: []string{"back.local.local"},
				},
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromIngress(ingress, nil, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec and all annotations, priority to annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: "front.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
								{
									Path: "/api",
								},
							},
						},
					},
				},
				{
					Host: "back.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts: []string{"back.local.local"},
				},
			},
		},
	}
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
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
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromIngress(ingress, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec without annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
		},
		Spec: networkv1.IngressSpec{
			Rules: []networkv1.IngressRule{
				{
					Host: "front.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
								{
									Path: "/api",
								},
							},
						},
					},
				},
				{
					Host: "back.local.local",
					IngressRuleValue: networkv1.IngressRuleValue{
						HTTP: &networkv1.HTTPIngressRuleValue{
							Paths: []networkv1.HTTPIngressPath{
								{
									Path: "/",
								},
							},
						},
					},
				},
			},
			TLS: []networkv1.IngressTLS{
				{
					Hosts: []string{"back.local.local"},
				},
			},
		},
	}
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
			Template:        "template",
			NameTemplate:    "name-<label.app>-<label.env>-<namespace>",
			DefaultHost:     "localhost",
			ActivateService: true,
			ServiceGroups:   []string{"sg1"},
			Macros: map[string]string{
				"SCHEME": "<rule.0.scheme>",
				"URL":    "<rule.0.host><rule.0.path.0>",
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "name-appTest-dev-default",
			Template: "template",
			Macros: map[string]string{
				"SCHEME": "http",
				"URL":    "front.local.local/",
			},
			Activated: true,
			Groups:    []string{"sg1"},
		},
	}
	cs, err = centreonServiceFromIngress(ingress, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

}

/*
func (t *ControllerTestSuite) TestCentreonServiceFromIngress() {
	var (
		ingress    *networkv1.Ingress
		cs         *v1alpha1.CentreonService
		expectedCS *v1alpha1.CentreonService
		err        error
	)

	// When ingress is nil
	t.
}
*/
