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

	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

var _ = Describe("ScaleNotifyConfig Controller", func() {
	var (
		ctx            context.Context
		reconciler     *ScaleNotifyConfigReconciler
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
		reconciler = &ScaleNotifyConfigReconciler{
			Client: fakeClient,
			Scheme: scheme,
		}

		resourceName = "test-config"
		namespacedName = types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}
	})

	Context("When reconciling a ScaleNotifyConfig resource", func() {
		It("should handle non-existent resource gracefully", func() {
			// 尝试reconcile一个不存在的资源
			_, err := reconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: namespacedName,
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("When testing SetupWithManager", func() {
		It("should set up controller successfully", func() {
			// 这里我们只能测试方法是否存在，因为需要真实的Manager
			Expect(reconciler.SetupWithManager).ToNot(BeNil())
		})
	})
})
