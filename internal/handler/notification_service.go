package handler

import (
	"bytes"
	"context"
	"fmt"
	"text/template"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	opsv1beta1 "udesk.cn/ops/api/v1beta1"
	"udesk.cn/ops/internal/strategy"
	scaletypes "udesk.cn/ops/internal/types"
)

// NotificationService 处理通知相关逻辑
type NotificationService struct {
	k8sClient client.Client
}

// NewNotificationService 创建通知服务
func NewNotificationService(k8sClient client.Client) *NotificationService {
	return &NotificationService{
		k8sClient: k8sClient,
	}
}

// TemplateData 模板渲染数据
type TemplateData struct {
	// AlertScale 相关字段
	ScaleReason       string `json:"scaleReason"`
	ScaleDuration     string `json:"scaleDuration"`
	ScaleThreshold    int32  `json:"scaleThreshold"`
	ScaleTimeout      string `json:"scaleTimeout"`
	ScaleAutoApproval bool   `json:"scaleAutoApproval"`

	// ScaleTarget 相关字段
	ScaleTarget struct {
		Name       string `json:"name"`
		Kind       string `json:"kind"`
		Namespace  string `json:"namespace"`
		APIVersion string `json:"apiVersion"`
	} `json:"scaleTarget"`

	// ScaleStatus 相关字段
	Status         string    `json:"status"`
	OriginReplicas int32     `json:"originReplicas"`
	ScaledReplicas int32     `json:"scaledReplicas"`
	ScaleBeginTime time.Time `json:"scaleBeginTime"`
	ScaleEndTime   time.Time `json:"scaleEndTime"`

	// 额外字段
	Timestamp time.Time `json:"timestamp"`
	Operator  string    `json:"operator"`
}

// SendNotification 发送通知
func (ns *NotificationService) SendNotification(ctx context.Context, scaleCtx *scaletypes.ScaleContext, phase string) error {
	log := logf.FromContext(ctx)

	// 检查是否配置了通知类型
	if scaleCtx.AlertScale.Spec.ScaleNotificationType == "" {
		log.V(1).Info("No notification type configured, skipping notification")
		return nil
	}

	// 获取通知客户端
	notifyClient := strategy.DefaultNotifyClientMap[scaleCtx.AlertScale.Spec.ScaleNotificationType]
	if notifyClient == nil {
		log.Info("No notification client found", "type", scaleCtx.AlertScale.Spec.ScaleNotificationType)
		return nil
	}

	// 准备模板数据
	templateData := ns.prepareTemplateData(scaleCtx)

	// 渲染消息内容
	message, err := ns.renderMessage(ctx, scaleCtx, templateData)
	if err != nil {
		log.Error(err, "Failed to render notification message")
		return err
	}

	// 发送通知
	if err := notifyClient.SendNotify(ctx, message); err != nil {
		log.Error(err, "Failed to send notification", "alertScale", scaleCtx.AlertScale.Name)
		return err
	}

	log.Info("Notification sent successfully", "alertScale", scaleCtx.AlertScale.Name)
	return nil
}

// prepareTemplateData 准备模板数据
func (ns *NotificationService) prepareTemplateData(scaleCtx *scaletypes.ScaleContext) *TemplateData {
	data := &TemplateData{
		ScaleReason:       scaleCtx.AlertScale.Spec.ScaleReason,
		ScaleDuration:     scaleCtx.AlertScale.Spec.ScaleDuration,
		ScaleThreshold:    scaleCtx.AlertScale.Spec.ScaleThreshold,
		ScaleTimeout:      scaleCtx.AlertScale.Spec.ScaleTimeout,
		ScaleAutoApproval: scaleCtx.AlertScale.Spec.ScaleAutoApproval,
		Status:            scaleCtx.AlertScale.Status.ScaleStatus.Status,
		OriginReplicas:    scaleCtx.AlertScale.Status.ScaleStatus.OriginReplicas,
		ScaledReplicas:    scaleCtx.AlertScale.Status.ScaleStatus.ScaledReplicas,
		Timestamp:         time.Now(),
		Operator:          "system", // 可以从 context 中获取
	}

	// 设置 ScaleTarget 信息
	data.ScaleTarget.Name = scaleCtx.AlertScale.Spec.ScaleTarget.Name
	data.ScaleTarget.Kind = scaleCtx.AlertScale.Spec.ScaleTarget.Kind
	data.ScaleTarget.Namespace = scaleCtx.AlertScale.Spec.ScaleTarget.Namespace
	data.ScaleTarget.APIVersion = scaleCtx.AlertScale.Spec.ScaleTarget.APIVersion

	// 设置时间信息
	if !scaleCtx.AlertScale.Status.ScaleStatus.ScaleBeginTime.IsZero() {
		data.ScaleBeginTime = scaleCtx.AlertScale.Status.ScaleStatus.ScaleBeginTime.Time
	}
	if !scaleCtx.AlertScale.Status.ScaleStatus.ScaleEndTime.IsZero() {
		data.ScaleEndTime = scaleCtx.AlertScale.Status.ScaleStatus.ScaleEndTime.Time
	}

	return data
}

