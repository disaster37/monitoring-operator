package controllers

import (
	"context"
	"errors"
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
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	condition "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestCentreonServiceGroupController() {
	key := types.NamespacedName{
		Name:      "t-csg-" + helpers.RandomString(10),
		Namespace: "default",
	}
	csg := &v1alpha1.CentreonServiceGroup{}
	data := map[string]any{}

	testCase := test.NewTestCase(t.T(), t.k8sClient, key, csg, 5*time.Second, data)
	testCase.Steps = []test.TestStep{
		doCreateCentreonServiceGroupStep(),
		doUpdateCentreonServiceGroupStep(),
		doDeleteCentreonServiceGroupStep(),
	}
	testCase.PreTest = doMockCentreonServiceGroup(t.mockCentreonHandler)

	testCase.Run()

}

func doMockCentreonServiceGroup(mockCSG *mocks.MockCentreonHandler) func(stepName *string, data map[string]any) error {
	return func(stepName *string, data map[string]any) (err error) {
		isCreated := false
		isUpdated := false

		serviceGroupToCreate := &centreonhandler.CentreonServiceGroup{
			Name:        "sg1",
			Activated:   "1",
			Comment:     "Managed by monitoring-operator",
			Description: "my sg",
		}

		serviceGroupToUpdate := &centreonhandler.CentreonServiceGroupDiff{
			IsDiff: true,
			Name:   "sg1",
			ParamsToSet: map[string]string{
				"alias": "my sg2",
			},
		}

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
			}
			return nil, nil
		})

		mockCSG.EXPECT().DiffServiceGroup(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(actual, expected *centreonhandler.CentreonServiceGroup) (diff *centreonhandler.CentreonServiceGroupDiff, err error) {
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
			}
			return nil, nil
		})

		mockCSG.EXPECT().CreateServiceGroup(gomock.Eq(serviceGroupToCreate)).AnyTimes().DoAndReturn(func(serviceGroup *centreonhandler.CentreonServiceGroup) (err error) {
			data["isCreated"] = true
			isCreated = true
			return nil
		})

		mockCSG.EXPECT().UpdateServiceGroup(gomock.Eq(serviceGroupToUpdate)).AnyTimes().DoAndReturn(func(serviceGroup *centreonhandler.CentreonServiceGroupDiff) (err error) {
			data["isUpdated"] = true
			isUpdated = true
			return nil
		})

		mockCSG.EXPECT().DeleteServiceGroup(gomock.Eq("sg1")).AnyTimes().DoAndReturn(func(name string) (err error) {
			data["isDeleted"] = true
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

			csg := &v1alpha1.CentreonServiceGroup{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: v1alpha1.CentreonServiceGroupSpec{
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
			csg := &v1alpha1.CentreonServiceGroup{}
			isCreated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
					t.Fatal("Centreon serviceGroup not found")
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
				t.Fatalf("Failed to get Centreon serviceGroup: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, CentreonServiceGroupCondition, metav1.ConditionTrue))
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
			csg := o.(*v1alpha1.CentreonServiceGroup)
			csg.Spec.Description = "my sg2"

			if err = c.Update(context.Background(), csg); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &v1alpha1.CentreonServiceGroup{}
			isUpdated := false

			isTimeout, err := RunWithTimeout(func() error {
				if err := c.Get(context.Background(), key, csg); err != nil {
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
				t.Fatalf("Failed to get Centreon serviceGroup: %s", err.Error())
			}
			assert.True(t, condition.IsStatusConditionPresentAndEqual(csg.Status.Conditions, CentreonServiceGroupCondition, metav1.ConditionTrue))
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
			csg := o.(*v1alpha1.CentreonServiceGroup)

			wait := int64(0)
			if err = c.Delete(context.Background(), csg, &client.DeleteOptions{GracePeriodSeconds: &wait}); err != nil {
				return err
			}

			return nil
		},
		Check: func(t *testing.T, c client.Client, key types.NamespacedName, o client.Object, data map[string]any) (err error) {
			csg := &v1alpha1.CentreonServiceGroup{}
			isDeleted := false

			isTimeout, err := RunWithTimeout(func() error {
				if err = c.Get(context.Background(), key, csg); err != nil {
					if k8serrors.IsNotFound(err) {
						isDeleted = true
						return nil
					}
					t.Fatal(err)
				}

				return errors.New("Not yet deleted")
			}, time.Second*30, time.Second*1)

			if err != nil || isTimeout {
				t.Fatalf("Centreon serviceGroup not deleted: %s", err.Error())
			}
			assert.True(t, isDeleted)
			return nil
		},
	}
}
