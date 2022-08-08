package controllers

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/test"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
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
	}
	testCase.PreTest = doMockCentreonService(t.mockCentreonHandler)

	testCase.Run()

}

func doMockCentreonService(mockCS *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		serviceToCreate := &centreonhandler.CentreonService{
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
		}

		serviceToUpdate := &centreonhandler.CentreonServiceDiff{
			IsDiff: true,
			Host:   "central",
			Name:   "ping",
			ParamsToSet: map[string]string{
				"template": "template2",
			},
		}

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
			}
			return nil, nil
		})

		mockCS.EXPECT().DiffService(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *centreonhandler.CentreonService) (diff *centreonhandler.CentreonServiceDiff, err error) {
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
			}
			return nil, nil
		})

		mockCS.EXPECT().CreateService(gomock.Eq(serviceToCreate)).AnyTimes().DoAndReturn(func(service *centreonhandler.CentreonService) (err error) {
			data["isCreated"] = true
			isCreated = true
			return nil
		})

		mockCS.EXPECT().UpdateService(gomock.Eq(serviceToUpdate)).AnyTimes().DoAndReturn(func(service *centreonhandler.CentreonServiceDiff) (err error) {
			data["isUpdated"] = true
			isUpdated = true
			return nil
		})

		mockCS.EXPECT().DeleteService(gomock.Eq("central"), gomock.Eq("ping")).AnyTimes().DoAndReturn(func(host, service string) (err error) {
			data["isDeleted"] = true
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
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Centreon service not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}
