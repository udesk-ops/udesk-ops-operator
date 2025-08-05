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
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("AlertScale CRD", func() {
	Describe("AlertScale Type", func() {
		var alertScale *AlertScale

		BeforeEach(func() {
			alertScale = &AlertScale{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ops.udesk.cn/v1beta1",
					Kind:       "AlertScale",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-alert-scale",
					Namespace: "default",
				},
				Spec: AlertScaleSpec{
					ScaleReason:            "Memory pressure detected",
					ScaleAutoApproval:      true,
					ScaleDuration:          "5m",
					ScaleThreshold:         80,
					ScaleNotificationType:  "WXWorkRobot",
					ScaleNotifyMsgTemplate: "default-template",
					ScaleTimeout:           "10m",
					ScaleTarget: ScaleTarget{
						Kind:      "Deployment",
						Name:      "web-app",
						Namespace: "default",
					},
				},
			}
		})

		Context("when creating a new AlertScale", func() {
			It("should have correct TypeMeta", func() {
				Expect(alertScale.APIVersion).To(Equal("ops.udesk.cn/v1beta1"))
				Expect(alertScale.Kind).To(Equal("AlertScale"))
			})

			It("should have correct ObjectMeta", func() {
				Expect(alertScale.Name).To(Equal("test-alert-scale"))
				Expect(alertScale.Namespace).To(Equal("default"))
			})

			It("should have valid spec fields", func() {
				Expect(alertScale.Spec.ScaleReason).To(Equal("Memory pressure detected"))
				Expect(alertScale.Spec.ScaleAutoApproval).To(BeTrue())
				Expect(alertScale.Spec.ScaleDuration).To(Equal("5m"))
				Expect(alertScale.Spec.ScaleThreshold).To(Equal(int32(80)))
				Expect(alertScale.Spec.ScaleTimeout).To(Equal("10m"))
			})

			It("should have valid ScaleTarget", func() {
				target := alertScale.Spec.ScaleTarget
				Expect(target.Kind).To(Equal("Deployment"))
				Expect(target.Name).To(Equal("web-app"))
				Expect(target.Namespace).To(Equal("default"))
			})

			It("should have valid notification settings", func() {
				Expect(alertScale.Spec.ScaleNotificationType).To(Equal("WXWorkRobot"))
				Expect(alertScale.Spec.ScaleNotifyMsgTemplate).To(Equal("default-template"))
			})
		})

		Context("when checking runtime.Object interface", func() {
			It("should implement runtime.Object", func() {
				var obj runtime.Object = alertScale
				Expect(obj).NotTo(BeNil())
			})
		})
	})
})

var _ = Describe("ScaleNotifyConfig CRD", func() {
	Describe("ScaleNotifyConfig Type", func() {
		var config *ScaleNotifyConfig

		BeforeEach(func() {
			config = &ScaleNotifyConfig{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ops.udesk.cn/v1beta1",
					Kind:       "ScaleNotifyConfig",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-notify-config",
					Namespace: "default",
				},
				Spec: ScaleNotifyConfigSpec{
					Default: true,
					Type:    "WXWorkRobot",
					Config: runtime.RawExtension{
						Raw: []byte(`{"webhookURL": "https://example.com/webhook"}`),
					},
				},
			}
		})

		Context("when creating a new ScaleNotifyConfig", func() {
			It("should have correct TypeMeta", func() {
				Expect(config.APIVersion).To(Equal("ops.udesk.cn/v1beta1"))
				Expect(config.Kind).To(Equal("ScaleNotifyConfig"))
			})

			It("should have valid spec", func() {
				Expect(config.Spec.Default).To(BeTrue())
				Expect(config.Spec.Type).To(Equal("WXWorkRobot"))
				Expect(config.Spec.Config.Raw).NotTo(BeEmpty())
			})
		})

		Context("when implementing runtime.Object", func() {
			It("should implement runtime.Object interface", func() {
				var obj runtime.Object = config
				Expect(obj).NotTo(BeNil())
			})
		})
	})
})

var _ = Describe("ScaleNotifyMsgTemplate CRD", func() {
	Describe("ScaleNotifyMsgTemplate Type", func() {
		var template *ScaleNotifyMsgTemplate

		BeforeEach(func() {
			template = &ScaleNotifyMsgTemplate{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "ops.udesk.cn/v1beta1",
					Kind:       "ScaleNotifyMsgTemplate",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-template",
					Namespace: "default",
				},
				Spec: ScaleNotifyMsgTemplateSpec{
					Title:   "Scaling Alert",
					Content: "Scaling operation initiated for {{.TargetName}}",
				},
			}
		})

		Context("when creating a new ScaleNotifyMsgTemplate", func() {
			It("should have correct TypeMeta", func() {
				Expect(template.APIVersion).To(Equal("ops.udesk.cn/v1beta1"))
				Expect(template.Kind).To(Equal("ScaleNotifyMsgTemplate"))
			})

			It("should have valid spec", func() {
				Expect(template.Spec.Title).To(Equal("Scaling Alert"))
				Expect(template.Spec.Content).To(Equal("Scaling operation initiated for {{.TargetName}}"))
			})
		})

		Context("when implementing runtime.Object", func() {
			It("should implement runtime.Object interface", func() {
				var obj runtime.Object = template
				Expect(obj).NotTo(BeNil())
			})
		})
	})
})
