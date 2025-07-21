package controller

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// ScaleStrategy 定义扩缩容策略接口
type ScaleStrategy interface {
	Scale(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error
	GetCurrentReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error)
	GetAvailableReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error)
}

// DeploymentStrategy Deployment 扩缩容策略
type DeploymentStrategy struct{}

func (s *DeploymentStrategy) Scale(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
	deployment := &appv1.Deployment{}
	key := client.ObjectKey{Name: target.Name, Namespace: target.Namespace}

	if err := client.Get(ctx, key, deployment); err != nil {
		return err
	}

	patch := client.MergeFrom(deployment.DeepCopy())
	deployment.Spec.Replicas = &replicas

	return client.Patch(ctx, deployment, patch)
}

func (s *DeploymentStrategy) GetCurrentReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	deployment := &appv1.Deployment{}
	key := client.ObjectKey{Name: target.Name, Namespace: target.Namespace}

	if err := client.Get(ctx, key, deployment); err != nil {
		return 0, err
	}

	if deployment.Spec.Replicas == nil {
		return 0, nil
	}

	return *deployment.Spec.Replicas, nil
}

func (s *DeploymentStrategy) GetAvailableReplicas(ctx context.Context, client client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	deployment := &appv1.Deployment{}
	key := client.ObjectKey{Name: target.Name, Namespace: target.Namespace}

	if err := client.Get(ctx, key, deployment); err != nil {
		return 0, err
	}

	return deployment.Status.AvailableReplicas, nil
}

// StatefulSetStrategy StatefulSet 扩缩容策略
type StatefulSetStrategy struct{}

// 实现 StatefulSet 的扩缩容逻辑...
