package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodRebalance 实现 ApprovableResource 接口

// GetAutoApproval 获取是否自动审批
func (pr *PodRebalance) GetAutoApproval() bool {
	return pr.Spec.AutoApproval
}

// GetTimeout 获取超时设置
func (pr *PodRebalance) GetTimeout() string {
	return pr.Spec.Timeout
}

// GetStatus 获取当前状态
func (pr *PodRebalance) GetStatus() string {
	return pr.Status.Status
}

// SetStatus 设置状态
func (pr *PodRebalance) SetStatus(status string) {
	pr.Status.Status = status
}

// GetBeginTime 获取开始时间
func (pr *PodRebalance) GetBeginTime() *metav1.Time {
	if pr.Status.RebalanceBeginTime.IsZero() {
		return nil
	}
	return &pr.Status.RebalanceBeginTime
}

// SetBeginTime 设置开始时间
func (pr *PodRebalance) SetBeginTime(time metav1.Time) {
	pr.Status.RebalanceBeginTime = time
}
