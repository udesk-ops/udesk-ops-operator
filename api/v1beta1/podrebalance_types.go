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

// PodRebalanceSpec defines the desired state of PodRebalance.
type PodRebalanceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Namespace where to perform rebalancing
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Selector for target pods
	// +kubebuilder:validation:Required
	Selector metav1.LabelSelector `json:"selector"`

	// RebalanceStrategy defines how pods should be rebalanced
	// +kubebuilder:validation:Required
	Strategy PodRebalanceStrategy `json:"strategy"`

	// AutoApproval enables automatic approval for rebalancing
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	AutoApproval bool `json:"autoApproval,omitempty"`

	// Timeout for the approval and rebalancing process
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="30m"
	// +kubebuilder:validation:Pattern=`^(\d+)([smhdw])$`
	Timeout string `json:"timeout,omitempty"`

	// DryRun mode for testing rebalancing without actual execution
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:=false
	DryRun bool `json:"dryRun,omitempty"`

	// NotificationType defines notification method
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Email;WXWorkRobot
	NotificationType string `json:"notificationType,omitempty"`

	// NotifyMsgTemplate is the reference to notification template
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	NotifyMsgTemplate string `json:"notifyMsgTemplate,omitempty"`
}

// PodRebalanceStrategy defines rebalancing strategy
type PodRebalanceStrategy struct {
	// Type of rebalancing strategy (NodeBalance, ResourceBalance, etc.)
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=NodeBalance;ResourceBalance;AntiAffinity
	Type string `json:"type"`

	// Parameters for the strategy
	// +kubebuilder:validation:Optional
	Parameters map[string]string `json:"parameters,omitempty"`

	// Threshold configuration for triggering rebalancing
	// +kubebuilder:validation:Optional
	Threshold *RebalanceThreshold `json:"threshold,omitempty"`
}

// RebalanceThreshold defines thresholds for triggering rebalancing
type RebalanceThreshold struct {
	// CPU usage threshold percentage
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	CPU *int32 `json:"cpu,omitempty"`

	// Memory usage threshold percentage
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	Memory *int32 `json:"memory,omitempty"`

	// Pod count imbalance threshold
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	PodCountImbalance *int32 `json:"podCountImbalance,omitempty"`
}

// PodRebalanceStatus defines the observed state of PodRebalance.
type PodRebalanceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status of the rebalancing process
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Approvaling;Approved;Rejected;Executing;Completed;Failed
	Status string `json:"status,omitempty"`

	// Message provides additional information about the current status
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`

	// RebalanceBeginTime records when rebalancing started
	// +kubebuilder:validation:Optional
	RebalanceBeginTime metav1.Time `json:"rebalanceBeginTime,omitempty"`

	// RebalanceEndTime records when rebalancing completed
	// +kubebuilder:validation:Optional
	RebalanceEndTime metav1.Time `json:"rebalanceEndTime,omitempty"`

	// RebalancedPods contains information about pods that were rebalanced
	// +kubebuilder:validation:Optional
	RebalancedPods []RebalancedPodInfo `json:"rebalancedPods,omitempty"`

	// Conditions represent the latest available observations
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// RebalancedPodInfo contains information about a rebalanced pod
type RebalancedPodInfo struct {
	// Name of the pod
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the pod
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`

	// Source node
	// +kubebuilder:validation:Optional
	SourceNode string `json:"sourceNode,omitempty"`

	// Target node
	// +kubebuilder:validation:Optional
	TargetNode string `json:"targetNode,omitempty"`

	// Rebalance status for this pod
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum=Pending;Moving;Completed;Failed
	Status string `json:"status,omitempty"`

	// Timestamp when this pod was rebalanced
	// +kubebuilder:validation:Optional
	Timestamp metav1.Time `json:"timestamp,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Namespaced,shortName=pr
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="Auto-Approval",type="boolean",JSONPath=".spec.autoApproval"
// +kubebuilder:printcolumn:name="Strategy",type="string",JSONPath=".spec.strategy.type"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PodRebalance is the Schema for the podrebalances API.
type PodRebalance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PodRebalanceSpec   `json:"spec,omitempty"`
	Status PodRebalanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PodRebalanceList contains a list of PodRebalance.
type PodRebalanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PodRebalance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodRebalance{}, &PodRebalanceList{})
}
