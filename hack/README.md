# API 审批接口调用脚本

本目录包含了用于调用 udesk-ops-operator API 审批接口的多个脚本工具。

## 脚本说明

### 1. api_approval_script.sh - 完整功能脚本
功能最全面的API调用脚本，支持所有API接口操作。

**特性:**
- 支持所有API端点（健康检查、列表、获取、审批、拒绝）
- 彩色输出和美化的JSON响应
- 交互式模式
- 批量操作示例
- 完整的错误处理

**使用方法:**
```bash
# 查看帮助
./hack/api_approval_script.sh help

# 检查API健康状态
./hack/api_approval_script.sh health

# 列出所有AlertScale
./hack/api_approval_script.sh list

# 获取特定AlertScale
./hack/api_approval_script.sh get default my-alertscale

# 审批通过
./hack/api_approval_script.sh approve default my-alertscale admin@company.com

# 拒绝审批
./hack/api_approval_script.sh reject default my-alertscale admin@company.com

# 交互式模式
./hack/api_approval_script.sh interactive

# 批量操作示例
./hack/api_approval_script.sh batch
```

### 2. quick_approval.sh - 快速审批脚本
专门用于审批操作的简化脚本，使用更方便。

**特性:**
- 专注于审批操作（approve/reject）
- 支持默认审批人设置
- 自动参数调整
- API连通性检查

**使用方法:**
```bash
# 基本审批（使用默认审批人）
./hack/quick_approval.sh approve default my-alertscale

# 指定审批人
./hack/quick_approval.sh approve default my-alertscale admin@company.com

# 带理由的审批
./hack/quick_approval.sh approve default my-alertscale admin@company.com "High CPU usage"

# 拒绝审批
./hack/quick_approval.sh reject default my-alertscale admin@company.com "Policy violation"

# 设置默认审批人
export APPROVER="your-email@company.com"
./hack/quick_approval.sh approve default my-alertscale
```

### 3. api_examples.sh - 示例演示脚本
包含各种API调用的示例，用于学习和测试。

**使用方法:**
```bash
# 运行所有示例
./hack/api_examples.sh
```

## 环境配置

### API服务器地址
默认API服务器地址为 `http://localhost:8088/api/v1`，可以通过环境变量修改：

```bash
export API_BASE="http://your-server:8088/api/v1"
```

### curl选项
可以通过环境变量添加额外的curl选项：

```bash
export CURL_OPTS="-k --connect-timeout 10"
```

### 默认审批人
为quick_approval.sh设置默认审批人：

```bash
export APPROVER="your-email@company.com"
```

## 依赖要求

### 必需依赖
- `curl` - 用于HTTP请求
- `bash` - 脚本运行环境

### 可选依赖
- `jq` - JSON美化输出（推荐安装）

安装jq：
```bash
# Ubuntu/Debian
sudo apt-get install jq

# macOS
brew install jq

# CentOS/RHEL
sudo yum install jq
```

## API端点说明

### 健康检查
```
GET /api/v1/health
```

### 列出AlertScale
```
GET /api/v1/alertscales
```

### 获取特定AlertScale
```
GET /api/v1/alertscales/{namespace}/{name}
```

### 审批通过
```
POST /api/v1/alertscales/{namespace}/{name}/approve
Content-Type: application/json

{
  "approver": "admin@company.com",
  "reason": "Scale approved for high load",
  "comment": "Scaling approved during peak hours"
}
```

### 拒绝审批
```
POST /api/v1/alertscales/{namespace}/{name}/reject
Content-Type: application/json

{
  "approver": "admin@company.com",
  "reason": "Scale rejected due to policy",
  "comment": "Not approved during maintenance window"
}
```

## 错误排查

### 常见问题

1. **连接失败**
   ```
   错误: 无法连接到API服务器
   ```
   解决方案：
   - 检查API服务器是否运行
   - 确认API_BASE环境变量设置正确
   - 检查网络连接

2. **权限拒绝**
   ```
   HTTP 403 Forbidden
   ```
   解决方案：
   - 确认API服务器配置正确
   - 检查网络策略和防火墙设置

3. **JSON解析错误**
   ```
   parse error: Invalid numeric literal
   ```
   解决方案：
   - 安装jq工具
   - 检查API响应格式

### 调试模式

启用详细输出：
```bash
export CURL_OPTS="-v"
./hack/api_approval_script.sh health
```

## 使用示例

### 典型工作流程
```bash
# 1. 检查API服务状态
./hack/api_approval_script.sh health

# 2. 查看待审批的AlertScale
./hack/api_approval_script.sh list

# 3. 获取特定AlertScale详情
./hack/api_approval_script.sh get default my-alertscale

# 4. 进行审批操作
./hack/quick_approval.sh approve default my-alertscale admin@company.com "Load testing approved"
```

### 批量审批
```bash
# 创建批量审批脚本
cat > batch_approve.sh << 'EOF'
#!/bin/bash
ALERTSCALES=("app1-scale" "app2-scale" "app3-scale")
for alertscale in "${ALERTSCALES[@]}"; do
    ./hack/quick_approval.sh approve default "$alertscale" admin@company.com "Batch approval"
    sleep 1
done
EOF

chmod +x batch_approve.sh
./batch_approve.sh
```

## 集成到CI/CD

### GitHub Actions示例
```yaml
- name: Approve AlertScale
  run: |
    ./hack/quick_approval.sh approve default my-alertscale ci@company.com "Automated approval"
  env:
    API_BASE: https://ops-api.company.com/api/v1
```

### Jenkins示例
```groovy
stage('Approve Scale') {
    steps {
        script {
            sh './hack/quick_approval.sh approve default my-alertscale jenkins@company.com "Pipeline approval"'
        }
    }
    environment {
        API_BASE = 'https://ops-api.company.com/api/v1'
    }
}
```
