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

package handler

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

var _ = Describe("Scale State Handler", func() {
	Describe("BaseStateHandler", func() {
		var handler BaseStateHandler
		var alertScale *opsv1beta1.AlertScale

		BeforeEach(func() {
			handler = BaseStateHandler{}

			alertScale = &opsv1beta1.AlertScale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-alert",
					Namespace: "default",
				},
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason:       "Memory pressure",
					ScaleAutoApproval: true,
					ScaleDuration:     "5m",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      "Deployment",
						Name:      "web-app",
						Namespace: "default",
					},
				},
			}
		})

		Context("when handling duration parsing", func() {
			It("should parse valid duration strings", func() {
				duration, err := handler.parseDuration("5m")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(5 * time.Minute))
			})

			It("should handle empty duration with default", func() {
				duration, err := handler.parseDuration("")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(5 * time.Minute))
			})

			It("should handle invalid duration strings", func() {
				_, err := handler.parseDuration("invalid")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when working with AlertScale objects", func() {
			It("should have valid AlertScale structure", func() {
				Expect(alertScale.Name).To(Equal("test-alert"))
				Expect(alertScale.Namespace).To(Equal("default"))
				Expect(alertScale.Spec.ScaleReason).To(Equal("Memory pressure"))
				Expect(alertScale.Spec.ScaleAutoApproval).To(BeTrue())
			})

			It("should have valid ScaleTarget", func() {
				target := alertScale.Spec.ScaleTarget
				Expect(target.Kind).To(Equal("Deployment"))
				Expect(target.Name).To(Equal("web-app"))
				Expect(target.Namespace).To(Equal("default"))
			})
		})
	})

	Describe("Duration parsing utility", func() {
		Context("when parsing duration strings", func() {
			It("should parse minutes correctly", func() {
				duration, err := parseDuration("10m")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(10 * time.Minute))
			})

			It("should parse hours correctly", func() {
				duration, err := parseDuration("2h")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(2 * time.Hour))
			})

			It("should parse seconds correctly", func() {
				duration, err := parseDuration("30s")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(30 * time.Second))
			})

			It("should use default duration for empty string", func() {
				duration, err := parseDuration("")
				Expect(err).ToNot(HaveOccurred())
				Expect(duration).To(Equal(5 * time.Minute))
			})

			It("should return error for invalid format", func() {
				_, err := parseDuration("invalid-format")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
