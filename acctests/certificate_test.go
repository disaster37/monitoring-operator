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
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func (t *AccTestSuite) TestCertificate() {

	var (
		cs        *api.CentreonService
		ucs       *unstructured.Unstructured
		s         *centreonhandler.CentreonService
		expectedS *centreonhandler.CentreonService
		certificate   *core.Secret
		err       error
	)

	centreonServiceGVR := api.GroupVersion.WithResource("centreonservices")
	templateCentreonServiceGVR := api.GroupVersion.WithResource("templates")

	/***
	 * Create new template dedicated for certificate test
	 */
	tcs := &api.Template{
		TypeMeta: v1.TypeMeta{
			Kind:       "Template",
			APIVersion: fmt.Sprintf("%s/%s", api.GroupVersion.Group, api.GroupVersion.Version),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "check-certificate",
		},
		Spec: api.TemplateSpec{
			Type: "CentreonService",
			Template: `
host: "localhost"
name: "test-certificate-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
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
	 * Create new certificate
	 */
	certificate = &core.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-certificate",
			Annotations: map[string]string{
				"monitor.k8s.webcenter.fr/templates": `[{"namespace":"default", "name": "check-certificate"}]`,
			},
			Labels: map[string]string{
				"foo": "bar",
			},
		},
		Type: core.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.key": []byte("test"),
			"tls.crt": []byte(`
-----BEGIN CERTIFICATE-----
MIIGSTCCBDGgAwIBAgITPQAAtLDzkxCyNoADlAABAAC0sDANBgkqhkiG9w0BAQsF
ADA/MRIwEAYKCZImiZPyLGQBGRYCQUQxEjAQBgoJkiaJk/IsZAEZFgJETTEVMBMG
A1UEAxMMUEtJIFNJSE0gT1BFMB4XDTIyMDgwMTA3MzUyOVoXDTIzMDgwMTA3MzUy
OVowgd0xCzAJBgNVBAYTAkZSMRYwFAYDVQQIEw1JbGUtZGUtRnJhbmNlMQ4wDAYD
VQQHEwVQYXJpczE5MDcGA1UEChMwU1lTVEVNRVMgSU5GT1JNQVRJT04gSEFSTU9O
SUUgTVVUVUVMTEVTIFNJSE0gR0lFMR0wGwYDVQQLExREaXJlY3Rpb24gUHJvZHVj
dGlvbjEUMBIGA1UEAxMLcmFuY2hlci1wcmQxNjA0BgkqhkiG9w0BCQEWJ2NvbnRh
Y3QuY2VydGlmaWNhdEBoYXJtb25pZS1tdXR1ZWxsZS5mcjCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAMYBYVnQdg42BosCJhB+Gteu9ozjgOfrqeNDiuA1
Tq1ialZeFU1vNvMp2v2GBxkZxIRnRImz4I41LpddAceKvkfFebhuVsH5OX6ENGH/
FDjpe6hd3AHDNVzZMybv5aP/FPphD9DL3YkYasnG+a5qJ/l+jZ7FxIVSnHj31IN+
2gZ4rvXdEBpsUx6v1uImZgcUd96cwXT3ec/qvgefaitwCjqyywH+wV64cVWumzdX
JVZZxOKlWvH4LFvfLGdeArcATZq2vbdx6VIFlQKfTm9akLaZTUz3v/ArUhVL0gya
DXjDXJ5zR74M1SvDb2Bb5CgpzsLk9TcapCp0mEcqmS+xPOcCAwEAAaOCAZ0wggGZ
MA4GA1UdDwEB/wQEAwIFoDBEBgNVHREEPTA7ggtyYW5jaGVyLXByZIIUcmFuY2hl
ci1wcmQuaG0uZG0uYWSCFioucmFuY2hlci1wcmQuaG0uZG0uYWQwHQYDVR0OBBYE
FJMq5SHDaRTpb8qVT8qKiSccH/PpMB8GA1UdIwQYMBaAFOSLR3dD9ClRyoAxxVwI
aMcM/xBjMD8GA1UdHwQ4MDYwNKAyoDCGLmh0dHA6Ly9wa2kuc2lobS5mci9jcmwv
UEtJJTIwU0lITSUyME9QRSgxKS5jcmwwPAYJKwYBBAGCNxUHBC8wLQYlKwYBBAGC
NxUIpL4Ehe2AKIPtnweBy/Mog8qXB4EykrMXg8SdaAIBZAIBCDATBgNVHSUEDDAK
BggrBgEFBQcDATAbBgkrBgEEAYI3FQoEDjAMMAoGCCsGAQUFBwMBMFAGA1UdIARJ
MEcwRQYMKwYBBAGC8QUBAQEBMDUwMwYIKwYBBQUHAgEWJ2h0dHA6Ly9wa2kuc2lo
bS5mci9QQ19BQ19UZWNobmlxdWUucGRmADANBgkqhkiG9w0BAQsFAAOCAgEAMcFs
oYb2VEN22+4szhOgliF1vAMmxGxgEj4tZIRMje/stMCBF2q2W46akT8TQCZOc/fB
pinsL/XizUhccf5G+g6XjTgbDP9sFPxIz8T4g6xNUxxXLtvLDWrucecTrlDOBxPT
QcKs0FypKVDM5D47ooqoczvvJoHJv3GIBgP8OR2X//K/Hvz4Ac0oO9XAD/cn85pK
5k4WKE/L4I10Ffu1kYS6qIblwKJep49ekU7uwO7fjnnzueaG/x+BXSBZZlxf0eMq
xiySvCVO5yJS9FysANKy4ilnoSXvfZPG44tXdF2B9RFWzeVbzj9MGPu+L0cjAxUC
sYtMqUQ7XsgiiizrGz6nvFwgm3J9howazMCfKQIzVvE0QaVmTWThjY1FwOEgwG9O
fKRzAAJ06XOT4j6ApmoOKx/bs+IhLx1655JdKThqQefUYHJBsO0gE84YnAjgGZXC
lHXO5SldffCKri8lTHMOCsz8b2O+DXl8oyD/WO7Yx8YFUH9/c+lo6yISCJe3Kh8c
MWoiGtTPopx2B7e3+We08NCQC7+kNkb4sd9fJUg4A+hX82TH7z+deLV7vfODlwsi
f2i6gQxarKRKxP4hzwWLutgWsMT6fYb6xK18uF0OHBpcKORlRgKCfOiOoli+mb/n
t92lHyOldNtNHZtcJSRqLQNZo7MixgHDJ/jLAio=
-----END CERTIFICATE-----`)},
	}
	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-certificate-ping",
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
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.CoreV1().Secrets("default").Create(context.Background(), certificate, v1.CreateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check that CentreonService created and in right status
	cs = &api.CentreonService{}
	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-certificate", v1.GetOptions{})
	if err != nil {
		assert.Fail(t.T(), err.Error())
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		assert.Fail(t.T(), err.Error())
	}
	assert.Equal(t.T(), "localhost", cs.Status.Host)
	assert.Equal(t.T(), "test-certificate-ping", cs.Status.ServiceName)
	assert.True(t.T(), condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, controllers.CentreonServiceCondition, v1.ConditionTrue))

	// Check ressource created on Centreon
	s, err = t.centreon.GetService("localhost", "test-certificate-ping")
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
	 * Update certificate
	 */
	time.Sleep(30 * time.Second)
	certificate, err = t.k8sclientStd.CoreV1().Secrets("default").Get(context.Background(), "test-certificate", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	certificate.Labels = map[string]string{"foo": "bar2"}

	expectedS = &centreonhandler.CentreonService{
		Host:                "localhost",
		Name:                "test-certificate-ping",
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
		},
		Activated: "1",
	}
	_, err = t.k8sclientStd.CoreV1().Secrets("default").Update(context.Background(), certificate, v1.UpdateOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-certificate", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "bar2", cs.Spec.Macros["LABEL"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-certificate-ping")
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
	 * Update certificate template
	 */
	time.Sleep(30 * time.Second)
	tcsu, err = t.k8sclient.Resource(templateCentreonServiceGVR).Namespace("default").Get(context.Background(), "check-certificate", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(tcsu, tcs); err != nil {
		t.T().Fatal(err)
	}
	tcs.Spec.Template = `
host: "localhost"
name: "test-certificate-ping"
template: "template-test"
checkCommand: "ping"
macros:
  LABEL: "{{ .labels.foo }}"
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
		Name:                "test-certificate-ping",
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
				Name:   "TEST",
				Value:  "plop",
				Source: "direct",
			},
		},
		Activated: "1",
	}
	time.Sleep(20 * time.Second)

	ucs, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-certificate", v1.GetOptions{})
	if err != nil {
		t.T().Fatal(err)
	}
	if err = unstructuredToStructured(ucs, cs); err != nil {
		t.T().Fatal(err)
	}
	assert.Equal(t.T(), "plop", cs.Spec.Macros["TEST"])

	// Check service updated on Centreon
	s, err = t.centreon.GetService("localhost", "test-certificate-ping")
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
	if err = t.k8sclientStd.CoreV1().Secrets("default").Delete(context.Background(), "test-certificate", *metav1.NewDeleteOptions(0)); err != nil {
		t.T().Fatal(err)
	}
	time.Sleep(20 * time.Second)

	// Check CentreonService delete on k8s
	_, err = t.k8sclient.Resource(centreonServiceGVR).Namespace("default").Get(context.Background(), "check-certificate", v1.GetOptions{})
	if err == nil || !errors.IsNotFound(err) {
		assert.Fail(t.T(), "CentreonService not delete on k8s after delete certificate")
	}

	// Check service is delete from centreon
	s, err = t.centreon.GetService("localhost", "test-certificate-ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), s)
}
