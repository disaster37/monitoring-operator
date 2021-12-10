package centreonhandler

import (
	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/api/v1alpha1"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func (t *CentreonHandlerTestSuite) TestCreateService() {
	toCreate := &v1alpha1.CentreonServiceSpec{
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
	}

	// Mock add service on Centreon
	t.mockService.EXPECT().
		Add(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("my-template")).
		Return(nil)

	// Mock set params on Centreon
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("check_command"), gomock.Eq("ping")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("normal_check_interval"), gomock.Eq("30s")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("retry_check_interval"), gomock.Eq("1s")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("max_check_attempts"), gomock.Eq("3")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("check_command_arguments"), gomock.Eq("!arg1")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("activate"), gomock.Eq("1")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("active_checks_enabled"), gomock.Eq("default")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("passive_checks_enabled"), gomock.Eq("default")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("comment"), gomock.Any()).
		Return(nil)

	// Mock set service groups
	t.mockService.EXPECT().
		SetServiceGroups(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"sg1"})).
		Return(nil)

	// Mock set categories
	t.mockService.EXPECT().
		SetCategories(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"cat1"})).
		Return(nil)

	//Mock set macros
	t.mockService.EXPECT().
		SetMacro(gomock.Eq("central"), gomock.Eq("ping"), gomock.Any()).
		Return(nil)

	err := t.client.CreateService(toCreate)
	assert.NoError(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestUpdateService() {
	toUpdate := &v1alpha1.CentreonServiceSpec{
		Name:         "ping",
		Host:         "central",
		Template:     "my-template2",
		CheckCommand: "ping2",
		Arguments:    []string{"arg2"},
		Groups:       []string{"sg2"},
		Categories:   []string{"cat2"},
		Macros: map[string]string{
			"macro2": "value",
		},
		Activated:           false,
		NormalCheckInterval: "35s",
		RetryCheckInterval:  "2s",
		MaxCheckAttempts:    "4",
	}
	cs := &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			CheckCommand:        "ping",
			Activated:           "1",
			ActiveCheckEnabled:  "default",
			PassiveCheckEnabled: "default",
			CheckCommandArgs:    "!arg1",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "3",
		},
	}

	// Mock get
	t.mockService.EXPECT().
		Get(gomock.Eq("central"), gomock.Eq("ping")).
		Return(cs, nil)

	// Mock get params
	t.mockService.EXPECT().
		GetParam(gomock.Eq("central"), gomock.Eq("ping"), []string{"template"}).
		Return(map[string]string{"template": "my-template"}, nil)

		// Mock get macros
	t.mockService.EXPECT().
		GetMacros(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]*models.Macro{
			{
				Name:       "macro1",
				Value:      "value",
				IsPassword: "0",
			},
		}, nil)

		// Mock get Categories
	t.mockService.EXPECT().
		GetCategories(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]string{"cat1"}, nil)

		// Mock get service groups
	t.mockService.EXPECT().
		GetServiceGroups(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]string{"sg1"}, nil)

	// Mock set params on Centreon
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("check_command"), gomock.Eq("ping2")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("normal_check_interval"), gomock.Eq("35s")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("retry_check_interval"), gomock.Eq("2s")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("max_check_attempts"), gomock.Eq("4")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("check_command_arguments"), gomock.Eq("!arg2")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("activate"), gomock.Eq("0")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("template"), gomock.Eq("my-template2")).
		Return(nil)

	// Mock set service groups
	t.mockService.EXPECT().
		SetServiceGroups(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"sg2"})).
		Return(nil)
	t.mockService.EXPECT().
		DeleteServiceGroups(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"sg1"})).
		Return(nil)

	// Mock set categories
	t.mockService.EXPECT().
		SetCategories(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"cat2"})).
		Return(nil)
	t.mockService.EXPECT().
		DeleteCategories(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq([]string{"cat1"})).
		Return(nil)

	//Mock set macros
	t.mockService.EXPECT().
		SetMacro(gomock.Eq("central"), gomock.Eq("ping"), gomock.Any()).
		Return(nil)
	t.mockService.EXPECT().
		DeleteMacro(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("macro1")).
		Return(nil)

	err := t.client.UpdateService(toUpdate)
	assert.NoError(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestDeleteService() {
	toDelete := &v1alpha1.CentreonServiceSpec{
		Name: "ping",
		Host: "central",
	}

	// Mock delete
	t.mockService.EXPECT().
		Delete(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil)

	err := t.client.DeleteService(toDelete)
	assert.NoError(t.T(), err)
}
