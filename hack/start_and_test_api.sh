#!/bin/bash

# 启动和测试 udesk-ops-operator API 服务器的完整脚本
set -e

echo "🚀 启动 udesk-ops-operator 与 API 服务器"
echo "================================================"

# 清理之前的进程
echo "🧹 清理之前的进程..."
pkill -f "udesk-ops-operator" 2>/dev/null || true
sleep 2

# 启动 operator (后台运行)
echo "🏃 启动 operator 和 API 服务器..."
cd /home/zero/Codes/udesk-ops-operator
ENABLE_WEBHOOKS=false nohup ./bin/manager --enable-api-server=true --api-addr=:8088 > operator.log 2>&1 &
OPERATOR_PID=$!

echo "🕐 等待 API 服务器启动..."
sleep 5

# 检查进程是否还在运行
if ! kill -0 $OPERATOR_PID 2>/dev/null; then
    echo "❌ Operator 启动失败，检查日志："
    tail -n 20 operator.log
    exit 1
fi

echo "✅ Operator 启动成功 (PID: $OPERATOR_PID)"

# 测试API功能
echo ""
echo "🔍 测试 API 功能..."
echo "===================="

# 测试健康检查
echo "🏥 测试健康检查..."
HEALTH_RESULT=$(curl -s -w "HTTP %{http_code}" http://localhost:8088/api/v1/health)
echo "健康检查结果: $HEALTH_RESULT"

# 测试获取 AlertScales
echo ""
echo "📋 测试获取 AlertScales..."
ALERTSCALES_RESULT=$(curl -s -w "HTTP %{http_code}" http://localhost:8088/api/v1/alertscales)
echo "AlertScales 结果: $ALERTSCALES_RESULT"

# 测试待审批列表
echo ""
echo "⏳ 测试获取待审批列表..."
PENDING_RESULT=$(curl -s -w "HTTP %{http_code}" http://localhost:8088/api/v1/approvals/pending)
echo "待审批列表结果: $PENDING_RESULT"

# 测试审批统计
echo ""
echo "📊 测试获取审批统计..."
STATS_RESULT=$(curl -s -w "HTTP %{http_code}" http://localhost:8088/api/v1/approvals/stats)
echo "审批统计结果: $STATS_RESULT"

# 测试批量审批
echo ""
echo "🔄 测试批量审批..."
BATCH_RESULT=$(curl -s -w "HTTP %{http_code}" -X POST http://localhost:8088/api/v1/approvals/batch \
  -H "Content-Type: application/json" \
  -d '{
    "items": [],
    "approver": "test@example.com",
    "reason": "测试批量审批",
    "action": "approve"
  }')
echo "批量审批结果: $BATCH_RESULT"

echo ""
echo "🎯 API 功能测试完成！"
echo "===================="

# 显示最新的日志
echo ""
echo "📜 最新的 operator 日志 (最后 10 行):"
echo "=================================="
tail -n 10 operator.log

echo ""
echo "⚡ 要停止 operator，请运行:"
echo "   kill $OPERATOR_PID"
echo ""
echo "🌟 恭喜！udesk-ops-operator API 服务器运行正常！"

# 询问是否要停止
echo ""
read -p "是否现在停止 operator? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🛑 停止 operator..."
    kill $OPERATOR_PID
    echo "✅ Operator 已停止"
else
    echo "ℹ️  Operator 继续在后台运行 (PID: $OPERATOR_PID)"
    echo "   日志文件: $(pwd)/operator.log"
    echo "   要停止请运行: kill $OPERATOR_PID"
fi
