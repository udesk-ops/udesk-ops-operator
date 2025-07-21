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

// AlertSpec defines the desired state of Alert.
type AlertSpec struct {
	// AlertStatus is the status of the alert.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=active;resolved
	// where the value must be one of the predefined statuses.
	AlertStatus string `json:"alertStatus,omitempty"`

	// Description is a brief description of the alert.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MaxLength=256
	// where the length must not exceed 256 characters.
	Description string `json:"description,omitempty"`
}

// AlertStatus defines the observed state of Alert.
type AlertStatus struct {
	// ScaleStatus indicates the current scaling status of the alert.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=pending;scaling;completed;failed
	// where the value must be one of the predefined statuses.
	ScaleStatus string `json:"scaleStatus,omitempty"`

	// ScaleTarget is the target resource for scaling operations.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Optional
	ScaleTarget ScaleTarget `json:"scaleTarget,omitempty"`

	// ScaleDuration is the duration for which the scaling should be applied.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^(\d+)([smhdw])$`
	// where s=seconds, m=minutes, h=hours, d=days, w=weeks
	ScaleDuration string `json:"scaleDuration,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Alert Status",type=string,JSONPath=`.spec.alertStatus`
// +kubebuilder:printcolumn:name="Name",type=string,JSONPath=`.metadata.name`
// +kubebuilder:printcolumn:name="Description",type=string,JSONPath=`.spec.description`
// +kubebuilder:printcolumn:name="Scale Target",type=string,JSONPath=`.status.scaleTarget.name`
// +kubebuilder:printcolumn:name="Scale Duration",type=string,JSONPath=`.status.scaleDuration`
// Alert is the Schema for the alerts API.
type Alert struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertSpec   `json:"spec,omitempty"`
	Status AlertStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertList contains a list of Alert.
type AlertList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Alert `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Alert{}, &AlertList{})
}
