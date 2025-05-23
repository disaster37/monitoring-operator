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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TemplateSpec defines the desired state of Template
// +k8s:openapi-gen=true
type TemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Deprecated: Use full template instead to set the type
	// Type is the object type it generate from template
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Type string `json:"type"`

	// Deprecated: Use the full template instead to set the name
	// Name is the resource name generated from template
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	Name string `json:"name,omitempty"`

	// Template is the template to render. You can use the golang template syntaxe with sprig function
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	Template string `json:"template"`

	// TemplateDelimiter is the delimiter to use when render template
	// It can be usefull if you use helm on top of them
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +optional
	TemplateDelimiter *TemplateTemplateDelimiter `json:"templateDelimiter,omitempty"`
}

type TemplateTemplateDelimiter struct {
	// Right is the right delimiter
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:MinLength:=1
	Right string `json:"right"`

	// Left is the left delimiter
	// +operator-sdk:csv:customresourcedefinitions:type=spec
	// +kubebuilder:validation:MinLength:=1
	Left string `json:"left"`
}

// TemplateStatus defines the observed state of Template
type TemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Fake status to generate bundle manifest without error
	// +operator-sdk:csv:customresourcedefinitions:type=status
	Status string `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

// Template is the Schema for the templates API
// +operator-sdk:csv:customresourcedefinitions:resources={{CentreonService,v1,centreonService},{CentreonServiceGroup,v1,centreonServiceGroup}}
// +kubebuilder:resource:shortName=mtmpl
// +kubebuilder:printcolumn:name="Type",type="string",JSONPath=".spec.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TemplateSpec   `json:"spec,omitempty"`
	Status TemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
