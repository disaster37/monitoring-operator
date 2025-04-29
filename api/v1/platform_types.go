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
	"github.com/disaster37/operator-sdk-extra/pkg/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PlatformSpec defines the desired state of Platform
// +k8s:openapi-gen=true
type PlatformSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// IsDefault is set to tru to use this plateform when is not specify on resource to create
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	IsDefault bool `json:"isDefault"`

	// PlatformType is the platform type.
	// It support only `centreon` at this time
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:Enum=centreon
	PlatformType string `json:"type"`

	// CentreonSettings is the setting for Centreon plateform type
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	CentreonSettings *PlatformSpecCentreonSettings `json:"centreonSettings,omitempty"`

	// Debug permit to enable debug log on client that call the plateform API
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Debug *bool `json:"debug,omitempty"`
}

type PlatformSpecCentreonSettings struct {
	// URL is the full URL to access on Centreon API
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	URL string `json:"url"`

	// SelfSignedCertificat is true if you shouldn't check Centreon API certificate
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	SelfSignedCertificate bool `json:"selfSignedCertificat"`

	// Secret is the secret that store the username and password to access on Centreon API
	// It need to have `username` and `password` key
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Secret string `json:"secret"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	apis.BasicRemoteObjectStatus `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Platform is the Schema for the platforms API
// +operator-sdk:csv:customresourcedefinitions:resources={{Secret,v1,monitoring-credentials}}
// +kubebuilder:printcolumn:name="Sync",type="boolean",JSONPath=".status.isSync"
// +kubebuilder:printcolumn:name="Error",type="boolean",JSONPath=".status.isOnError",description="Is on error"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="health"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Platform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlatformSpec   `json:"spec,omitempty"`
	Status PlatformStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlatformList contains a list of Platform
type PlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Platform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Platform{}, &PlatformList{})
}
