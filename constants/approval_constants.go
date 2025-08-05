package v1beta1

// 审批流相关的注解常量 - 供所有CRD复用
const (
	// ApprovalDecisionAnnotation 存储审批决策 (approve/reject)
	ApprovalDecisionAnnotation = "ops.udesk.cn/approval-decision"

	// ApprovalTimestampAnnotation 存储审批时间戳
	ApprovalTimestampAnnotation = "ops.udesk.cn/approval-timestamp"

	// ApprovalOperatorAnnotation 存储审批操作员
	ApprovalOperatorAnnotation = "ops.udesk.cn/approval-operator"

	// ApprovalReasonAnnotation 存储审批原因
	ApprovalReasonAnnotation = "ops.udesk.cn/approval-reason"

	// ApprovalCommentAnnotation 存储审批备注
	ApprovalCommentAnnotation = "ops.udesk.cn/approval-comment"

	// ApprovalProcessingAnnotation 存储审批处理状态
	ApprovalProcessingAnnotation = "ops.udesk.cn/approval-processing"
)

// 审批决策值常量
const (
	// ApprovalDecisionApprove 批准
	ApprovalDecisionApprove = "approve"

	// ApprovalDecisionReject 拒绝
	ApprovalDecisionReject = "reject"
)

// 审批处理状态常量
const (
	// ApprovalProcessingPending 等待处理
	ApprovalProcessingPending = "pending"

	// ApprovalProcessingCompleted 处理完成
	ApprovalProcessingCompleted = "completed"
)
