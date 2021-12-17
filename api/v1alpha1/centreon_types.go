/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CentreonSpec defines the desired state of Centreon
// +k8s:openapi-gen=true
type CentreonSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The endpoint default setting
	Endpoints *CentreonSpecEndpoint `json:"endpoint"`
}

// General configuration setting when handle monitring service from endpoint (Ingress / Route)
// +k8s:openapi-gen=true
type CentreonSpecEndpoint struct {

	// The default service template to use when create service from endpoint
	// It normally optional, but Centreon bug impose to set an existed template
	Template string `json:"template"`

	// The default template name when create service from endpoint
	// +optional
	NameTemplate string `json:"nameTemplate"`

	// The default host to attach service
	// +optional
	DefaultHost string `json:"defaultHost"`

	// The default macro to set when create service
	// You can use special tag to generate value on the flow
	// +optional
	Macros map[string]string `json:"macros;omitempty"`

	// The default command arguements to set when create service
	// You can use special tag to generate value on the flow
	// +optional
	Arguments []string `json:"args;omitempty"`

	// By default, activate service when created it
	// Default to true
	// +optional
	ActivateService bool `json:"activeService;omitempty"`

	// Default service groups
	// +optional
	ServiceGroups []string `json:"serviceGroups;omitempty"`

	// Default categories
	// +optional
	Categories []string `json:"categories;omitempty"`
}

// CentreonStatus defines the observed state of Centreon
type CentreonStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The date when item is handle by operator
	UpdatedAt string `json:"updatedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Centreon is the Schema for the centreons API
type Centreon struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CentreonSpec   `json:"spec,omitempty"`
	Status CentreonStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CentreonList contains a list of Centreon
type CentreonList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Centreon `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Centreon{}, &CentreonList{})
}

// IsBeingDeleted returns true if a deletion timestamp is set
func (c *Centreon) IsBeingDeleted() bool {
	return !c.ObjectMeta.DeletionTimestamp.IsZero()
}

// IsSubmitted return true if service has been submitted to Centreon
func (c *Centreon) IsSubmitted() bool {
	return c.Status.UpdatedAt != ""
}
