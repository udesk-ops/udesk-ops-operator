# udesk-ops-operator Development Prompt

You are an AI assistant helping with the development of a Kubernetes Operator project called **udesk-ops-operator**. This operator provides automated scaling and approval workflows for Kubernetes workloads.

## Project Overview

This is a Go-based Kubernetes Operator built with:
- **Framework**: Kubebuilder/controller-runtime
- **Language**: Go 1.21+
- **Testing**: Ginkgo v2 BDD framework
- **API**: REST API server for external integrations
- **CRDs**: AlertScale, ScaleNotifyConfig, ScaleNotifyMsgTemplate

## Mandatory Requirements

### Code Quality Gates
Before any code changes are considered complete, you MUST:

1. **Execute Tests**: Run `make test` and ensure all tests pass
2. **Execute Linting**: Run `make lint` and fix all issues
3. **No Exceptions**: Both commands must succeed with zero errors

### Documentation Standards
- **All documentation** goes in the `docs/` directory
- **Project overview** is maintained in `README.md` (root level)
- **API documentation** should be in `docs/api/`
- **Development guides** should be in `docs/development/`
- **Deployment guides** should be in `docs/deployment/`

### Testing Framework
- **Framework**: Ginkgo v2 BDD testing only
- **No standard Go testing**: All tests must use Ginkgo
- **Structure**: Each package has `*_suite_test.go` and `*_test.go` files
- **Coverage**: Maintain meaningful test coverage for all packages

## Project Structure

```
udesk-ops-operator/
├── api/v1beta1/              # CRD definitions
├── cmd/                      # Main application entry point
├── config/                   # Kubernetes manifests and configuration
├── docs/                     # ALL documentation files
├── hack/                     # Build scripts and utilities
├── internal/
│   ├── controller/           # Kubernetes controllers
│   ├── handler/              # State handlers for AlertScale
│   ├── server/               # REST API server
│   │   └── handlers/         # API endpoint handlers
│   ├── strategy/             # Scaling strategies
│   ├── types/                # Internal type definitions
│   └── webhook/              # Admission webhooks
├── test/                     # Integration and e2e tests
└── README.md                 # Project overview
```

## Development Guidelines

### Code Standards
- **Go Modules**: Use Go 1.21+ with proper module management
- **Imports**: Group imports (std, 3rd party, internal)
- **Error Handling**: Always handle errors appropriately
- **Logging**: Use controller-runtime logging (logr)
- **Context**: Pass context.Context for cancellation
- **Resource Management**: Properly handle Kubernetes client resources

### Kubernetes Operator Best Practices
- **Reconciliation**: Implement idempotent reconcile loops
- **Status Updates**: Use separate status client for status updates
- **Finalizers**: Implement proper cleanup with finalizers
- **RBAC**: Define minimal required permissions
- **Validation**: Use admission webhooks for validation
- **Observability**: Add metrics and health checks

### API Design Principles
- **REST**: Follow RESTful conventions
- **Versioning**: Use `/api/v1/` prefix
- **Error Handling**: Return consistent error responses
- **Authentication**: Consider authentication for production
- **Documentation**: Maintain OpenAPI/Swagger docs

### Testing Approach
- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test controller reconciliation
- **E2E Tests**: Test complete workflows
- **BDD Style**: Use descriptive Ginkgo specs
- **Mocking**: Use fake clients for unit tests

## Common Tasks

### Adding New CRD
1. Define types in `api/v1beta1/`
2. Add controller in `internal/controller/`
3. Generate manifests: `make manifests`
4. Create Ginkgo tests
5. Update RBAC permissions
6. Document in `docs/api/`

### Adding API Endpoint
1. Create handler in `internal/server/handlers/`
2. Register in handler's `init()` function
3. Add validation and error handling
4. Create Ginkgo tests
5. Document endpoint in `docs/api/`

### Adding State Handler
1. Implement handler in `internal/handler/`
2. Add to state machine registration
3. Create comprehensive Ginkgo tests
4. Document state transitions

## Code Examples

### Ginkgo Test Structure
```go
var _ = Describe("ComponentName", func() {
    var (
        ctx        context.Context
        fakeClient client.Client
    )

    BeforeEach(func() {
        ctx = context.Background()
        fakeClient = fake.NewClientBuilder().Build()
    })

    Describe("MethodName", func() {
        Context("when condition is met", func() {
            It("should perform expected behavior", func() {
                // Test implementation
                Expect(result).To(Succeed())
            })
        })
    })
})
```

### Error Handling Pattern
```go
if err != nil {
    log.Error(err, "Failed to perform operation")
    return ctrl.Result{}, err
}
```

### Status Update Pattern
```go
// Update status separately from spec
alertScale.Status.ScaleStatus.Status = newStatus
if err := r.Status().Update(ctx, &alertScale); err != nil {
    return ctrl.Result{}, err
}
```

## Build and Development Commands

```bash
# Development workflow
make manifests generate fmt vet test lint

# Run locally
make run-without-webhook

# Build binary
make build

# Docker operations
make docker-build docker-push

# Deploy to cluster
make deploy

# Run e2e tests
make test-e2e
```

## Quality Assurance Checklist

Before considering any feature complete:
- [ ] All new code has Ginkgo tests
- [ ] `make test` passes (100% success rate required)
- [ ] `make lint` passes (0 issues required)
- [ ] Documentation updated in `docs/`
- [ ] RBAC permissions updated if needed
- [ ] Error handling implemented
- [ ] Logging added for observability
- [ ] Status updates use proper patterns

## Documentation Requirements

### API Documentation
- Endpoint descriptions
- Request/response schemas
- Error codes and meanings
- Authentication requirements
- Usage examples

### Developer Documentation
- Architecture overview
- State machine diagrams
- Database/storage patterns
- Configuration options
- Troubleshooting guides

### Deployment Documentation
- Installation instructions
- Configuration examples
- Monitoring setup
- Backup/restore procedures
- Upgrade procedures

## Common Patterns to Follow

### Controller Pattern
- Implement `Reconcile(ctx context.Context, req ctrl.Request)`
- Handle not found errors gracefully
- Use exponential backoff for retries
- Implement proper status reporting

### API Handler Pattern
- Validate input parameters
- Use proper HTTP status codes
- Return consistent JSON responses
- Log all operations
- Handle context cancellation

### State Handler Pattern
- Implement `Handle(ctx *types.ScaleContext) (ctrl.Result, error)`
- Use declarative state transitions
- Avoid direct status field conflicts
- Implement timeout handling

Remember: This is a production Kubernetes operator. All code must be robust, well-tested, and follow cloud-native best practices. Always prioritize reliability and observability over feature velocity.
