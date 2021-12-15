package controllers

import (
	"errors"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func (t *ControllerTestSuite) TestCentreonServiceSetLogger() {
	log := logrus.NewEntry(logrus.New())
	t.service.SetLogger(log)
}

func (t *ControllerTestSuite) TestCentreonServiceDelete() {

	var err error
	cs := &v1alpha1.CentreonService{
		Spec: v1alpha1.CentreonServiceSpec{
			Host: "central",
			Name: "ping",
		},
	}

	// When service no exist on Centreon
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, nil)
	err = t.service.Delete(cs)
	assert.NoError(t.T(), err)

	// When service exist on Centreon and no error
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(&centreonhandler.CentreonService{
			Host: "central",
			Name: "ping",
		}, nil)
	t.mockCentreonHandler.EXPECT().
		DeleteService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil)
	err = t.service.Delete(cs)
	assert.NoError(t.T(), err)

	// Error when  get service
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, errors.New("fake error"))
	err = t.service.Delete(cs)
	assert.Error(t.T(), err)

	// Error when delete service
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(&centreonhandler.CentreonService{
			Host: "central",
			Name: "ping",
		}, nil)
	t.mockCentreonHandler.EXPECT().
		DeleteService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(errors.New("fake error"))
	err = t.service.Delete(cs)
	assert.Error(t.T(), err)
}

func (t *ControllerTestSuite) TestCentreonServiceReconcile() {

	var (
		err        error
		isCreated  bool
		isUpdated  bool
		csExpected *v1alpha1.CentreonService
		csActual   *centreonhandler.CentreonService
	)
	enabled := true
	csExpected = &v1alpha1.CentreonService{
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
	csActual = &centreonhandler.CentreonService{
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

	// When no change
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(csActual, nil)
	t.mockCentreonHandler.EXPECT().
		DiffService(gomock.Eq(csActual), gomock.Any()).
		Return(&centreonhandler.CentreonServiceDiff{
			IsDiff: false,
			Host:   "central",
			Name:   "ping",
		}, nil)
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isCreated)
	assert.False(t.T(), isUpdated)

	// When no change with error
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, errors.New("fake error"))
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.Error(t.T(), err)

	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(csActual, nil)
	t.mockCentreonHandler.EXPECT().
		DiffService(gomock.Eq(csActual), gomock.Any()).
		Return(nil, errors.New("fake error"))
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.Error(t.T(), err)

	// When create
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, nil)
	t.mockCentreonHandler.EXPECT().
		CreateService(gomock.Eq(csActual)).
		Return(nil)
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.NoError(t.T(), err)
	assert.True(t.T(), isCreated)
	assert.False(t.T(), isUpdated)

	// When create with error
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, nil)
	t.mockCentreonHandler.EXPECT().
		CreateService(gomock.Eq(csActual)).
		Return(errors.New("fake error"))
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.Error(t.T(), err)

	// When update
	diff := &centreonhandler.CentreonServiceDiff{
		IsDiff: true,
		Host:   "central",
		Name:   "ping",
	}
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(csActual, nil)
	t.mockCentreonHandler.EXPECT().
		DiffService(gomock.Eq(csActual), gomock.Any()).
		Return(diff, nil)
	t.mockCentreonHandler.EXPECT().
		UpdateService(gomock.Eq(diff)).
		Return(nil)
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.NoError(t.T(), err)
	assert.False(t.T(), isCreated)
	assert.True(t.T(), isUpdated)

	// When update with error
	t.mockCentreonHandler.EXPECT().
		GetService(gomock.Eq("central"), gomock.Eq("ping")).
		Return(csActual, nil)
	t.mockCentreonHandler.EXPECT().
		DiffService(gomock.Eq(csActual), gomock.Any()).
		Return(diff, nil)
	t.mockCentreonHandler.EXPECT().
		UpdateService(gomock.Eq(diff)).
		Return(errors.New("fake error"))
	isCreated, isUpdated, err = t.service.Reconcile(csExpected)
	assert.Error(t.T(), err)
}
