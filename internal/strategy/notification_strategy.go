package strategy

import (
	"context"
	"fmt"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"udesk.cn/ops/internal/types"
)

var DefaultNotificationClient *types.ScaleNotificationClient

// WXWorkRobotNotificationClient represents the configuration for WeChat Work notifications.
type WXWorkRobotNotificationClient struct {
	WebhookURL      string   `json:"webhookURL,omitempty"`
	Secret          string   `json:"secret,omitempty"`
	MessageTemplate string   `json:"messageTemplate,omitempty"`
	AtUsers         []string `json:"atUsers,omitempty"`
	AtAll           bool     `json:"atAll,omitempty"`
}

// Validate validates the WXWorkRobotConfig.
func (c *WXWorkRobotNotificationClient) Validate() error {
	if c.WebhookURL == "" {
		return fmt.Errorf("webhookURL is required")
	}
	return nil
}

func (c *WXWorkRobotNotificationClient) SendNotification(ctx context.Context, message string) error {
	if err := c.Validate(); err != nil {
		return err
	}

	// 这里可以添加发送微信工作通知的逻辑
	// 例如使用 HTTP POST 请求发送到 c.WebhookURL
	log := logf.FromContext(ctx)
	log.Info("Sending WXWorkRobot notification", "message", message)
	return nil
}

// EmailNotificationClient represents the configuration for email notifications.
type EmailNotificationClient struct {
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
func (c *EmailNotificationClient) Validate(ctx context.Context) error {
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

func (c *EmailNotificationClient) SendNotification(ctx context.Context, message string) error {
	if err := c.Validate(ctx); err != nil {
		return err
	}

	// 这里可以添加发送邮件通知的逻辑
	// 例如使用 SMTP 客户端发送邮件
	log := logf.FromContext(ctx)
	log.Info("Sending email notification", "to", c.ToEmail, "message", message)
	return nil
}
