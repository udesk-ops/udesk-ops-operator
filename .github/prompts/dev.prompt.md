# udesk-ops-operator Development Prompt

You are an AI assistant helping with the development of **udesk-ops-operator**, a production-grade Kubernetes Operator that provides intelligent scaling management with approval workflows and multi-channel notification systems.

## Project Overview

**udesk-ops-operator** is a Go-based Kubernetes Operator built with:
- **Framework**: Kubebuilder/controller-runtime
- **Language**: Go 1.24+
- **Testing**: Ginkgo v2 BDD framework exclusively
- **API**: REST API server for external integrations
- **CRDs**: AlertScale, ScaleNotifyConfig, ScaleNotifyMsgTemplate

## 🚨 MANDATORY REQUIREMENTS

### Code Quality Gates (NON-NEGOTIABLE)
Every code change MUST complete ALL of these steps before being considered done:

1. **Add Test Cases**: Write comprehensive Ginkgo tests for new functionality
2. **Execute Tests**: Run `make test` and ensure 100% pass rate
3. **Execute Linting**: Run `make lint` and achieve 0 issues
4. **NO EXCEPTIONS**: All three steps must succeed with zero errors/failures

### Documentation Standards
- **ALL documentation** files go in the `docs/` directory
- **Project overview** is maintained in root `README.md`
- **API docs** → `docs/api/`
- **Development guides** → `docs/development/`
- **Deployment guides** → `docs/deployment/`
- **Architecture docs** → `docs/architecture/`

### Testing Framework Requirements
- **ONLY Ginkgo v2 BDD**: No standard Go testing allowed
- **Test Structure**: Each package requires `*_suite_test.go` + `*_test.go`
- **Coverage**: Meaningful test coverage for all new functionality
- **BDD Style**: Use descriptive, behavior-driven test specifications

## 📁 Project Architecture

```
udesk-ops-operator/
├── api/v1beta1/              # Custom Resource Definitions
│   ├── alertscale_types.go
│   ├── scalenotifyconfig_types.go
│   └── scalenotifymsgtemplate_types.go
├── cmd/
│   └── main.go               # Application entry point
├── config/                   # Kubernetes manifests
│   ├── crd/                  # CRD definitions
│   ├── rbac/                 # RBAC configurations
│   ├── manager/              # Deployment configs
│   └── samples/              # Example resources
├── docs/                     # ALL documentation (REQUIRED)
│   ├── api/                  # API documentation
│   ├── architecture/         # Architecture diagrams
│   ├── deployment/           # Installation guides
│   └── development/          # Development guides
├── internal/
│   ├── controller/           # Kubernetes controllers
│   │   ├── alertscale_controller.go
│   │   ├── scalenotifyconfig_controller.go
│   │   └── scalenotifymsgtemplate_controller.go
│   ├── handler/              # State machine handlers
│   │   └── scale_state_handler.go
│   ├── server/               # REST API server
│   │   ├── server.go
│   │   └── handlers/         # HTTP endpoint handlers
│   ├── strategy/             # Scaling strategies
│   │   ├── scale_strategy.go
│   │   └── notify_strategy.go
│   ├── types/                # Internal type definitions
│   │   └── scale_types.go
│   └── webhook/              # Admission webhooks
│       └── v1beta1/
├── test/                     # Integration and E2E tests
│   ├── e2e/                  # End-to-end tests
│   └── utils/                # Test utilities
└── README.md                 # Project overview
```

## 🏗️ Development Guidelines

### Kubernetes Operator Best Practices

#### Controller Development
- **Reconciliation**: Implement idempotent reconcile loops
- **Error Handling**: Use controller-runtime error patterns
- **Status Updates**: Always use separate status client
- **Finalizers**: Implement proper cleanup mechanisms
- **Events**: Record meaningful Kubernetes events
- **Metrics**: Add Prometheus metrics for observability

#### CRD Design Principles
- **Validation**: Use OpenAPI schema validation
- **Status Subresource**: Always implement status reporting
- **Conditions**: Use standard Kubernetes condition types
- **Defaults**: Set sensible default values
- **Immutability**: Mark appropriate fields as immutable

#### State Management
- **State Machine**: Use clear state transitions
- **Status Reporting**: Provide detailed status information
- **Error Recovery**: Implement robust error recovery
- **Timeout Handling**: Set appropriate timeouts

### API Server Best Practices

