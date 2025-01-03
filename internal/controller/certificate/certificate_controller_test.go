package certificate

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *CertificateControllerTestSuite) TestCertificateCentreonController() {
	key := types.NamespacedName{
		Name:      "t-certificate-" + helpers.RandomString(10),
		Namespace: "default",
	}
	secret := &corev1.Secret{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, secret, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateCertificateOldStep(),
		doUpdateCertificateOldStep(),
		doDeleteCertificateOldStep(),
		doCreateCertificateStep(),
		doUpdateCertificateStep(),
		doDeleteCertificateStep(),
	}

	testCase.Run()
}

func doCreateCertificateOldStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &centreoncrd.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-certificate",
					Namespace: "default",
				},
				Spec: centreoncrd.TemplateSpec{
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

			certificate := &corev1.Secret{
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
				Type: corev1.SecretTypeTLS,
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
			cs := &centreoncrd.CentreonService{}

			// Get service generated by template-certificate
			isTimeout, err := test.RunWithTimeout(func() error {
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
			expectedCSSpec := centreoncrd.CentreonServiceSpec{
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
			assert.Equal(t, "default.template-certificate", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			return nil
		},
	}
}

func doUpdateCertificateOldStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {
			logrus.Info("Update template template-certificate")
			template := &centreoncrd.Template{}
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
			certificate := o.(*corev1.Secret)

			certificate.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &centreoncrd.CentreonService{}
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
			cs := &centreoncrd.CentreonService{}

			version := data["version"].(string)
			time.Sleep(5 * time.Second)

			// Get service generated by template-certificate
			isTimeout, err := test.RunWithTimeout(func() error {
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
			expectedCSSpec := centreoncrd.CentreonServiceSpec{
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

func doDeleteCertificateOldStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Certificate %s/%s ===", key.Namespace, key.Name)
			if o == nil {
				return errors.New("Certiifcateis null")
			}
			certificate := o.(*corev1.Secret)

			wait := int64(0)
			if err = c.Delete(context.Background(), certificate, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			certificate := &corev1.Secret{}
			isDeleted := false

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			// Object can be deleted or marked as deleted
			isTimeout, err := test.RunWithTimeout(func() error {
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

func doCreateCertificateStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &centreoncrd.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-certificate2",
					Namespace: "default",
				},
				Spec: centreoncrd.TemplateSpec{
					Template: `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
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
			logrus.Infof("Create template template-certificate2")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Certificate %s/%s ===", key.Namespace, key.Name)

			certificate := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-certificate2\"}]",
					},
				},
				Type: corev1.SecretTypeTLS,
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
			cs := &centreoncrd.CentreonService{}

			// Get service generated by template-certificate
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate2"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-certificate2: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-certificate2: %s", err.Error())
			}
			expectedCSSpec := centreoncrd.CentreonServiceSpec{
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
			assert.Equal(t, "template-certificate2", cs.Name)
			assert.Equal(t, "default.template-certificate2", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
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
			logrus.Info("Update template template-certificate2")
			template := &centreoncrd.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-certificate2"}, template); err != nil {
				return err
			}

			template.Spec.Template = `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
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
			certificate := o.(*corev1.Secret)

			certificate.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &centreoncrd.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate2"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), certificate); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &centreoncrd.CentreonService{}

			version := data["version"].(string)
			time.Sleep(5 * time.Second)

			// Get service generated by template-certificate
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-certificate2"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-certificate2: %s", err.Error())
			}
			expectedCSSpec := centreoncrd.CentreonServiceSpec{
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
			certificate := o.(*corev1.Secret)

			wait := int64(0)
			if err = c.Delete(context.Background(), certificate, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			certificate := &corev1.Secret{}
			isDeleted := false

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			// Object can be deleted or marked as deleted
			isTimeout, err := test.RunWithTimeout(func() error {
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
