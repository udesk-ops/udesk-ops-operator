package strategy

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// DeploymentStrategy Deployment 扩缩容策略
type DeploymentStrategy struct{}

func (s *DeploymentStrategy) Scale(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
	deployment := &appv1.Deployment{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, deployment); err != nil {
		return err
	}

	patch := client.MergeFrom(deployment.DeepCopy())
	deployment.Spec.Replicas = &replicas

	return r.Patch(ctx, deployment, patch)
}

func (s *DeploymentStrategy) GetCurrentReplicas(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	deployment := &appv1.Deployment{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, deployment); err != nil {
		return 0, err
	}

	if deployment.Spec.Replicas == nil {
		return 0, nil
	}

	return *deployment.Spec.Replicas, nil
}

func (s *DeploymentStrategy) GetAvailableReplicas(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	deployment := &appv1.Deployment{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, deployment); err != nil {
		return 0, err
	}

	return deployment.Status.AvailableReplicas, nil
}

// StatefulSetStrategy StatefulSet 扩缩容策略
type StatefulSetStrategy struct{}

func (s *StatefulSetStrategy) Scale(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
	statefulSet := &appv1.StatefulSet{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, statefulSet); err != nil {
		return err
	}

	patch := client.MergeFrom(statefulSet.DeepCopy())
	statefulSet.Spec.Replicas = &replicas

	return r.Patch(ctx, statefulSet, patch)
}
func (s *StatefulSetStrategy) GetCurrentReplicas(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	statefulSet := &appv1.StatefulSet{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, statefulSet); err != nil {
		return 0, err
	}

	if statefulSet.Spec.Replicas == nil {
		return 0, nil
	}

	return *statefulSet.Spec.Replicas, nil
}
func (s *StatefulSetStrategy) GetAvailableReplicas(ctx context.Context, r client.Client, target *opsv1beta1.ScaleTarget) (int32, error) {
	statefulSet := &appv1.StatefulSet{}
	key := types.NamespacedName{Name: target.Name, Namespace: target.Namespace}

	if err := r.Get(ctx, key, statefulSet); err != nil {
		return 0, err
	}

	return statefulSet.Status.ReadyReplicas, nil
}
