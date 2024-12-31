package centreon

import "github.com/disaster37/monitoring-operator/pkg/centreonhandler"

// CentreonServiceGroup wrap the original model because we haven't unique model on each step.
// Sometime, we need to have centreonhandler.CentreonServiceGroup, sometime we need to have centreonhandler.CentreonServiceGroupDiff
type CentreonServiceGroup struct {
	*centreonhandler.CentreonServiceGroup
	*centreonhandler.CentreonServiceGroupDiff
}