#### REST API Design
- **Versioning**: Use `/api/v1/` prefix
- **HTTP Methods**: Follow RESTful conventions
- **Status Codes**: Use appropriate HTTP status codes
- **Error Responses**: Return consistent error formats
- **Authentication**: Implement proper auth for production

#### Handler Implementation
- **Input Validation**: Validate all input parameters
- **Context Propagation**: Pass context for cancellation
- **Logging**: Log all operations with structured logging
- **Error Handling**: Return meaningful error messages
- **Rate Limiting**: Consider rate limiting for production

### Code Quality Standards

#### Go Language Standards
```go
// Package imports organization
import (
    // Standard library
    "context"
    "fmt"
    "time"

    // Third-party packages
    "github.com/go-logr/logr"
    "k8s.io/api/apps/v1"
    "sigs.k8s.io/controller-runtime"

    // Internal packages
    "udesk.cn/ops/api/v1beta1"
    "udesk.cn/ops/internal/types"
)
```

#### Error Handling Patterns
```go
// Controller error handling
if err != nil {
    log.Error(err, "Failed to perform operation", "resource", req.NamespacedName)
    return ctrl.Result{RequeueAfter: time.Minute * 5}, err
}

// API handler error handling
if err != nil {
    http.Error(w, fmt.Sprintf("Internal server error: %v", err), http.StatusInternalServerError)
    return
}
```

#### Logging Standards
```go
// Use structured logging
log := ctrl.Log.WithName("alertscale-controller")
log.Info("Reconciling AlertScale", 
    "namespace", alertScale.Namespace,
    "name", alertScale.Name,
    "generation", alertScale.Generation)
```

## 🧪 Testing Requirements

### Ginkgo Test Structure
```go
var _ = Describe("AlertScale Controller", func() {
    var (
        ctx        context.Context
        cancel     context.CancelFunc
        k8sClient  client.Client
        testEnv    *envtest.Environment
    )

    BeforeEach(func() {
        ctx, cancel = context.WithCancel(context.Background())
        // Setup test environment
    })

    AfterEach(func() {
        cancel()
        // Cleanup
    })

    Describe("Reconciling AlertScale", func() {
        Context("when AlertScale is created", func() {
            It("should update status to pending", func() {
                // Test implementation
                Expect(alertScale.Status.State).To(Equal("pending"))
            })

            It("should send notification", func() {
                // Test notification behavior
                Eventually(func() bool {
                    // Check notification was sent
                    return true
                }).Should(BeTrue())
            })
        })

        Context("when auto-approval is enabled", func() {
            BeforeEach(func() {
                alertScale.Spec.ScaleAutoApproval = true
            })

            It("should transition to approved state", func() {
                // Test auto-approval logic
            })
        })
    })
})
```

### Test Categories
1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test controller reconciliation with fake clients
3. **Webhook Tests**: Test admission webhook validation logic
4. **API Tests**: Test REST API endpoints
5. **E2E Tests**: Test complete workflows in real cluster

### Test Coverage Requirements
- **New Features**: 100% test coverage required
- **Bug Fixes**: Add regression tests
- **Edge Cases**: Test error conditions and timeouts
- **Mock Objects**: Use fake clients for unit tests
- **Real Cluster**: Use envtest for integration tests

## 🔄 Development Workflow

### Adding New Custom Resource
1. **Define Types**: Add to `api/v1beta1/`
2. **Generate Code**: Run `make generate manifests`
3. **Add Controller**: Implement in `internal/controller/`
4. **Add Validation**: Implement admission webhook
5. **Write Tests**: Create comprehensive Ginkgo tests
6. **Update RBAC**: Add necessary permissions
7. **Document**: Add to `docs/api/`
8. **Verify**: Run `make test lint`

### Adding API Endpoint
1. **Define Handler**: Create in `internal/server/handlers/`
2. **Register Route**: Add to server router
3. **Add Validation**: Validate input parameters
4. **Error Handling**: Implement proper error responses
5. **Write Tests**: Create Ginkgo tests
6. **Document**: Add to `docs/api/`
7. **Verify**: Run `make test lint`

### Adding State Handler
1. **Implement Handler**: Add to `internal/handler/`
2. **Update State Machine**: Register new state
3. **Add Transitions**: Define valid state transitions
4. **Write Tests**: Test all state transitions
5. **Document**: Update state machine docs
6. **Verify**: Run `make test lint`

## 🛠️ Development Commands

