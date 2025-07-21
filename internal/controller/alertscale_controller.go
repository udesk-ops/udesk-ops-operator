package controller

import (
	"context"

	"github.com/go-logr/logr"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// StateHandler 定义状态处理接口
type StateHandler interface {
	Handle(ctx *ScaleContext) (ctrl.Result, error)
	CanTransition(toState string) bool
}

// ScaleContext 包含所有状态处理所需的上下文
type ScaleContext struct {
	AlertScale    *opsv1beta1.AlertScale
	Reconciler    *AlertScaleReconciler
	Request       ctrl.Request
	Context       context.Context
	Logger        logr.Logger
	ScaleStrategy ScaleStrategy // 添加扩缩容策略
}

// 状态常量
const (
	ScaleStatusPending   = "Pending"
	ScaleStatusScaling   = "Scaling"
	ScaleStatusScaled    = "Scaled"
	ScaleStatusCompleted = "Completed"
	ScaleStatusFailed    = "Failed"
	ScaleStatusArchived  = "Archived"
)

// AlertScaleReconciler reconciles a AlertScale object
type AlertScaleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	StateHandlers map[string]StateHandler
}

func (r *AlertScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 获取 AlertScale 资源
	alertScale := &opsv1beta1.AlertScale{}
	if err := r.Get(ctx, req.NamespacedName, alertScale); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 根据目标类型选择策略
	var scaleStrategy ScaleStrategy
	switch alertScale.Spec.ScaleTarget.Kind {
	case "Deployment":
		scaleStrategy = &DeploymentStrategy{}
	case "StatefulSet":
		scaleStrategy = &StatefulSetStrategy{}
	default:
		log.Info("Unsupported scale target kind", "kind", alertScale.Spec.ScaleTarget.Kind)
		return ctrl.Result{}, nil
	}

	// 创建上下文
	scaleContext := &ScaleContext{
		AlertScale:    alertScale,
		Reconciler:    r,
		Request:       req,
		Context:       ctx,
		Logger:        log,
		ScaleStrategy: scaleStrategy,
	}

	// 获取当前状态处理器
	currentStatus := alertScale.Status.ScaleStatus.Status
	handler, exists := r.StateHandlers[currentStatus]
	if !exists {
		handler = r.StateHandlers["default"]
	}

	return handler.Handle(scaleContext)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertScaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// 初始化状态处理器
	r.StateHandlers = map[string]StateHandler{
		ScaleStatusPending:   &PendingHandler{},
		ScaleStatusScaling:   &ScalingHandler{},
		ScaleStatusScaled:    &ScaledHandler{},
		ScaleStatusCompleted: &CompletedHandler{},
		ScaleStatusFailed:    &FailedHandler{},
		ScaleStatusArchived:  &ArchivedHandler{},
		"default":            &DefaultHandler{},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.AlertScale{}).
		Owns(&appv1.Deployment{}).
		Named("alertscale").
		Complete(r)
}
