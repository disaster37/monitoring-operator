package centreonhandler

import (
	"encoding/json"

	"github.com/disaster37/go-centreon-rest/v21/models"
)

type CentreonService struct {
	Host                string
	Name                string
	CheckCommand        string
	CheckCommandArgs    string
	NormalCheckInterval string
	RetryCheckInterval  string
	MaxCheckAttempts    string
	ActiveCheckEnabled  string
	PassiveCheckEnabled string
	Activated           string
	Template            string
	Comment             string
	Groups              []string
	Categories          []string
	Macros              []*models.Macro
}

type CentreonServiceDiff struct {
	Host               string
	Name               string
	IsDiff             bool
	GroupsToSet        []string
	GroupsToDelete     []string
	CategoriesToSet    []string
	CategoriesToDelete []string
	MacrosToSet        []*models.Macro
	MacrosToDelete     []*models.Macro
	ParamsToSet        map[string]string
	HostToSet          string
}

func (cs *CentreonService) String() string {
	b, err := json.Marshal(cs)
	if err != nil {
		return ""
	}

	return string(b)
}

func (csd *CentreonServiceDiff) String() string {
	b, err := json.Marshal(csd)
	if err != nil {
		return ""
	}

	return string(b)
}
