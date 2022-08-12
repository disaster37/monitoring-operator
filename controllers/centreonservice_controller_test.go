package controllers

import (
	"context"
	"testing"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/shared"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/thoas/go-funk"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestCentreonServiceController() {
	key := types.NamespacedName{
		Name:      "t-cs-" + helpers.RandomString(10),
		Namespace: "default",
	}
	cs := &v1alpha1.CentreonService{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, cs, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateCentreonServiceStep(),
		doUpdateCentreonServiceStep(),
		doDeleteCentreonServiceStep(),
		doPolicyNoCreateCentreonServiceStep(),
		doPolicyNoUpdateCentreonServiceStep(),
		doPolicyExcludeFieldsCentreonServiceStep(),
		doPolicyNoDeleteCentreonServiceStep(),
	}
	testCase.PreTest = doMockCentreonService(t.mockCentreonHandler)

	testCase.Run()

}

func doMockCentreonService(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false
		isCreatedPolicyNoCreate := false
		isUpdatedPolicyNoUpdate := false
		isUpdatedPolicyExcludeFields := false

		mockCS.EXPECT().GetService(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(host, name string) (service *centreonhandler.CentreonService, err error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template1",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				}
			case "update":
				if !isUpdated {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template1",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				} else {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template2",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				}
			case "policyNoCreate":
				if !isCreatedPolicyNoCreate {
					return nil, nil
				} else {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template1",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				}
			case "policyNoUpdate":
				if !isUpdatedPolicyNoUpdate {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template1",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				} else {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template2",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				}
			case "policyExcludeFields":
				if !isUpdatedPolicyExcludeFields {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template1",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				} else {
					return &centreonhandler.CentreonService{
						Host:                "central",
						Name:                "ping",
						Template:            "template2",
						NormalCheckInterval: "30s",
						CheckCommand:        "check_ping",
						CheckCommandArgs:    "!arg1",
						RetryCheckInterval:  "5s",
						MaxCheckAttempts:    "3",
						ActiveCheckEnabled:  "1",
						PassiveCheckEnabled: "1",
						Activated:           "1",
						Comment:             "Managed by monitoring-operator",
						Groups:              []string{"sg1"},
						Categories:          []string{"cat1"},
						Macros: []*models.Macro{
							{
								Name:       "MACRO1",
								Value:      "value1",
								IsPassword: "0",
							},
						},
					}, nil
				}
			default:
				return &centreonhandler.CentreonService{
					Host:                "central",
					Name:                "ping",
					Template:            "template1",
					NormalCheckInterval: "30s",
					CheckCommand:        "check_ping",
					CheckCommandArgs:    "!arg1",
					RetryCheckInterval:  "5s",
					MaxCheckAttempts:    "3",
					ActiveCheckEnabled:  "1",
					PassiveCheckEnabled: "1",
					Activated:           "1",
					Comment:             "Managed by monitoring-operator",
					Groups:              []string{"sg1"},
					Categories:          []string{"cat1"},
					Macros: []*models.Macro{
						{
							Name:       "MACRO1",
							Value:      "value1",
							IsPassword: "0",
						},
					},
				}, nil
			}
		})

		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *centreonhandler.CentreonService, ignoreFields []string) (diff *centreonhandler.CentreonServiceDiff, err error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: true,
						Host:   "central",
						Name:   "ping",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: false,
						Host:   "central",
						Name:   "ping",
					}, nil
				}
			case "update":
				if !isUpdated {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: true,
						Host:   "central",
						Name:   "ping",
						ParamsToSet: map[string]string{
							"template": "template2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: false,
						Host:   "central",
						Name:   "ping",
					}, nil
				}
			case "policyNoCreate":
				if !isCreatedPolicyNoCreate {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: true,
						Host:   "central",
						Name:   "ping",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: false,
						Host:   "central",
						Name:   "ping",
					}, nil
				}
			case "policyNoUpdate":
				if !isUpdatedPolicyNoUpdate {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: true,
						Host:   "central",
						Name:   "ping",
						ParamsToSet: map[string]string{
							"template": "template2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: false,
						Host:   "central",
						Name:   "ping",
					}, nil
				}
			case "policyExcludeFields":
				if !funk.Contains(ignoreFields, "template") {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: true,
						Host:   "central",
						Name:   "ping",
						ParamsToSet: map[string]string{
							"template": "template2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceDiff{
						IsDiff: false,
						Host:   "central",
						Name:   "ping",
					}, nil
				}
			}
			return nil, errors.Errorf("Unnatented test: %s", *stepName)
		})

		mockCS.EXPECT().CreateService(gomock.Any()).AnyTimes().DoAndReturn(func(service *centreonhandler.CentreonService) (err error) {
			switch *stepName {
			case "create":
				data["isCreated"] = true
				isCreated = true
			case "policyNoCreate":
				data["isCreatedPolicyNoCreate"] = true
				isCreatedPolicyNoCreate = true
			}

			return nil
		})

		mockCS.EXPECT().UpdateService(gomock.Any()).AnyTimes().DoAndReturn(func(service *centreonhandler.CentreonServiceDiff) (err error) {
			switch *stepName {
			case "update":
				data["isUpdated"] = true
				isUpdated = true
			case "policyNoUpdate":
				data["isUpdatedPolicyNoUpdate"] = true
				isUpdatedPolicyNoUpdate = true
			case "policyExcludeFields":
				data["isUpdatedPolicyExcludeFields"] = true
				isUpdatedPolicyExcludeFields = true
			}

			return nil
		})

		mockCS.EXPECT().DeleteService(gomock.Eq("central"), gomock.Eq("ping")).AnyTimes().DoAndReturn(func(host, service string) (err error) {
			switch *stepName {
			case "delete":
				data["isDeleted"] = true
			case "policyNoDelete":
				data["isDeletedPolicyNoDelete"] = true
			}

			return nil
		})

		return nil
	}
}

func doCreateCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Centreon Service %s/%s ===", key.Namespace, key.Name)

			enabled := true
			cs := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.CentreonServiceSpec{
					Host:                "central",
					Name:                "ping",
					Template:            "template1",
					NormalCheckInterval: "30s",
					CheckCommand:        "check_ping",
					RetryCheckInterval:  "5s",
					MaxCheckAttempts:    "3",
					ActiveCheckEnabled:  &enabled,
					PassiveCheckEnabled: &enabled,
					Activated:           true,
					Groups:              []string{"sg1"},
					Macros: map[string]string{
						"macro1": "value1",
					},
					Arguments:  []string{"arg1"},
					Categories: []string{"cat1"},
				},
			}

			if err = c.Create(context.Background(), cs); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatal("Centreon service not found")
				}
				if b, ok := data["isCreated"]; ok {
					isCreated = b.(bool)
				}
				if !isCreated {
					return errors.New("Not yet created")
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, CentreonServiceCondition, metav1.ConditionTrue))
			assert.Equal(t, "central", cs.Status.Host)
			assert.Equal(t, "ping", cs.Status.ServiceName)
			return nil
		},
	}
}

func doUpdateCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon Service %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon service is null")
			}
			cs := o.(*v1alpha1.CentreonService)
			cs.Spec.Template = "template2"

			if err = c.Update(context.Background(), cs); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdated"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon service: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(cs.Status.Conditions, CentreonServiceCondition, metav1.ConditionTrue))
			assert.Equal(t, "central", cs.Status.Host)
			assert.Equal(t, "ping", cs.Status.ServiceName)
			return nil
		},
	}
}

func doDeleteCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Centreon Service %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon service is null")
			}
			cs := o.(*v1alpha1.CentreonService)

			wait := int64(0)
			if err = c.Delete(context.Background(), cs, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, cs); err != nil {
					if !k8serrors.IsNotFound(err) {
						t.Fatal(err)
					}
				}

				if b, ok := data["isDeleted"]; ok {
					isDeleted = b.(bool)
				}

				if !isDeleted {
					return errors.New("Not yet delete")
				}

				return nil
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Centreon service not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}

func doPolicyNoCreateCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoCreate",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Centreon Service %s/%s (policyNoCreate) ===", key.Namespace, key.Name)

			enabled := true
			cs := &v1alpha1.CentreonService{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.CentreonServiceSpec{
					Host:                "central",
					Name:                "ping",
					Template:            "template1",
					NormalCheckInterval: "30s",
					CheckCommand:        "check_ping",
					RetryCheckInterval:  "5s",
					MaxCheckAttempts:    "3",
					ActiveCheckEnabled:  &enabled,
					PassiveCheckEnabled: &enabled,
					Activated:           true,
					Groups:              []string{"sg1"},
					Macros: map[string]string{
						"macro1": "value1",
					},
					Arguments:  []string{"arg1"},
					Categories: []string{"cat1"},
					Policy: shared.Policy{
						NoCreate:            true,
						NoUpdate:            true,
						NoDelete:            true,
						ExcludeFieldsOnDiff: []string{"template"},
					},
				},
			}

			if err = c.Create(context.Background(), cs); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatal("Centreon service not found")
				}
				if b, ok := data["isCreatedPolicyNoCreate"]; ok {
					isCreated = b.(bool)
				}
				if !isCreated {
					return errors.New("Not yet created")
				}
				return nil
			}, time.Second*10, time.Second*1)
			assert.True(t, isTimeout)
			return nil
		},
	}
}

func doPolicyNoUpdateCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoUpdate",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon Service %s/%s (policyNoUpdate) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon service is null")
			}
			cs := o.(*v1alpha1.CentreonService)
			cs.Spec.Template = "template2"

			if err = c.Update(context.Background(), cs); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdatedPolicyNoUpdate"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*10, time.Second*1)

			assert.True(t, isTimeout)
			return nil
		},
	}
}

func doPolicyExcludeFieldsCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "policyExcludeFields",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon Service %s/%s (policyExcludeFields) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon service is null")
			}
			cs := o.(*v1alpha1.CentreonService)
			cs.Spec.Policy.NoUpdate = false
			cs.Spec.Template = "template2"

			if err = c.Update(context.Background(), cs); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, cs); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdatedPolicyExcludeFields"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*10, time.Second*1)

			assert.True(t, isTimeout)
			return nil
		},
	}
}

func doPolicyNoDeleteCentreonServiceStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoDelete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Centreon Service %s/%s (policyNoDelete) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon service is null")
			}
			cs := o.(*v1alpha1.CentreonService)

			wait := int64(0)
			if err = c.Delete(context.Background(), cs, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			cs := &v1alpha1.CentreonService{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, cs); err != nil {
					if !k8serrors.IsNotFound(err) {
						t.Fatal(err)
					}
				}

				if b, ok := data["isDeletedPolicyNoDelete"]; ok {
					isDeleted = b.(bool)
				}

				if !isDeleted {
					return errors.New("Not yet delete")
				}

				return nil
			}, time.Second*10, time.Second*1)

			assert.True(t, isTimeout)
			return nil
		},
	}
}
