package strategy

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"udesk.cn/ops/internal/types"
)

var DefaultNotifyClientMap = make(map[string]types.ScaleNotifyClient)

func NewScaleNotifyClient(name string, config runtime.RawExtension) types.ScaleNotifyClient {
	switch name {
	case types.NotifyTypeWXWorkRobot:
		client, err := NewWXWorkRobotNotificationClient(config)
		if err != nil {
			logf.Log.Error(err, "Failed to create WeChat Work notification client", "type", name)
			return nil
		}
		return client
	case types.NotifyTypeEmail:
		client, err := NewEmailNotificationClient(config)
		if err != nil {
			logf.Log.Error(err, "Failed to create Email notification client", "type", name)
			return nil
		}
		return client
	default:
		logf.Log.Error(fmt.Errorf("unknown notification type: %s", name), "Failed to create notification client")
		return nil
	}
}

// WXWorkRobotNotificationClient represents the configuration for WeChat Work notifications.
type WXWorkRobotNotificationClient struct {
	WebhookURL string   `json:"webhookURL,omitempty"`
	Secret     string   `json:"secret,omitempty"`
	AtUsers    []string `json:"atUsers,omitempty"`
	AtAll      bool     `json:"atAll,omitempty"`
}

func NewWXWorkRobotNotificationClient(config runtime.RawExtension) (*WXWorkRobotNotificationClient, error) {
	notifyClient := &WXWorkRobotNotificationClient{}
	if err := json.Unmarshal(config.Raw, notifyClient); err != nil {
		logf.Log.Error(err, "Failed to unmarshal config", "type", "WxWorkRobot")
		return nil, err
	}
	return notifyClient, nil
}

// Validate validates the WXWorkRobotConfig.
func (c *WXWorkRobotNotificationClient) Validate(ctx context.Context) error {
	if c.WebhookURL == "" {
		// If WebhookURL is not provided, return an error
		// This is a required field for sending notifications
		return fmt.Errorf("webhookURL is required")
	}
	return nil
}

func (c *WXWorkRobotNotificationClient) SendNotify(ctx context.Context, message string) error {
	// Validate the configuration before sending the notification
	if err := c.Validate(ctx); err != nil {
		return err
	}

	log := logf.FromContext(ctx)
	log.Info("Sending WeChat Work notification", "webhookURL", c.WebhookURL, "message", message)

	payload := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": message,
		},
	}

	// 添加 @ 用户功能
	if len(c.AtUsers) > 0 || c.AtAll {
		payload["markdown"].(map[string]interface{})["mentioned_list"] = c.AtUsers
		if c.AtAll {
			payload["markdown"].(map[string]interface{})["mentioned_mobile_list"] = []string{"@all"}
		}
	}

	// 如果配置了 Secret，添加签名
	webhookURL := c.WebhookURL
	if c.Secret != "" {
		timestamp := time.Now().Unix()
		stringToSign := fmt.Sprintf("%d\n%s", timestamp, c.Secret)

		h := hmac.New(sha256.New, []byte(c.Secret))
		h.Write([]byte(stringToSign))
		signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", c.WebhookURL, timestamp, signature)
	}

	// 序列化消息体
	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Error(err, "Failed to marshal WeChat Work message")
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// 发送 HTTP 请求
	resp, err := http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Error(err, "Failed to send WeChat Work notification")
		return fmt.Errorf("failed to send notification: %v", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Error(closeErr, "Failed to close response body")
		}
	}()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(err, "Failed to read WeChat Work response")
		return fmt.Errorf("failed to read response: %v", err)
	}

	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		log.Error(nil, "WeChat Work notification failed", "statusCode", resp.StatusCode, "response", string(body))
		return fmt.Errorf("notification failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应以检查是否成功
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Error(err, "Failed to parse WeChat Work response")
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// 检查错误码 - 企业微信 API 中 errcode=0 表示成功
	if errCode, ok := result["errcode"]; ok {
		if errCodeInt, ok := errCode.(float64); ok && errCodeInt != 0 {
			errMsg := result["errmsg"]
			log.Error(nil, "WeChat Work API returned error", "errcode", errCode, "errmsg", errMsg)
			return fmt.Errorf("WeChat Work API error %v: %v", errCode, errMsg)
		}
	}

	log.Info("WeChat Work notification sent successfully")
	return nil
}

