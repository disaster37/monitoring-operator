package controllers

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestCertificateCentreonController() {
	key := types.NamespacedName{
		Name:      "t-certificate-" + helpers.RandomString(10),
		Namespace: "default",
	}
	secret := &core.Secret{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, secret, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateCertificateStep(),
		doUpdateCertificateStep(),
		doDeleteCertificateStep(),
	}
	testCase.PreTest = doMockCertificate(t.mockCentreonHandler)

	os.Setenv("OPERATOR_NAMESPACE", "default")

	testCase.Run()

}

func doMockCertificate(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {

		mockCS.EXPECT().GetService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&centreonhandler.CentreonServiceDiff{
			IsDiff: false,
		}, nil)

		mockCS.EXPECT().CreateService(gomock.Any()).AnyTimes().Return(nil)

		mockCS.EXPECT().UpdateService(gomock.Any()).AnyTimes().Return(nil)

		mockCS.EXPECT().DeleteService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		return nil
	}
}

func doCreateCertificateStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-certificate",
					Namespace: "default",
				},
				Spec: monitorapi.TemplateSpec{
					Type: "CentreonService",
					Template: `
host: "localhost"
name: "ping1"
template: "template1"
macros:
  name: "{{ .name }}"
  namespace: "{{ .namespace }}"
arguments:
  - "arg1"
  - "arg2"
activate: true
groups:
  - "sg1"
categories:
  - "cat1"`,
				},
			}
			if err := c.Create(context.Background(), template); err != nil {
				return err
			}
			logrus.Infof("Create template template-certificate")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Certificate %s/%s ===", key.Namespace, key.Name)

			certificate := &core.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-certificate\"}]",
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
-----END CERTIFICATE-----`),
				},
			}

			if err = c.Create(context.Background(), certificate); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			// Get service generated by template-certificate
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-certificate: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-certificate: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping1",
				Template: "template1",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Namespace,
				},
				Arguments:  []string{"arg1", "arg2"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "template-certificate", cs.Name)
			assert.Equal(t, "template-certificate", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-certificate\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			return nil
		},
	}
}

func doUpdateCertificateStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {

			logrus.Info("Update template template-certificate")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-certificate"}, template); err != nil {
				return err
			}

			template.Spec.Template = `
host: "localhost"
name: "ping1"
template: "template1"
macros:
  name: "{{ .name }}"
  namespace: "{{ .namespace }}"
arguments:
  - "arg11"
  - "arg21"
activate: true
groups:
  - "sg1"
categories:
  - "cat1"`
			if err := c.Update(context.Background(), template); err != nil {
				return err
			}

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Certificate %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Certificate is null")
			}
			certificate := o.(*core.Secret)

			certificate.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &monitorapi.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), certificate); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			version := data["version"].(string)

			// Get service generated by template-certificate
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						t.Fatalf("Error when get Centreon service: %s", err.Error())
					}
					if cs.ResourceVersion == version {
						return errors.New("Not yet updated")
					}
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-certificate: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping1",
				Template: "template1",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Namespace,
				},
				Arguments:  []string{"arg11", "arg21"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, expectedCSSpec, cs.Spec)

			return nil
		},
	}
}

func doDeleteCertificateStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Certificate %s/%s ===", key.Namespace, key.Name)
			if o == nil {
				return errors.New("Certiifcateis null")
			}
			certificate := o.(*core.Secret)

			wait := int64(0)
			if err = c.Delete(context.Background(), certificate, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			certificate := &core.Secret{}
			isDeleted := false

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			// Object can be deleted or marked as deleted
			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, certificate); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return nil

			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Certificate not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}

func TestGeneratePlaceholdersCertificate(t *testing.T) {

	var (
		certificate *core.Secret
		ph          map[string]any
		expectedPh  map[string]any
	)

	// When certificate is nil
	ph, err := generatePlaceholdersCertificate(nil)
	assert.NoError(t, err)
	assert.Empty(t, ph)

	// When all properties
	certificate = &core.Secret{
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
-----END CERTIFICATE-----`),
		},
	}

	expectedPh = map[string]any{
		"name":      "test",
		"namespace": "default",
		"labels": map[string]string{
			"app": "appTest",
			"env": "dev",
		},
		"annotations": map[string]string{
			"anno1": "value1",
			"anno2": "value2",
		},
	}

	ph, err = generatePlaceholdersCertificate(certificate)
	assert.NoError(t, err)
	assert.NotEmpty(t, ph["certificates"])
	delete(ph, "certificates")
	assert.Equal(t, expectedPh, ph)

}
