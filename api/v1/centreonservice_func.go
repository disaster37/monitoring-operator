package v1

import (
	"github.com/disaster37/operator-sdk-extra/pkg/helper"
	"github.com/disaster37/operator-sdk-extra/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetStatus return the status object
func (o *CentreonService) GetStatus() object.RemoteObjectStatus {
	return &o.Status
}

// GetExternalName return the role name
// If name is empty, it use the ressource name
func (o *CentreonService) GetExternalName() string {
	if o.Spec.Name == "" {
		return o.Name
	}

	return o.Spec.Name
}

func (o *CentreonService) GetPlatform() string {
	if o.Spec.PlatformRef == "" {
		return "default"
	}

	return o.Spec.PlatformRef
}

// IsValid check Centreon service is valid for Centreon
func (c *CentreonService) IsValid() bool {
	if c.Spec.Host == "" || c.Spec.Name == "" || c.Spec.Template == "" {
		return false
	}

	return true
}

// GetItems permit to get items
func (o *CentreonServiceList) GetItems() []client.Object {
	return helper.ToSliceOfObject(o.Items)
}
