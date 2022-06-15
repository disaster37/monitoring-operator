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

// PlatformSpec defines the desired state of Platform
type PlatformSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Name is the unique name for platform
	Name string `json:"name"`

	// IsDefault is set to tru to use this plateform when is not specify on resource to create
	IsDefault bool `json:"isDefault"`

	// PlatformType is the platform type.
	// It support only `centreon` at this time
	PlatformType string `json:"type"`

	// CentreonSettings is the setting for Centreon plateform type
	// +optional
	CentreonSettings *PlatformSpecCentreonSettings `json:"centreonSettings,omitempty"`
}

type PlatformSpecCentreonSettings struct {

	// URL is the full URL to access on Centreon API
	URL string `json:"url"`

	// SelfSignedCertificat is true if you shouldn't check Centreon API certificate
	SelfSignedCertificate bool `json:"selfSignedCertificat"`

	// Secret is the secret that store the username and password to access on Centreon API
	// It need to have `username` and `password` key
	Secret string `json:"secret"`

	// The endpoint default setting
	// +optional
	Endpoint *CentreonSpecEndpoint `json:"endpoint,omitempty"`
}

// General configuration setting when handle monitring service from endpoint (Ingress / Route)
// +k8s:openapi-gen=true
type CentreonSpecEndpoint struct {

	// The default service template to use when create service from endpoint
	// It normally optional, but Centreon bug impose to set an existed template
	Template string `json:"template"`

	// The default template name when create service from endpoint
	// +optional
	NameTemplate string `json:"nameTemplate,omitempty"`

	// The default host to attach service
	// +optional
	DefaultHost string `json:"defaultHost,omitempty"`

	// The default macro to set when create service
	// You can use special tag to generate value on the flow
	// +optional
	Macros map[string]string `json:"macros,omitempty"`

	// The default command arguements to set when create service
	// You can use special tag to generate value on the flow
	// +optional
	Arguments []string `json:"arguments,omitempty"`

	// By default, activate service when created it
	// Default to true
	// +optional
	ActivateService bool `json:"activeService,omitempty"`

	// Default service groups
	// +optional
	ServiceGroups []string `json:"serviceGroups,omitempty"`

	// Default categories
	// +optional
	Categories []string `json:"categories,omitempty"`
}

// PlatformStatus defines the observed state of Platform
type PlatformStatus struct {
	Conditions []metav1.Condition `json:"conditions"`

	// SecretHash contain a secret hash to help controller to know if secret has been changed
	// +optional
	SecretHash string `json:"secretHash,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Platform is the Schema for the platforms API
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
