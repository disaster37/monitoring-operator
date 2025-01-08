package v1

import (
	"context"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (t *APITestSuite) TestSetupTemplateWebhook() {
	var (
		o   *Template
		err error
	)

	// Need Work when template is valid yaml
	o = &Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook",
			Namespace: "default",
		},
		Spec: TemplateSpec{
			Template: `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
  host: "localhost"
  name: "test-certificate-ping"
  template: "template-test"
  checkCommand: "ping"
  macros:
  LABEL: "{{ .labels.foo }}"
  activate: true`,
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.NoError(t.T(), err)

	// Need failed when yaml is invalid (bad indent)
	o = &Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook2",
			Namespace: "default",
		},
		Spec: TemplateSpec{
			Template: `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
  spec:
   host: "localhost"`,
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

	// Need failed when kind and api version not provided
	o = &Template{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-webhook3",
			Namespace: "default",
		},
		Spec: TemplateSpec{
			Template: `
spec:
  host: "localhost"
  name: "test-certificate-ping"
  template: "template-test"
  checkCommand: "ping"
  macros:
  LABEL: "{{ .labels.foo }}"
  activate: true`,
		},
	}
	err = t.k8sClient.Create(context.Background(), o)
	assert.Error(t.T(), err)

}
