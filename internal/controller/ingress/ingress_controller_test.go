package ingress

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	networkv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *IngressControllerTestSuite) TestIngressCentreonController() {
	key := types.NamespacedName{
		Name:      "t-ingress-" + helpers.RandomString(10),
		Namespace: "default",
	}
	ingress := &networkv1.Ingress{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, ingress, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateIngressOldStep(),
		doUpdateIngressOldStep(),
		doDeleteIngressOldStep(),
		doCreateIngressStep(),
		doUpdateIngressStep(),
		doDeleteIngressStep(),
	}

	testCase.Run()
}

func doCreateIngressOldStep() test.TestStep {
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
			isTimeout, err := test.RunWithTimeout(func() error {
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
			assert.Equal(t, "default.template-ingress1", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-ingress2
			isTimeout, err = test.RunWithTimeout(func() error {
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
			assert.Equal(t, "default.template-ingress2", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateIngressOldStep() test.TestStep {
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

			time.Sleep(5 * time.Second)

			// Get service generated by template-ingress1
			isTimeout, err := test.RunWithTimeout(func() error {
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
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			return nil
		},
	}
}

func doDeleteIngressOldStep() test.TestStep {
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
			isTimeout, err := test.RunWithTimeout(func() error {
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

func doCreateIngressStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-ingress3",
					Namespace: "default",
				},
				Spec: monitorapi.TemplateSpec{
					Template: `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
  host: "localhost"
  name: "ping3"
  template: "template3"
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
			logrus.Infof("Create template template-ingress3")

			template = &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-ingress4",
					Namespace: "default",
				},
				Spec: monitorapi.TemplateSpec{
					Template: `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
  host: "localhost"
  name: "ping4"
  template: "template4"
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
			logrus.Infof("Create template template-ingress4")

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
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-ingress3\"}, {\"namespace\":\"default\", \"name\": \"template-ingress4\"}]",
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

			// Get service generated by template-ingress3
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress3"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-ingress3: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-ingress3: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping3",
				Template: "template3",
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
			assert.Equal(t, "template-ingress3", cs.Name)
			assert.Equal(t, "default.template-ingress3", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-ingress2
			isTimeout, err = test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress4"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-ingress4: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-ingress4: %s", err.Error())
			}
			expectedCSSpec = monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping4",
				Template: "template4",
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
			assert.Equal(t, "default.template-ingress4", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Namespace, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
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
			logrus.Info("Update CentreonServiceTemplate template-ingress3")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-ingress3"}, template); err != nil {
				return err
			}

			template.Spec.Template = `
apiVersion: monitor.k8s.webcenter.fr/v1
kind: CentreonService
spec:
  host: "localhost"
  name: "ping3"
  template: "template3"
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
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress3"}, cs); err != nil {
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

			time.Sleep(5 * time.Second)

			// Get service generated by template-ingress1
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-ingress3"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-ingress3: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping3",
				Template: "template3",
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
			isTimeout, err := test.RunWithTimeout(func() error {
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
