/*
Copyright 2025.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScaleNotifyConfigSpec defines the desired state of ScaleNotifyConfig.
type ScaleNotifyConfigSpec struct {
	// Default indicates whether this configuration is the default one.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=boolean
	// where the value must be a boolean.
	// +kubebuilder:default=false
	// Example: true
	Default bool `json:"default,omitempty"`
	// Type indicates the type of scaling notification.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Email;WXWorkRobot
	Type string `json:"type,omitempty"`
	// Config contains the configuration details for the scaling notification.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=object
	Config runtime.RawExtension `json:"config,omitempty"`
}

// ScaleNotifyConfigStatus defines the observed state of ScaleNotifyConfig.
type ScaleNotifyConfigStatus struct {
	// ValidationStatus indicates the validation status of the configuration.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Valid;Invalid;Pending
	// where the value must be one of the predefined statuses.
	// +kubebuilder:default=Pending
	// Example: Valid
	ValidationStatus string `json:"validationStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=scn;
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Default",type=boolean,JSONPath=`.spec.default`
// +kubebuilder:printcolumn:name="Validation Status",type=string,JSONPath=`.status.validationStatus`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// ScaleNotifyConfig is the Schema for the scalenotifyconfigs API.
type ScaleNotifyConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScaleNotifyConfigSpec   `json:"spec,omitempty"`
	Status ScaleNotifyConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScaleNotifyConfigList contains a list of ScaleNotifyConfig.
type ScaleNotifyConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScaleNotifyConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScaleNotifyConfig{}, &ScaleNotifyConfigList{})
}

// ValidateDefault validates that only one default config exists per type
func (c *ScaleNotifyConfig) ValidateDefault() error {
	// This method can be used by webhooks or other validation logic
	// The actual validation logic is implemented in the controller
	return nil
}
