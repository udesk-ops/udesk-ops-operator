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

package strategy

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

var _ = Describe("Scale Strategy", func() {
	var (
		strategy   *DeploymentStrategy
		fakeClient client.Client
		ctx        context.Context
		deployment *appv1.Deployment
		target     *opsv1beta1.ScaleTarget
	)

	BeforeEach(func() {
		strategy = &DeploymentStrategy{}
		ctx = context.Background()

		// Create test scheme
		scheme := runtime.NewScheme()
		_ = appv1.AddToScheme(scheme)
		_ = opsv1beta1.AddToScheme(scheme)

		// Create test deployment
		deployment = &appv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-deployment",
				Namespace: "default",
			},
			Spec: appv1.DeploymentSpec{
				Replicas: int32Ptr(3),
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "test"},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"app": "test"},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "test-container",
								Image: "nginx:latest",
							},
						},
					},
				},
			},
		}

		// Create fake client with deployment
		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(deployment).
			Build()

		// Create scale target
		target = &opsv1beta1.ScaleTarget{
			Kind:      "Deployment",
			Name:      "test-deployment",
			Namespace: "default",
		}
	})

	Describe("DeploymentStrategy", func() {
		Context("when scaling deployment", func() {
			It("should scale deployment to specified replicas", func() {
				err := strategy.Scale(ctx, fakeClient, target, 5)
				Expect(err).NotTo(HaveOccurred())

				// Verify the deployment was scaled
				updatedDeployment := &appv1.Deployment{}
				key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}
				err = fakeClient.Get(ctx, key, updatedDeployment)
				Expect(err).NotTo(HaveOccurred())
				Expect(*updatedDeployment.Spec.Replicas).To(Equal(int32(5)))
			})

			It("should return error when deployment not found", func() {
				target.Name = "non-existent-deployment"
				err := strategy.Scale(ctx, fakeClient, target, 5)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when getting current replicas", func() {
			It("should return current replica count", func() {
				replicas, err := strategy.GetCurrentReplicas(ctx, fakeClient, target)
				Expect(err).NotTo(HaveOccurred())
				Expect(replicas).To(Equal(int32(3)))
			})

			It("should return error when deployment not found", func() {
				target.Name = "non-existent-deployment"
				_, err := strategy.GetCurrentReplicas(ctx, fakeClient, target)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

// Helper function to create int32 pointer
func int32Ptr(i int32) *int32 {
	return &i
}
