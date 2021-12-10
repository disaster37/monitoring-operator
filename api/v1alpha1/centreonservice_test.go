package v1alpha1

import (
	"time"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *V1alpha1TestSuite) TestCentreonServiceCRUD() {
	var (
		key              types.NamespacedName
		created, fetched *CentreonService
		err              error
	)

	key = types.NamespacedName{
		Name:      "foo-" + helpers.RandomString(5),
		Namespace: "default",
	}

	// Create object
	created = &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: CentreonServiceSpec{
			Host: "central",
			Name: "ping",
		},
	}
	err = t.k8sClient.Create(context.Background(), created)
	assert.NoError(t.T(), err)

	// Get object
	fetched = &CentreonService{}
	err = t.k8sClient.Get(context.Background(), key, fetched)
	assert.NoError(t.T(), err)
	assert.Equal(t.T(), created, fetched)

	// Delete object
	err = t.k8sClient.Delete(context.Background(), created)
	assert.NoError(t.T(), err)
	err = t.k8sClient.Get(context.Background(), key, created)
	assert.Error(t.T(), err)
}

func (t *V1alpha1TestSuite) TestCentreonServiceIsSubmitted() {
	centreonService := &CentreonService{}
	assert.False(t.T(), centreonService.IsSubmitted())

	centreonService.Status.ID = "test"
	assert.True(t.T(), centreonService.IsSubmitted())
}

func (t *V1alpha1TestSuite) TestCentreonServiceIsBeingDeleted() {
	centreonService := &CentreonService{
		ObjectMeta: metav1.ObjectMeta{
			DeletionTimestamp: &metav1.Time{
				Time: time.Now(),
			},
		},
	}
	assert.True(t.T(), centreonService.IsBeingDeleted())
}

func (t *V1alpha1TestSuite) TestCentreonServiceFinalizer() {
	centreonService := &CentreonService{}

	centreonService.AddFinalizer()
	assert.Equal(t.T(), 1, len(centreonService.GetFinalizers()))
	assert.True(t.T(), centreonService.HasFinalizer())

	centreonService.RemoveFinalizer()
	assert.Equal(t.T(), 0, len(centreonService.GetFinalizers()))
	assert.False(t.T(), centreonService.HasFinalizer())
}

func (t *V1alpha1TestSuite) TestCentreonServiceNeedUpdate() {

	// When no need update and extra infos is nil
	centreonService := &CentreonService{
		Spec: CentreonServiceSpec{
			Host:      "central",
			Name:      "ping",
			Activated: true,
		},
	}
	service := &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "default",
			PassiveCheckEnabled: "default",
			Activated:           "1",
		},
	}
	assert.False(t.T(), centreonService.NeedUpdate(service, nil, nil, nil, nil))

	// When no need update with all extra infos
	trueValue := true
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1", "sg2"},
			Categories:          []string{"cat1", "cat2"},
			Macros: map[string]string{
				"macro1": "value1",
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups := []string{"sg1", "sg2"}
	cats := []string{"cat1", "cat2"}
	macros := []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params := map[string]string{
		"template": "template",
	}
	assert.False(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))

	// When service change
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1", "sg2"},
			Categories:          []string{"cat1", "cat2"},
			Macros: map[string]string{
				"macro1": "value1",
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups = []string{"sg1", "sg2"}
	cats = []string{"cat1", "cat2"}
	macros = []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params = map[string]string{
		"template": "template",
	}
	assert.True(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))

	// When params change
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template2",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1", "sg2"},
			Categories:          []string{"cat1", "cat2"},
			Macros: map[string]string{
				"macro1": "value1",
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups = []string{"sg1", "sg2"}
	cats = []string{"cat1", "cat2"}
	macros = []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params = map[string]string{
		"template": "template",
	}
	assert.True(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))

	// When groups change
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1"},
			Categories:          []string{"cat1", "cat2"},
			Macros: map[string]string{
				"macro1": "value1",
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups = []string{"sg1", "sg2"}
	cats = []string{"cat1", "cat2"}
	macros = []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params = map[string]string{
		"template": "template",
	}
	assert.True(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))

	// When category change
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1", "sg2"},
			Categories:          []string{"cat2"},
			Macros: map[string]string{
				"macro1": "value1",
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups = []string{"sg1", "sg2"}
	cats = []string{"cat1", "cat2"}
	macros = []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params = map[string]string{
		"template": "template",
	}
	assert.True(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))

	// When macro change
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:                "central",
			Name:                "ping",
			Activated:           true,
			Template:            "template",
			Arguments:           []string{"arg1", "arg2"},
			CheckCommand:        "ping",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
			ActiveCheckEnabled:  &trueValue,
			PassiveCheckEnabled: &trueValue,
			Groups:              []string{"sg1", "sg2"},
			Categories:          []string{"cat1", "cat2"},
			Macros: map[string]string{
				"macro2": "value2",
			},
		},
	}
	service = &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			HostName:            "central",
			Name:                "ping",
			ActiveCheckEnabled:  "1",
			PassiveCheckEnabled: "1",
			Activated:           "1",
			ID:                  "id",
			HostId:              "id",
			CheckCommand:        "ping",
			CheckCommandArgs:    "!arg1!arg2",
			NormalCheckInterval: "30s",
			RetryCheckInterval:  "1s",
			MaxCheckAttempts:    "5",
		},
	}
	groups = []string{"sg1", "sg2"}
	cats = []string{"cat1", "cat2"}
	macros = []*models.Macro{
		{
			Name:       "macro1",
			Value:      "value1",
			IsPassword: "0",
		},
		{
			Name:       "macro2",
			Value:      "value2",
			IsPassword: "0",
		},
	}
	params = map[string]string{
		"template": "template",
	}
	assert.True(t.T(), centreonService.NeedUpdate(service, params, groups, cats, macros))
}
