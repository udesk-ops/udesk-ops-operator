#!/bin/bash

# udesk-ops-operator API 审批接口调用脚本
# 使用方法: ./api_approval_script.sh [command] [args...]

# API 服务器配置
API_BASE="http://localhost:8088/api/v1"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印帮助信息
show_help() {
    echo -e "${BLUE}udesk-ops-operator API 审批接口调用脚本${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 health                                    # 检查API服务器健康状态"
    echo "  $0 list                                      # 列出所有AlertScale资源"
    echo "  $0 get <namespace> <name>                    # 获取指定AlertScale详情"
    echo "  $0 approve <namespace> <name> <approver>     # 审批通过AlertScale"
    echo "  $0 reject <namespace> <name> <approver>      # 拒绝AlertScale"
    echo ""
    echo "示例:"
    echo "  $0 health"
    echo "  $0 list"
    echo "  $0 get default my-alertscale"
    echo "  $0 approve default my-alertscale admin@company.com"
    echo "  $0 reject default my-alertscale admin@company.com"
    echo ""
    echo "环境变量:"
    echo "  API_BASE    - API服务器地址 (默认: http://localhost:8088/api/v1)"
    echo "  CURL_OPTS   - 额外的curl选项"
}

# 发送HTTP请求的通用函数
send_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local content_type="$4"
    
    local curl_cmd="curl -s -X $method"
    
    # 添加Content-Type头
    if [[ -n "$content_type" ]]; then
        curl_cmd="$curl_cmd -H 'Content-Type: $content_type'"
    fi
    
    # 添加数据
    if [[ -n "$data" ]]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    # 添加额外选项
    if [[ -n "$CURL_OPTS" ]]; then
        curl_cmd="$curl_cmd $CURL_OPTS"
    fi
    
    # 执行请求
    curl_cmd="$curl_cmd '$API_BASE$endpoint'"
    
    echo -e "${YELLOW}执行请求: $method $API_BASE$endpoint${NC}"
    if [[ -n "$data" ]]; then
        echo -e "${YELLOW}请求数据: $data${NC}"
    fi
    echo ""
    
    local response=$(eval $curl_cmd)
    local exit_code=$?
    
    if [[ $exit_code -ne 0 ]]; then
        echo -e "${RED}错误: curl请求失败 (退出代码: $exit_code)${NC}"
        return 1
    fi
    
    # 美化JSON输出
    if command -v jq >/dev/null 2>&1; then
        echo "$response" | jq .
    else
        echo "$response"
    fi
    
    # 检查响应中的success字段
    if command -v jq >/dev/null 2>&1; then
        local success=$(echo "$response" | jq -r '.success // empty')
        if [[ "$success" == "true" ]]; then
            echo -e "\n${GREEN}✓ 请求成功${NC}"
        elif [[ "$success" == "false" ]]; then
            echo -e "\n${RED}✗ 请求失败${NC}"
        fi
    fi
    
    return 0
}

# 检查健康状态
check_health() {
    echo -e "${BLUE}检查API服务器健康状态...${NC}"
    send_request "GET" "/health"
}

# 列出所有AlertScale
list_alertscales() {
    echo -e "${BLUE}获取所有AlertScale资源...${NC}"
    send_request "GET" "/alertscales"
}

# 获取指定AlertScale
get_alertscale() {
    local namespace="$1"
    local name="$2"
    
    if [[ -z "$namespace" || -z "$name" ]]; then
        echo -e "${RED}错误: 需要提供namespace和name参数${NC}"
        echo "使用方法: $0 get <namespace> <name>"
        return 1
    fi
    
    echo -e "${BLUE}获取AlertScale: $namespace/$name${NC}"
    send_request "GET" "/alertscales/$namespace/$name"
}

