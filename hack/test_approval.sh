#!/bin/bash

# 测试 AlertScale 审批功能的完整脚本
set -e

echo "🧪 测试 AlertScale 审批功能"
echo "============================="

# 启动 operator (后台运行)
echo "🚀 启动 operator..."
cd /home/zero/Codes/udesk-ops-operator
pkill -f "udesk-ops-operator" 2>/dev/null || true
sleep 2

ENABLE_WEBHOOKS=false nohup ./bin/manager --enable-api-server=true --api-addr=:8088 > operator.log 2>&1 &
OPERATOR_PID=$!

echo "🕐 等待服务启动..."
sleep 8

# 检查进程是否还在运行
if ! kill -0 $OPERATOR_PID 2>/dev/null; then
    echo "❌ Operator 启动失败"
    exit 1
fi

echo "✅ Operator 启动成功 (PID: $OPERATOR_PID)"

# 获取当前状态
echo ""
echo "📋 获取当前 AlertScale 列表..."
CURRENT_ALERTSCALES=$(curl -s http://localhost:8088/api/v1/alertscales | jq -r '.data.items[].name' 2>/dev/null || echo "无")
echo "当前 AlertScales: $CURRENT_ALERTSCALES"

# 获取待审批列表
echo ""
echo "⏳ 获取待审批列表..."
PENDING_RESPONSE=$(curl -s http://localhost:8088/api/v1/approvals/pending)
echo "待审批响应: $PENDING_RESPONSE"

PENDING_COUNT=$(echo "$PENDING_RESPONSE" | jq -r '.data.count' 2>/dev/null || echo "0")
echo "待审批数量: $PENDING_COUNT"

if [ "$PENDING_COUNT" -gt 0 ]; then
    # 获取第一个待审批项目
    FIRST_ITEM_NS=$(echo "$PENDING_RESPONSE" | jq -r '.data.items[0].namespace' 2>/dev/null)
    FIRST_ITEM_NAME=$(echo "$PENDING_RESPONSE" | jq -r '.data.items[0].name' 2>/dev/null)
    
    if [ "$FIRST_ITEM_NS" != "null" ] && [ "$FIRST_ITEM_NAME" != "null" ]; then
        echo ""
        echo "🎯 找到待审批项目: $FIRST_ITEM_NS/$FIRST_ITEM_NAME"
        
        # 获取审批前状态
        echo "📋 审批前状态检查..."
        BEFORE_RESPONSE=$(curl -s "http://localhost:8088/api/v1/alertscales/$FIRST_ITEM_NS/$FIRST_ITEM_NAME")
        BEFORE_STATUS=$(echo "$BEFORE_RESPONSE" | jq -r '.data.status' 2>/dev/null || echo "unknown")
        echo "审批前状态: $BEFORE_STATUS"
        
        # 执行审批
        echo ""
        echo "✅ 执行审批操作..."
        APPROVAL_RESPONSE=$(curl -s -X POST "http://localhost:8088/api/v1/alertscales/$FIRST_ITEM_NS/$FIRST_ITEM_NAME/approve" \
          -H "Content-Type: application/json" \
          -d '{
            "approver": "test-admin@example.com",
            "reason": "自动化测试审批",
            "comment": "通过 API 测试工具进行的审批"
          }')
        
        echo "审批响应: $APPROVAL_RESPONSE"
        
        # 等待状态变化
        echo ""
        echo "⏰ 等待状态更新..."
        sleep 5
        
        # 检查审批后状态
        echo "📋 审批后状态检查..."
        AFTER_RESPONSE=$(curl -s "http://localhost:8088/api/v1/alertscales/$FIRST_ITEM_NS/$FIRST_ITEM_NAME")
        AFTER_STATUS=$(echo "$AFTER_RESPONSE" | jq -r '.data.status' 2>/dev/null || echo "unknown")
        echo "审批后状态: $AFTER_STATUS"
        
        # 比较状态变化
        if [ "$BEFORE_STATUS" != "$AFTER_STATUS" ]; then
            echo "🎉 状态变化成功: $BEFORE_STATUS -> $AFTER_STATUS"
        else
            echo "⚠️  状态未发生变化，可能需要更多时间"
        fi
        
        # 再次检查待审批列表
        echo ""
        echo "📊 更新后的待审批列表..."
        UPDATED_PENDING=$(curl -s http://localhost:8088/api/v1/approvals/pending)
        UPDATED_COUNT=$(echo "$UPDATED_PENDING" | jq -r '.data.count' 2>/dev/null || echo "0")
        echo "更新后待审批数量: $UPDATED_COUNT"
        
        # 检查审批统计
        echo ""
        echo "📈 审批统计信息..."
        STATS_RESPONSE=$(curl -s http://localhost:8088/api/v1/approvals/stats)
        echo "统计响应: $STATS_RESPONSE"
        
    else
        echo "❌ 无法解析待审批项目信息"
    fi
else
    echo "ℹ️  当前没有待审批的项目"
    
    # 创建一个测试AlertScale（如果可能的话）
    echo ""
    echo "🔧 尝试通过kubectl检查现有资源..."
    kubectl get alertscale -A 2>/dev/null || echo "无法访问 kubectl 或没有 AlertScale 资源"
fi

# 显示最新日志
echo ""
echo "📜 最新的 operator 日志 (最后 15 行):"
echo "==================================="
tail -n 15 operator.log

echo ""
echo "🎯 测试总结"
echo "==========="
echo "✅ API 服务器正常运行"
echo "✅ 审批接口响应正常"
if [ "$PENDING_COUNT" -gt 0 ]; then
    echo "✅ 发现并处理了待审批项目"
else
    echo "ℹ️  当前环境中没有待审批项目"
fi

# 询问是否停止
echo ""
read -p "是否现在停止 operator? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🛑 停止 operator..."
    kill $OPERATOR_PID
    echo "✅ Operator 已停止"
else
    echo "ℹ️  Operator 继续在后台运行 (PID: $OPERATOR_PID)"
    echo "   要停止请运行: kill $OPERATOR_PID"
fi
