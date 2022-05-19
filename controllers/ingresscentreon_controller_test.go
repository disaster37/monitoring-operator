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

		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any()).AnyTimes().Return(&centreonhandler.CentreonServiceDiff{
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
		Pre: func(c client.Client, data map[string]any) (err error) {

			// Delete CentreonSpec if already exist
			// Maybee it created by others tests
			centreon := &v1alpha1.Centreon{}
			if err = c.Get(context.Background(), types.NamespacedName{Namespace: "default", Name: CentreonResourceName}, centreon); err != nil {
				if !k8serrors.IsNotFound(err) {
					return err
				}
				centreon = nil
			}

			if centreon != nil {
				wait := int64(0)
				if err = c.Delete(context.Background(), centreon, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
					return err
				}
			}

			return nil
		},
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Ingress %s/%s ===", key.Namespace, key.Name)

			//Create Centreon
			centreon := &v1alpha1.Centreon{
				ObjectMeta: metav1.ObjectMeta{
					Name:      CentreonResourceName,
					Namespace: "default",
				},
				Spec: v1alpha1.CentreonSpec{
					Endpoints: &v1alpha1.CentreonSpecEndpoint{
						Template:     "template",
						DefaultHost:  "localhost",
						NameTemplate: "ping",
						Macros: map[string]string{
							"mac1": "value1",
							"mac2": "value2",
						},
						Arguments:       []string{"arg1", "arg2"},
						ActivateService: true,
						ServiceGroups:   []string{"sg1"},
						Categories:      []string{"cat1"},
					},
				},
			}
			if err = c.Create(context.Background(), centreon); err != nil {
				return err
			}

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
						"monitor.k8s.webcenter.fr/discover": "true",
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
			cs := &v1alpha1.CentreonService{}

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					if k8serrors.IsNotFound(err) {
						return errors.New("Not yet created")
					}
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service: %s", err.Error())
			}

			expectedCS := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/discover": "true",
					},
				},
				Spec: v1alpha1.CentreonServiceSpec{
					Host:     "localhost",
					Name:     "ping",
					Template: "template",
					Macros: map[string]string{
						"mac1": "value1",
						"mac2": "value2",
					},
					Arguments:  []string{"arg1", "arg2"},
					Activated:  true,
					Groups:     []string{"sg1"},
					Categories: []string{"cat1"},
				},
			}

			assert.Equal(t, expectedCS.Labels, cs.Labels)
			assert.Equal(t, expectedCS.Annotations, cs.Annotations)
			assert.Equal(t, expectedCS.Spec, cs.Spec)
			assert.NotEmpty(t, cs.OwnerReferences)
			return nil
		},
	}
}

func doUpdateIngressStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Ingress %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Ingress is null")
			}
			ingress := o.(*networkv1.Ingress)

			ingress.Annotations = map[string]string{
				"monitor.k8s.webcenter.fr/discover":                       "true",
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template2",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value11", "mac2": "value22"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg11, arg22",
				"centreon.monitor.k8s.webcenter.fr/activated":             "0",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg2",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat2",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "1",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "1",
			}

			// Get version of current CentreonService object
			cs := &v1alpha1.CentreonService{}
			if err := c.Get(context.Background(), key, cs); err != nil {
				return err
			}

			data["version"] = cs.ResourceVersion

			if err = c.Update(context.Background(), ingress); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}

			activate := true

			version := data["version"].(string)

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatalf("Error when get Centreon service: %s", err.Error())
				}

				if cs.ResourceVersion == version {
					return errors.New("Not yet updated")
				}

				return nil

			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service: %s", err.Error())
			}

			expectedCS := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
					Labels: map[string]string{
						"app": "appTest",
						"env": "dev",
					},
					Annotations: map[string]string{
						"monitor.k8s.webcenter.fr/discover":                       "true",
						"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
						"centreon.monitor.k8s.webcenter.fr/template":              "template2",
						"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
						"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value11", "mac2": "value22"}`,
						"centreon.monitor.k8s.webcenter.fr/arguments":             "arg11, arg22",
						"centreon.monitor.k8s.webcenter.fr/activated":             "0",
						"centreon.monitor.k8s.webcenter.fr/groups":                "sg2",
						"centreon.monitor.k8s.webcenter.fr/categories":            "cat2",
						"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
						"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
						"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
						"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "1",
						"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "1",
					},
				},
				Spec: v1alpha1.CentreonServiceSpec{
					Host:     "localhost",
					Name:     "ping",
					Template: "template2",
					Macros: map[string]string{
						"mac1": "value11",
						"mac2": "value22",
					},
					Arguments:           []string{"arg11", "arg22"},
					Activated:           false,
					Groups:              []string{"sg2"},
					Categories:          []string{"cat2"},
					NormalCheckInterval: "30",
					RetryCheckInterval:  "10",
					MaxCheckAttempts:    "5",
					ActiveCheckEnabled:  &activate,
					PassiveCheckEnabled: &activate,
				},
			}

			assert.Equal(t, expectedCS.Labels, cs.Labels)
			assert.Equal(t, expectedCS.Annotations, cs.Annotations)
			assert.Equal(t, expectedCS.Spec, cs.Spec)
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

func TestGeneratePlaceholdersIngressCentreonService(t *testing.T) {

	var (
		ingress    *networkv1.Ingress
		ph         map[string]string
		expectedPh map[string]string
	)

	// When ingress is nil
	ph = generatePlaceholdersIngressCentreonService(nil)
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

	expectedPh = map[string]string{
		"name":             "test",
		"namespace":        "default",
		"rule.0.host":      "front.local.local",
		"rule.0.scheme":    "http",
		"rule.0.path.0":    "/",
		"rule.0.path.1":    "/api",
		"rule.1.host":      "back.local.local",
		"rule.1.scheme":    "https",
		"rule.1.path.0":    "/",
		"label.app":        "appTest",
		"label.env":        "dev",
		"annotation.anno1": "value1",
		"annotation.anno2": "value2",
	}

	ph = generatePlaceholdersIngressCentreonService(ingress)
	assert.Equal(t, expectedPh, ph)

}

/*

func TestCentreonServiceFromIngress(t *testing.T) {

	var (
		ingress      *networkv1.Ingress
		cs           *v1alpha1.CentreonService
		expectedCs   *v1alpha1.CentreonService
		centreonSpec *v1alpha1.CentreonSpec
		err          error
	)

	// When ingress is nil
	_, err = centreonServiceFromIngress(nil, nil, nil)
	assert.Error(t, err)

	// When no centreonSpec and not all annotations
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
	_, err = centreonServiceFromIngress(ingress, nil, runtime.NewScheme())
	assert.Error(t, err)

	// When no centreonSpec and all annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
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
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromIngress(ingress, nil, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec and all annotations, priority to annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
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
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
			Template:        "template",
			NameTemplate:    "name",
			DefaultHost:     "localhost",
			ActivateService: true,
			Arguments:       []string{"arg1"},
			ServiceGroups:   []string{"sg1"},
			Categories:      []string{"cat1"},
			Macros: map[string]string{
				"MACRO1": "value1",
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
			Annotations: map[string]string{
				"centreon.monitor.k8s.webcenter.fr/name":                  "ping",
				"centreon.monitor.k8s.webcenter.fr/template":              "template",
				"centreon.monitor.k8s.webcenter.fr/host":                  "localhost",
				"centreon.monitor.k8s.webcenter.fr/macros":                `{"mac1": "value1", "mac2": "value2"}`,
				"centreon.monitor.k8s.webcenter.fr/arguments":             "arg1, arg2",
				"centreon.monitor.k8s.webcenter.fr/activated":             "1",
				"centreon.monitor.k8s.webcenter.fr/groups":                "sg1",
				"centreon.monitor.k8s.webcenter.fr/categories":            "cat1",
				"centreon.monitor.k8s.webcenter.fr/normal-check-interval": "30",
				"centreon.monitor.k8s.webcenter.fr/retry-check-interval":  "10",
				"centreon.monitor.k8s.webcenter.fr/max-check-attempts":    "5",
				"centreon.monitor.k8s.webcenter.fr/active-check-enabled":  "2",
				"centreon.monitor.k8s.webcenter.fr/passive-check-enabled": "2",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
			Macros: map[string]string{
				"mac1": "value1",
				"mac2": "value2",
			},
			Arguments:           []string{"arg1", "arg2"},
			Activated:           true,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1"},
			NormalCheckInterval: "30",
			RetryCheckInterval:  "10",
			MaxCheckAttempts:    "5",
		},
	}
	cs, err = centreonServiceFromIngress(ingress, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

	// When centreonSpec without annotations
	ingress = &networkv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
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
	centreonSpec = &v1alpha1.CentreonSpec{
		Endpoints: &v1alpha1.CentreonSpecEndpoint{
			Template:        "template",
			NameTemplate:    "name-<label.app>-<label.env>-<namespace>",
			DefaultHost:     "localhost",
			ActivateService: true,
			ServiceGroups:   []string{"sg1"},
			Macros: map[string]string{
				"SCHEME": "<rule.0.scheme>",
				"URL":    "<rule.0.host><rule.0.path.0>",
			},
		},
	}
	expectedCs = &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "default",
			Labels: map[string]string{
				"app": "appTest",
				"env": "dev",
			},
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Host:     "localhost",
			Name:     "name-appTest-dev-default",
			Template: "template",
			Macros: map[string]string{
				"SCHEME": "http",
				"URL":    "front.local.local/",
			},
			Activated: true,
			Groups:    []string{"sg1"},
		},
	}
	cs, err = centreonServiceFromIngress(ingress, centreonSpec, runtime.NewScheme())
	assert.NoError(t, err)
	assert.Equal(t, expectedCs, cs)

}

*/