### Daily Development Workflow
```bash
# Complete development cycle (MANDATORY for every change)
make manifests generate fmt vet test lint

# Individual commands
make manifests          # Generate CRD manifests
make generate          # Generate deepcopy methods
make fmt               # Format Go code
make vet               # Run go vet
make test              # Run all tests (REQUIRED)
make lint              # Run golangci-lint (REQUIRED)
```

### Local Development
```bash
# Run operator locally (without webhooks)
make run

# Run with webhooks (requires cert-manager)
make run-webhooks

# Install CRDs to cluster
make install

# Deploy complete operator
make deploy

# Undeploy
make undeploy
```

### Testing Commands
```bash
# Run all tests
make test

# Run specific package tests
go test ./internal/controller/... -v

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out
```

## 📋 Quality Assurance Checklist

Before ANY pull request or commit, verify:

### Testing Requirements
- [ ] New functionality has comprehensive Ginkgo tests
- [ ] All existing tests continue to pass
- [ ] Edge cases and error conditions are tested
- [ ] `make test` returns 0 exit code

### Code Quality Requirements
- [ ] `make lint` returns 0 issues
- [ ] Code follows Go conventions
- [ ] Error handling is implemented
- [ ] Logging is added for observability
- [ ] Context cancellation is handled

### Documentation Requirements
- [ ] New features documented in `docs/`
- [ ] API changes documented
- [ ] Configuration options explained
- [ ] Examples provided

### Kubernetes Requirements
- [ ] RBAC permissions updated if needed
- [ ] CRD validation schemas updated
- [ ] Status reporting implemented
- [ ] Finalizers added for cleanup

## 🎯 Common Implementation Patterns

### Controller Reconcile Pattern
```go
func (r *AlertScaleReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
    log := r.Log.WithValues("alertscale", req.NamespacedName)
    
    // Fetch the AlertScale instance
    alertScale := &opsv1beta1.AlertScale{}
    if err := r.Get(ctx, req.NamespacedName, alertScale); err != nil {
        if errors.IsNotFound(err) {
            return ctrl.Result{}, nil
        }
        return ctrl.Result{}, err
    }
    
    // Handle deletion
    if alertScale.DeletionTimestamp != nil {
        return r.handleDeletion(ctx, alertScale)
    }
    
    // Add finalizer if needed
    if !controllerutil.ContainsFinalizer(alertScale, AlertScaleFinalizer) {
        controllerutil.AddFinalizer(alertScale, AlertScaleFinalizer)
        return ctrl.Result{}, r.Update(ctx, alertScale)
    }
    
    // Main reconciliation logic
    return r.reconcileNormal(ctx, alertScale)
}
```

### API Handler Pattern
```go
func (h *AlertScaleHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    
    // Parse request
    var req CreateAlertScaleRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }
    
    // Validate request
    if err := h.validateCreateRequest(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Create resource
    alertScale, err := h.createAlertScale(ctx, &req)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // Return response
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(alertScale)
}
```

### State Handler Pattern
```go
func (h *PendingStateHandler) Handle(ctx *types.ScaleContext) (ctrl.Result, error) {
    log := ctrl.LoggerFrom(ctx.Context).WithName("pending-handler")
    
    // Validate state transition
    if !h.CanTransition(ctx.AlertScale.Status.State) {
        return ctrl.Result{}, fmt.Errorf("invalid state transition")
    }
    
    // Execute state logic
    if err := h.executeStateBehavior(ctx); err != nil {
        return ctrl.Result{}, err
    }
    
    // Update status
    return h.updateStatus(ctx, "processing")
}
```

## 🚀 Performance and Scalability

### Resource Management
- Use resource limits and requests
- Implement graceful shutdown
- Handle high-volume events efficiently
- Use work queues for async processing

### Monitoring and Observability
- Add Prometheus metrics
- Implement health checks
- Use structured logging
- Add distributed tracing

### Security Considerations
- Follow principle of least privilege
- Validate all inputs
- Secure API endpoints
- Use service accounts properly

---

## 🎯 Success Criteria

A feature is considered complete ONLY when:
1. ✅ Comprehensive Ginkgo tests are written and passing
2. ✅ `make test` returns zero failures
3. ✅ `make lint` returns zero issues
4. ✅ Documentation is updated in `docs/`
5. ✅ Code follows all established patterns
6. ✅ RBAC permissions are properly configured

**Remember**: This is a production Kubernetes operator. Quality, reliability, and maintainability are paramount. Never compromise on testing or code quality for speed of delivery.
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
