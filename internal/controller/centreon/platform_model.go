package centreon

import centreoncrd "github.com/disaster37/monitoring-operator/api/v1"

type ComputedPlatform struct {
	client   any
	platform *centreoncrd.Platform
	hash     string
}
