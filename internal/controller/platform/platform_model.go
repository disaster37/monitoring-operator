package platform

import centreoncrd "github.com/disaster37/monitoring-operator/api/v1"

type ComputedPlatform struct {
	Client   any
	Platform *centreoncrd.Platform
	Hash     string
}
