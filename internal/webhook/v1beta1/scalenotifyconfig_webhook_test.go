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
	"context"
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// createRawExtension is a helper function to create runtime.RawExtension from map
func createRawExtension(data map[string]interface{}) runtime.RawExtension {
	bytes, _ := json.Marshal(data)
	return runtime.RawExtension{Raw: bytes}
}

var _ = Describe("ScaleNotifyConfig Webhook", func() {
	var (
		ctx       context.Context
		validator *ScaleNotifyConfigCustomValidator
		scheme    *runtime.Scheme
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
		Expect(opsv1beta1.AddToScheme(scheme)).To(Succeed())
	})

	Context("ValidateCreate", func() {
		It("should allow creating the first default config of a type", func() {
			// Setup: empty client (no existing configs)
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			config := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-email-1",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.example.com",
						"smtpPort":   587,
						"username":   "test@example.com",
						"password":   "password123",
						"fromEmail":  "noreply@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, config)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject creating duplicate default config of same type", func() {
			// Setup: client with existing default Email config
			existingConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-email-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.existing.com",
						"smtpPort":   587,
						"username":   "existing@example.com",
						"password":   "password123",
						"fromEmail":  "existing@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			duplicateConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-email-2",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp2.example.com",
						"smtpPort":   587,
						"username":   "test2@example.com",
						"password":   "password123",
						"fromEmail":  "noreply2@example.com",
						"toEmails":   []string{"admin2@example.com"},
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, duplicateConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("a default ScaleNotifyConfig of type 'Email' already exists"))
			Expect(err.Error()).To(ContainSubstring("existing-email-default"))
		})

		It("should allow creating default config of different type", func() {
			// Setup: client with existing default Email config
			existingEmailConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-email-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.existing.com",
						"smtpPort":   587,
						"username":   "existing@example.com",
						"password":   "password123",
						"fromEmail":  "existing@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingEmailConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			wxworkConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-wxwork-1",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "WXWorkRobot",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"webhookURL":      "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=example",
						"secret":          "secret123",
						"messageTemplate": "Alert: {{.Message}}",
						"atUsers":         []string{"user1", "user2"},
						"atAll":           false,
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, wxworkConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should allow creating non-default config of same type", func() {
			// Setup: client with existing default Email config
			existingDefaultConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-email-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.existing.com",
						"smtpPort":   587,
						"username":   "existing@example.com",
						"password":   "password123",
						"fromEmail":  "existing@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingDefaultConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			nonDefaultConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-email-non-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: false, // Non-default should be allowed
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp2.example.com",
						"smtpPort":   587,
						"username":   "test2@example.com",
						"password":   "password123",
						"fromEmail":  "noreply2@example.com",
						"toEmails":   []string{"admin2@example.com"},
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, nonDefaultConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject config with invalid type", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			invalidConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-invalid",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "InvalidType",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"invalid": "config",
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, invalidConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.type must be one of"))
		})

		It("should reject Email config with missing required fields", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			incompleteConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-incomplete-email",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.example.com",
						// Missing required fields: username, password, fromEmail, toEmails
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, incompleteConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid Email configuration"))
		})

		It("should reject WXWorkRobot config with missing required fields", func() {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			incompleteConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-incomplete-wxwork",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "WXWorkRobot",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"secret": "secret123",
						// Missing required field: webhookURL
					}),
				},
			}

			_, err := validator.ValidateCreate(ctx, incompleteConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid WXWorkRobot configuration"))
			Expect(err.Error()).To(ContainSubstring("webhookURL is required"))
		})
	})

	Context("ValidateUpdate", func() {
		It("should allow updating non-default config to default when no other default exists", func() {
			// Setup: client with non-default Email config
			existingConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-email-non-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: false,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.existing.com",
						"smtpPort":   587,
						"username":   "existing@example.com",
						"password":   "password123",
						"fromEmail":  "existing@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			updatedConfig := existingConfig.DeepCopy()
			updatedConfig.Spec.Default = true // Change to default

			_, err := validator.ValidateUpdate(ctx, existingConfig, updatedConfig)
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reject updating non-default config to default when another default exists", func() {
			// Setup: client with existing default and non-default Email configs
			existingDefaultConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.default.com",
						"smtpPort":   587,
						"username":   "default@example.com",
						"password":   "password123",
						"fromEmail":  "default@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			existingNonDefaultConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-non-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: false,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.nondefault.com",
						"smtpPort":   587,
						"username":   "nondefault@example.com",
						"password":   "password123",
						"fromEmail":  "nondefault@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingDefaultConfig, existingNonDefaultConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			updatedConfig := existingNonDefaultConfig.DeepCopy()
			updatedConfig.Spec.Default = true // Try to change to default

			_, err := validator.ValidateUpdate(ctx, existingNonDefaultConfig, updatedConfig)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("a default ScaleNotifyConfig of type 'Email' already exists"))
			Expect(err.Error()).To(ContainSubstring("existing-default"))
		})

		It("should allow updating default config configuration without changing default status", func() {
			existingDefaultConfig := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-default",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.old.com",
						"smtpPort":   587,
						"username":   "old@example.com",
						"password":   "password123",
						"fromEmail":  "old@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				WithObjects(existingDefaultConfig).
				Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			updatedConfig := existingDefaultConfig.DeepCopy()
			updatedConfig.Spec.Config = createRawExtension(map[string]interface{}{
				"smtpServer": "smtp.new.com", // Updated server
				"smtpPort":   587,
				"username":   "new@example.com", // Updated user
				"password":   "newpassword123",
				"fromEmail":  "new@example.com",
				"toEmails":   []string{"admin@example.com"},
			})

			_, err := validator.ValidateUpdate(ctx, existingDefaultConfig, updatedConfig)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("ValidateDelete", func() {
		It("should allow deleting any config", func() {
			config := &opsv1beta1.ScaleNotifyConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-config",
				},
				Spec: opsv1beta1.ScaleNotifyConfigSpec{
					Type:    "Email",
					Default: true,
					Config: createRawExtension(map[string]interface{}{
						"smtpServer": "smtp.example.com",
						"smtpPort":   587,
						"username":   "test@example.com",
						"password":   "password123",
						"fromEmail":  "test@example.com",
						"toEmails":   []string{"admin@example.com"},
					}),
				},
			}

			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			validator = &ScaleNotifyConfigCustomValidator{Client: fakeClient}

			_, err := validator.ValidateDelete(ctx, config)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
