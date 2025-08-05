# 部署指南 - udesk-ops-operator 通用API服务

本指南详细说明如何部署带有通用API服务的 udesk-ops-operator。

## 🏗️ 构建和部署

### 1. 构建镜像

```bash
# 构建 operator 镜像
make docker-build IMG=your-registry/udesk-ops-operator:latest

# 推送镜像到仓库
make docker-push IMG=your-registry/udesk-ops-operator:latest
```

### 2. 部署到 Kubernetes

```bash
# 安装 CRD
make install

# 部署 operator 到集群
make deploy IMG=your-registry/udesk-ops-operator:latest
```

### 3. 启用 API 服务器

修改部署配置以启用 API 服务：

```yaml
# config/manager/manager.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
  namespace: system
spec:
  template:
    spec:
      containers:
      - name: manager
        image: your-registry/udesk-ops-operator:latest
        args:
        - --leader-elect
        - --enable-api-server=true
        - --api-addr=:8088
        env:
        - name: ENABLE_API_SERVER
          value: "true"
        - name: API_ADDR  
          value: ":8088"
        ports:
        - containerPort: 8088
          name: api
          protocol: TCP
        - containerPort: 9443
          name: webhook-server
          protocol: TCP
        - containerPort: 8080
          name: metrics
          protocol: TCP
```

## 🌐 网络配置

### Service 配置

创建 Service 以暴露 API 端点：

```yaml
# config/default/api_service.yaml
apiVersion: v1
kind: Service
metadata:
  name: udesk-ops-operator-api
  namespace: udesk-ops-operator-system
spec:
  selector:
    control-plane: controller-manager
  ports:
  - name: api
    port: 8088
    targetPort: 8088
    protocol: TCP
  type: ClusterIP
```

### Ingress 配置

```yaml
# config/default/api_ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: udesk-ops-api-ingress
  namespace: udesk-ops-operator-system
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
    nginx.ingress.kubernetes.io/cors-allow-origin: "*"
    nginx.ingress.kubernetes.io/cors-allow-methods: "GET, POST, PUT, DELETE, OPTIONS"
    nginx.ingress.kubernetes.io/cors-allow-headers: "Content-Type, Authorization"
spec:
  rules:
  - host: udesk-ops-api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: udesk-ops-operator-api
            port:
              number: 8088
  tls:
  - hosts:
    - udesk-ops-api.example.com
    secretName: udesk-ops-api-tls
```

### NetworkPolicy 配置

```yaml
# config/network-policy/allow-api-traffic.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-api-traffic
  namespace: udesk-ops-operator-system
spec:
  podSelector:
    matchLabels:
      control-plane: controller-manager
  policyTypes:
  - Ingress
  ingress:
  - from: []
    ports:
    - protocol: TCP
      port: 8088
```

## 🔐 安全配置

### RBAC 权限

API 服务需要额外的 RBAC 权限：

```yaml
# config/rbac/api_role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: api-server-role
rules:
- apiGroups: ["ops.udesk.cn"]
  resources: ["alertscales"]
  verbs: ["get", "list", "create", "update", "patch", "watch"]
- apiGroups: ["ops.udesk.cn"]
  resources: ["alertscales/status"]
  verbs: ["get", "update", "patch"]
- apiGroups: [""]
  resources: ["events"]
  verbs: ["create", "patch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: api-server-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: api-server-role
subjects:
- kind: ServiceAccount
  name: controller-manager
  namespace: udesk-ops-operator-system
```

### TLS 证书配置

为 API 端点配置 TLS：

```yaml
# config/certmanager/api-certificate.yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: api-server-cert
  namespace: udesk-ops-operator-system
spec:
  secretName: udesk-ops-api-tls
  issuerRef:
    kind: ClusterIssuer
    name: letsencrypt-prod
  dnsNames:
  - udesk-ops-api.example.com
```

## 📊 监控配置

### ServiceMonitor 配置

```yaml
# config/prometheus/api_monitor.yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: udesk-ops-api-monitor
  namespace: udesk-ops-operator-system
spec:
  selector:
    matchLabels:
      app: udesk-ops-operator-api
  endpoints:
  - port: api
    path: /api/v1/health
    interval: 30s
    scrapeTimeout: 10s
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "id": null,
    "title": "udesk-ops-operator API Server",
    "tags": ["kubernetes", "operator", "api"],
    "panels": [
      {
        "title": "API Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"udesk-ops-api\"}[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "API Response Time",
        "type": "graph", 
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"udesk-ops-api\"}[5m]))",
            "legendFormat": "95th percentile"
          },
          {
            "expr": "histogram_quantile(0.50, rate(http_request_duration_seconds_bucket{job=\"udesk-ops-api\"}[5m]))",
            "legendFormat": "50th percentile"
          }
        ]
      },
      {
        "title": "Approval Statistics",
        "type": "stat",
        "targets": [
          {
            "expr": "alertscale_pending_count",
            "legendFormat": "Pending"
          },
          {
            "expr": "alertscale_approved_count", 
            "legendFormat": "Approved"
          },
          {
            "expr": "alertscale_rejected_count",
            "legendFormat": "Rejected"
          }
        ]
      }
    ]
  }
}
```

