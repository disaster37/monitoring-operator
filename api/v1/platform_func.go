package v1

import "github.com/disaster37/operator-sdk-extra/pkg/object"

// GetStatus implement the object.MultiPhaseObject
func (h *Platform) GetStatus() object.RemoteObjectStatus {
	return &h.Status
}

// GetExternalName return the role name
// If name is empty, it use the ressource name
func (o *Platform) GetExternalName() string {
	return o.Name
}


// IsDebug return true if debug field is true
// else it return false
func (h *Platform) IsDebug() bool {
	if h.Spec.Debug != nil && *h.Spec.Debug {
		return true
	}

	return false
}