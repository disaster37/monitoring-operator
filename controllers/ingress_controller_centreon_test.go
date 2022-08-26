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
	networkv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestIngressCentreonController() {
	key := types.NamespacedName{
		Name:      "t-ingress-" + helpers.RandomString(10),
		Namespace: "default",
	}
	ingress := &networkv1.Ingress{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, ingress, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateIngressStep(),
		doUpdateIngressStep(),
		doDeleteIngressStep(),
	}
	testCase.PreTest = doMockIngress(t.mockCentreonHandler)

	os.Setenv("OPERATOR_NAMESPACE", "default")

	testCase.Run()

}

func doMockIngress(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
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

func doCreateIngressStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-ingress1",
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
			logrus.Infof("Create template template-ingress1")

			template = &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-ingress2",
					Namespace: "default",
				},
				Spec: monitorapi.TemplateSpec{
					Type: "CentreonService",
					Template: `
host: "localhost"
name: "ping2"
template: "template2"
macros:
  name: "{{ .name }}"
  namespace: "{{ .namespace }}"
arguments:
  - "arg1"
  - "arg2"
activate:  true
groups:
  - "sg1"
categories:
  - "cat1"`,
				},
			}
			if err := c.Create(context.Background(), template); err != nil {
				return err
			}
			logrus.Infof("Create template template-ingress2")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Ingress %s/%s ===", key.Namespace, key.Name)

			// Create ingress without annotations
			pathType := networkv1.PathTypePrefix
			ingress := &networkv1.Ingress{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-ingress1\"}, {\"namespace\":\"default\", \"name\": \"template-ingress2\"}]",
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

			if err = c.Create(context.Background(), ingress); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			// Get service generated by template-ingress1
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress1"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-ingress1: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-ingress1: %s", err.Error())
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
			assert.Equal(t, "template-ingress1", cs.Name)
			assert.Equal(t, "template-ingress1", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-ingress1\"}, {\"namespace\":\"default\", \"name\": \"template-ingress2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-ingress2
			isTimeout, err = RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress2"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-ingress2: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-ingress2: %s", err.Error())
			}
			expectedCSSpec = monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping2",
				Template: "template2",
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
			assert.Equal(t, "template-ingress2", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-ingress1\"}, {\"namespace\":\"default\", \"name\": \"template-ingress2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateIngressStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {

			logrus.Info("Update CentreonServiceTemplate template-ingress1")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-ingress1"}, template); err != nil {
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
			logrus.Infof("=== Update Ingress %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Ingress is null")
			}
			ingress := o.(*networkv1.Ingress)

			ingress.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &monitorapi.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress1"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), ingress); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			version := data["version"].(string)

			// Get service generated by template-ingress1
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress1"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-ingress1: %s", err.Error())
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
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-ingress1\"}, {\"namespace\":\"default\", \"name\": \"template-ingress2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			return nil
		},
	}
}

func doDeleteIngressStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Ingress %s/%s ===", key.Namespace, key.Name)
			if o == nil {
				return errors.New("Ingress is null")
			}
			ingress := o.(*networkv1.Ingress)

			wait := int64(0)
			if err = c.Delete(context.Background(), ingress, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			ingress := &networkv1.Ingress{}
			isDeleted := false

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			// Object can be deleted or marked as deleted
			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, ingress); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return nil

			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Ingress not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}

func TestGeneratePlaceholdersIngress(t *testing.T) {

	var (
		ingress    *networkv1.Ingress
		ph         map[string]any
		expectedPh map[string]any
	)

	// When ingress is nil
	ph = generatePlaceholdersIngress(nil)
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
		"rules": []map[string]any{
			{
				"host":   "front.local.local",
				"scheme": "http",
				"paths": []string{
					"/",
					"/api",
				},
			},
			{
				"host":   "back.local.local",
				"scheme": "https",
				"paths": []string{
					"/",
				},
			},
		},
	}

	ph = generatePlaceholdersIngress(ingress)
	assert.Equal(t, expectedPh, ph)

}
