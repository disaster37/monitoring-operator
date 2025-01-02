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
