package centreon

import (
	"testing"

	"github.com/disaster37/go-centreon-rest/v21/models"
	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
)

func TestCentreonServiceBuild(t *testing.T) {

	client := &centreonServiceApiClient{}

	o := &centreoncrd.CentreonService{
		Spec: centreoncrd.CentreonServiceSpec{
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

	expectedCS := &centreonhandler.CentreonService{
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

	cs, err := client.Build(o)
	assert.NoError(t, err)
	assert.Equal(t, expectedCS, cs.CentreonService)


}
