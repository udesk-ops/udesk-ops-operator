package controller

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"

	"udesk.cn/ops/internal/handler"
	"udesk.cn/ops/internal/strategy"
	"udesk.cn/ops/internal/types"
)

//+kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ops.udesk.cn,resources=alertscales/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

const (
	// ResourceKindDeployment represents Deployment resource kind
	ResourceKindDeployment = "Deployment"
	// ResourceKindStatefulSet represents StatefulSet resource kind
	ResourceKindStatefulSet = "StatefulSet"
)

// AlertScaleReconciler reconciles a AlertScale object
type AlertScaleReconciler struct {
	client.Client
	Scheme        *runtime.Scheme
	StateHandlers map[string]types.StateHandler
}

// 确保 AlertScaleReconciler 实现了 ScaleReconciler 接口

func (r *AlertScaleReconciler) Status() client.StatusWriter {
	return r.Client.Status()
}

func (r *AlertScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	// 获取 AlertScale 资源
	alertScale := &opsv1beta1.AlertScale{}
	if err := r.Get(ctx, req.NamespacedName, alertScale); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 根据目标类型选择策略
	var scaleStrategy types.ScaleStrategy
	switch alertScale.Spec.ScaleTarget.Kind {
	case ResourceKindDeployment:
		scaleStrategy = &strategy.DeploymentStrategy{}
	case ResourceKindStatefulSet:
		scaleStrategy = &strategy.StatefulSetStrategy{}
	default:
		log.Info("Unsupported scale target kind", "kind", alertScale.Spec.ScaleTarget.Kind)
		return ctrl.Result{}, nil
	}

	// 创建上下文
	scaleContext := &types.ScaleContext{
		AlertScale:    alertScale,
		Client:        r.Client,
		Request:       req,
		Context:       ctx,
		ScaleStrategy: scaleStrategy,
	}

	// 获取当前状态处理器
	currentStatus := alertScale.Status.ScaleStatus.Status
	stateHandler, exists := r.StateHandlers[currentStatus]
	if !exists {
		stateHandler = r.StateHandlers["default"]
	}

	return stateHandler.Handle(scaleContext)
}

// SetupWithManager sets up the controller with the Manager.
func (r *AlertScaleReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// 初始化状态处理器
	r.StateHandlers = map[string]types.StateHandler{
		types.ScaleStatusPending:   &handler.PendingHandler{},
		types.ScaleStatusScaling:   &handler.ScalingHandler{},
		types.ScaleStatusScaled:    &handler.ScaledHandler{},
		types.ScaleStatusCompleted: &handler.CompletedHandler{},
		types.ScaleStatusFailed:    &handler.FailedHandler{},
		types.ScaleStatusArchived:  &handler.ArchivedHandler{},
		"default":                  &handler.DefaultHandler{},
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1beta1.AlertScale{}).
		Owns(&appv1.Deployment{}).
		Named("alertscale").
		Complete(r)
}
