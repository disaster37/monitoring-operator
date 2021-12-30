package acctests

import (
	"context"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
	networkv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (t *AccTestSuite) TestIngress() {

	var (
		cs        *v1alpha1.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		ingress   *networkv1.Ingress
		err       error
	)

	centreonServiceGVR := schema.GroupVersionResource{
		Group:    "monitor.k8s.webcenter.fr",
		Version:  "v1alpha1",
		Resource: "centreonservices",
	}

	/***
	 * Create new ingress
	 */
	pathType := networkv1.PathTypePrefix
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ingress",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/discover":               "true",
				"centreon.monitor.k8s.webcenter.fr/name":          "test-ingress-ping",
				"centreon.monitor.k8s.webcenter.fr/template":      "template-test",
				"centreon.monitor.k8s.webcenter.fr/host":          "localhost",
				"centreon.monitor.k8s.webcenter.fr/activated":     "1",
				"centreon.monitor.k8s.webcenter.fr/check-command": "ping",
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
		Macros:              []*models.Macro{},
		Activated:           "1",
	}
	_, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Create(context.Background(), ingress, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &v1alpha1.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-ingress", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.NotEmpty(t.T(), cs.Status.ID)
	assert.NotEmpty(t.T(), cs.Status.CreatedAt)

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
	assert.Equal(t.T(), expectedS, s)

	/***
	 * Update Ingress
	 */
	time.Sleep(30 * time.Second)
	ingress, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Get(context.Background(), "test-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/groups"] = "sg1"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/categories"] = "Ping"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/arguments"] = "arg1"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/normal-check-interval"] = "60"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/retry-check-interval"] = "10"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/max-check-attempts"] = "2"
	ingress.Annotations["centreon.monitor.k8s.webcenter.fr/macros"] = `{"MAC1": "value"}`

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-ingress-ping",
		CheckCommand:        "ping",
		Template:            "template-test",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "Managed by monitoring-operator",
		Groups:              []string{"sg1"},
		Categories:          []string{"Ping"},
		Macros: []*models.Macro{
			{
				Name:   "MAC1",
				Value:  "value",
				Source: "direct",
			},
		},
		Activated:           "1",
		NormalCheckInterval: "60",
		RetryCheckInterval:  "10",
		MaxCheckAttempts:    "2",
		CheckCommandArgs:    "!arg1",
	}
	_, err = t.k8sclientStd.NetworkingV1().Ingresses("default").Update(context.Background(), ingress, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that status is updated
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-ingress", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.NotEmpty(t.T(), cs.Status.UpdatedAt)

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	if err != nil {
		t.T().Fatal(err)
	}
	assert.NotNil(t.T(), s)
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
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "test-ingress", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not delete on k8s after delete ingress")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-ingress-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
