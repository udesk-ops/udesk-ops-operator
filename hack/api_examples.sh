#!/bin/bash

# udesk-ops-operator API 常用操作示例
# 这个脚本包含了常见的API调用示例

API_BASE="http://localhost:8088/api/v1"

echo "=== udesk-ops-operator API 使用示例 ==="
echo ""

# 1. 健康检查
echo "1. 健康检查"
echo "curl -X GET $API_BASE/health"
curl -X GET "$API_BASE/health" | jq . 2>/dev/null || curl -X GET "$API_BASE/health"
echo ""
echo ""

# 2. 列出所有AlertScale
echo "2. 列出所有AlertScale"
echo "curl -X GET $API_BASE/alertscales"
curl -X GET "$API_BASE/alertscales" | jq . 2>/dev/null || curl -X GET "$API_BASE/alertscales"
echo ""
echo ""

# 3. 获取特定AlertScale
echo "3. 获取特定AlertScale"
echo "curl -X GET $API_BASE/alertscales/default/my-alertscale"
curl -X GET "$API_BASE/alertscales/default/my-alertscale" | jq . 2>/dev/null || curl -X GET "$API_BASE/alertscales/default/my-alertscale"
echo ""
echo ""

# 4. 审批通过AlertScale
echo "4. 审批通过AlertScale"
cat << 'EOF'
curl -X POST http://localhost:8088/api/v1/alertscales/default/my-alertscale/approve \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "CPU usage is high, scaling approved",
    "comment": "Approved during business hours"
  }'
EOF

curl -X POST "$API_BASE/alertscales/default/my-alertscale/approve" \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "CPU usage is high, scaling approved",
    "comment": "Approved during business hours"
  }' | jq . 2>/dev/null || curl -X POST "$API_BASE/alertscales/default/my-alertscale/approve" \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "CPU usage is high, scaling approved",
    "comment": "Approved during business hours"
  }'
echo ""
echo ""

# 5. 拒绝AlertScale
echo "5. 拒绝AlertScale"
cat << 'EOF'
curl -X POST http://localhost:8088/api/v1/alertscales/default/another-alertscale/reject \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "Policy violation - scaling not allowed during maintenance",
    "comment": "Rejected due to scheduled maintenance window"
  }'
EOF

curl -X POST "$API_BASE/alertscales/default/another-alertscale/reject" \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "Policy violation - scaling not allowed during maintenance",
    "comment": "Rejected due to scheduled maintenance window"
  }' | jq . 2>/dev/null || curl -X POST "$API_BASE/alertscales/default/another-alertscale/reject" \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "admin@company.com",
    "reason": "Policy violation - scaling not allowed during maintenance",
    "comment": "Rejected due to scheduled maintenance window"
  }'
echo ""
echo ""

echo "=== 示例完成 ==="
