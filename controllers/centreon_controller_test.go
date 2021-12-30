package controllers

import (
	"context"
	"errors"
	"time"

	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (t *ControllerTestSuite) TestCentreonController() {
	var (
		err       error
		fetched   *v1alpha1.Centreon
		isTimeout bool
	)

	centreonName := "t-centreon-" + helpers.RandomString(10)
	key := types.NamespacedName{
		Name:      centreonName,
		Namespace: "default",
	}

	//Create new CR Centreon
	toCreate := &v1alpha1.Centreon{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: v1alpha1.CentreonSpec{
			Endpoints: &v1alpha1.CentreonSpecEndpoint{
				Template:     "my-template",
				NameTemplate: "ping",
				DefaultHost:  "localhost",
				Macros: map[string]string{
					"macro1": "value",
				},
				Arguments:       []string{"arg1"},
				ServiceGroups:   []string{"sg1"},
				Categories:      []string{"cat1"},
				ActivateService: true,
			},
		},
	}
	err = t.k8sClient.Create(context.Background(), toCreate)
	assert.NoError(t.T(), err)
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.Centreon{}
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
	assert.True(t.T(), fetched.HasFinalizer())
	assert.NotNil(t.T(), t.a.Load())
	assert.Equal(t.T(), &toCreate.Spec, t.a.Load())

	//Update CR Centreon
	fetched = &v1alpha1.Centreon{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	fetched.Spec.Endpoints.Template = "my template 2"
	err = t.k8sClient.Update(context.Background(), fetched)
	assert.NoError(t.T(), err)
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.Centreon{}
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
	assert.NotNil(t.T(), t.a.Load())
	assert.Equal(t.T(), &fetched.Spec, t.a.Load())

	// Delete CR Centreon
	wait := int64(0)
	fetched = &v1alpha1.Centreon{}
	if err := t.k8sClient.Get(context.Background(), key, fetched); err != nil {
		t.T().Fatal(err)
	}
	err = t.k8sClient.Delete(context.Background(), fetched, &client.DeleteOptions{
		GracePeriodSeconds: &wait,
	})
	assert.NoError(t.T(), err)
	isTimeout, err = RunWithTimeout(func() error {
		fetched = &v1alpha1.Centreon{}
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
	assert.Nil(t.T(), t.a.Load())

}
