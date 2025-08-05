#!/bin/bash

# 快速审批脚本 - 专门用于AlertScale审批操作
# 使用方法: ./quick_approval.sh [approve|reject] <namespace> <name> <approver> [reason] [comment]

set -e

API_BASE="${API_BASE:-http://localhost:8088/api/v1}"
APPROVER="${APPROVER:-admin@company.com}"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 显示使用帮助
usage() {
    echo "快速审批脚本 - AlertScale审批操作"
    echo ""
    echo "使用方法:"
    echo "  $0 approve <namespace> <name> <approver> [reason] [comment]"
    echo "  $0 reject <namespace> <name> <approver> [reason] [comment]"
    echo ""
    echo "简化使用方法 (使用默认审批人):"
    echo "  $0 approve <namespace> <name> [reason] [comment]"
    echo "  $0 reject <namespace> <name> [reason] [comment]"
    echo ""
    echo "示例:"
    echo "  $0 approve default my-alertscale"
    echo "  $0 approve default my-alertscale admin@company.com"
    echo "  $0 approve default my-alertscale admin@company.com 'High CPU usage'"
    echo "  $0 reject default my-alertscale admin@company.com 'Policy violation'"
    echo ""
    echo "环境变量:"
    echo "  API_BASE  - API服务器地址 (默认: http://localhost:8088/api/v1)"
    echo "  APPROVER  - 默认审批人 (默认: admin@company.com)"
}

# 检查API服务器连通性
check_api_server() {
    echo -e "${BLUE}检查API服务器连通性...${NC}"
    if ! curl -s --connect-timeout 5 "$API_BASE/health" > /dev/null; then
        echo -e "${RED}错误: 无法连接到API服务器 $API_BASE${NC}"
        echo "请确保API服务器正在运行"
        exit 1
    fi
    echo -e "${GREEN}✓ API服务器连接正常${NC}"
}

# 审批AlertScale
approve_alertscale() {
    local namespace="$1"
    local name="$2"
    local approver="$3"
    local reason="$4"
    local comment="$5"
    
    # 设置默认值
    reason="${reason:-Scale approved for operational requirements}"
    comment="${comment:-Approved via quick approval script}"
    
    echo -e "${BLUE}审批通过AlertScale${NC}"
    echo -e "  命名空间: ${YELLOW}$namespace${NC}"
    echo -e "  名称: ${YELLOW}$name${NC}"
    echo -e "  审批人: ${YELLOW}$approver${NC}"
    echo -e "  理由: ${YELLOW}$reason${NC}"
    echo -e "  备注: ${YELLOW}$comment${NC}"
    echo ""
    
    local payload=$(cat <<EOF
{
  "approver": "$approver",
  "reason": "$reason",
  "comment": "$comment"
}
EOF
)
    
    echo -e "${BLUE}发送审批请求...${NC}"
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$API_BASE/alertscales/$namespace/$name/approve")
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✓ 请求发送成功${NC}"
        echo ""
        echo "响应:"
        if command -v jq >/dev/null 2>&1; then
            echo "$response" | jq .
        else
            echo "$response"
        fi
    else
        echo -e "${RED}✗ 请求发送失败${NC}"
        exit 1
    fi
}

# 拒绝AlertScale
reject_alertscale() {
    local namespace="$1"
    local name="$2"
    local approver="$3"
    local reason="$4"
    local comment="$5"
    
    # 设置默认值
    reason="${reason:-Scale rejected due to policy constraints}"
    comment="${comment:-Rejected via quick approval script}"
    
    echo -e "${BLUE}拒绝AlertScale${NC}"
    echo -e "  命名空间: ${YELLOW}$namespace${NC}"
    echo -e "  名称: ${YELLOW}$name${NC}"
    echo -e "  审批人: ${YELLOW}$approver${NC}"
    echo -e "  理由: ${YELLOW}$reason${NC}"
    echo -e "  备注: ${YELLOW}$comment${NC}"
    echo ""
    
    local payload=$(cat <<EOF
{
  "approver": "$approver",
  "reason": "$reason",
  "comment": "$comment"
}
EOF
)
    
    echo -e "${BLUE}发送拒绝请求...${NC}"
    local response=$(curl -s -X POST \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "$API_BASE/alertscales/$namespace/$name/reject")
    
    if [[ $? -eq 0 ]]; then
        echo -e "${GREEN}✓ 请求发送成功${NC}"
        echo ""
        echo "响应:"
        if command -v jq >/dev/null 2>&1; then
            echo "$response" | jq .
        else
            echo "$response"
        fi
    else
        echo -e "${RED}✗ 请求发送失败${NC}"
        exit 1
    fi
}

# 主函数
main() {
    local action="$1"
    local namespace="$2"
    local name="$3"
    local approver="$4"
    local reason="$5"
    local comment="$6"
    
    # 检查参数
    if [[ -z "$action" || -z "$namespace" || -z "$name" ]]; then
        usage
        exit 1
    fi
    
    # 如果没有提供approver，检查是否第4个参数看起来像email或者使用默认值
    if [[ -z "$approver" ]]; then
        approver="$APPROVER"
    elif [[ "$approver" != *"@"* ]]; then
        # 第4个参数不像email，可能是reason，重新排列参数
        comment="$reason"
        reason="$approver"
        approver="$APPROVER"
    fi
    
    # 检查API服务器
    check_api_server
    echo ""
    
    case "$action" in
        "approve")
            approve_alertscale "$namespace" "$name" "$approver" "$reason" "$comment"
            ;;
        "reject")
            reject_alertscale "$namespace" "$name" "$approver" "$reason" "$comment"
            ;;
        *)
            echo -e "${RED}错误: 未知操作 '$action'${NC}"
            echo "支持的操作: approve, reject"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
