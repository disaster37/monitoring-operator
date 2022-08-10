package controllers

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/golang/mock/gomock"
	routev1 "github.com/openshift/api/route/v1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestRouteCentreonController() {
	key := types.NamespacedName{
		Name:      "t-route-" + helpers.RandomString(10),
		Namespace: "default",
	}
	route := &routev1.Route{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, route, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateRouteStep(),
		doUpdateRouteStep(),
		doDeleteRouteStep(),
	}
	testCase.PreTest = doMockRoute(t.mockCentreonHandler)

	os.Setenv("OPERATOR_NAMESPACE", "default")

	testCase.Run()
}

func doMockRoute(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {

		mockCS.EXPECT().GetService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil, nil)

		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any()).AnyTimes().Return(&centreonhandler.CentreonServiceDiff{
			IsDiff: false,
		}, nil)

		mockCS.EXPECT().CreateService(gomock.Any()).AnyTimes().Return(nil)

		mockCS.EXPECT().UpdateService(gomock.Any()).AnyTimes().Return(nil)

		mockCS.EXPECT().DeleteService(gomock.Any(), gomock.Any()).AnyTimes().Return(nil)

		return nil
	}
}

func doCreateRouteStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Pre: func(c client.Client, data map[string]any) error {
			template := &v1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-route1",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
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
			logrus.Infof("Create template template-route1")

			template = &v1alpha1.Template{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "template-route2",
					Namespace: "default",
				},
				Spec: v1alpha1.TemplateSpec{
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
			logrus.Infof("Create template template-route2")

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Route %s/%s ===", key.Namespace, key.Name)

			// Create route without annotations
			route := &routev1.Route{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/templates": "[{\"namespace\":\"default\", \"name\": \"template-route1\"}, {\"namespace\":\"default\", \"name\": \"template-route2\"}]",
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

			if err = c.Create(context.Background(), route); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-route1"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-route1: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-route1: %s", err.Error())
			}

			expectedCSSpec := v1alpha1.CentreonServiceSpec{
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
			assert.Equal(t, "template-route1", cs.Name)
			assert.Equal(t, "template-route1", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-route1\"}, {\"namespace\":\"default\", \"name\": \"template-route2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			// Get service generated by template-ingress2
			isTimeout, err = RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-route2"}, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service template-route2: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service template-route2: %s", err.Error())
			}
			expectedCSSpec = v1alpha1.CentreonServiceSpec{
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
			assert.Equal(t, "template-route2", cs.Labels["monitor.k8s.webcenter.fr/template-name"])
			assert.Equal(t, "default", cs.Labels["monitor.k8s.webcenter.fr/template-namespace"])
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-route1\"}, {\"namespace\":\"default\", \"name\": \"template-route2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateRouteStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Pre: func(c client.Client, data map[string]any) error {

			logrus.Info("Update CentreonServiceTemplate template-route1")
			template := &v1alpha1.Template{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: "template-route1"}, template); err != nil {
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
			logrus.Infof("=== Update Route %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Route is null")
			}
			route := o.(*routev1.Route)

			route.Annotations["test"] = "update"

			// Get version of current CentreonService object
			cs := &v1alpha1.CentreonService{}
			if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-route1"}, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), route); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}

			version := data["version"].(string)

			// Get service generated by template-ingress1
			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), types.NamespacedName{Namespace: key.Namespace, Name: "template-route1"}, cs); err != nil {
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
				t.Fatalf("Failed to get Centreon service template-route1: %s", err.Error())
			}
			expectedCSSpec := v1alpha1.CentreonServiceSpec{
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
			assert.Equal(t, "[{\"namespace\":\"default\", \"name\": \"template-route1\"}, {\"namespace\":\"default\", \"name\": \"template-route2\"}]", cs.Annotations["monitor.k8s.webcenter.fr/templates"])
			assert.Equal(t, expectedCSSpec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)

			return nil
		},
	}
}

func doDeleteRouteStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Route %s/%s ===", key.Namespace, key.Name)
			if o == nil {
				return errors.New("Route is null")
			}
			route := o.(*routev1.Route)

			wait := int64(0)
			if err = c.Delete(context.Background(), route, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			route := &routev1.Route{}
			isDeleted := false

			// We can't test in envtest that the children is deleted
			// https://stackoverflow.com/questions/64821970/operator-controller-could-not-delete-correlated-resources

			// Object can be deleted or marked as deleted
			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, route); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return nil

			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Route not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)

			return nil
		},
	}
}

func TestGeneratePlaceholdersRouteCentreonService(t *testing.T) {
	var (
		route      *routev1.Route
		ph         map[string]any
		expectedPh map[string]any
	)

	// When route is nil
	ph = generatePlaceholdersRoute(nil)
	assert.Empty(t, ph)

	// When all no path
	route = &routev1.Route{
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
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
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
				},
			},
		},
	}

	ph = generatePlaceholdersRoute(route)
	assert.Equal(t, expectedPh, ph)

	// When all properties with tls
	route = &routev1.Route{
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
		Spec: routev1.RouteSpec{
			Host: "front.local.local",
			Path: "/",
			TLS: &routev1.TLSConfig{
				Termination: routev1.TLSTerminationEdge,
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
				"scheme": "https",
				"paths": []string{
					"/",
				},
			},
		},
	}

	ph = generatePlaceholdersRoute(route)
	assert.Equal(t, expectedPh, ph)

}
