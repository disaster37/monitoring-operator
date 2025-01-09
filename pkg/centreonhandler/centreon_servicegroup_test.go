package centreonhandler

import (
	"testing"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func (t *CentreonHandlerTestSuite) TestCreateServiceGroup() {
	toCreate := &CentreonServiceGroup{
		Name:        "sg1",
		Description: "my sg",
		Activated:   "1",
		Comment:     "some comments",
	}

	// Mock add serviceGroup on Centreon
	t.mockServiceGroup.EXPECT().
		Add(gomock.Eq("sg1"), gomock.Eq("my sg")).
		Return(nil)

	// Mock set params on Centreon
	t.mockServiceGroup.EXPECT().
		SetParam(gomock.Eq("sg1"), gomock.Eq("activate"), gomock.Eq("1")).
		Return(nil)
	t.mockServiceGroup.EXPECT().
		SetParam(gomock.Eq("sg1"), gomock.Eq("comment"), gomock.Eq("some comments")).
		Return(nil)

	err := t.client.CreateServiceGroup(toCreate)
	assert.NoError(t.T(), err)

	// When use bad parameterr
	err = t.client.CreateServiceGroup(nil)
	assert.Error(t.T(), err)

	// When no name
	toCreate = &CentreonServiceGroup{
		Description: "my sg",
		Activated:   "1",
		Comment:     "some comments",
	}
	err = t.client.CreateServiceGroup(toCreate)
	assert.Error(t.T(), err)

	// When no description
	toCreate = &CentreonServiceGroup{
		Name:      "sg1",
		Activated: "1",
		Comment:   "some comments",
	}
	err = t.client.CreateServiceGroup(toCreate)
	assert.Error(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestUpdateServiceGroup() {
	toUpdate := &CentreonServiceGroupDiff{
		IsDiff: true,
		Name:   "sg1",
		ParamsToSet: map[string]string{
			"param1": "value1",
			"param2": "value2",
			"name":   "sg2",
		},
	}

	// Mock set params on Centreon
	t.mockServiceGroup.EXPECT().
		SetParam(gomock.Any(), gomock.Eq("param1"), gomock.Eq("value1")).
		Return(nil)
	t.mockServiceGroup.EXPECT().
		SetParam(gomock.Any(), gomock.Eq("param2"), gomock.Eq("value2")).
		Return(nil)
	t.mockServiceGroup.EXPECT().
		SetParam(gomock.Eq("sg1"), gomock.Eq("name"), gomock.Eq("sg2")).
		Return(nil)

	err := t.client.UpdateServiceGroup(toUpdate)
	assert.NoError(t.T(), err)

	// When no diff
	toUpdate = &CentreonServiceGroupDiff{
		IsDiff: false,
		Name:   "sg1",
		ParamsToSet: map[string]string{
			"param1": "value1",
			"param2": "value2",
			"name":   "sg2",
		},
	}
	err = t.client.UpdateServiceGroup(toUpdate)
	assert.NoError(t.T(), err)

	// When bad parameters
	err = t.client.UpdateServiceGroup(nil)
	assert.Error(t.T(), err)

	toUpdate = &CentreonServiceGroupDiff{
		IsDiff: true,
		ParamsToSet: map[string]string{
			"param1": "value1",
			"param2": "value2",
			"name":   "sg2",
		},
	}
	err = t.client.UpdateServiceGroup(toUpdate)
	assert.Error(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestDeleteServiceGroup() {
	// Mock delete
	t.mockServiceGroup.EXPECT().
		Delete(gomock.Eq("sg1")).
		Return(nil)

	err := t.client.DeleteServiceGroup("sg1")
	assert.NoError(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestGetServiceGroup() {
	expected := &CentreonServiceGroup{
		Name:        "sg1",
		Description: "my sg",
		Activated:   "1",
		Comment:     "my comment",
	}

	sg := &models.ServiceGroup{
		Name:        "sg1",
		Description: "my sg",
	}

	// Mock get
	t.mockServiceGroup.EXPECT().
		Get(gomock.Eq("sg1")).
		Return(sg, nil)

	// Mock get params
	t.mockServiceGroup.EXPECT().
		GetParam(gomock.Eq("sg1"), []string{"activate", "comment"}).
		Return(map[string]string{"activate": "1", "comment": "my comment"}, nil)

	serviceGroup, err := t.client.GetServiceGroup("sg1")
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), expected, serviceGroup)

	// When not found
	t.mockServiceGroup.EXPECT().
		Get(gomock.Eq("sg1")).
		Return(nil, nil)
	serviceGroup, err = t.client.GetServiceGroup("sg1")
	assert.NoError(t.T(), err)
	assert.Nil(t.T(), serviceGroup)

	// When bad parameters
	_, err = t.client.GetServiceGroup("")
	assert.Error(t.T(), err)
}

func (t *CentreonHandlerTestSuite) TestDiffServiceGroup() {
	tests := []struct {
		Name            string
		ActualService   *CentreonServiceGroup
		ExpectedService *CentreonServiceGroup
		ExpectedDiff    *CentreonServiceGroupDiff
		IgnoreFields    []string
	}{
		{
			Name: "no need update and extra infos is nil",
			ActualService: &CentreonServiceGroup{
				Name: "sg1",
			},
			ExpectedService: &CentreonServiceGroup{
				Name: "sg1",
			},
			ExpectedDiff: &CentreonServiceGroupDiff{
				IsDiff:      false,
				Name:        "sg1",
				ParamsToSet: map[string]string{},
			},
		},
		{
			Name: "No Need update and all properties set",
			ActualService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "1",
				Comment:     "comment2",
				Description: "my sg",
			},
			ExpectedService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "1",
				Comment:     "comment2",
				Description: "my sg",
			},
			ExpectedDiff: &CentreonServiceGroupDiff{
				IsDiff:      false,
				Name:        "sg1",
				ParamsToSet: map[string]string{},
			},
		},
		{
			Name: "Need update all properties",
			ActualService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "0",
				Comment:     "comment",
				Description: "my sg",
			},
			ExpectedService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "1",
				Comment:     "comment2",
				Description: "my sg2",
			},
			ExpectedDiff: &CentreonServiceGroupDiff{
				IsDiff: true,
				Name:   "sg1",
				ParamsToSet: map[string]string{
					"activate": "1",
					"comment":  "comment2",
					"alias":    "my sg2",
				},
			},
		},
		{
			Name: "Need update name",
			ActualService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "0",
				Comment:     "comment",
				Description: "my sg",
			},
			ExpectedService: &CentreonServiceGroup{
				Name:        "sg2",
				Activated:   "0",
				Comment:     "comment",
				Description: "my sg",
			},
			ExpectedDiff: &CentreonServiceGroupDiff{
				IsDiff: true,
				Name:   "sg1",
				ParamsToSet: map[string]string{
					"name": "sg2",
				},
			},
		},
		{
			Name: "Need update all properties but all fields ignored",
			ActualService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "0",
				Comment:     "comment",
				Description: "my sg",
			},
			ExpectedService: &CentreonServiceGroup{
				Name:        "sg1",
				Activated:   "1",
				Comment:     "comment2",
				Description: "my sg2",
			},
			IgnoreFields: []string{
				"name",
				"activate",
				"description",
				"comment",
			},
			ExpectedDiff: &CentreonServiceGroupDiff{
				IsDiff:      false,
				Name:        "sg1",
				ParamsToSet: map[string]string{},
			},
		},
	}

	for _, test := range tests {
		diff, err := t.client.DiffServiceGroup(test.ActualService, test.ExpectedService, test.IgnoreFields)
		assert.NoErrorf(t.T(), err, test.Name)
		assert.Equalf(t.T(), test.ExpectedDiff, diff, test.Name)
	}
}

func TestCentreonServiceGroupToString(t *testing.T) {
	sg := &CentreonServiceGroup{
		Name:        "sg1",
		Description: "desc",
		Activated:   "1",
		Comment:     "foo",
	}

	assert.NotEmpty(t, sg.String())
}

func TestCentreonServiceGroupDiffToString(t *testing.T) {
	sg := &CentreonServiceGroupDiff{
		Name:   "sg1",
		IsDiff: true,
		ParamsToSet: map[string]string{
			"param1": "val1",
		},
	}

	assert.NotEmpty(t, sg.String())
}
