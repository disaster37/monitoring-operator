package v1

import (
	"context"
	"fmt"

	routev1 "github.com/openshift/api/route/v1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	networkv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

func (t *APITestSuite) TestSetupCertificateIndexer() {
	secret := &corev1.Secret{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Annotations: map[string]string{
				fmt.Sprintf("%s/templates", MonitoringAnnotationKey): `[{"namespace": "default", "name": "template1"}, {"namespace": "default", "name": "template2"}]`,
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
	err := t.k8sClient.Create(context.Background(), secret)
	assert.NoError(t.T(), err)
}

func (t *APITestSuite) TestSetupIngressIndexer() {
	ingress := &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Annotations: map[string]string{
				fmt.Sprintf("%s/templates", MonitoringAnnotationKey): `[{"namespace": "default", "name": "template1"}, {"namespace": "default", "name": "template2"}]`,
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
									PathType: ptr.To(networkv1.PathTypePrefix),
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
		},
	}

	err := t.k8sClient.Create(context.Background(), ingress)
	assert.NoError(t.T(), err)

}

func (t *APITestSuite) TestSetupNamespaceIndexer() {
	namespace := &corev1.Namespace{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				fmt.Sprintf("%s/templates", MonitoringAnnotationKey): `[{"namespace": "default", "name": "template1"}, {"namespace": "default", "name": "template2"}]`,
			},
		},
	}

	err := t.k8sClient.Create(context.Background(), namespace)
	assert.NoError(t.T(), err)

}

func (t *APITestSuite) TestSetupNodeIndexer() {
	node := &corev1.Node{
		ObjectMeta: v1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				fmt.Sprintf("%s/templates", MonitoringAnnotationKey): `[{"namespace": "default", "name": "template1"}, {"namespace": "default", "name": "template2"}]`,
			},
		},
	}

	err := t.k8sClient.Create(context.Background(), node)
	assert.NoError(t.T(), err)

}

func (t *APITestSuite) TestSetupRouteIndexer() {
	route := &routev1.Route{
		ObjectMeta: v1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Annotations: map[string]string{
				fmt.Sprintf("%s/templates", MonitoringAnnotationKey): `[{"namespace": "default", "name": "template1"}, {"namespace": "default", "name": "template2"}]`,
			},
		},
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			To: routev1.RouteTargetReference{
				Kind: "Service",
				Name: "fake",
			},
			Port: &routev1.RoutePort{
				TargetPort: intstr.FromString("8080"),
			},
		},
	}

	err := t.k8sClient.Create(context.Background(), route)
	assert.NoError(t.T(), err)

}
