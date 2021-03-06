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
	"strings"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/pkg/centreonhandler"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
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
}

// CentreonServiceStatus defines the observed state of CentreonService
type CentreonServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The host affected to service on Centreon
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Host string `json:"host,omitempty"`

	// The service name
	// +operator-sdk:csv:customresourcedefinitions:type=status
	ServiceName string `json:"serviceName,omitempty"`

	// List of conditions
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Conditions []metav1.Condition `json:"conditions"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CentreonService is the Schema for the centreonservices API
// +operator-sdk:csv:customresourcedefinitions:resources={{None,None,None}}
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

// IsValid check Centreon service is valid for Centreon
func (c *CentreonService) IsValid() bool {
	if c.Spec.Host == "" || c.Spec.Name == "" || c.Spec.Template == "" {
		return false
	}

	return true
}

func (h *CentreonService) ToCentreoonService() (*centreonhandler.CentreonService, error) {
	cs := &centreonhandler.CentreonService{
		Host:                h.Spec.Host,
		Name:                h.Spec.Name,
		CheckCommand:        h.Spec.CheckCommand,
		CheckCommandArgs:    helpers.CheckArgumentsToString(h.Spec.Arguments),
		NormalCheckInterval: h.Spec.NormalCheckInterval,
		RetryCheckInterval:  h.Spec.RetryCheckInterval,
		MaxCheckAttempts:    h.Spec.MaxCheckAttempts,
		ActiveCheckEnabled:  helpers.BoolToString(h.Spec.ActiveCheckEnabled),
		PassiveCheckEnabled: helpers.BoolToString(h.Spec.PassiveCheckEnabled),
		Activated:           helpers.BoolToString(&h.Spec.Activated),
		Template:            h.Spec.Template,
		Comment:             "Managed by monitoring-operator",
		Groups:              h.Spec.Groups,
		Categories:          h.Spec.Categories,
		Macros:              make([]*models.Macro, 0, len(h.Spec.Macros)),
	}
	for name, value := range h.Spec.Macros {
		macro := &models.Macro{
			Name:       strings.ToUpper(name),
			Value:      value,
			IsPassword: "0",
		}
		cs.Macros = append(cs.Macros, macro)
	}

	return cs, nil

}