## 🔄 Health Checks

### Kubernetes Probes

```yaml
# config/manager/manager.yaml 中添加健康检查
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  template:
    spec:
      containers:
      - name: manager
        livenessProbe:
          httpGet:
            path: /api/v1/health
            port: 8088
          initialDelaySeconds: 30
          periodSeconds: 10
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /api/v1/health
            port: 8088
          initialDelaySeconds: 5
          periodSeconds: 5
          timeoutSeconds: 3
          failureThreshold: 3
```

## 📝 配置文件更新

更新 kustomization 文件以包含新的资源：

```yaml
# config/default/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
- ../crd
- ../rbac
- ../manager
- ../webhook
- ../certmanager
- ../prometheus
- ../network-policy
- api_service.yaml
- api_ingress.yaml

patches:
- path: manager_api_patch.yaml
  target:
    kind: Deployment
    name: controller-manager
```

```yaml
# config/default/manager_api_patch.yaml
- op: add
  path: /spec/template/spec/containers/0/args/-
  value: --enable-api-server=true
- op: add
  path: /spec/template/spec/containers/0/args/-
  value: --api-addr=:8088
- op: add
  path: /spec/template/spec/containers/0/ports/-
  value:
    containerPort: 8088
    name: api
    protocol: TCP
```

## 🚀 部署验证

### 1. 检查部署状态

```bash
# 检查 operator 状态
kubectl get pods -n udesk-ops-operator-system

# 检查 service 状态
kubectl get svc -n udesk-ops-operator-system

# 检查 ingress 状态
kubectl get ingress -n udesk-ops-operator-system
```

### 2. 测试 API 连接

```bash
# 内部测试（通过 Service）
kubectl port-forward -n udesk-ops-operator-system svc/udesk-ops-operator-api 8088:8088

# 测试健康检查
curl http://localhost:8088/api/v1/health

# 外部测试（通过 Ingress）
curl https://udesk-ops-api.example.com/api/v1/health
```

### 3. 功能验证

```bash
# 创建测试 AlertScale
kubectl apply -f - <<EOF
apiVersion: ops.udesk.cn/v1beta1
kind: AlertScale
metadata:
  name: test-alertscale
  namespace: default
spec:
  scaleReason: "Test scale operation"
  scaleDuration: "10m"
  template: "test-template"
  autoApproval: false
EOF

# 通过 API 查看
curl http://localhost:8088/api/v1/alertscales

# 通过 API 审批
curl -X POST http://localhost:8088/api/v1/alertscales/default/test-alertscale/approve \
  -H "Content-Type: application/json" \
  -d '{
    "approver": "test@company.com",
    "reason": "Test approval",
    "comment": "Testing API functionality"
  }'
```

## 🔧 故障排除

### 常见问题

1. **API 服务器无法启动**
   ```bash
   # 检查日志
   kubectl logs -n udesk-ops-operator-system deployment/controller-manager
   
   # 检查端口绑定
   kubectl describe pod -n udesk-ops-operator-system -l control-plane=controller-manager
   ```

2. **API 请求超时**
   ```bash
   # 检查网络策略
   kubectl get networkpolicy -n udesk-ops-operator-system
   
   # 检查服务端点
   kubectl get endpoints -n udesk-ops-operator-system
   ```

3. **权限错误**
   ```bash
   # 检查 RBAC 权限
   kubectl auth can-i get alertscales --as=system:serviceaccount:udesk-ops-operator-system:controller-manager
   ```

### 日志收集

```bash
# 收集 operator 日志
kubectl logs -n udesk-ops-operator-system deployment/controller-manager > operator.log

# 收集事件
kubectl get events -n udesk-ops-operator-system --sort-by='.lastTimestamp'

# 收集资源状态
kubectl get all -n udesk-ops-operator-system -o yaml > cluster-state.yaml
```

## 📈 生产环境考虑

### 高可用配置

```yaml
# config/manager/manager.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: controller-manager
spec:
  replicas: 3
  template:
    spec:
      affinity:
        podAntiAffinity:
          requiredDuringSchedulingIgnoredDuringExecution:
          - labelSelector:
              matchLabels:
                control-plane: controller-manager
            topologyKey: kubernetes.io/hostname
```

### 资源限制

```yaml
resources:
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 200m
    memory: 256Mi
```

### 日志配置

```yaml
env:
- name: LOG_LEVEL
  value: "info"
- name: LOG_FORMAT
  value: "json"
```

通过以上配置，您的 udesk-ops-operator 将具备完整的通用 API 服务能力，支持扩容审批工作流！🎉
