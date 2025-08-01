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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ScaleNotifyMsgTemplateSpec defines the desired state of ScaleNotifyMsgTemplate.
type ScaleNotifyMsgTemplateSpec struct {

	// Title is the message title template.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// where the value must be a string with a minimum length of 1 and a maximum length of 255 characters.
	Title string `json:"title,omitempty"`
	// Content is the message content template. go templates can be used to format the content.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=5000
	// where the value must be a string with a minimum length of 1 and a maximum length of 5000 characters.
	// Example: "Scale notification: {{.ScaleName}}
	Content string `json:"content,omitempty"`
}

// ScaleNotifyMsgTemplateStatus defines the observed state of ScaleNotifyMsgTemplate.
type ScaleNotifyMsgTemplateStatus struct {
	// ValidationStatus indicates the validation status of the message template.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Valid;Invalid;Pending
	ValidationStatus string `json:"validationStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ScaleNotifyMsgTemplate is the Schema for the scalenotifymsgtemplates API.
type ScaleNotifyMsgTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScaleNotifyMsgTemplateSpec   `json:"spec,omitempty"`
	Status ScaleNotifyMsgTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScaleNotifyMsgTemplateList contains a list of ScaleNotifyMsgTemplate.
type ScaleNotifyMsgTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScaleNotifyMsgTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScaleNotifyMsgTemplate{}, &ScaleNotifyMsgTemplateList{})
}