# 审批通过AlertScale
approve_alertscale() {
    local namespace="$1"
    local name="$2"
    local approver="$3"
    local reason="$4"
    local comment="$5"
    
    if [[ -z "$namespace" || -z "$name" || -z "$approver" ]]; then
        echo -e "${RED}错误: 需要提供namespace、name和approver参数${NC}"
        echo "使用方法: $0 approve <namespace> <name> <approver> [reason] [comment]"
        return 1
    fi
    
    # 设置默认值
    reason="${reason:-Scale approved for operational needs}"
    comment="${comment:-Approved via API script}"
    
    local data=$(cat <<EOF
{
  "approver": "$approver",
  "reason": "$reason",
  "comment": "$comment"
}
EOF
)
    
    echo -e "${BLUE}审批通过AlertScale: $namespace/$name${NC}"
    echo -e "${YELLOW}审批人: $approver${NC}"
    echo -e "${YELLOW}审批理由: $reason${NC}"
    echo -e "${YELLOW}备注: $comment${NC}"
    echo ""
    
    send_request "POST" "/alertscales/$namespace/$name/approve" "$data" "application/json"
}

# 拒绝AlertScale
reject_alertscale() {
    local namespace="$1"
    local name="$2"
    local approver="$3"
    local reason="$4"
    local comment="$5"
    
    if [[ -z "$namespace" || -z "$name" || -z "$approver" ]]; then
        echo -e "${RED}错误: 需要提供namespace、name和approver参数${NC}"
        echo "使用方法: $0 reject <namespace> <name> <approver> [reason] [comment]"
        return 1
    fi
    
    # 设置默认值
    reason="${reason:-Scale rejected due to policy violation}"
    comment="${comment:-Rejected via API script}"
    
    local data=$(cat <<EOF
{
  "approver": "$approver",
  "reason": "$reason",
  "comment": "$comment"
}
EOF
)
    
    echo -e "${BLUE}拒绝AlertScale: $namespace/$name${NC}"
    echo -e "${YELLOW}审批人: $approver${NC}"
    echo -e "${YELLOW}拒绝理由: $reason${NC}"
    echo -e "${YELLOW}备注: $comment${NC}"
    echo ""
    
    send_request "POST" "/alertscales/$namespace/$name/reject" "$data" "application/json"
}

# 批量操作示例
batch_example() {
    echo -e "${BLUE}批量操作示例${NC}"
    echo ""
    
    # 1. 检查健康状态
    echo -e "${YELLOW}1. 检查API健康状态${NC}"
    check_health
    echo ""
    
    # 2. 列出所有AlertScale
    echo -e "${YELLOW}2. 列出所有AlertScale${NC}"
    list_alertscales
    echo ""
    
    # 3. 获取特定AlertScale (示例)
    echo -e "${YELLOW}3. 获取示例AlertScale${NC}"
    get_alertscale "default" "example-alertscale"
    echo ""
    
    echo -e "${GREEN}批量操作示例完成${NC}"
}

# 交互式模式
interactive_mode() {
    echo -e "${BLUE}进入交互式模式${NC}"
    echo "输入 'help' 查看可用命令，输入 'exit' 退出"
    echo ""
    
    while true; do
        echo -n -e "${GREEN}udesk-ops-api> ${NC}"
        read -r input
        
        if [[ -z "$input" ]]; then
            continue
        fi
        
        case "$input" in
            "help")
                show_help
                ;;
            "exit"|"quit")
                echo -e "${BLUE}退出交互式模式${NC}"
                break
                ;;
            "health")
                check_health
                ;;
            "list")
                list_alertscales
                ;;
            *)
                echo -e "${RED}未知命令: $input${NC}"
                echo "输入 'help' 查看可用命令"
                ;;
        esac
        echo ""
    done
}

# 主函数
main() {
    local command="$1"
    
    case "$command" in
        "health")
            check_health
            ;;
        "list")
            list_alertscales
            ;;
        "get")
            get_alertscale "$2" "$3"
            ;;
        "approve")
            approve_alertscale "$2" "$3" "$4" "$5" "$6"
            ;;
        "reject")
            reject_alertscale "$2" "$3" "$4" "$5" "$6"
            ;;
        "batch")
            batch_example
            ;;
        "interactive"|"i")
            interactive_mode
            ;;
        "help"|"-h"|"--help"|"")
            show_help
            ;;
        *)
            echo -e "${RED}错误: 未知命令 '$command'${NC}"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

# 检查jq是否安装
if ! command -v jq >/dev/null 2>&1; then
    echo -e "${YELLOW}警告: 未找到jq命令，JSON输出将不会被美化${NC}"
    echo -e "${YELLOW}建议安装jq: sudo apt-get install jq (Ubuntu/Debian) 或 brew install jq (macOS)${NC}"
    echo ""
fi

# 执行主函数
main "$@"
