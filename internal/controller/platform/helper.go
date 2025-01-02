package platform

import (
	"emperror.dev/errors"
	monitorapi "github.com/disaster37/monitoring-operator/api/v1"
)

// GetClient premit to get client to connect on monitoring platform
func GetClient(platformRef string, platforms map[string]*ComputedPlatform) (meta any, platform *monitorapi.Platform, err error) {
	if platformRef == "" {
		if p, ok := platforms["default"]; ok {
			return p.Client, p.Platform, nil
		}

		return nil, nil, errors.New("No default platform")
	}

	if p, ok := platforms[platformRef]; ok {
		return p.Client, p.Platform, nil
	}

	return nil, nil, errors.Errorf("Platform %s not found", platformRef)
}