// EmailNotificationClient represents the configuration for email notifications.
type EmailNotificationClient struct {
	SMTPServer string   `json:"smtpServer,omitempty"`
	SMTPPort   int32    `json:"smtpPort,omitempty"`
	Username   string   `json:"username,omitempty"`
	Password   string   `json:"password,omitempty"`
	FromEmail  string   `json:"fromEmail,omitempty"`
	ToEmails   []string `json:"toEmails,omitempty"`
	Subject    string   `json:"subject,omitempty"`
}

func NewEmailNotificationClient(config runtime.RawExtension) (*EmailNotificationClient, error) {
	notifyClient := &EmailNotificationClient{}
	if err := json.Unmarshal(config.Raw, notifyClient); err != nil {
		logf.Log.Error(err, "Failed to unmarshal config", "type", "Email")
		return nil, err
	}
	return notifyClient, nil
}

// Validate validates the EmailNotificationConfig.
// It checks that SMTPServer, FromEmail, and ToEmails are provided.
// SMTPPort is optional and defaults to 587 if not specified.
func (c *EmailNotificationClient) Validate(ctx context.Context) error {
	if c.SMTPServer == "" || c.FromEmail == "" || len(c.ToEmails) == 0 {
		return fmt.Errorf("SMTPServer, FromEmail, and ToEmails are required")
	}
	if c.SMTPPort == 0 {
		c.SMTPPort = 587 // Default SMTP port if not specified
	}
	if c.SMTPPort < 1 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTPPort must be between 1 and 65535")
	}
	// 验证邮件地址格式（简单验证）
	for _, email := range c.ToEmails {
		if !strings.Contains(email, "@") {
			return fmt.Errorf("invalid email address: %s", email)
		}
	}
	if !strings.Contains(c.FromEmail, "@") {
		return fmt.Errorf("invalid from email address: %s", c.FromEmail)
	}
	return nil
}

func (c *EmailNotificationClient) SendNotify(ctx context.Context, message string) error {
	// Validate the configuration before sending the notification
	if err := c.Validate(ctx); err != nil {
		return err
	}

	log := logf.FromContext(ctx)
	log.Info("Sending email notification", "smtpServer", c.SMTPServer, "smtpPort", c.SMTPPort, "fromEmail", c.FromEmail, "toEmails", c.ToEmails)

	// 如果没有邮件主题，使用默认主题
	subject := c.Subject
	if subject == "" {
		subject = "AlertScale Notification"
	}

	// 创建 SMTP 认证
	var auth smtp.Auth
	if c.Username != "" && c.Password != "" {
		auth = smtp.PlainAuth("", c.Username, c.Password, c.SMTPServer)
	}

	// 构建邮件头
	headers := make(map[string]string)
	headers["From"] = c.FromEmail
	headers["To"] = strings.Join(c.ToEmails, ",")
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/plain; charset=UTF-8"
	headers["Date"] = time.Now().Format(time.RFC1123Z)

	// 构建邮件消息
	var msg bytes.Buffer
	for k, v := range headers {
		msg.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	msg.WriteString("\r\n")
	msg.WriteString(message)

	// 发送邮件
	addr := fmt.Sprintf("%s:%d", c.SMTPServer, c.SMTPPort)

	// 对于每个收件人发送邮件
	for _, to := range c.ToEmails {
		err := smtp.SendMail(addr, auth, c.FromEmail, []string{to}, msg.Bytes())
		if err != nil {
			log.Error(err, "Failed to send email notification", "to", to)
			return fmt.Errorf("failed to send email to %s: %v", to, err)
		}
	}

	log.Info("Email notification sent successfully", "recipients", len(c.ToEmails))
	return nil
}
