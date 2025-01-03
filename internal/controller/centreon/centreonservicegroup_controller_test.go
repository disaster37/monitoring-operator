package centreon

import (
	"context"
	"testing"
	"time"

	"github.com/disaster37/monitoring-operator/api/shared"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/disaster37/monitoring-operator/pkg/mocks"
	"github.com/disaster37/operator-sdk-extra/pkg/controller"
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

func (t *CentreonControllerTestSuite) TestCentreonServiceGroupController() {
	key := types.NamespacedName{
		Name:      "t-csg-" + helpers.RandomString(10),
		Namespace: "default",
	}
	csg := &monitorapi.CentreonServiceGroup{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, csg, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateCentreonServiceGroupStep(),
		doUpdateCentreonServiceGroupStep(),
		doDeleteCentreonServiceGroupStep(),
		doPolicyNoCreateCentreonServiceGroupStep(),
		doPolicyNoUpdateCentreonServiceGroupStep(),
		doPolicyExcludeFieldsCentreonServiceGroupStep(),
		doPolicyNoDeleteCentreonServiceGroupStep(),
	}
	testCase.PreTest = doMockCentreonServiceGroup(t.mockCentreonHandler)

	testCase.Run()
}

func doMockCentreonServiceGroup(mockCSG *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false
		isCreatedPolicyNoCreate := false
		isUpdatedPolicyNoUpdate := false
		isUpdatedPolicyExcludeFields := false

		mockCSG.EXPECT().GetServiceGroup(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (serviceGroup *centreonhandler.CentreonServiceGroup, err error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return nil, nil
				} else {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg",
					}, nil
				}
			case "update":
				if !isUpdated {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg2",
					}, nil
				}
			case "policyNocreate":
				if !isCreatedPolicyNoCreate {
					return nil, nil
				} else {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg",
					}, nil
				}
			case "policyNoUpdate":
				if !isUpdatedPolicyNoUpdate {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg2",
					}, nil
				}
			case "policyExcludeFields":
				if !isUpdatedPolicyExcludeFields {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroup{
						Name:        "sg1",
						Activated:   "1",
						Comment:     "Managed by monitoring-operator",
						Description: "my sg2",
					}, nil
				}
			default:
				return &centreonhandler.CentreonServiceGroup{
					Name:        "sg1",
					Activated:   "1",
					Comment:     "Managed by monitoring-operator",
					Description: "my sg",
				}, nil
			}
		})

		mockCSG.EXPECT().DiffServiceGroup(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *centreonhandler.CentreonServiceGroup, ignoreFields []string) (diff *centreonhandler.CentreonServiceGroupDiff, err error) {
			switch *stepName {
			case "create":
				if !isCreated {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: true,
						Name:   "sg1",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: false,
						Name:   "sg1",
					}, nil
				}
			case "update":
				if !isUpdated {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: true,
						Name:   "sg1",
						ParamsToSet: map[string]string{
							"alias": "my sg2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: false,
						Name:   "sg1",
					}, nil
				}
			case "policyNoCreate":
				if !isCreatedPolicyNoCreate {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: true,
						Name:   "sg1",
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: false,
						Name:   "sg1",
					}, nil
				}
			case "policyNoUpdate":
				if !isUpdatedPolicyNoUpdate {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: true,
						Name:   "sg1",
						ParamsToSet: map[string]string{
							"alias": "my sg2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: false,
						Name:   "sg1",
					}, nil
				}
			case "policyExcludeFields":
				if !funk.Contains(ignoreFields, "description") {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: true,
						Name:   "sg1",
						ParamsToSet: map[string]string{
							"alias": "my sg2",
						},
					}, nil
				} else {
					return &centreonhandler.CentreonServiceGroupDiff{
						IsDiff: false,
						Name:   "sg1",
					}, nil
				}
			}
			return nil, errors.Errorf("Unnatented test: %s", *stepName)
		})

		mockCSG.EXPECT().CreateServiceGroup(gomock.Any()).AnyTimes().DoAndReturn(func(serviceGroup *centreonhandler.CentreonServiceGroup) (err error) {
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

		mockCSG.EXPECT().UpdateServiceGroup(gomock.Any()).AnyTimes().DoAndReturn(func(serviceGroup *centreonhandler.CentreonServiceGroupDiff) (err error) {
			switch *stepName {
			case "update":
				data["isUpdated"] = true
				isUpdated = true
			case "policyNoUpdate":
				data["isUpdatedPolicyNoUpdate"] = true
				isUpdatedPolicyNoUpdate = true
			case "policyExcludeFields":
				data["isUpdatedPolicyExcludeFields"] = true
				isUpdatedPolicyNoUpdate = true
			}

			return nil
		})

		mockCSG.EXPECT().DeleteServiceGroup(gomock.Any()).AnyTimes().DoAndReturn(func(name string) (err error) {
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

func doCreateCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "create",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Centreon ServiceGroup %s/%s ===", key.Namespace, key.Name)

			csg := &monitorapi.CentreonServiceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: monitorapi.CentreonServiceGroupSpec{
					Name:        "sg1",
					Description: "my sg",
					Activated:   true,
				},
			}

			if err = c.Create(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isCreated := false

			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
					t.Fatal("Centreon serviceGroup not found")
				}
				if b, ok := data["isCreated"]; ok {
					isCreated = b.(bool)
				}
				if !isCreated || csg.GetStatus().GetObservedGeneration() == 0 {
					return errors.New("Not yet created")
				}
				return nil
			}, time.Second*30, time.Second*1)
			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon serviceGroup: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, controller.ReadyCondition.String(), metav1.ConditionTrue))
			assert.Equal(t, "sg1", csg.Status.ServiceGroupName)
			return nil
		},
	}
}

func doUpdateCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "update",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon ServiceGroup %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon serviceGroup is null")
			}
			csg := o.(*monitorapi.CentreonServiceGroup)

			data["lastGeneration"] = csg.GetStatus().GetObservedGeneration()
			csg.Spec.Description = "my sg2"
			if err = c.Update(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isUpdated := false
			lastGeneration := data["lastGeneration"].(int64)

			isTimeout, err := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
					t.Fatal(err)
				}
				if b, ok := data["isUpdated"]; ok {
					isUpdated = b.(bool)
				}
				if !isUpdated || lastGeneration == csg.GetStatus().GetObservedGeneration() {
					return errors.New("Not yet updated")
				}
				return nil
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Failed to get Centreon serviceGroup: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, controller.ReadyCondition.String(), metav1.ConditionTrue))
			assert.Equal(t, "sg1", csg.Status.ServiceGroupName)
			return nil
		},
	}
}

func doDeleteCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "delete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Centreon ServiceGroup %s/%s ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon serviceGroup is null")
			}
			csg := o.(*monitorapi.CentreonServiceGroup)

			wait := int64(0)
			if err = c.Delete(context.Background(), csg, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isDeleted := false

			isTimeout, err := test.RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, csg); err != nil {
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
				t.Fatalf("Centreon serviceGroup not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}

func doPolicyNoCreateCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoCreate",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Add new Centreon ServiceGroup %s/%s (policyNoCreate) ===", key.Namespace, key.Name)

			csg := &monitorapi.CentreonServiceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: monitorapi.CentreonServiceGroupSpec{
					Name:        "sg1",
					Description: "my sg",
					Activated:   true,
					Policy: shared.Policy{
						NoCreate:            true,
						NoUpdate:            true,
						NoDelete:            true,
						ExcludeFieldsOnDiff: []string{"description"},
					},
				},
			}

			if err = c.Create(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isCreated := false

			isTimeout, _ := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
					t.Fatal("Centreon serviceGroup not found")
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

func doPolicyNoUpdateCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoUpdate",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon ServiceGroup %s/%s (policyNoUpdate) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon serviceGroup is null")
			}
			csg := o.(*monitorapi.CentreonServiceGroup)

			if err := c.Get(context.Background(), key, csg); err != nil {
				return err
			}

			data["lastGeneration"] = csg.GetStatus().GetObservedGeneration()
			csg.Spec.Description = "my sg3"
			if err = c.Update(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isUpdated := false

			isTimeout, _ := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
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

func doPolicyExcludeFieldsCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "policyExcludeFields",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Update Centreon ServiceGroup %s/%s (policyExcludeFields) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon serviceGroup is null")
			}
			csg := o.(*monitorapi.CentreonServiceGroup)

			if err := c.Get(context.Background(), key, csg); err != nil {
				return err
			}

			data["lastGeneration"] = csg.GetStatus().GetObservedGeneration()
			csg.Spec.Policy.NoUpdate = false
			csg.Spec.Description = "my sg4"
			if err = c.Update(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isUpdated := false

			isTimeout, _ := test.RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
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

func doPolicyNoDeleteCentreonServiceGroupStep() test.TestStep {
	return test.TestStep{
		Name: "policyNoDelete",
		Do: func(c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			logrus.Infof("=== Delete Centreon ServiceGroup %s/%s (policyNoDelete) ===", key.Namespace, key.Name)

			if o == nil {
				return errors.New("Centreon serviceGroup is null")
			}
			csg := o.(*monitorapi.CentreonServiceGroup)

			wait := int64(0)
			if err = c.Delete(context.Background(), csg, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &monitorapi.CentreonServiceGroup{}
			isDeleted := false

			isTimeout, err := test.RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, csg); err != nil {
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
