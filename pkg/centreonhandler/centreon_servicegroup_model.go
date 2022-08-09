package centreonhandler

import (
	"encoding/json"
)

type CentreonServiceGroup struct {
	Name        string
	Activated   string
	Comment     string
	Description string
}

type CentreonServiceGroupDiff struct {
	Name        string
	IsDiff      bool
	ParamsToSet map[string]string
}

func (cs *CentreonServiceGroup) String() string {
	b, err := json.Marshal(cs)
	if err != nil {
		return ""
	}

	return string(b)
}

func (csd *CentreonServiceGroupDiff) String() string {
	b, err := json.Marshal(csd)
	if err != nil {
		return ""
	}

	return string(b)
}
