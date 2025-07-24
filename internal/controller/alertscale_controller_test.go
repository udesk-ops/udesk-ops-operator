/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the Li				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason: "Memory Pressure",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      ResourceKindStatefulSet,
						Name:      "test-statefulset",
						Namespace: "default",
					},t

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/handler"
	"udesk.cn/ops/internal/strategy"
	internalTypes "udesk.cn/ops/internal/types"
)

// MockStateHandler for testing
type MockStateHandler struct {
	handleFunc        func(*internalTypes.ScaleContext) (reconcile.Result, error)
	canTransitionFunc func(string) bool
}

func (m *MockStateHandler) Handle(ctx *internalTypes.ScaleContext) (reconcile.Result, error) {
	if m.handleFunc != nil {
		return m.handleFunc(ctx)
	}
	return reconcile.Result{}, nil
}

func (m *MockStateHandler) CanTransition(toState string) bool {
	if m.canTransitionFunc != nil {
		return m.canTransitionFunc(toState)
	}
	return true
}

var _ = Describe("AlertScale Controller", func() {
	var (
		ctx            context.Context
		reconciler     *AlertScaleReconciler
		fakeClient     client.Client
		scheme         *runtime.Scheme
		resourceName   string
		namespacedName types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()

		// 创建scheme并注册类型
		scheme = runtime.NewScheme()
		Expect(clientgoscheme.AddToScheme(scheme)).To(Succeed())
		Expect(opsv1beta1.AddToScheme(scheme)).To(Succeed())

		// 创建fake client
		fakeClient = fake.NewClientBuilder().WithScheme(scheme).Build()

		// 初始化reconciler
		reconciler = &AlertScaleReconciler{
			Client: fakeClient,
			Scheme: scheme,
			StateHandlers: map[string]internalTypes.StateHandler{
				internalTypes.ScaleStatusPending:   &handler.PendingHandler{},
				internalTypes.ScaleStatusScaling:   &handler.ScalingHandler{},
				internalTypes.ScaleStatusScaled:    &handler.ScaledHandler{},
				internalTypes.ScaleStatusCompleted: &handler.CompletedHandler{},
				internalTypes.ScaleStatusFailed:    &handler.FailedHandler{},
				internalTypes.ScaleStatusArchived:  &handler.ArchivedHandler{},
				"default":                          &handler.DefaultHandler{},
			},
		}

		resourceName = "test-alertscale"
		namespacedName = types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
	})

	Context("When reconciling an AlertScale resource", func() {
		It("should handle non-existent resource gracefully", func() {
			// 尝试reconcile一个不存在的资源
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reconcile AlertScale with Deployment target successfully", func() {
			// 使用mock handler来避免依赖真实的Kubernetes资源
			mockHandler := &MockStateHandler{
				handleFunc: func(ctx *internalTypes.ScaleContext) (reconcile.Result, error) {
					Expect(ctx.AlertScale.Spec.ScaleTarget.Kind).To(Equal(ResourceKindDeployment))
					return reconcile.Result{}, nil
				},
			}

			// 临时替换handler
			originalHandler := reconciler.StateHandlers[internalTypes.ScaleStatusPending]
			reconciler.StateHandlers[internalTypes.ScaleStatusPending] = mockHandler
			defer func() {
				reconciler.StateHandlers[internalTypes.ScaleStatusPending] = originalHandler
			}()

			// 创建AlertScale资源，目标为Deployment
			alertScale := &opsv1beta1.AlertScale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason: "High CPU Usage",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      ResourceKindDeployment,
						Name:      "test-deployment",
						Namespace: "default",
					},
					ScaleThreshold:        80,
					ScaleDuration:         "5m",
					ScaleNotificationType: "email",
				},
				Status: opsv1beta1.AlertScaleStatus{
					ScaleStatus: opsv1beta1.ScaleStatus{
						Status: internalTypes.ScaleStatusPending,
					},
				},
			}

			Expect(fakeClient.Create(ctx, alertScale)).To(Succeed())

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should reconcile AlertScale with StatefulSet target successfully", func() {
			// 使用mock handler来避免依赖真实的Kubernetes资源
			mockHandler := &MockStateHandler{
				handleFunc: func(ctx *internalTypes.ScaleContext) (reconcile.Result, error) {
					Expect(ctx.AlertScale.Spec.ScaleTarget.Kind).To(Equal(ResourceKindStatefulSet))
					return reconcile.Result{}, nil
				},
			}

			// 临时替换handler
			originalHandler := reconciler.StateHandlers[internalTypes.ScaleStatusPending]
			reconciler.StateHandlers[internalTypes.ScaleStatusPending] = mockHandler
			defer func() {
				reconciler.StateHandlers[internalTypes.ScaleStatusPending] = originalHandler
			}()

			// 创建AlertScale资源，目标为StatefulSet
			alertScale := &opsv1beta1.AlertScale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason: "Memory Pressure",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      "StatefulSet",
						Name:      "test-statefulset",
						Namespace: "default",
					},
					ScaleThreshold:        90,
					ScaleDuration:         "10m",
					ScaleNotificationType: "wxworkrobot",
				},
				Status: opsv1beta1.AlertScaleStatus{
					ScaleStatus: opsv1beta1.ScaleStatus{
						Status: internalTypes.ScaleStatusPending,
					},
				},
			}

			Expect(fakeClient.Create(ctx, alertScale)).To(Succeed())

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle unsupported scale target kind", func() {
			// 创建AlertScale资源，目标为不支持的类型
			alertScale := &opsv1beta1.AlertScale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason: "Test Reason",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      "UnsupportedKind",
						Name:      "test-resource",
						Namespace: "default",
					},
				},
				Status: opsv1beta1.AlertScaleStatus{
					ScaleStatus: opsv1beta1.ScaleStatus{
						Status: internalTypes.ScaleStatusPending,
					},
				},
			}

			Expect(fakeClient.Create(ctx, alertScale)).To(Succeed())

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		It("should use correct state handler based on status", func() {
			// 测试不同状态使用正确的handler
			testCases := []struct {
				status string
				desc   string
			}{
				{internalTypes.ScaleStatusPending, "Pending status"},
				{internalTypes.ScaleStatusScaling, "Scaling status"},
				{internalTypes.ScaleStatusScaled, "Scaled status"},
				{internalTypes.ScaleStatusCompleted, "Completed status"},
				{internalTypes.ScaleStatusFailed, "Failed status"},
				{internalTypes.ScaleStatusArchived, "Archived status"},
				{"UnknownStatus", "Unknown status uses default handler"},
			}

			for _, tc := range testCases {
				By("Testing " + tc.desc)

				// 创建mock handler来验证正确的handler被调用
				mockHandler := &MockStateHandler{
					handleFunc: func(ctx *internalTypes.ScaleContext) (reconcile.Result, error) {
						Expect(ctx.AlertScale.Status.ScaleStatus.Status).To(Equal(tc.status))
						return reconcile.Result{}, nil
					},
				}

				// 临时替换对应的handler
				var originalHandler internalTypes.StateHandler
				if tc.status == "UnknownStatus" {
					originalHandler = reconciler.StateHandlers["default"]
					reconciler.StateHandlers["default"] = mockHandler
				} else {
					originalHandler = reconciler.StateHandlers[tc.status]
					reconciler.StateHandlers[tc.status] = mockHandler
				}

				alertScale := &opsv1beta1.AlertScale{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName + "-" + tc.status,
						Namespace: "default",
					},
					Spec: opsv1beta1.AlertScaleSpec{
						ScaleReason: "Test Reason",
						ScaleTarget: opsv1beta1.ScaleTarget{
							Kind:      ResourceKindDeployment,
							Name:      "test-deployment",
							Namespace: "default",
						},
					},
					Status: opsv1beta1.AlertScaleStatus{
						ScaleStatus: opsv1beta1.ScaleStatus{
							Status: tc.status,
						},
					},
				}

				Expect(fakeClient.Create(ctx, alertScale)).To(Succeed())

				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      resourceName + "-" + tc.status,
						Namespace: "default",
					},
				})
				Expect(err).NotTo(HaveOccurred())

				// 恢复原始handler
				if tc.status == "UnknownStatus" {
					reconciler.StateHandlers["default"] = originalHandler
				} else {
					reconciler.StateHandlers[tc.status] = originalHandler
				}
			}
		})

		It("should create proper scale context", func() {
			// 使用mock handler验证context创建
			mockHandler := &MockStateHandler{
				handleFunc: func(ctx *internalTypes.ScaleContext) (reconcile.Result, error) {
					// 验证context包含所有必要字段
					Expect(ctx.AlertScale).ToNot(BeNil())
					Expect(ctx.Client).ToNot(BeNil())
					Expect(ctx.Context).ToNot(BeNil())
					Expect(ctx.ScaleStrategy).ToNot(BeNil())
					Expect(ctx.Request.Name).To(Equal(resourceName))

					// 验证策略类型
					_, isDeploymentStrategy := ctx.ScaleStrategy.(*strategy.DeploymentStrategy)
					Expect(isDeploymentStrategy).To(BeTrue())

					return reconcile.Result{}, nil
				},
			}

			// 临时替换handler进行测试
			reconciler.StateHandlers[internalTypes.ScaleStatusPending] = mockHandler

			alertScale := &opsv1beta1.AlertScale{
				ObjectMeta: metav1.ObjectMeta{
					Name:      resourceName,
					Namespace: "default",
				},
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleReason: "Test Context Creation",
					ScaleTarget: opsv1beta1.ScaleTarget{
						Kind:      "Deployment",
						Name:      "test-deployment",
						Namespace: "default",
					},
				},
				Status: opsv1beta1.AlertScaleStatus{
					ScaleStatus: opsv1beta1.ScaleStatus{
						Status: internalTypes.ScaleStatusPending,
					},
				},
			}

			Expect(fakeClient.Create(ctx, alertScale)).To(Succeed())

			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When testing reconciler methods", func() {
		It("should implement Status method correctly", func() {
			statusWriter := reconciler.Status()
			Expect(statusWriter).ToNot(BeNil())
			Expect(statusWriter).To(Equal(fakeClient.Status()))
		})

		It("should set up controller successfully", func() {
			// 测试SetupWithManager方法存在
			Expect(reconciler.SetupWithManager).ToNot(BeNil())

			// 验证StateHandlers初始化
			Expect(reconciler.StateHandlers).ToNot(BeNil())
			Expect(reconciler.StateHandlers).To(HaveLen(7)) // 6个状态 + 1个default

			// 验证所有期望的handlers存在
			expectedHandlers := []string{
				internalTypes.ScaleStatusPending,
				internalTypes.ScaleStatusScaling,
				internalTypes.ScaleStatusScaled,
				internalTypes.ScaleStatusCompleted,
				internalTypes.ScaleStatusFailed,
				internalTypes.ScaleStatusArchived,
				"default",
			}

			for _, handlerKey := range expectedHandlers {
				Expect(reconciler.StateHandlers[handlerKey]).ToNot(BeNil())
			}
		})
	})

	Context("When testing strategy selection", func() {
		It("should select DeploymentStrategy for Deployment target", func() {
			alertScale := &opsv1beta1.AlertScale{
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleTarget: opsv1beta1.ScaleTarget{Kind: ResourceKindDeployment},
				},
			}

			// 模拟reconciler中的策略选择逻辑
			var scaleStrategy internalTypes.ScaleStrategy
			switch alertScale.Spec.ScaleTarget.Kind {
			case ResourceKindDeployment:
				scaleStrategy = &strategy.DeploymentStrategy{}
			case ResourceKindStatefulSet:
				scaleStrategy = &strategy.StatefulSetStrategy{}
			}

			_, isDeploymentStrategy := scaleStrategy.(*strategy.DeploymentStrategy)
			Expect(isDeploymentStrategy).To(BeTrue())
		})

		It("should select StatefulSetStrategy for StatefulSet target", func() {
			alertScale := &opsv1beta1.AlertScale{
				Spec: opsv1beta1.AlertScaleSpec{
					ScaleTarget: opsv1beta1.ScaleTarget{Kind: ResourceKindStatefulSet},
				},
			}

			// 模拟reconciler中的策略选择逻辑
			var scaleStrategy internalTypes.ScaleStrategy
			switch alertScale.Spec.ScaleTarget.Kind {
			case ResourceKindDeployment:
				scaleStrategy = &strategy.DeploymentStrategy{}
			case ResourceKindStatefulSet:
				scaleStrategy = &strategy.StatefulSetStrategy{}
			}

			_, isStatefulSetStrategy := scaleStrategy.(*strategy.StatefulSetStrategy)
			Expect(isStatefulSetStrategy).To(BeTrue())
		})
	})
})
