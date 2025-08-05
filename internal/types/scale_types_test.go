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

package types

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

var _ = Describe("Scale Types", func() {
	var (
		scaleContext *ScaleContext
		alertScale   *opsv1beta1.AlertScale
		fakeClient   client.Client
		ctx          context.Context
		mockStrategy *MockScaleStrategy
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Create test scheme
		scheme := runtime.NewScheme()
		_ = opsv1beta1.AddToScheme(scheme)

		// Create fake client
		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		// Create test AlertScale
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

		// Create mock strategy
		mockStrategy = &MockScaleStrategy{}

		// Create scale context
		scaleContext = &ScaleContext{
			AlertScale:    alertScale,
			Client:        fakeClient,
			Request:       ctrl.Request{NamespacedName: client.ObjectKeyFromObject(alertScale)},
			Context:       ctx,
			ScaleStrategy: mockStrategy,
		}
	})

	Describe("ScaleContext", func() {
		Context("when creating a new ScaleContext", func() {
			It("should have all required fields", func() {
				Expect(scaleContext.AlertScale).NotTo(BeNil())
				Expect(scaleContext.Client).NotTo(BeNil())
				Expect(scaleContext.Context).NotTo(BeNil())
				Expect(scaleContext.ScaleStrategy).NotTo(BeNil())
			})

			It("should have valid AlertScale", func() {
				Expect(scaleContext.AlertScale.Name).To(Equal("test-alert"))
				Expect(scaleContext.AlertScale.Namespace).To(Equal("default"))
				Expect(scaleContext.AlertScale.Spec.ScaleReason).To(Equal("Memory pressure"))
			})

			It("should have valid Request", func() {
				Expect(scaleContext.Request.Name).To(Equal("test-alert"))
				Expect(scaleContext.Request.Namespace).To(Equal("default"))
			})
		})
	})

	Describe("ScaleStrategy Interface", func() {
		Context("when implementing ScaleStrategy", func() {
			It("should implement all required methods", func() {
				var strategy ScaleStrategy = mockStrategy
				Expect(strategy).NotTo(BeNil())
			})

			It("should scale successfully", func() {
				target := &opsv1beta1.ScaleTarget{
					Kind:      "Deployment",
					Name:      "test-deployment",
					Namespace: "default",
				}

				err := mockStrategy.Scale(ctx, fakeClient, target, 5)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should get current replicas", func() {
				target := &opsv1beta1.ScaleTarget{
					Kind:      "Deployment",
					Name:      "test-deployment",
					Namespace: "default",
				}

				replicas, err := mockStrategy.GetCurrentReplicas(ctx, fakeClient, target)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).To(Equal(int32(3))) // Mock returns 3
			})

			It("should get available replicas", func() {
				target := &opsv1beta1.ScaleTarget{
					Kind:      "Deployment",
					Name:      "test-deployment",
					Namespace: "default",
				}

				replicas, err := mockStrategy.GetAvailableReplicas(ctx, fakeClient, target)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).To(Equal(int32(3))) // Mock returns 3
			})
		})
	})

	Describe("StateHandler Interface", func() {
		Context("when implementing StateHandler", func() {
			It("should implement all required methods", func() {
				handler := &MockStateHandler{}
				var stateHandler StateHandler = handler
				Expect(stateHandler).NotTo(BeNil())
			})

			It("should handle state transitions", func() {
				handler := &MockStateHandler{}
				result, err := handler.Handle(scaleContext)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(ctrl.Result{}))
			})

			It("should check transition permissions", func() {
				handler := &MockStateHandler{}
				canTransition := handler.CanTransition("approved")
				Expect(canTransition).To(BeTrue())
			})
		})
	})
})

// MockScaleStrategy is a mock implementation for testing
type MockScaleStrategy struct{}

func (m *MockScaleStrategy) Scale(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
	return nil
}

func (m *MockScaleStrategy) GetCurrentReplicas(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	return 3, nil
}

func (m *MockScaleStrategy) GetAvailableReplicas(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	return 3, nil
}

// MockStateHandler is a mock implementation for testing
type MockStateHandler struct{}

func (m *MockStateHandler) Handle(ctx *ScaleContext) (ctrl.Result, error) {
	return ctrl.Result{}, nil
}

func (m *MockStateHandler) CanTransition(toState string) bool {
	return true
}
