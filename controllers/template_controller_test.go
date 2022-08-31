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

func (t *ControllerTestSuite) TestTemplateController() {
	key := types.NamespacedName{
		Name:      "t-template-" + helpers.RandomString(10),
		Namespace: "default",
	}
	ingress := &networkv1.Ingress{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, ingress, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateTemplateStep(),
		doUpdateTemplateStep(),
		doDeleteTemplateStep(),
	}
	testCase.PreTest = doMockTemplate(t.mockCentreonHandler)

	os.Setenv("OPERATOR_NAMESPACE", "default")

	testCase.Run()

}

func doMockTemplate(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {

		// Service
		mockCS.EXPECT().GetService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)
		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&centreonhandler.CentreonServiceDiff{
			IsDiff: false,
		}, nil)
		mockCS.EXPECT().CreateService(gomock.Any()).AnyTimes().Return(nil)
		mockCS.EXPECT().UpdateService(gomock.Any()).AnyTimes().Return(nil)
		mockCS.EXPECT().DeleteService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		// ServiceGroup
		mockCS.EXPECT().GetServiceGroup(gomock.Any()).AnyTimes().Return(nil, nil)
		mockCS.EXPECT().DiffServiceGroup(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&centreonhandler.CentreonServiceGroupDiff{
			IsDiff: false,
		}, nil)
		mockCS.EXPECT().CreateServiceGroup(gomock.Any()).AnyTimes().Return(nil)
		mockCS.EXPECT().UpdateServiceGroup(gomock.Any()).AnyTimes().Return(nil)
		mockCS.EXPECT().DeleteServiceGroup(gomock.Any()).AnyTimes().Return(nil)

		return nil
	}
}

func doCreateTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-service",
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
			logrus.Infof("Create template template-service")

			template = &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-sg",
					Namespace: "default",
				},
				Spec: monitorapi.TemplateSpec{
					Type: "CentreonServiceGroup",
					Template: `
name: "sg-test"
description: "my sg {{ .namespace }}"
activate:  true`,
				},
			}
			if err := c.Create(context.Background(), template); err != nil {
				return err
			}
			logrus.Infof("Create template template-sg")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Ingress from template test %s/%s ===", key.Namespace, key.Name)

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
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-service\"}, {\"namespace\":\"default\", \"name\": \"template-sg\"}]",
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
			csg := &monitorapi.CentreonServiceGroup{}

			// Get service generated by template-service
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-service"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-service: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-service: %s", err.Error())
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
			assert.Equal(t, "template-service", cs.Name)
			assert.Equal(t, "template-service", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-service\"}, {\"namespace\":\"default\", \"name\": \"template-sg\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get serviceGroup generated by template-sg
			isTimeout, err = RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-sg"}, csg); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon serviceGroup template-sg: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon serviceGroup template-sg: %s", err.Error())
			}
			expectedCSGSpec := monitorapi.CentreonServiceGroupSpec{
				Name:        "sg-test",
				Description: "my sg default",
				Activated:   true,
			}
			assert.Equal(t, "appTest", csg.Labels["app"])
			assert.Equal(t, "template-sg", csg.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", csg.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-service\"}, {\"namespace\":\"default\", \"name\": \"template-sg\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSGSpec, csg.Spec)
			assert.NotEmpty(t, csg.OwnerReferences)
			return nil
		},
	}
}

func doUpdateTemplateStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {

			// Get version of current CentreonService object
			cs := &monitorapi.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-service"}, cs); err != nil {
				return err
			}
			data["csVersion"] = cs.ResourceVersion

			// Get version of current CentreonService object
			csg := &monitorapi.CentreonServiceGroup{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-sg"}, csg); err != nil {
				return err
			}
			data["csgVersion"] = cs.ResourceVersion

			logrus.Info("Update template template-service")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-service"}, template); err != nil {
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

			logrus.Info("Update template template-sg")
			template = &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-sg"}, template); err != nil {
				return err
			}

			template.Spec.Template = `
name: "sg-test"
description: "my sg bis"
activate:  true`
			if err := c.Update(context.Background(), template); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}
			csg := &monitorapi.CentreonServiceGroup{}

			csVersion := data["csVersion"].(string)
			csgVersion := data["csgVersion"].(string)

			// Get service generated by template-service
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-service"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						t.Fatalf("Error when get Centreon service: %s", err.Error())
					}
					if cs.ResourceVersion == csVersion {
						return errors.New("Not yet updated")
					}
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-service: %s", err.Error())
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

			// Get service generated by template-sg
			isTimeout, err = RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-sg"}, csg); err != nil {
					if k8serrors.IsNotFound(err) {
						t.Fatalf("Error when get Centreon serviceGroup: %s", err.Error())
					}
					if csg.ResourceVersion == csgVersion {
						return errors.New("Not yet updated")
					}
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon serviceGroup template-sg: %s", err.Error())
			}
			expectedCSGSpec := monitorapi.CentreonServiceGroupSpec{
				Name:        "sg-test",
				Description: "my sg bis",
				Activated:   true,
			}
			assert.Equal(t, expectedCSGSpec, csg.Spec)

			return nil
		},
	}
}

func doDeleteTemplateStep() test.TestStep {
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

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			return nil
		},
	}
}
