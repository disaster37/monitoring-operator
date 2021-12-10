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
	"reflect"

	"github.com/disaster37/go-centreon-rest/v21/models"
	"github.com/disaster37/monitoring-operator/pkg/helpers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const centreonServicedFinalizer = "service.monitor.k8s.webcenter.fr/finalizer"

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CentreonServiceSpec defines the desired state of CentreonService
// +k8s:openapi-gen=true
type CentreonServiceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The service name
	Name string `json:"name"`

	// The host to attach the service
	Host string `json:"host"`

	// The service templates
	// +optional
	Template string `json:"template,omitempty"`

	// The list of service groups
	// +optional
	Groups []string `json:"groups,omitempty"`

	// The map of macros
	// +optional
	Macros map[string]string `json:"macros,omitempty"`

	// The list of arguments
	// +optional
	Arguments []string `json:"arguments,omitempty"`

	// The list of categories
	// +optional
	Categories []string `json:"categories,omitempty"`

	// The check command
	// +optional
	CheckCommand string `json:"checkCommand,omitempty"`

	// The normal check interval
	// +optional
	NormalCheckInterval string `json:"normalCheckInterval,omitempty"`

	// The retry check interval
	// +optional
	RetryCheckInterval string `json:"retryCheckInterval,omitempty"`

	// The max check attemps
	// +optional
	MaxCheckAttempts string `json:"maxCheckAttempts,omitempty"`

	// The active check enable
	// +optional
	ActiveCheckEnabled *bool `json:"activeChecksEnabled,omitempty"`

	// The passive check enable
	// +optional
	PassiveCheckEnabled *bool `json:"passiveChecksEnabled,omitempty"`

	// Activate or disable service
	// +optional
	Activated bool `json:"activate,omitempty"`
}

// CentreonServiceStatus defines the observed state of CentreonService
type CentreonServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// The service ID on Centreon
	ID string `json:"id,omitempty"`

	// The date when service is modified on Centreon by operator
	UpdatedAt string `json:"updatedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// CentreonService is the Schema for the centreonservices API
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

// IsBeingDeleted returns true if a deletion timestamp is set
func (c *CentreonService) IsBeingDeleted() bool {
	return !c.ObjectMeta.DeletionTimestamp.IsZero()
}

// HasFinalizer returns true if the item has the specified finalizer
func (c *CentreonService) HasFinalizer() bool {
	return controllerutil.ContainsFinalizer(c, centreonServicedFinalizer)
}

// IsSubmitted return true if service has been submitted to Centreon
func (c *CentreonService) IsSubmitted() bool {
	return c.Status.ID != ""
}

// AddFinalizer adds the specified finalizer
func (c *CentreonService) AddFinalizer() {
	controllerutil.AddFinalizer(c, centreonServicedFinalizer)
}

// RemoveFinalizer removes the specified finalizer
func (c *CentreonService) RemoveFinalizer() {
	controllerutil.RemoveFinalizer(c, centreonServicedFinalizer)
}

func (c *CentreonService) NeedUpdate(actualService *models.ServiceGet, params map[string]string, actualGroups, actualCategories []string, actualMacros []*models.Macro) bool {

	expectedService := &models.ServiceGet{
		ServiceBaseGet: &models.ServiceBaseGet{
			ID:                  actualService.ID,
			HostId:              actualService.HostId,
			HostName:            c.Spec.Host,
			Name:                c.Spec.Name,
			CheckCommand:        c.Spec.CheckCommand,
			CheckCommandArgs:    helpers.CheckArgumentsToString(c.Spec.Arguments),
			NormalCheckInterval: c.Spec.NormalCheckInterval,
			RetryCheckInterval:  c.Spec.RetryCheckInterval,
			MaxCheckAttempts:    c.Spec.MaxCheckAttempts,
			ActiveCheckEnabled:  helpers.BoolToString(c.Spec.ActiveCheckEnabled),
			PassiveCheckEnabled: helpers.BoolToString(c.Spec.PassiveCheckEnabled),
			Activated:           helpers.BoolToString(&c.Spec.Activated),
		},
	}
	if !reflect.DeepEqual(actualService, expectedService) {
		return true
	}

	// Check params
	actualParams := map[string]string{}
	if c.Spec.Template != "" {
		actualParams["template"] = c.Spec.Template
	}
	if params == nil {
		params = map[string]string{}
	}
	if !reflect.DeepEqual(actualParams, params) {
		return true
	}

	// Check groups
	if !reflect.DeepEqual(actualGroups, c.Spec.Groups) {
		return true
	}

	// Check categories
	if !reflect.DeepEqual(actualCategories, c.Spec.Categories) {
		return true
	}

	// Check macro
	var expectedMacros []*models.Macro = nil
	if c.Spec.Macros != nil {
		expectedMacros = make([]*models.Macro, 0, len(c.Spec.Macros))
		for name, value := range c.Spec.Macros {
			macro := &models.Macro{
				Name:       name,
				Value:      value,
				IsPassword: "0",
			}
			expectedMacros = append(expectedMacros, macro)
		}
	}

	if !reflect.DeepEqual(actualMacros, expectedMacros) {
		return true
	}

	return false
}
