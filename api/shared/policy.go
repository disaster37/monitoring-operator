package shared

// Policy define the policy that controller need to respect when it reconcile resource
type Policy struct {

	// NoDelete is true if controller can't delete resource on remote provider
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NoDelete bool `json:"noDelete,omitempty"`

	// NoCreate is true if controller can't create resource on remote provider
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NoCreate bool `json:"noCreate,omitempty"`

	// NoUpdate is true if controller can't update resource on remote provider
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NoUpdate bool `json:"noUpdate,omitempty"`

	// ExcludeFieldsOnDiff is the list of fields to exclude when diff step is processing
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ExcludeFieldsOnDiff []string `json:"excludeFields,omitempty"`
}
