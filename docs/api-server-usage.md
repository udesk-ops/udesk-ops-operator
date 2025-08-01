# API Server Usage Guide

The udesk-ops-operator now includes a REST API server that provides external access to AlertScale operations with approval workflows.

## Configuration

The API server can be enabled via command-line flags:

```bash
./bin/manager \
  --enable-api-server=true \
  --api-addr=:8088
```

Or via environment variables:
- `ENABLE_API_SERVER=true`
- `API_ADDR=:8088`

## API Endpoints

### Health Check
```http
GET /api/v1/health
```

Response:
```json
{
  "success": true,
  "message": "API server is healthy",
  "data": {
    "status": "healthy",
    "timestamp": "2023-10-01T12:00:00Z",
    "version": "v1.0.0",
    "server": "udesk-ops-operator-api-server"
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

### List AlertScales
```http
GET /api/v1/alertscales
```

Response:
```json
{
  "success": true,
  "message": "AlertScales retrieved successfully",
  "data": {
    "items": [
      {
        "name": "my-alertscale",
        "namespace": "default",
        "reason": "CPU utilization high",
        "status": "Approvaling",
        "duration": "5m",
        "template": "default-template"
      }
    ],
    "count": 1
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

### Get Specific AlertScale
```http
GET /api/v1/alertscales/{namespace}/{name}
```

Response:
```json
{
  "success": true,
  "message": "AlertScale retrieved successfully",
  "data": {
    "name": "my-alertscale",
    "namespace": "default",
    "reason": "CPU utilization high",
    "status": "Approvaling",
    "duration": "5m",
    "template": "default-template",
    "autoApproval": false,
    "createdAt": "2023-10-01T11:00:00Z"
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

### Approve AlertScale
```http
POST /api/v1/alertscales/{namespace}/{name}/approve
Content-Type: application/json

{
  "approver": "admin@company.com",
  "reason": "Scale approved for high load",
  "comment": "Scaling approved during peak hours"
}
```

Response:
```json
{
  "success": true,
  "message": "AlertScale approved successfully",
  "data": {
    "namespace": "default",
    "name": "my-alertscale",
    "status": "Approved",
    "approver": "admin@company.com"
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

### Reject AlertScale
```http
POST /api/v1/alertscales/{namespace}/{name}/reject
Content-Type: application/json

{
  "approver": "admin@company.com",
  "reason": "Scale rejected due to policy",
  "comment": "Not approved during maintenance window"
}
```

Response:
```json
{
  "success": true,
  "message": "AlertScale rejected successfully",
  "data": {
    "namespace": "default",
    "name": "my-alertscale",
    "status": "Rejected",
    "rejector": "admin@company.com"
  },
  "timestamp": "2023-10-01T12:00:00Z"
}
```

## Security Considerations

The API server currently runs without authentication. In production environments, you should:

1. **Use HTTPS**: Deploy behind a reverse proxy with TLS termination
2. **Add Authentication**: Implement API key or OAuth2 authentication
3. **Network Security**: Restrict access via firewall rules or network policies
4. **Rate Limiting**: Implement rate limiting to prevent abuse

## Error Responses

All error responses follow this format:
```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error message",
  "timestamp": "2023-10-01T12:00:00Z"
}
```

Common HTTP status codes:
- `400 Bad Request`: Invalid parameters or request body
- `404 Not Found`: AlertScale resource not found
- `500 Internal Server Error`: Server-side errors

## Integration Examples

### cURL Examples

```bash
# Check health
curl -X GET http://localhost:8088/api/v1/health

# List all AlertScales
curl -X GET http://localhost:8088/api/v1/alertscales

# Get specific AlertScale
curl -X GET http://localhost:8088/api/v1/alertscales/default/my-alertscale

# Approve an AlertScale
curl -X POST http://localhost:8088/api/v1/alertscales/default/my-alertscale/approve \
  -H "Content-Type: application/json" \
  -d '{"approver": "admin@company.com", "reason": "Approved for scaling"}'

# Reject an AlertScale
curl -X POST http://localhost:8088/api/v1/alertscales/default/my-alertscale/reject \
  -H "Content-Type: application/json" \
  -d '{"approver": "admin@company.com", "reason": "Policy violation"}'
```

### Python Example

```python
import requests
import json

api_base = "http://localhost:8088/api/v1"

# List AlertScales
response = requests.get(f"{api_base}/alertscales")
alertscales = response.json()

# Approve an AlertScale
approval_data = {
    "approver": "admin@company.com",
    "reason": "Approved for scaling",
    "comment": "Load testing approved"
}

response = requests.post(
    f"{api_base}/alertscales/default/my-alertscale/approve",
    json=approval_data
)

print(response.json())
```

## Deployment Considerations

When integrating the API server into your deployment:

1. **Service Configuration**: Expose the API server port in your Kubernetes service
2. **Ingress Setup**: Configure ingress rules for external access
3. **Monitoring**: Add health check endpoints to your monitoring system
4. **Logging**: API requests are logged with method, path, and duration

Example Kubernetes service:
```yaml
apiVersion: v1
kind: Service
metadata:
  name: udesk-ops-operator-api
spec:
  selector:
    app: udesk-ops-operator
  ports:
  - name: api
    port: 8088
    targetPort: 8088
  type: ClusterIP
```
