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
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CentreonServiceGroupSpec defines the desired state of CentreonServiceGroup
// +k8s:openapi-gen=true
type CentreonServiceGroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PlatformRef is the target platform where to create serviceGroup
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PlatformRef string `json:"platformRef,omitempty"`

	// The serviceGroup name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Name string `json:"name"`

	// The serviceGroup description
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Description string `json:"description"`

	// Activate or disable service
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Activated bool `json:"activate,omitempty"`
}

// CentreonServiceGroupStatus defines the observed state of CentreonServiceGroup
type CentreonServiceGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The service group name
	// +operator-sdk:csv:customresourcedefinitions:type=status
	ServiceGroupName string `json:"serviceGroupName,omitempty"`

	// List of conditions
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CentreonServiceGroup is the Schema for the centreonservicegroups API
// +operator-sdk:csv:customresourcedefinitions:resources={{None,None,None}}
type CentreonServiceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CentreonServiceGroupSpec   `json:"spec,omitempty"`
	Status CentreonServiceGroupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CentreonServiceGroupList contains a list of CentreonServiceGroup
type CentreonServiceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CentreonServiceGroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CentreonServiceGroup{}, &CentreonServiceGroupList{})
}

// ToCentreonServiceGroup permit to convert current spec to centreonServiceGroup object
func (h *CentreonServiceGroup) ToCentreonServiceGroup() (*centreonhandler.CentreonServiceGroup, error) {
	csg := &centreonhandler.CentreonServiceGroup{
		Name:        h.Spec.Name,
		Activated:   helpers.BoolToString(&h.Spec.Activated),
		Comment:     "Managed by monitoring-operator",
		Description: h.Spec.Description,
	}

	return csg, nil

}
