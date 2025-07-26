package strategy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"udesk.cn/ops/internal/types"
)

func TestNewScaleNotifyClient(t *testing.T) {
	tests := []struct {
		name          string
		notifyType    string
		config        runtime.RawExtension
		expectedType  interface{}
		expectedNil   bool
		expectedError bool
	}{
		{
			name:       "Valid WXWorkRobot client",
			notifyType: types.NotifyTypeWXWorkRobot,
			config: runtime.RawExtension{
				Raw: []byte(`{"webhookURL": "https://example.com/webhook"}`),
			},
			expectedType: &WXWorkRobotNotificationClient{},
			expectedNil:  false,
		},
		{
			name:       "Valid Email client",
			notifyType: types.NotifyTypeEmail,
			config: runtime.RawExtension{
				Raw: []byte(`{"smtpServer": "smtp.example.com", "fromEmail": "test@example.com", "toEmails": ["admin@example.com"]}`),
			},
			expectedType: &EmailNotificationClient{},
			expectedNil:  false,
		},
		{
			name:       "Invalid notification type",
			notifyType: "InvalidType",
			config: runtime.RawExtension{
				Raw: []byte(`{}`),
			},
			expectedNil: true,
		},
		{
			name:       "Invalid JSON config for WXWorkRobot",
			notifyType: types.NotifyTypeWXWorkRobot,
			config: runtime.RawExtension{
				Raw: []byte(`invalid json`),
			},
			expectedNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewScaleNotifyClient(tt.notifyType, tt.config)

			if tt.expectedNil {
				if client != nil {
					t.Errorf("Expected nil client, got %T", client)
				}
				return
			}

			if client == nil {
				t.Errorf("Expected non-nil client, got nil")
				return
			}

			// Type assertion check
			switch tt.expectedType.(type) {
			case *WXWorkRobotNotificationClient:
				if _, ok := client.(*WXWorkRobotNotificationClient); !ok {
					t.Errorf("Expected *WXWorkRobotNotificationClient, got %T", client)
				}
			case *EmailNotificationClient:
				if _, ok := client.(*EmailNotificationClient); !ok {
					t.Errorf("Expected *EmailNotificationClient, got %T", client)
				}
			}
		})
	}
}

