package v1

import (
	"testing"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	"github.com/stretchr/testify/assert"

	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func (t *APITestSuite) TestCentreonServiceCRUD() {
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

func TestCentreonServiceIsValid(t *testing.T) {
	var centreonService *CentreonService

	// When is valid
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "template",
		},
	}
	assert.True(t, centreonService.IsValid())

	// When invalid
	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "",
			Name:     "ping",
			Template: "template",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "",
			Template: "template",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "localhost",
			Name:     "ping",
			Template: "",
		},
	}
	assert.False(t, centreonService.IsValid())

	centreonService = &CentreonService{}
	assert.False(t, centreonService.IsValid())
}

func TestToCentreonService(t *testing.T) {

	s := &CentreonService{
		Spec: CentreonServiceSpec{
			Host:     "host1",
			Name:     "s1",
			Template: "template1",
			Groups:   []string{"group1"},
			Macros: map[string]string{
				"MAC1": "value1",
			},
			Arguments:           []string{"arg1"},
			Categories:          []string{"cat1"},
			CheckCommand:        "check",
			NormalCheckInterval: "1s",
			RetryCheckInterval:  "2s",
			MaxCheckAttempts:    "3s",
			ActiveCheckEnabled:  nil,
			PassiveCheckEnabled: nil,
			Activated:           true,
		},
	}

	expectedS := &centreonhandler.CentreonService{
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

	currentS, err := s.ToCentreonService()
	assert.NoError(t, err)
	assert.Equal(t, expectedS, currentS)
}
