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

// CentreonServiceSpec defines the desired state of CentreonService
// +k8s:openapi-gen=true
type CentreonServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PlatformRef is the target platform where to create service
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PlatformRef string `json:"platformRef,omitempty"`

	// The service name
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Name string `json:"name"`

	// The host to attach the service
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Host string `json:"host"`

	// The service templates
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Template string `json:"template,omitempty"`

	// The list of service groups
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Groups []string `json:"groups,omitempty"`

	// The map of macros
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Macros map[string]string `json:"macros,omitempty"`

	// The list of arguments
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Arguments []string `json:"arguments,omitempty"`

	// The list of categories
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Categories []string `json:"categories,omitempty"`

	// The check command
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	CheckCommand string `json:"checkCommand,omitempty"`

	// The normal check interval
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	NormalCheckInterval string `json:"normalCheckInterval,omitempty"`

	// The retry check interval
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	RetryCheckInterval string `json:"retryCheckInterval,omitempty"`

	// The max check attemps
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	MaxCheckAttempts string `json:"maxCheckAttempts,omitempty"`

	// The active check enable
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	ActiveCheckEnabled *bool `json:"activeChecksEnabled,omitempty"`

	// The passive check enable
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	PassiveCheckEnabled *bool `json:"passiveChecksEnabled,omitempty"`

	// Activate or disable service
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Activated bool `json:"activate,omitempty"`

	// Policy define the policy that controller need to respect when it reconcile resource
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Policy shared.Policy `json:"policy,omitempty"`
}

// CentreonServiceStatus defines the observed state of CentreonService
type CentreonServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	apis.BasicRemoteObjectStatus `json:",inline"`

	// The host affected to service on Centreon
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Host string `json:"host,omitempty"`

	// The service name
	// +operator-sdk:csv:customresourcedefinitions:type=status
	ServiceName string `json:"serviceName,omitempty"`

	// The platform ref
	// +operator-sdk:csv:customresourcedefinitions:type=status
	PlatformRef string `json:"platformRef,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// CentreonService is the Schema for the centreonservices API
// +operator-sdk:csv:customresourcedefinitions:resources={{None,None,None}}
// +kubebuilder:resource:shortName=mcs
// +kubebuilder:printcolumn:name="Sync",type="boolean",JSONPath=".status.isSync"
// +kubebuilder:printcolumn:name="Error",type="boolean",JSONPath=".status.isOnError",description="Is on error"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status",description="health"
// +kubebuilder:printcolumn:name="Host",type="string",JSONPath=".status.host"
// +kubebuilder:printcolumn:name="Service",type="string",JSONPath=".status.serviceName"
// +kubebuilder:printcolumn:name="Platform",type="string",JSONPath=".status.platformRef"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type CentreonService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CentreonServiceSpec   `json:"spec,omitempty"`
	Status CentreonServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CentreonServiceList contains a list of CentreonService
type CentreonServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CentreonService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CentreonService{}, &CentreonServiceList{})
}