// renderMessage 渲染消息内容
func (ns *NotificationService) renderMessage(ctx context.Context, scaleCtx *scaletypes.ScaleContext, data *TemplateData) (string, error) {
	// 如果指定了消息模板，使用模板渲染
	if scaleCtx.AlertScale.Spec.ScaleNotifyMsgTemplate != "" {
		return ns.renderWithTemplate(ctx, scaleCtx, data)
	}

	// 否则使用默认消息格式
	return ns.renderDefaultMessage(data), nil
}

// renderWithTemplate 使用指定模板渲染消息
func (ns *NotificationService) renderWithTemplate(ctx context.Context, scaleCtx *scaletypes.ScaleContext, data *TemplateData) (string, error) {
	log := logf.FromContext(ctx)

	// 获取消息模板
	msgTemplate := &opsv1beta1.ScaleNotifyMsgTemplate{}
	templateKey := types.NamespacedName{
		Name:      scaleCtx.AlertScale.Spec.ScaleNotifyMsgTemplate,
		Namespace: scaleCtx.AlertScale.Namespace, // 假设模板在同一命名空间
	}

	if err := ns.k8sClient.Get(ctx, templateKey, msgTemplate); err != nil {
		log.Error(err, "Failed to get message template", "template", scaleCtx.AlertScale.Spec.ScaleNotifyMsgTemplate)
		// 降级到默认消息
		return ns.renderDefaultMessage(data), nil
	}

	// 渲染标题
	titleTmpl, err := template.New("title").Parse(msgTemplate.Spec.Title)
	if err != nil {
		log.Error(err, "Failed to parse title template")
		return "", err
	}

	var titleBuf bytes.Buffer
	if err := titleTmpl.Execute(&titleBuf, data); err != nil {
		log.Error(err, "Failed to execute title template")
		return "", err
	}

	// 渲染内容
	contentTmpl, err := template.New("content").Parse(msgTemplate.Spec.Content)
	if err != nil {
		log.Error(err, "Failed to parse content template")
		return "", err
	}

	var contentBuf bytes.Buffer
	if err := contentTmpl.Execute(&contentBuf, data); err != nil {
		log.Error(err, "Failed to execute content template")
		return "", err
	}

	// 组合标题和内容
	return fmt.Sprintf("**%s**\n\n%s", titleBuf.String(), contentBuf.String()), nil
}

// renderDefaultMessage 渲染默认消息
func (ns *NotificationService) renderDefaultMessage(data *TemplateData) string {
	return fmt.Sprintf(`**扩缩容操作通知**

**目标资源:** %s/%s
**命名空间:** %s
**操作原因:** %s
**当前状态:** %s
**原始副本数:** %d
**目标副本数:** %d
**触发阈值:** %d%%
**持续时间:** %s
**自动审批:** %t

**时间信息:**
- 开始时间: %s
- 当前时间: %s

请及时关注系统状态！`,
		data.ScaleTarget.Kind,
		data.ScaleTarget.Name,
		data.ScaleTarget.Namespace,
		data.ScaleReason,
		data.Status,
		data.OriginReplicas,
		data.ScaledReplicas,
		data.ScaleThreshold,
		data.ScaleDuration,
		data.ScaleAutoApproval,
		data.ScaleBeginTime.Format("2006-01-02 15:04:05"),
		data.Timestamp.Format("2006-01-02 15:04:05"),
	)
}
