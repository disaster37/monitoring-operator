package centreon

import "github.com/disaster37/monitoring-operator/pkg/centreonhandler"

// CentreonService wrap the original model because we haven't unique model on each step.
// Sometime, we need to have centreonhandler.CentreonService, sometime we need to have centreonhandler.CentreonServiceDiff
type CentreonService struct {
	*centreonhandler.CentreonService
	*centreonhandler.CentreonServiceDiff
}
