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

type ScaleStatus struct {
	// Status indicates the current status of the scaling operation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Pending;Scaling;Scaled;Failed;Completed;Archived;Approvaling;Approved;Rejected
	// where the value must be one of the predefined statuses.
	Status string `json:"status,omitempty"`
	// ScaleBeginTime is the time when the scaling operation began.
	// +kuberbuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	// where the value must be in RFC3339 format.
	// Example: "2023-10-01T12:00:00Z"
	ScaleBeginTime metav1.Time `json:"scaleBeginTime,omitempty"`
	// ScaleEndTime is the time when the scaling operation ended.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Format=date-time
	// where the value must be in RFC3339 format.
	// Example: "2023-10-01T12:00:00Z"
	ScaleEndTime metav1.Time `json:"scaleEndTime,omitempty"`
	// OriginReplicas is the original number of replicas before scaling.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Minimum=0
	// where the value must be a non-negative integer.
	OriginReplicas int32 `json:"originReplicas,omitempty"`
	// ScaledReplicas is the number of replicas after scaling.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Minimum=0
	// where the value must be a non-negative integer.
	ScaledReplicas int32 `json:"scaledReplicas,omitempty"`
}

// ScaleTarget defines the target resource for scaling operations.
type ScaleTarget struct {
	// Name is the name of the target resource.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9]([-a-zA-Z0-9]*[a-zA-Z0-9])?$`
	// where a-z, A-Z, 0-9, and '-' are allowed,
	// and must start and end with an alphanumeric character.
	Name string `json:"name,omitempty"`
	// Namespace is the namespace of the target resource.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	// where a-z, 0-9, and '-' are allowed,
	// and must start and end with a lowercase alphanumeric character.
	Namespace string `json:"namespace,omitempty"`
	// Kind is the kind of the target resource (e.g., Deployment, StatefulSet).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[A-Z][a-zA-Z0-9]*$`
	// where the first character is uppercase and the rest are alphanumeric
	Kind string `json:"kind,omitempty"`
	// APIVersion is the API version of the target resource.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`
	// where the first part is the group and the second part is the version,
	// both can contain alphanumeric characters, dots, underscores, and hyphens.
	APIVersion string `json:"apiVersion,omitempty"`
	// Labels is a map of labels for the target resource.
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AlertScaleSpec defines the desired state of AlertScale.
type AlertScaleSpec struct {
	// ScaleReason is the reason for scaling, e.g., "High CPU Usage".
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MaxLength=1024
	// where the length must not exceed 1024 characters.
	ScaleReason string `json:"scaleReason,omitempty"` // Reason for scaling, e.g., "High CPU Usage"
	// ScaleDuration is the duration for which the scaling should be applied.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^(\d+)([smhdw])$`
	// where s=seconds, m=minutes, h=hours, d=days, w=weeks
	ScaleDuration string `json:"scaleDuration,omitempty"`
	// ScaleTarget is the target resource for scaling.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=object
	// +kubebuilder:validation:Required
	ScaleTarget ScaleTarget `json:"scaleTarget,omitempty"`
	// ScaleThreshold is the threshold for scaling alerts.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=number
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	ScaleThreshold int32 `json:"scaleThreshold,omitempty"`
	// ScaleNotification defines the notification settings for scaling alerts.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Email;WXWorkRobot
	ScaleNotificationType string `json:"scaleNotificationType,omitempty"`
	// ScaleNotifyMsgTemplate is the reference to the message template for notifications.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	// where the value must be a valid Kubernetes resource name.
	// Example: "my-notification-template"
	ScaleNotifyMsgTemplate string `json:"scaleNotifyMsgTemplate,omitempty"`
	// ScaleTimeout is the timeout for the scaling operation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^(\d+)([smhdw])$`
	// where s=seconds, m=minutes, h=hours, d=days, w=weeks
	ScaleTimeout string `json:"scaleTimeout,omitempty"`

	// ScaleAutoApproval indicates whether the scaling operation requires auto-approval.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=boolean
	// +kubebuilder:default=false
	// Example: true
	ScaleAutoApproval bool `json:"scaleAutoApproval,omitempty"`
}

// AlertScaleStatus defines the observed state of AlertScale.
type AlertScaleStatus struct {
	// ScaleStatus is the status of the scaling operation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=object
	ScaleStatus ScaleStatus `json:"scaleStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=as;ascale
// +kubebuilder:printcolumn:name="Target",type=string,JSONPath=`.spec.scaleTarget.name`
// +kubebuilder:printcolumn:name="AutoApproval",type=boolean,JSONPath=`.spec.scaleAutoApproval`
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.scaleStatus.status`
// +kubebuilder:printcolumn:name="Origin-Replicas",type=integer,JSONPath=`.status.scaleStatus.originReplicas`
// +kubebuilder:printcolumn:name="Scaled-Replicas",type=integer,JSONPath=`.status.scaleStatus.scaledReplicas`
// +kubebuilder:printcolumn:name="Scaled-Duration",type=string,JSONPath=`.spec.scaleDuration`
// +kubebuilder:printcolumn:name="Threshold",type=integer,JSONPath=`.spec.scaleThreshold`
// +kubebuilder:printcolumn:name="NotificationType",type=string,JSONPath=`.spec.scaleNotificationType`
// +kubebuilder:printcolumn:name="MsgTemplate",type=string,JSONPath=`.spec.scaleNotifyMsgTemplate`
// +kubebuilder:printcolumn:name="Reason",type=string,JSONPath=`.spec.scaleReason`
// AlertScale is the Schema for the alertscales API.
type AlertScale struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AlertScaleSpec   `json:"spec,omitempty"`
	Status AlertScaleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AlertScaleList contains a list of AlertScale.
type AlertScaleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AlertScale `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AlertScale{}, &AlertScaleList{})
}
