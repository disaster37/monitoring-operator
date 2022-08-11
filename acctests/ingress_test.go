package acctests

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	api "github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/controllers"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (t *AccTestSuite) TestIngress() {

	var (
		cs        *api.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		ingress   *networkv1.Ingress
		err       error
	)

	centreonServiceGVR := api.GroupVersion.WithResource("centreonservices")
	templateCentreonServiceGVR := api.GroupVersion.WithResource("templates")

	/***
	 * Create new template dedicated for ingress test
	 */
	tcs := &api.Template{
		TypeMeta: v1.TypeMeta{
			Kind:       "Template",
			APIVersion: fmt.Sprintf("%s/%s", api.GroupVersion.Group, api.GroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "check-ingress",
		},
		Spec: api.TemplateSpec{
			Type: "CentreonService",
			Template: `
{{ $rule := index .rules 0}}
{{ $path := index $rule.paths 0}}
host: "localhost"
name: "test-ingress-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
  SCHEME: "{{ $rule.scheme }}"
  HOST: "{{ $rule.host }}"
  PATH: "{{ $path }}"
activate: true`,
		},
	}
	tcsu, err := structuredToUntructured(tcs)
	if err != nil {
		t.T().Fatal(err)
	}
	if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Create(context.Background(), tcsu, v1.CreateOptions{}); err != nil {
		t.T().Fatal(err)
	}

	/***
	 * Create new ingress
	 */
	pathType := networkv1.PathTypePrefix
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ingress",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-ingress"}]`,
			},
			Labels: map[string]string{
				"foo": "bar",
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
									Path:     "/",
									PathType: &pathType,
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: "test",
											Port: networkv1.ServiceBackendPort{Number: 80},
										},
									},
								},
								{
									Path:     "/api",
									PathType: &pathType,
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: "test",
											Port: networkv1.ServiceBackendPort{Number: 80},
										},
									},
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
									Path:     "/",
									PathType: &pathType,
									Backend: networkv1.IngressBackend{
										Service: &networkv1.IngressServiceBackend{
											Name: "test",
											Port: networkv1.ServiceBackendPort{Number: 80},
										},
									},
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
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ingress-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros: []*models.Macro{
			{
				Name:   "LABEL",
				Value:  "bar",
				Source: "direct",
			},
			{
				Name:   "SCHEME",
				Value:  "http",
				Source: "direct",
			},
			{
				Name:   "HOST",
				Value:  "front.local.local",
				Source: "direct",
			},
			{
				Name:   "PATH",
				Value:  "/",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Create(context.Background(), ingress, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &api.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-ingress", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-ingress-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)

	// Sort macro to fix test
	sort.Slice(expectedS.Macros, func(i, j int) bool {
		return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	})
	sort.Slice(s.Macros, func(i, j int) bool {
		return s.Macros[i].Name < s.Macros[j].Name
	})
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update Ingress
	 */
	time.Sleep(30 * time.Second)
	ingress, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Get(context.Background(), "test-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	ingress.Labels = map[string]string{"foo": "bar2"}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ingress-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros: []*models.Macro{
			{
				Name:   "LABEL",
				Value:  "bar2",
				Source: "direct",
			},
			{
				Name:   "SCHEME",
				Value:  "http",
				Source: "direct",
			},
			{
				Name:   "HOST",
				Value:  "front.local.local",
				Source: "direct",
			},
			{
				Name:   "PATH",
				Value:  "/",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Update(context.Background(), ingress, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "bar2", cs.Spec.Macros["LABEL"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)

	// Sort macro to fix test
	sort.Slice(expectedS.Macros, func(i, j int) bool {
		return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	})
	sort.Slice(s.Macros, func(i, j int) bool {
		return s.Macros[i].Name < s.Macros[j].Name
	})
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update ingress template
	 */
	time.Sleep(30 * time.Second)
	tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Get(context.Background(), "check-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(tcsu, tcs); err != nil {
		t.T().Fatal(err)
	}
	tcs.Spec.Template = `
{{ $rule := index .rules 0}}
{{ $path := index $rule.paths 0}}
host: "localhost"
name: "test-ingress-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
  SCHEME: "{{ $rule.scheme }}"
  HOST: "{{ $rule.host }}"
  PATH: "{{ $path }}"
  TEST: "plop"
activate: true`

	tcsu, err = structuredToUntructured(tcs)
	if err != nil {
		t.T().Fatal(err)
	}
	if _, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Update(context.Background(), tcsu, v1.UpdateOptions{}); err != nil {
		t.T().Fatal(err)
	}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ingress-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{},
		Categories:          []string{},
		Macros: []*models.Macro{
			{
				Name:   "LABEL",
				Value:  "bar2",
				Source: "direct",
			},
			{
				Name:   "SCHEME",
				Value:  "http",
				Source: "direct",
			},
			{
				Name:   "HOST",
				Value:  "front.local.local",
				Source: "direct",
			},
			{
				Name:   "PATH",
				Value:  "/",
				Source: "direct",
			},
			{
				Name:   "TEST",
				Value:  "plop",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "plop", cs.Spec.Macros["TEST"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	// Sort macro to fix test
	sort.Slice(expectedS.Macros, func(i, j int) bool {
		return expectedS.Macros[i].Name < expectedS.Macros[j].Name
	})
	sort.Slice(s.Macros, func(i, j int) bool {
		return s.Macros[i].Name < s.Macros[j].Name
	})
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Delete service
	 */
	time.Sleep(20 * time.Second)
	if err = t.k8sclientStd.NetworkingV1().Ingresses("default").Delete(context.Background(), "test-ingress", *metav1.NewDeleteOptions(0)); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService delete on k8s
	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-ingress", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not delete on k8s after delete ingress")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
