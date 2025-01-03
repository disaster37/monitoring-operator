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

package v1

import (
	"github.com/disaster37/monitoring-operator/api/shared"
	"github.com/disaster37/operator-sdk-extra/pkg/apis"
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

	// Policy define the policy that controller need to respect when it reconcile resource
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Policy shared.Policy `json:"policy,omitempty"`
}

// CentreonServiceGroupStatus defines the observed state of CentreonServiceGroup
type CentreonServiceGroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.BasicRemoteObjectStatus `json:",inline"`

	// The service group name
	// +operator-sdk:csv:customresourcedefinitions:type=status
	ServiceGroupName string `json:"serviceGroupName,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CentreonServiceGroup is the Schema for the centreonservicegroups API
// +operator-sdk:csv:customresourcedefinitions:resources={{None,None,None}}
// +kubebuilder:resource:shortName=mcsg
// +kubebuilder:printcolumn:name="Sync",type="boolean",JSONPath=".status.isSync"
// +kubebuilder:printcolumn:name="Error",type="boolean",JSONPath=".status.isOnError",description="Is on error"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="health"
// +kubebuilder:printcolumn:name="ServiceGroup",type="string",JSONPath=".status.serviceGroupName"
// +kubebuilder:printcolumn:name="Platform",type="string",JSONPath=".spec.platformRef"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
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
