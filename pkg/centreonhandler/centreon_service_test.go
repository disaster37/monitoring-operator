package centreonhandler

import (
	"testing"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func (t *CentreonHandlerTestSuite) TestCreateService() {
	macro1 := &models.Macro{

		Name:       "macro1",
		Value:      "value",
		IsPassword: "0",
	}
	toCreate := &CentreonService{
		Name:                "ping",
		Host:                "central",
		Template:            "my-template",
		CheckCommand:        "ping",
		CheckCommandArgs:    "!arg1",
		Groups:              []string{"sg1"},
		Categories:          []string{"cat1"},
		Macros:              []*models.Macro{macro1},
		Activated:           "1",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "some comments",
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
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("active_checks_enabled"), gomock.Eq("2")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("passive_checks_enabled"), gomock.Eq("2")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("comment"), gomock.Eq("some comments")).
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
		SetMacro(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq(macro1)).
		Return(nil)

	err := t.client.CreateService(toCreate)
	assert.NoError(t.T(), err)

	// When bad parameters
	err = t.client.CreateService(nil)
	assert.Error(t.T(), err)

	toCreate = &CentreonService{
		Host:                "central",
		Template:            "my-template",
		CheckCommand:        "ping",
		CheckCommandArgs:    "!arg1",
		Groups:              []string{"sg1"},
		Categories:          []string{"cat1"},
		Macros:              []*models.Macro{macro1},
		Activated:           "1",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "some comments",
		NormalCheckInterval: "30s",
		RetryCheckInterval:  "1s",
		MaxCheckAttempts:    "3",
	}
	err = t.client.CreateService(toCreate)
	assert.Error(t.T(), err)

	toCreate = &CentreonService{
		Name:                "ping",
		Template:            "my-template",
		CheckCommand:        "ping",
		CheckCommandArgs:    "!arg1",
		Groups:              []string{"sg1"},
		Categories:          []string{"cat1"},
		Macros:              []*models.Macro{macro1},
		Activated:           "1",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "some comments",
		NormalCheckInterval: "30s",
		RetryCheckInterval:  "1s",
		MaxCheckAttempts:    "3",
	}
	err = t.client.CreateService(toCreate)
	assert.Error(t.T(), err)

}

func (t *CentreonHandlerTestSuite) TestUpdateService() {
	macro1 := &models.Macro{
		Name:       "macro1",
		Value:      "value1",
		IsPassword: "0",
	}
	macro2 := &models.Macro{
		Name:       "macro2",
		Value:      "value2",
		IsPassword: "0",
	}
	toUpdate := &CentreonServiceDiff{
		IsDiff:             true,
		Name:               "ping",
		Host:               "central",
		GroupsToSet:        []string{"sg2"},
		GroupsToDelete:     []string{"sg1"},
		CategoriesToSet:    []string{"cat2"},
		CategoriesToDelete: []string{"cat1"},
		MacrosToSet:        []*models.Macro{macro2},
		MacrosToDelete:     []*models.Macro{macro1},
		ParamsToSet: map[string]string{
			"param1":      "value1",
			"param2":      "value2",
			"description": "ping2",
		},
		HostToSet: "central2",
	}

	// Mock set params on Centreon
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Any(), gomock.Eq("param1"), gomock.Eq("value1")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Any(), gomock.Eq("param2"), gomock.Eq("value2")).
		Return(nil)
	t.mockService.EXPECT().
		SetParam(gomock.Eq("central"), gomock.Eq("ping"), gomock.Eq("description"), gomock.Eq("ping2")).
		Return(nil)

		// Mock set service groups
	t.mockService.EXPECT().
		SetServiceGroups(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq([]string{"sg2"})).
		Return(nil)
	t.mockService.EXPECT().
		DeleteServiceGroups(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq([]string{"sg1"})).
		Return(nil)

	// Mock set categories
	t.mockService.EXPECT().
		SetCategories(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq([]string{"cat2"})).
		Return(nil)
	t.mockService.EXPECT().
		DeleteCategories(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq([]string{"cat1"})).
		Return(nil)

	//Mock set macros
	t.mockService.EXPECT().
		SetMacro(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq(macro2)).
		Return(nil)
	t.mockService.EXPECT().
		DeleteMacro(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq(macro1.Name)).
		Return(nil)

	// Mock set host
	t.mockService.EXPECT().
		SetHost(gomock.Eq("central"), gomock.Eq("ping2"), gomock.Eq("central2")).
		Return(nil)

	err := t.client.UpdateService(toUpdate)
	assert.NoError(t.T(), err)

	// When no diff
	toUpdate = &CentreonServiceDiff{
		IsDiff:             false,
		Name:               "ping",
		Host:               "central",
		GroupsToSet:        []string{"sg2"},
		GroupsToDelete:     []string{"sg1"},
		CategoriesToSet:    []string{"cat2"},
		CategoriesToDelete: []string{"cat1"},
		MacrosToSet:        []*models.Macro{macro2},
		MacrosToDelete:     []*models.Macro{macro1},
		ParamsToSet: map[string]string{
			"param1":      "value1",
			"param2":      "value2",
			"description": "ping2",
		},
		HostToSet: "central2",
	}
	err = t.client.UpdateService(toUpdate)
	assert.NoError(t.T(), err)

	// When diff is bad
	err = t.client.UpdateService(nil)
	assert.Error(t.T(), err)

	toUpdate = &CentreonServiceDiff{
		IsDiff:             false,
		Host:               "central",
		GroupsToSet:        []string{"sg2"},
		GroupsToDelete:     []string{"sg1"},
		CategoriesToSet:    []string{"cat2"},
		CategoriesToDelete: []string{"cat1"},
		MacrosToSet:        []*models.Macro{macro2},
		MacrosToDelete:     []*models.Macro{macro1},
		ParamsToSet: map[string]string{
			"param1":      "value1",
			"param2":      "value2",
			"description": "ping2",
		},
		HostToSet: "central2",
	}
	err = t.client.UpdateService(toUpdate)
	assert.Error(t.T(), err)

	toUpdate = &CentreonServiceDiff{
		IsDiff:             false,
		Name:               "ping",
		GroupsToSet:        []string{"sg2"},
		GroupsToDelete:     []string{"sg1"},
		CategoriesToSet:    []string{"cat2"},
		CategoriesToDelete: []string{"cat1"},
		MacrosToSet:        []*models.Macro{macro2},
		MacrosToDelete:     []*models.Macro{macro1},
		ParamsToSet: map[string]string{
			"param1":      "value1",
			"param2":      "value2",
			"description": "ping2",
		},
		HostToSet: "central2",
	}
	err = t.client.UpdateService(toUpdate)
	assert.Error(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestDeleteService() {

	// Mock delete
	t.mockService.EXPECT().
		Delete(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil)

	err := t.client.DeleteService("central", "ping")
	assert.NoError(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestGetService() {

	macro1 := &models.Macro{
		Name:       "macro1",
		Value:      "value1",
		IsPassword: "0",
	}
	expected := &CentreonService{
		Name:                "ping",
		Host:                "central",
		Template:            "my-template",
		CheckCommand:        "ping",
		CheckCommandArgs:    "!arg1",
		Groups:              []string{"sg1"},
		Categories:          []string{"cat1"},
		Macros:              []*models.Macro{macro1},
		Activated:           "1",
		PassiveCheckEnabled: "2",
		ActiveCheckEnabled:  "2",
		Comment:             "my comment",
		NormalCheckInterval: "30s",
		RetryCheckInterval:  "1s",
		MaxCheckAttempts:    "3",
	}

	cs := &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			CheckCommand:        "ping",
			Activated:           "1",
			ActiveCheckEnabled:  "2",
			PassiveCheckEnabled: "2",
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
		GetParam(gomock.Eq("central"), gomock.Eq("ping"), []string{"template", "comment"}).
		Return(map[string]string{"template": "my-template", "comment": "my comment"}, nil)

	// Mock get macros
	t.mockService.EXPECT().
		GetMacros(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]*models.Macro{macro1}, nil)

	// Mock get Categories
	t.mockService.EXPECT().
		GetCategories(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]string{"cat1"}, nil)

	// Mock get service groups
	t.mockService.EXPECT().
		GetServiceGroups(gomock.Eq("central"), gomock.Eq("ping")).
		Return([]string{"sg1"}, nil)

	service, err := t.client.GetService("central", "ping")
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), expected, service)

	// When not found
	t.mockService.EXPECT().
		Get(gomock.Eq("central"), gomock.Eq("ping")).
		Return(nil, nil)
	service, err = t.client.GetService("central", "ping")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), service)

	// When bad parameters
	_, err = t.client.GetService("", "ping")
	assert.Error(t.T(), err)

	_, err = t.client.GetService("central", "")
	assert.Error(t.T(), err)

}