func TestWXWorkRobotNotificationClient_Validate(t *testing.T) {
	tests := []struct {
		name        string
		client      *WXWorkRobotNotificationClient
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
			},
			expectError: false,
		},
		{
			name: "Missing webhook URL",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "",
			},
			expectError: true,
			errorMsg:    "webhookURL is required",
		},
		{
			name: "Valid configuration with optional fields",
			client: &WXWorkRobotNotificationClient{
				WebhookURL:      "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				Secret:          "secret123",
				MessageTemplate: "Alert: {{.Message}}",
				AtUsers:         []string{"user1", "user2"},
				AtAll:           true,
			},
			expectError: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Validate(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestWXWorkRobotNotificationClient_SendNotify(t *testing.T) {
	tests := []struct {
		name           string
		client         *WXWorkRobotNotificationClient
		message        string
		mockResponse   string
		mockStatusCode int
		expectError    bool
		errorMsg       string
	}{
		{
			name: "Successful notification without secret",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "test-url", // Will be replaced by test server URL
			},
			message:        "Test message",
			mockResponse:   `{"errcode": 0, "errmsg": "ok"}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name: "Successful notification with template",
			client: &WXWorkRobotNotificationClient{
				WebhookURL:      "test-url",
				MessageTemplate: "Alert: {{.Message}} at {{.Time}}",
			},
			message:        "Test message",
			mockResponse:   `{"errcode": 0, "errmsg": "ok"}`,
			mockStatusCode: http.StatusOK,
			expectError:    false,
		},
		{
			name: "API error response",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "test-url",
			},
			message:        "Test message",
			mockResponse:   `{"errcode": 93000, "errmsg": "invalid webhook url"}`,
			mockStatusCode: http.StatusOK,
			expectError:    true,
			errorMsg:       "WeChat Work API error",
		},
		{
			name: "HTTP error status",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "test-url",
			},
			message:        "Test message",
			mockResponse:   `Server Error`,
			mockStatusCode: http.StatusInternalServerError,
			expectError:    true,
			errorMsg:       "notification failed with status 500",
		},
		{
			name: "Invalid webhook URL",
			client: &WXWorkRobotNotificationClient{
				WebhookURL: "",
			},
			message:     "Test message",
			expectError: true,
			errorMsg:    "webhookURL is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			var server *httptest.Server
			if tt.client.WebhookURL != "" {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.mockStatusCode)
					w.Write([]byte(tt.mockResponse))
				}))
				defer server.Close()
				tt.client.WebhookURL = server.URL
			}

			ctx := context.Background()
			err := tt.client.SendNotify(ctx, tt.message)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

func TestEmailNotificationClient_Validate(t *testing.T) {
	tests := []struct {
		name        string
		client      *EmailNotificationClient
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid configuration",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				SMTPPort:   587,
				FromEmail:  "test@example.com",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: false,
		},
		{
			name: "Missing SMTP server",
			client: &EmailNotificationClient{
				SMTPServer: "",
				FromEmail:  "test@example.com",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTPServer, FromEmail, and ToEmails are required",
		},
		{
			name: "Missing from email",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				FromEmail:  "",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTPServer, FromEmail, and ToEmails are required",
		},
		{
			name: "Empty to emails",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				FromEmail:  "test@example.com",
				ToEmails:   []string{},
			},
			expectError: true,
			errorMsg:    "SMTPServer, FromEmail, and ToEmails are required",
		},
		{
			name: "Invalid port range",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				SMTPPort:   99999,
				FromEmail:  "test@example.com",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: true,
			errorMsg:    "SMTPPort must be between 1 and 65535",
		},
		{
			name: "Invalid from email format",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				FromEmail:  "invalid-email",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: true,
			errorMsg:    "invalid from email address",
		},
		{
			name: "Invalid to email format",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				FromEmail:  "test@example.com",
				ToEmails:   []string{"invalid-email"},
			},
			expectError: true,
			errorMsg:    "invalid email address",
		},
		{
			name: "Default port assignment",
			client: &EmailNotificationClient{
				SMTPServer: "smtp.example.com",
				SMTPPort:   0, // Should default to 587
				FromEmail:  "test@example.com",
				ToEmails:   []string{"admin@example.com"},
			},
			expectError: false,
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Validate(ctx)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got %v", err)
				}
				// Check if default port was assigned
				if tt.client.SMTPPort == 0 {
					t.Errorf("Expected SMTPPort to be set to default 587, got %d", tt.client.SMTPPort)
				}
			}
		})
	}
}

func TestNewWXWorkRobotNotificationClient(t *testing.T) {
	tests := []struct {
		name        string
		config      runtime.RawExtension
		expectError bool
		expected    *WXWorkRobotNotificationClient
	}{
		{
			name: "Valid configuration",
			config: runtime.RawExtension{
				Raw: []byte(`{
					"webhookURL": "https://example.com/webhook",
					"secret": "test-secret",
					"messageTemplate": "Alert: {{.Message}}",
					"atUsers": ["user1", "user2"],
					"atAll": true
				}`),
			},
			expectError: false,
			expected: &WXWorkRobotNotificationClient{
				WebhookURL:      "https://example.com/webhook",
				Secret:          "test-secret",
				MessageTemplate: "Alert: {{.Message}}",
				AtUsers:         []string{"user1", "user2"},
				AtAll:           true,
			},
		},
		{
			name: "Invalid JSON",
			config: runtime.RawExtension{
				Raw: []byte(`invalid json`),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewWXWorkRobotNotificationClient(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if client == nil {
				t.Errorf("Expected non-nil client")
				return
			}

			// Compare fields
			if client.WebhookURL != tt.expected.WebhookURL {
				t.Errorf("Expected WebhookURL %s, got %s", tt.expected.WebhookURL, client.WebhookURL)
			}
			if client.Secret != tt.expected.Secret {
				t.Errorf("Expected Secret %s, got %s", tt.expected.Secret, client.Secret)
			}
			if client.MessageTemplate != tt.expected.MessageTemplate {
				t.Errorf("Expected MessageTemplate %s, got %s", tt.expected.MessageTemplate, client.MessageTemplate)
			}
		})
	}
}

func TestNewEmailNotificationClient(t *testing.T) {
	tests := []struct {
		name        string
		config      runtime.RawExtension
		expectError bool
		expected    *EmailNotificationClient
	}{
		{
			name: "Valid configuration",
			config: runtime.RawExtension{
				Raw: []byte(`{
					"smtpServer": "smtp.example.com",
					"smtpPort": 587,
					"username": "user@example.com",
					"password": "password123",
					"fromEmail": "noreply@example.com",
					"toEmails": ["admin@example.com", "ops@example.com"],
					"subject": "Test Subject",
					"messageTemplate": "Message: {{.Message}}"
				}`),
			},
			expectError: false,
			expected: &EmailNotificationClient{
				SMTPServer:      "smtp.example.com",
				SMTPPort:        587,
				Username:        "user@example.com",
				Password:        "password123",
				FromEmail:       "noreply@example.com",
				ToEmails:        []string{"admin@example.com", "ops@example.com"},
				Subject:         "Test Subject",
				MessageTemplate: "Message: {{.Message}}",
			},
		},
		{
			name: "Invalid JSON",
			config: runtime.RawExtension{
				Raw: []byte(`invalid json`),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewEmailNotificationClient(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Expected no error, got %v", err)
				return
			}

			if client == nil {
				t.Errorf("Expected non-nil client")
				return
			}

			// Compare fields
			if client.SMTPServer != tt.expected.SMTPServer {
				t.Errorf("Expected SMTPServer %s, got %s", tt.expected.SMTPServer, client.SMTPServer)
			}
			if client.SMTPPort != tt.expected.SMTPPort {
				t.Errorf("Expected SMTPPort %d, got %d", tt.expected.SMTPPort, client.SMTPPort)
			}
			if client.FromEmail != tt.expected.FromEmail {
				t.Errorf("Expected FromEmail %s, got %s", tt.expected.FromEmail, client.FromEmail)
			}
		})
	}
}
