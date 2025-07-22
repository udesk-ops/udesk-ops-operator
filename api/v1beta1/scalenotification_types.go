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
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"udesk.cn/ops/internal/notifications"
)

// WXWorkRobotNotificationConfig represents the configuration for WeChat Work notifications.
type WXWorkRobotNotificationConfig struct {
	WebhookURL      string   `json:"webhookURL,omitempty"`
	Secret          string   `json:"secret,omitempty"`
	MessageTemplate string   `json:"messageTemplate,omitempty"`
	AtUsers         []string `json:"atUsers,omitempty"`
	AtAll           bool     `json:"atAll,omitempty"`
}

// Validate validates the WXWorkRobotConfig.
func (c *WXWorkRobotNotificationConfig) Validate() error {
	if c.WebhookURL == "" {
		return fmt.Errorf("webhookURL is required")
	}
	return nil
}

// EmailNotificationConfig represents the configuration for email notifications.
type EmailNotificationConfig struct {
	SMTPServer   string `json:"smtpServer,omitempty"`
	SMTPPort     int32  `json:"smtpPort,omitempty"`
	SMTPUser     string `json:"smtpUser,omitempty"`
	SMTPPassword string `json:"smtpPassword,omitempty"`
	FromEmail    string `json:"fromEmail,omitempty"`
	ToEmail      string `json:"toEmail,omitempty"`
}

// Validate validates the EmailNotificationConfig.
// It checks that SMTPServer, FromEmail, and ToEmail are provided.
// SMTPPort is optional and defaults to 587 if not specified.
func (c *EmailNotificationConfig) Validate() error {
	if c.SMTPServer == "" || c.FromEmail == "" || c.ToEmail == "" {
		return fmt.Errorf("SMTPServer, FromEmail, and ToEmail are required")
	}
	if c.SMTPPort == 0 {
		c.SMTPPort = 587 // Default SMTP port if not specified
	}
	if c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTPPort must be between 1 and 65535")
	}
	if c.SMTPUser == "" || c.SMTPPassword == "" {
		return fmt.Errorf("SMTPUser and SMTPPassword are required")
	}
	// Additional validation can be added here, such as checking email formats.
	// For simplicity, we assume the email format is valid if not empty.
	// In a real-world application, you might want to use regex or a library to validate
	// the email format.
	// For example:
	// if !isValidEmail(c.FromEmail) {
	//     return fmt.Errorf("FromEmail is not a valid email address")
	// }
	// if !isValidEmail(c.ToEmail) {
	// 	return fmt.Errorf("ToEmail is not a valid email address")
	// }
	return nil
}

// UnmarshalConfig unmarshals the raw JSON config into the appropriate struct.
func (n *ScaleNotificationSpec) UnmarshalConfig() (notifications.ScaleNotificationConfig, error) {
	switch n.Type {
	case "email":
		var emailConfig EmailNotificationConfig
		if err := json.Unmarshal(n.Config.Raw, &emailConfig); err != nil {
			return nil, err
		}
		return &emailConfig, nil
	case "wxworkrobot":
		var wxConfig WXWorkRobotNotificationConfig
		if err := json.Unmarshal(n.Config.Raw, &wxConfig); err != nil {
			return nil, err
		}
		return &wxConfig, nil
	default:
		return nil, fmt.Errorf("unsupported notification type: %s", n.Type)
	}
}

// ScaleNotificationSpec defines the desired state of ScaleNotification.
type ScaleNotificationSpec struct {
	// Type is the type of notification (e.g., email, wxworkrobot).
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=email;wxworkrobot
	Type string `json:"type,omitempty"`
	// Config contains the specific configuration for the notification.
	// It uses runtime.RawExtension to support polymorphism.
	// +kubebuilder:pruning:PreserveUnknownFields
	Config runtime.RawExtension `json:"config,omitempty"`
	// default indicates if this notification is the default for the scale operation.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Type=boolean
	// +kubebuilder:default=false
	Default bool `json:"default,omitempty"` // Indicates if this notification is the default for the scale operation
}

// ScaleNotificationStatus defines the observed state of ScaleNotification.
type ScaleNotificationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Type",type=string,JSONPath=`.spec.type`
// +kubebuilder:printcolumn:name="Default",type=boolean,JSONPath=`.spec.default`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// ScaleNotification is the Schema for the scalenotifications API.
type ScaleNotification struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ScaleNotificationSpec   `json:"spec,omitempty"`
	Status ScaleNotificationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ScaleNotificationList contains a list of ScaleNotification.
type ScaleNotificationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ScaleNotification `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ScaleNotification{}, &ScaleNotificationList{})
}
