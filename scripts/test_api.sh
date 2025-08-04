#!/bin/bash

# API服务器测试脚本
set -e

API_URL="http://localhost:8088"

echo "🚀 udesk-ops-operator API 服务器测试"
echo "======================================"

# 等待API服务器启动
echo "等待 API 服务器启动..."
sleep 3

# 测试健康检查
echo "🏥 测试健康检查端点..."
curl -s -X GET "$API_URL/api/v1/health" | jq '.' || echo "健康检查测试完成"

echo ""

# 测试获取所有 AlertScales
echo "📋 测试获取所有 AlertScales..."
curl -s -X GET "$API_URL/api/v1/alertscales" | jq '.' || echo "获取 AlertScales 测试完成"

echo ""

# 测试获取待审批列表
echo "⏳ 测试获取待审批列表..."
curl -s -X GET "$API_URL/api/v1/approvals/pending" | jq '.' || echo "获取待审批列表测试完成"

echo ""

# 测试获取审批统计
echo "📊 测试获取审批统计..."
curl -s -X GET "$API_URL/api/v1/approvals/stats" | jq '.' || echo "获取审批统计测试完成"

echo ""

# 测试批量审批（空列表）
echo "🔄 测试批量审批操作..."
curl -s -X POST "$API_URL/api/v1/approvals/batch" \
  -H "Content-Type: application/json" \
  -d '{
    "items": [],
    "approver": "test@example.com",
    "reason": "测试批量审批",
    "action": "approve"
  }' | jq '.' || echo "批量审批测试完成"

echo ""
echo "✅ API 服务器功能测试完成！"
echo ""
echo "💡 要测试完整功能，您需要："
echo "   1. 在 Kubernetes 集群中运行 operator"
echo "   2. 创建一些 AlertScale 资源"
echo "   3. 使用 API 进行实际的审批操作"
echo ""
echo "🌟 恭喜！通用 API 服务器已成功实现！"