func (t *CentreonHandlerTestSuite) TestDiffService() {

	tests := []struct {
		Name            string
		ActualService   *CentreonService
		ExpectedService *CentreonService
		ExpectedDiff    *CentreonServiceDiff
		IgnoreFields    []string
	}{
		{
			Name: "no need update and extra infos is nil",
			ActualService: &CentreonService{
				Host: "central",
				Name: "ping",
			},
			ExpectedService: &CentreonService{
				Host: "central",
				Name: "ping",
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff:             false,
				Host:               "central",
				Name:               "ping",
				ParamsToSet:        map[string]string{},
				GroupsToSet:        make([]string, 0),
				GroupsToDelete:     make([]string, 0),
				CategoriesToSet:    make([]string, 0),
				CategoriesToDelete: make([]string, 0),
				MacrosToSet:        make([]*models.Macro, 0),
				MacrosToDelete:     make([]*models.Macro, 0),
			},
		},
		{
			Name: "No Need update and all properties set",
			ActualService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff:             false,
				Host:               "central",
				Name:               "ping",
				ParamsToSet:        map[string]string{},
				GroupsToSet:        make([]string, 0),
				GroupsToDelete:     make([]string, 0),
				CategoriesToSet:    make([]string, 0),
				CategoriesToDelete: make([]string, 0),
				MacrosToSet:        make([]*models.Macro, 0),
				MacrosToDelete:     make([]*models.Macro, 0),
			},
		},
		{
			Name: "Need update all properties",
			ActualService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "0",
				Template:            "template",
				CheckCommand:        "ping",
				NormalCheckInterval: "30s",
				RetryCheckInterval:  "1s",
				MaxCheckAttempts:    "5",
				ActiveCheckEnabled:  "0",
				PassiveCheckEnabled: "0",
				CheckCommandArgs:    "!arg1",
				Comment:             "comment",
				Groups:              []string{"sg1"},
				Categories:          []string{"cat1"},
				Macros: []*models.Macro{
					{
						Name:       "macro1",
						Value:      "value1",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff: true,
				Host:   "central",
				Name:   "ping",
				ParamsToSet: map[string]string{
					"retry_check_interval":    "2s",
					"max_check_attempts":      "6",
					"check_command_arguments": "!arg2",
					"activate":                "1",
					"template":                "template2",
					"check_command":           "ping2",
					"normal_check_interval":   "31s",
					"comment":                 "comment2",
					"active_checks_enabled":   "1",
					"passive_checks_enabled":  "1",
				},
				GroupsToSet:        []string{"sg2"},
				GroupsToDelete:     []string{"sg1"},
				CategoriesToSet:    []string{"cat2"},
				CategoriesToDelete: []string{"cat1"},
				MacrosToSet: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
				MacrosToDelete: []*models.Macro{
					{
						Name:       "macro1",
						Value:      "value1",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
		},
		{
			Name: "No Need update because of exclude macro",
			ActualService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "0",
				Template:            "template",
				CheckCommand:        "ping",
				NormalCheckInterval: "30s",
				RetryCheckInterval:  "1s",
				MaxCheckAttempts:    "5",
				ActiveCheckEnabled:  "0",
				PassiveCheckEnabled: "0",
				CheckCommandArgs:    "!arg1",
				Comment:             "comment",
				Groups:              []string{"sg1"},
				Categories:          []string{"cat1"},
				Macros: []*models.Macro{
					{
						Name:       "macro1",
						Value:      "value1",
						Source:     "template",
						IsPassword: "0",
					},
					{
						Name:       "macro2",
						Value:      "",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "0",
				Template:            "template",
				CheckCommand:        "ping",
				NormalCheckInterval: "30s",
				RetryCheckInterval:  "1s",
				MaxCheckAttempts:    "5",
				ActiveCheckEnabled:  "0",
				PassiveCheckEnabled: "0",
				CheckCommandArgs:    "!arg1",
				Comment:             "comment",
				Groups:              []string{"sg1"},
				Categories:          []string{"cat1"},
				Macros:              []*models.Macro{},
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff:             false,
				Host:               "central",
				Name:               "ping",
				ParamsToSet:        map[string]string{},
				GroupsToSet:        make([]string, 0),
				GroupsToDelete:     make([]string, 0),
				CategoriesToSet:    make([]string, 0),
				CategoriesToDelete: make([]string, 0),
				MacrosToSet:        make([]*models.Macro, 0),
				MacrosToDelete:     make([]*models.Macro, 0),
			},
		},
		{
			Name: "Need update all properties but all fields are exclued",
			ActualService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "0",
				Template:            "template",
				CheckCommand:        "ping",
				NormalCheckInterval: "30s",
				RetryCheckInterval:  "1s",
				MaxCheckAttempts:    "5",
				ActiveCheckEnabled:  "0",
				PassiveCheckEnabled: "0",
				CheckCommandArgs:    "!arg1",
				Comment:             "comment",
				Groups:              []string{"sg1"},
				Categories:          []string{"cat1"},
				Macros: []*models.Macro{
					{
						Name:       "macro1",
						Value:      "value1",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			IgnoreFields: []string{
				"host",
				"name",
				"activate",
				"template",
				"checkCommand",
				"normalCheckInterval",
				"retryCheckInterval",
				"maxCheckAttempts",
				"activeChecksEnabled",
				"passiveChecksEnabled",
				"arguments",
				"comment",
				"groups",
				"categories",
				"macros",
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff:             false,
				Host:               "central",
				Name:               "ping",
				ParamsToSet:        map[string]string{},
				GroupsToSet:        make([]string, 0),
				GroupsToDelete:     make([]string, 0),
				CategoriesToSet:    make([]string, 0),
				CategoriesToDelete: make([]string, 0),
				MacrosToSet:        make([]*models.Macro, 0),
				MacrosToDelete:     make([]*models.Macro, 0),
			},
		},
		{
			Name: "Rename service and host",
			ActualService: &CentreonService{
				Host:                "central",
				Name:                "ping",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedService: &CentreonService{
				Host:                "central2",
				Name:                "ping2",
				Activated:           "1",
				Template:            "template2",
				CheckCommand:        "ping2",
				NormalCheckInterval: "31s",
				RetryCheckInterval:  "2s",
				MaxCheckAttempts:    "6",
				ActiveCheckEnabled:  "1",
				PassiveCheckEnabled: "1",
				CheckCommandArgs:    "!arg2",
				Comment:             "comment2",
				Groups:              []string{"sg2"},
				Categories:          []string{"cat2"},
				Macros: []*models.Macro{
					{
						Name:       "macro2",
						Value:      "value2",
						Source:     "direct",
						IsPassword: "0",
					},
				},
			},
			ExpectedDiff: &CentreonServiceDiff{
				IsDiff:    true,
				Host:      "central",
				Name:      "ping",
				HostToSet: "central2",
				ParamsToSet: map[string]string{
					"description": "ping2",
				},
				GroupsToSet:        make([]string, 0),
				GroupsToDelete:     make([]string, 0),
				CategoriesToSet:    make([]string, 0),
				CategoriesToDelete: make([]string, 0),
				MacrosToSet:        make([]*models.Macro, 0),
				MacrosToDelete:     make([]*models.Macro, 0),
			},
		},
	}

	for _, test := range tests {
		diff, err := t.client.DiffService(test.ActualService, test.ExpectedService, test.IgnoreFields)
		assert.NoErrorf(t.T(), err, test.Name)
		assert.Equalf(t.T(), test.ExpectedDiff, diff, test.Name)
	}
}

func TestCentreonServiceToString(t *testing.T) {
	s := &CentreonService{
		Host:                "host1",
		Name:                "s1",
		CheckCommand:        "check",
		CheckCommandArgs:    "!arg1",
		NormalCheckInterval: "1s",
		RetryCheckInterval:  "2s",
		MaxCheckAttempts:    "3s",
		ActiveCheckEnabled:  "2",
		PassiveCheckEnabled: "2",
		Template:            "template1",
		Groups:              []string{"group1"},
		Categories:          []string{"cat1"},
		Macros: []*models.Macro{
			{
				Name:       "MAC1",
				Value:      "value1",
				IsPassword: "0",
			},
		},
		Activated: "1",
		Comment:   "Managed by monitoring-operator",
	}

	assert.NotEmpty(t, s.String())
}

func TestCentreonServiceDiffToString(t *testing.T) {
	s := &CentreonServiceDiff{
		Host:               "host1",
		Name:               "s1",
		IsDiff:             true,
		GroupsToSet:        []string{"group1"},
		GroupsToDelete:     []string{"group2"},
		CategoriesToSet:    []string{"cat1"},
		CategoriesToDelete: []string{"cat2"},
		MacrosToSet: []*models.Macro{
			{
				Name:  "MAC1",
				Value: "val1",
			},
		},
		MacrosToDelete: []*models.Macro{
			{
				Name:  "MAC2",
				Value: "val2",
			},
		},
		ParamsToSet: map[string]string{
			"param1": "val1",
		},
		HostToSet: "foo",
	}

	assert.NotEmpty(t, s.String())
}
