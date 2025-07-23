#!/bin/bash

set -e

echo "=== Testing Webhook Functionality ==="

# 先删除可能存在的测试资源
echo "Cleaning up any existing test resources..."
kubectl delete scalenotifyconfig test-email-1 test-email-2 test-wxwork-1 --ignore-not-found=true

echo ""
echo "1. Testing valid single default config (should succeed)..."
cat <<EOF | kubectl apply -f -
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: test-email-1
spec:
  type: Email
  default: true
  config:
    smtpServer: "smtp.example.com"
    smtpPort: 587
    smtpUser: "test@example.com"
    smtpPassword: "password"
    fromEmail: "noreply@example.com"
    toEmail: "admin@example.com"
EOF

if [ $? -eq 0 ]; then
    echo "✅ First default config created successfully"
else
    echo "❌ Failed to create first default config"
    exit 1
fi

echo ""
echo "2. Testing duplicate default config of same type (should fail)..."
cat <<EOF | kubectl apply -f -
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: test-email-2
spec:
  type: Email
  default: true  # This should be rejected by webhook
  config:
    smtpServer: "smtp2.example.com"
    smtpPort: 587
    smtpUser: "test2@example.com"
    smtpPassword: "password2"
    fromEmail: "noreply2@example.com"
    toEmail: "admin2@example.com"
EOF

if [ $? -ne 0 ]; then
    echo "✅ Duplicate default config correctly rejected by webhook"
else
    echo "❌ Webhook should have rejected duplicate default config"
fi

echo ""
echo "3. Testing different type default config (should succeed)..."
cat <<EOF | kubectl apply -f -
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: test-wxwork-1
spec:
  type: WXWorkRobot
  default: true
  config:
    webhookURL: "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=example"
    secret: "secret"
    messageTemplate: "Alert: {{.Message}}"
    atUsers: ["user1", "user2"]
    atAll: false
EOF

if [ $? -eq 0 ]; then
    echo "✅ Different type default config created successfully"
else
    echo "❌ Failed to create different type default config"
fi

echo ""
echo "4. Testing invalid config (should fail)..."
cat <<EOF | kubectl apply -f -
apiVersion: ops.udesk.cn/v1beta1
kind: ScaleNotifyConfig
metadata:
  name: test-invalid
spec:
  type: InvalidType
  default: true
  config:
    invalid: "config"
EOF

if [ $? -ne 0 ]; then
    echo "✅ Invalid config correctly rejected by webhook"
else
    echo "❌ Webhook should have rejected invalid config"
fi

echo ""
echo "=== Test Results ==="
echo "Existing ScaleNotifyConfigs:"
kubectl get scalenotifyconfig -o custom-columns=NAME:.metadata.name,TYPE:.spec.type,DEFAULT:.spec.default,STATUS:.status.validationStatus

echo ""
echo "=== Webhook Test Completed ==="
