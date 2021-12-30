package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestCentreonServiceController() {
	var (
		err       error
		step      *string
		fetched   *v1alpha1.CentreonService
		isTimeout bool
		isCreated bool = false
		isUpdated bool = false
	)
	activated := true
	centreonServiceName := "t-centreon-service-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      centreonServiceName,
		Namespace: "default",
	}

	t.mockCentreonService.EXPECT().
		Reconcile(gomock.Any()).AnyTimes().DoAndReturn(func(instance *v1alpha1.CentreonService) (bool, bool, error) {

		if *step == "create" {
			if !isCreated {
				isCreated = true
				return true, false, nil
			}

			return false, false, nil
		}

		if *step == "update" {
			if !isUpdated {
				isUpdated = true
				return false, true, nil
			}
			return false, false, nil
		}

		return false, false, nil

	})
	t.mockCentreonService.EXPECT().
		Delete(gomock.Any()).AnyTimes().Return(nil)

	//Create new CR ServiceCentreon
	create := "create"
	step = &create
	toCreate := &v1alpha1.CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: v1alpha1.CentreonServiceSpec{
			Name:         "ping",
			Host:         "central",
			Template:     "my-template",
			CheckCommand: "ping",
			Arguments:    []string{"arg1"},
			Groups:       []string{"sg1"},
			Categories:   []string{"cat1"},
			Macros: map[string]string{
				"macro1": "value",
			},
			Activated:           true,
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "3",
			ActiveCheckEnabled:  &activated,
			PassiveCheckEnabled: &activated,
		},
	}
	if err = t.k8sClient.Create(context.Background(), toCreate); err != nil {
		t.T().Fatal(err)
	}
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.CentreonService{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if !fetched.IsSubmitted() {
			return errors.New("Not yet created")
		}
		return nil
	}, time.Second*30, time.Second*1)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isTimeout)
	assert.NotEmpty(t.T(), fetched.Status.ID)
	assert.NotEmpty(t.T(), fetched.Status.CreatedAt)
	assert.Empty(t.T(), fetched.Status.UpdatedAt)
	assert.True(t.T(), fetched.HasFinalizer())
	time.Sleep(10 * time.Second)

	//Update CR ServiceCentreon
	fetched = &v1alpha1.CentreonService{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Template = "my template 2"
	update := "update"
	step = &update
	if err = t.k8sClient.Update(context.Background(), fetched); err != nil {
		t.T().Fatal(err)
	}
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.CentreonService{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			t.T().Fatal(err)
		}
		if fetched.Status.UpdatedAt == "" {
			return errors.New("Not yet updated")
		}

		return nil
	}, time.Second*30, time.Second*1)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isTimeout)
	assert.NotEmpty(t.T(), fetched.Status.ID)
	assert.NotEmpty(t.T(), fetched.Status.UpdatedAt)
	assert.True(t.T(), fetched.HasFinalizer())
	time.Sleep(10 * time.Second)

	// Delete CR ServiceCentreon
	fetched = &v1alpha1.CentreonService{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	delete := "delete"
	step = &delete
	wait := int64(0)
	if err = t.k8sClient.Delete(context.Background(), fetched, &client.DeleteOptions{
		GracePeriodSeconds: &wait,
	}); err != nil {
		t.T().Fatal(err)
	}
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.CentreonService{}
		if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
			if k8serrors.IsNotFound(err) {
				return nil
			}
			t.T().Fatal(err)
		}

		return errors.New("Not yet deleted")
	}, time.Second*30, time.Second*1)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isTimeout)
	time.Sleep(10 * time.Second)
}
