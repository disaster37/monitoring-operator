package centreon

import (
	"testing"

	centreoncrd "github.com/disaster37/monitoring-operator/api/v1"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/stretchr/testify/assert"
)

func TestCentreonServiceGroupBuild(t *testing.T) {
	client := &centreonServiceGroupApiClient{}

	o := &centreoncrd.CentreonServiceGroup{
		Spec: centreoncrd.CentreonServiceGroupSpec{
			Name:        "sg1",
			Description: "my sg",
			Activated:   true,
		},
	}

	expectedCSG := &centreonhandler.CentreonServiceGroup{
		Name:        "sg1",
		Description: "my sg",
		Activated:   "1",
		Comment:     "Managed by monitoring-operator",
	}

	csg, err := client.Build(o)
	assert.NoError(t, err)
	assert.Equal(t, expectedCSG, csg.CentreonServiceGroup)
}
