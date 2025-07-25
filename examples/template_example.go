package main

import (
	"context"
	"fmt"
	"log"

	"k8s.io/apimachinery/pkg/runtime"
	"udesk.cn/ops/internal/strategy"
)

func main() {
	ctx := context.Background()

	fmt.Println("=== 企业微信机器人模板渲染示例 ===")

	// 创建配置 - 使用 Go 模板语法
	config := runtime.RawExtension{
		Raw: []byte(`{
			"webhookURL": "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test-key",
			"messageTemplate": "🚨 AlertScale 通知\n\n消息: {{.Message}}\n时间: {{.Time}}\n\n请及时处理！",
			"atUsers": ["@all"],
			"atAll": true
		}`),
	}

	client, err := strategy.NewWXWorkRobotNotificationClient(config)
	if err != nil {
		log.Fatalf("Failed to create WeChat Work client: %v", err)
	}

	// 模拟发送通知（实际不会发送，因为URL是测试用的）
	message := "deployment/my-app 需要从 2 个副本扩容到 5 个副本"
	fmt.Printf("原始消息: %s\n", message)

	if err := client.SendNotify(ctx, message); err != nil {
		// 这里会失败是正常的，因为我们使用的是测试URL
		fmt.Printf("发送失败（预期的）: %v\n", err)
	}

	fmt.Println("\n=== 邮件通知模板渲染示例 ===")

	// 邮件配置 - 使用 Go 模板语法
	emailConfig := runtime.RawExtension{
		Raw: []byte(`{
			"smtpServer": "smtp.example.com",
			"smtpPort": 587,
			"username": "test@example.com",
			"password": "password",
			"fromEmail": "noreply@company.com",
			"toEmails": ["admin@company.com"],
			"subject": "AlertScale 扩容通知",
			"messageTemplate": "AlertScale 通知\n\n详情: {{.Message}}\n发送时间: {{.Time}}\n\n此邮件由系统自动发送。"
		}`),
	}

	emailClient, err := strategy.NewEmailNotificationClient(emailConfig)
	if err != nil {
		log.Fatalf("Failed to create Email client: %v", err)
	}

	// 模拟发送邮件（实际不会发送，因为SMTP配置是测试用的）
	emailMessage := "statefulset/database 需要从 3 个副本扩容到 6 个副本"
	fmt.Printf("原始消息: %s\n", emailMessage)

	if err := emailClient.SendNotify(ctx, emailMessage); err != nil {
		// 这里会失败是正常的，因为我们使用的是测试SMTP配置
		fmt.Printf("发送失败（预期的）: %v\n", err)
	}

	fmt.Println("\n=== 模板语法演示 ===")
	fmt.Println("支持的模板变量:")
	fmt.Println("- {{.Message}} : 扩容消息内容")
	fmt.Println("- {{.Time}}    : 发送时间 (RFC3339格式)")
	fmt.Println("")
	fmt.Println("示例模板:")
	fmt.Println(`"🚨 AlertScale 通知\n\n消息: {{.Message}}\n时间: {{.Time}}"`)
	fmt.Println("")
	fmt.Println("渲染结果类似:")
	fmt.Println("🚨 AlertScale 通知")
	fmt.Println("")
	fmt.Println("消息: deployment/my-app 需要从 2 个副本扩容到 5 个副本")
	fmt.Println("时间: 2025-07-25T20:35:00+08:00")
}
