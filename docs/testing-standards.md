# 测试编写规范文档

## 概述

本文档制定了 udesk-ops-operator 项目的测试编写标准和最佳实践，确保所有包使用统一的 Ginkgo BDD 测试框架，提供一致的测试结构和清晰的测试输出。

## 测试框架

### 统一测试框架：Ginkgo v2

所有包必须使用 [Ginkgo v2](https://github.com/onsi/ginkgo) 作为测试框架，配合 [Gomega](https://github.com/onsi/gomega) 作为断言库。

### 核心依赖

```go
import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)
```

## 文件结构规范

### 1. 测试套件文件 (`*_suite_test.go`)

每个包必须包含一个测试套件文件，负责初始化测试环境：

```go
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package packagename

import (
    "testing"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
)

func TestPackageName(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Package Name Suite")
}
```

### 2. 测试实现文件 (`*_test.go`)

具体的测试实现文件，包含实际的测试用例：

```go
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package packagename

import (
    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    // 其他必要的导入
)

var _ = Describe("功能模块名称", func() {
    // 测试实现
})
```

## BDD 测试结构

### 基本结构层次

使用 Ginkgo 的 BDD 语法组织测试：

```go
var _ = Describe("顶层功能描述", func() {
    var (
        // 声明测试变量
        testObject *SomeType
        mockClient client.Client
    )

    BeforeEach(func() {
        // 每个测试前的初始化
        testObject = &SomeType{}
        mockClient = fake.NewClientBuilder().Build()
    })

    Context("当满足特定条件时", func() {
        BeforeEach(func() {
            // 特定上下文的初始化
        })

        It("应该执行期望的行为", func() {
            // 具体测试实现
            result := testObject.SomeMethod()
            Expect(result).To(Equal(expectedValue))
        })

        It("应该处理错误情况", func() {
            // 错误场景测试
            err := testObject.ErrorMethod()
            Expect(err).To(HaveOccurred())
        })
    })

    Context("当满足其他条件时", func() {
        // 其他测试场景
    })
})
```

### 层次结构说明

1. **Describe**: 描述被测试的功能模块或组件
2. **Context**: 描述特定的测试环境或条件
3. **It**: 描述具体的行为期望
4. **BeforeEach**: 在每个测试前执行的初始化代码
5. **AfterEach**: 在每个测试后执行的清理代码

## 命名规范

### 描述性命名

- **Describe**: 使用被测试的类型名或功能模块名
- **Context**: 使用 "when/当" 开头，描述测试条件
- **It**: 使用 "should/应该" 开头，描述期望行为

### 示例

```go
var _ = Describe("AlertScale CRD", func() {
    Context("when creating a new AlertScale", func() {
        It("should have correct TypeMeta", func() {
            // 测试实现
        })
        
        It("should validate required fields", func() {
            // 测试实现
        })
    })
    
    Context("when updating AlertScale status", func() {
        It("should update status successfully", func() {
            // 测试实现
        })
    })
})
```

## Mock 和测试工具

### 使用 Fake Client

对于 Kubernetes 资源测试，使用 controller-runtime 的 fake client：

```go
import (
    "k8s.io/apimachinery/pkg/runtime"
    "sigs.k8s.io/controller-runtime/pkg/client/fake"
    opsv1beta1 "udesk.cn/ops/api/v1beta1"
)

// 在 BeforeEach 中设置
scheme := runtime.NewScheme()
_ = opsv1beta1.AddToScheme(scheme)

fakeClient := fake.NewClientBuilder().
    WithScheme(scheme).
    WithObjects(/* 预创建的对象 */).
    Build()
```

### Mock 对象命名

Mock 对象应以 `Mock` 前缀命名：

```go
type MockScaleStrategy struct{}

func (m *MockScaleStrategy) Scale(ctx context.Context, c client.Client, target *opsv1beta1.ScaleTarget, replicas int32) error {
    return nil
}
```

## 断言规范

### Gomega 断言风格

使用 Gomega 的流畅断言语法：

```go
// 正确示例
Expect(result).To(Equal(expected))
Expect(err).NotTo(HaveOccurred())
Expect(value).To(BeNil())
Expect(slice).To(HaveLen(3))
Expect(condition).To(BeTrue())

// 错误场景断言
Expect(err).To(HaveOccurred())
Expect(err.Error()).To(ContainSubstring("expected error message"))
```

### 常用断言

- `Equal(expected)`: 值相等
- `BeNil()`: 值为 nil
- `BeTrue()/BeFalse()`: 布尔值断言
- `HaveOccurred()`: 错误发生断言
- `HaveLen(n)`: 切片/数组长度断言
- `ContainSubstring(s)`: 字符串包含断言

## 异步测试

### 使用 Eventually 和 Consistently

对于异步操作，使用 Gomega 的异步断言：

```go
// 等待条件满足
Eventually(func() bool {
    err := client.Get(ctx, key, object)
    return err == nil
}, timeout, interval).Should(BeTrue())

// 确保条件持续满足
Consistently(func() int {
    return len(objects)
}, duration, interval).Should(Equal(expectedCount))
```

### 超时设置

```go
It("should complete within timeout", func(done Done) {
    go func() {
        defer GinkgoRecover()
        // 异步操作
        close(done)
    }()
}, 2) // 2秒超时
```

## 测试组织

### 按功能分组

每个功能模块应有对应的测试文件：

```
internal/
├── controller/
│   ├── alertscale_controller.go
│   ├── alertscale_controller_test.go
│   └── controller_suite_test.go
├── handler/
│   ├── scale_state_handler.go
│   ├── scale_state_handler_test.go
│   └── handler_suite_test.go
└── strategy/
    ├── scale_strategy.go
    ├── scale_strategy_test.go
    └── strategy_suite_test.go
```

### 测试覆盖范围

每个包应包含以下测试类型：

1. **单元测试**: 测试单个函数或方法
2. **集成测试**: 测试组件间交互
3. **错误场景测试**: 测试异常情况处理
4. **边界条件测试**: 测试极端输入情况

## 运行和输出

### 运行所有测试

```bash
make test
```

### 运行特定包测试

```bash
go test ./internal/controller/... -v
```

### Ginkgo 详细输出

```bash
ginkgo -v ./...
```

### 期望的测试输出格式

Ginkgo 提供结构化的测试输出：

```
Running Suite: Controller Suite
===============================
Random Seed: 1234567890

Will run 10 of 10 specs

AlertScale Controller
  when reconciling AlertScale
    should create AlertScale successfully ✓
    should update AlertScale status ✓
    should handle deletion ✓
  when handling errors
    should retry on transient errors ✓
    should fail on permanent errors ✓

Ran 5 of 5 Specs in 2.345 seconds
SUCCESS! -- 5 Passed | 0 Failed | 0 Pending | 0 Skipped
```

## 最佳实践

### 1. 测试隔离

每个测试应该是独立的，不依赖其他测试的状态：

```go
BeforeEach(func() {
    // 为每个测试创建新的对象
    testObject = NewTestObject()
})
```

### 2. 清晰的测试意图

测试名称应该清楚表达测试意图：

```go
It("should return error when deployment not found", func() {
    // 明确的错误场景测试
})
```

### 3. 适度的 Mock

只对必要的依赖进行 Mock，避免过度 Mock：

```go
// 好的做法：只 Mock 外部依赖
mockClient := fake.NewClientBuilder().Build()

// 避免：过度 Mock 内部逻辑
```

### 4. 测试数据管理

使用工厂函数创建测试数据：

```go
func createTestAlertScale(name, namespace string) *opsv1beta1.AlertScale {
    return &opsv1beta1.AlertScale{
        ObjectMeta: metav1.ObjectMeta{
            Name:      name,
            Namespace: namespace,
        },
        // 其他默认字段
    }
}
```

### 5. 错误测试

确保测试错误场景：

```go
Context("when handling errors", func() {
    It("should return appropriate error for invalid input", func() {
        err := handler.Process(invalidInput)
        Expect(err).To(HaveOccurred())
        Expect(err.Error()).To(ContainSubstring("invalid input"))
    })
})
```

## 持续集成

### 测试在 CI 中的执行

确保 CI 流水线包含测试步骤：

```yaml
# .github/workflows/test.yml
- name: Run Tests
  run: make test
```

### 测试覆盖率

定期检查测试覆盖率：

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 结论

遵循这些测试编写规范，可以确保：

1. **一致性**: 所有包使用统一的测试框架和结构
2. **可读性**: 清晰的 BDD 语法使测试易于理解
3. **可维护性**: 结构化的测试代码便于维护
4. **可靠性**: 全面的测试覆盖确保代码质量

定期回顾和更新这些规范，以适应项目的发展需求。
