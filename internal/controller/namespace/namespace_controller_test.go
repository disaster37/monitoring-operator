package namespace

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
	core "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *NamespaceControllerTestSuite) TestNamespaceCentreonController() {
	key := types.NamespacedName{
		Name:      "t-namespace-" + helpers.RandomString(10),
		Namespace: "default",
	}
	ns := &core.Namespace{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, ns, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateNamespaceOldStep(),
		doUpdateNamespaceOldStep(),
		doDeleteNamespaceOldStep(),
		doCreateNamespaceStep(),
		doUpdateNamespaceStep(),
		doDeleteNamespaceStep(),
	}

	testCase.Run()
}

func doCreateNamespaceOldStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-namespace1",
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
			logrus.Infof("Create template template-namespace1")

			template = &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-namespace2",
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
			logrus.Infof("Create template template-namespace2")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Namespace %s ===", key.Name)

			// Create namespace that refer 2 templates
			ns := &core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-namespace1\"}, {\"namespace\":\"default\", \"name\": \"template-namespace2\"}]",
					},
				},
			}

			if err = c.Create(context.Background(), ns); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			// Get service generated by template-namespace1
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace1"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-namespace1: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-namespace1: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping1",
				Template: "template1",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
				},
				Arguments:  []string{"arg1", "arg2"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "template-namespace1", cs.Name)
			assert.Equal(t, "default.template-namespace1", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Name, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-namespace2
			isTimeout, err = test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace2"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-namespace2: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-namespace2: %s", err.Error())
			}
			expectedCSSpec = monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping2",
				Template: "template2",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
				},
				Arguments:  []string{"arg1", "arg2"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "default.template-namespace2", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Name, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateNamespaceOldStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {
			logrus.Info("Update template template-namespace1")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-namespace1"}, template); err != nil {
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
			logrus.Infof("=== Update Namespace %s ===", key.Name)

			if o == nil {
				return errors.New("Namespace is null")
			}
			ns := o.(*core.Namespace)

			ns.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &monitorapi.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace1"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), ns); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &monitorapi.CentreonService{}

			version := data["version"].(string)

			time.Sleep(5 * time.Second)

			// Get service generated by template-namespace1
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace1"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-namespace1: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping1",
				Template: "template1",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
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

func doDeleteNamespaceOldStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Namespace %s ===", key.Name)
			if o == nil {
				return errors.New("Namespace is null")
			}
			ns := o.(*core.Namespace)

			wait := int64(0)
			if err = c.Delete(context.Background(), ns, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources
			// Namespace seems can't be deleted in envtest. Block on kubernetes finalizer

			return nil
		},
	}
}

func doCreateNamespaceStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-namespace3",
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
			logrus.Infof("Create template template-namespace3")

			template = &monitorapi.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-namespace4",
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
			logrus.Infof("Create template template-namespace4")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			key.Name = fmt.Sprintf("%s2", key.Name)
			logrus.Infof("=== Add new Namespace %s ===", key.Name)

			// Create namespace that refer 2 templates
			ns := &core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: key.Name,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-namespace3\"}, {\"namespace\":\"default\", \"name\": \"template-namespace4\"}]",
					},
				},
			}

			if err = c.Create(context.Background(), ns); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			key.Name = fmt.Sprintf("%s2", key.Name)
			cs := &monitorapi.CentreonService{}

			// Get service generated by template-namespace1
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace3"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-namespace3: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-namespace3: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping3",
				Template: "template3",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
				},
				Arguments:  []string{"arg1", "arg2"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "template-namespace3", cs.Name)
			assert.Equal(t, "default.template-namespace3", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Name, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-namespace2
			isTimeout, err = test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace4"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-namespace4: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-namespace4: %s", err.Error())
			}
			expectedCSSpec = monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping4",
				Template: "template4",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
				},
				Arguments:  []string{"arg1", "arg2"},
				Activated:  true,
				Groups:     []string{"sg1"},
				Categories: []string{"cat1"},
			}
			assert.Equal(t, "appTest", cs.Labels["app"])
			assert.Equal(t, "default.template-namespace4", cs.Labels["monitor.k8s.webcenter.fr/template"])
			assert.Equal(t, fmt.Sprintf("%s.%s", key.Name, key.Name), cs.Labels["monitor.k8s.webcenter.fr/parent"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateNamespaceStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {
			logrus.Info("Update template template-namespace3")
			template := &monitorapi.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-namespace3"}, template); err != nil {
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
			key.Name = fmt.Sprintf("%s2", key.Name)
			logrus.Infof("=== Update Namespace %s ===", key.Name)

			// Force read o because we overwrite the key.name
			if err = c.Get(context.Background(), types.NamespacedName{Name: key.Name}, o); err != nil {
				panic(err)
			}

			if o == nil {
				return errors.New("Namespace is null")
			}
			ns := o.(*core.Namespace)

			ns.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &monitorapi.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace3"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), ns); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			key.Name = fmt.Sprintf("%s2", key.Name)
			cs := &monitorapi.CentreonService{}

			version := data["version"].(string)

			time.Sleep(5 * time.Second)

			// Get service generated by template-namespace1
			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Name, Name: "template-namespace3"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-namespace3: %s", err.Error())
			}
			expectedCSSpec := monitorapi.CentreonServiceSpec{
				Host:     "localhost",
				Name:     "ping3",
				Template: "template3",
				Macros: map[string]string{
					"name":      key.Name,
					"namespace": key.Name,
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

func doDeleteNamespaceStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			key.Name = fmt.Sprintf("%s2", key.Name)
			logrus.Infof("=== Delete Namespace %s ===", key.Name)
			if o == nil {
				return errors.New("Namespace is null")
			}
			ns := o.(*core.Namespace)

			wait := int64(0)
			if err = c.Delete(context.Background(), ns, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources
			// Namespace seems can't be deleted in envtest. Block on kubernetes finalizer

			return nil
		},
	}
}
