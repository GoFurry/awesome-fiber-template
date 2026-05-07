# Go 代码审计报告

## Summary

本次审计聚焦 `fiberx` 当前已完成主线，重点查看生成后骨架代码的易用性、可维护性，以及会在真实项目中放大的安全性与正确性风险。

整体判断：

- 生成器主线功能已经形成闭环，项目生成、升级评估、构建辅助和 release-facing 文档都较完整。
- 生成骨架在“可跑起来”这一层已经达标，但在默认错误边界、超时包装一致性、缓存一致性、配置默认值语义和隐式全局状态方面，仍有几处需要继续打磨。
- 当前没有看到 P0 级别问题，但有 1 个 P1 和 3 个 P2 级别问题，足以影响生成项目上线后的安全感知和维护体验。

## Scope

- 项目类型：Go CLI 生成器，产物为 HTTP 服务骨架
- 运行上下文：生成项目默认面向生产或准生产 HTTP 服务
- 关键审查面：
  - 生成后的 HTTP 路由与中间件骨架
  - 默认错误处理与响应模型
  - timeout 路由包装
  - 示例业务服务层
  - 默认配置加载与默认值逻辑
- 报告目标：`docs/code-audit.md`

## Severity Overview

| Severity | Count | Meaning |
|---|---:|---|
| P0 | 0 | Critical |
| P1 | 1 | High |
| P2 | 3 | Medium |
| P3 | 1 | Low |

## Findings

### P0 - Critical

No findings.

### P1 - High

### P1-001: 默认错误处理会把内部错误直接暴露给客户端

- Severity: P1
- Category: Security / Reliability
- Location: `generator/assets/packs/preset-medium-v3/internal/transport/http/router/router.go.tmpl:39`，`generator/assets/packs/preset-medium-v3/internal/app/user/controller/user_controller.go.tmpl:111`
- Status: Open
- Confidence: High

#### Problem

生成骨架的默认 HTTP 错误处理会直接把 `err.Error()` 返回给客户端。对于未被识别的业务错误，这意味着数据库错误、SQL 细节、底层文件路径或内部实现信息可能被原样暴露。

#### Impact

一旦生成项目接入真实数据库、外部服务或更复杂的 DAO 层，这种默认行为会把内部实现细节暴露给外部调用方，增加信息泄露风险，也会让客户端错误语义不稳定。

#### Evidence

- `router.go.tmpl:42` 直接把 `message := err.Error()` 作为错误输出基础。
- `router.go.tmpl:48` 把该消息通过 `common.Error(...)` 返回。
- `user_controller.go.tmpl:117-118` 在默认分支直接返回 `err.Error()`。

#### Recommendation

把“默认未知错误”的对外文案统一收敛为通用消息，例如 `internal server error` 或 `request failed`，仅对显式识别的应用错误保留可读业务文案。

#### Suggested Change

- 在 `pkg/common/error.go` 中补一个“是否可公开”或“是否业务错误”的判断入口。
- `router.go` 的全局 `ErrorHandler` 对未知错误统一返回通用消息。
- controller 中的 `writeServiceError(...)` 默认分支不再直接透传 `err.Error()`。

#### Verification

- 生成一个样例项目并在 DAO 层手工返回带 SQL/底层信息的错误。
- 确认客户端只能看到通用错误文案，日志中仍能看到详细错误。

### P2 - Medium

### P2-001: timeout 路由包装只覆盖首个 handler，无法完整包住多 handler 链

- Severity: P2
- Category: Reliability / Correctness
- Location: `generator/assets/packs/preset-medium-v3/internal/transport/http/router/timeout_router.go.tmpl:102`
- Status: Open
- Confidence: High

#### Problem

当前 `wrapHandlers(...)` 只包装首个 handler，剩余 `handlers...` 原样透传。对于使用“路由级中间件 + 最终处理器”形式的业务路由，timeout 只覆盖第一段逻辑，后续链路可以绕过超时约束。

#### Impact

生成项目会给开发者一种“业务路由默认都受 timeout 保护”的预期，但一旦他们在路由上增加第二个或第三个 handler，timeout 语义就会变得不完整，导致行为不一致和调试困难。

#### Evidence

- `timeout_router.go.tmpl:102-107` 只对首个 `handler` 调用 `router.wrapHandler(...)`。
- `Get/Post/Put/...` 等方法都依赖这段逻辑。

#### Recommendation

把 timeout 语义提升为完整的 route/group middleware，而不是只包第一个 handler；或者显式遍历并包装所有 `fiber.Handler`。

#### Suggested Change

- 对 `handlers...` 中的每个 `fiber.Handler` 都进行包装。
- 更稳妥的做法是把 timeout 作为 group middleware 接到 `/api` 分组上，再通过 `ExcludePaths` 或独立 group 控制跳过范围。

#### Verification

- 在生成项目中为同一条业务路由挂两个 handler，其中第二个 handler 人工 `sleep` 超时。
- 确认超时响应仍然稳定触发，而不是只保护第一个 handler。

### P2-002: 示例业务缓存没有失效策略，CRUD 后会长期返回陈旧列表

- Severity: P2
- Category: Correctness / Maintainability
- Location: `generator/assets/packs/preset-medium-v3/internal/app/user/service/user_service.go.tmpl:43`
- Status: Open
- Confidence: High

#### Problem

示例业务 `List(...)` 会缓存分页结果，但 `Create(...)`、`Update(...)`、`Delete(...)` 完成后不会清理缓存，也没有 TTL 语义。

#### Impact

当后续 capability 或真实缓存实现接入后，示例业务会默认呈现“写成功但读出来还是旧数据”的状态。这会降低脚手架的可信度，也容易让使用者把问题误判为数据库事务或 Fiber 路由问题。

#### Evidence

- `user_service.go.tmpl:54-69` 对列表结果做缓存读写。
- `Create(...)`、`Update(...)`、`Delete(...)` 路径没有任何缓存清理逻辑。
- 当前 `cache.Store` 接口也没有失效或 TTL 能力。

#### Recommendation

至少为示例缓存补齐一种明确语义：

- 写后清理相关列表缓存，或
- 为缓存引入短 TTL，或
- 在默认 `noopStore` 之外明确说明“示例缓存未做一致性保证”

#### Suggested Change

- 给 `cache.Store` 增加 `Delete` 或 `DeletePrefix` 能力，并在写操作后失效 `users:*`。
- 如果不想引入更多接口，也可以先给缓存键加入时间分段或干脆移除默认列表缓存。

#### Verification

- 生成项目后先请求列表，再执行创建或更新，再重复请求列表。
- 确认返回结果不会长期保留旧数据。

### P2-003: 布尔配置默认值写法会吞掉显式 false，默认语义不稳定

- Severity: P2
- Category: Correctness / Maintainability
- Location: `generator/assets/packs/preset-medium-v3/config/config.go.tmpl:137`
- Status: Open
- Confidence: High

#### Problem

当前 `applyDefaults()` 用 `if !c.Log.LogCompress { c.Log.LogCompress = true }` 这类写法设置布尔默认值。由于 Go 的布尔零值和“用户显式配置 false”无法区分，最终会导致某些配置项根本无法通过 YAML 关闭。

#### Impact

这会让生成项目的配置行为和用户直觉不一致，尤其是日志压缩、日志行号、以及后续可能继续扩展的布尔选项。维护者很难解释“为什么配置成 false 还是被打开了”。

#### Evidence

- `config.go.tmpl:210-214` 对 `LogCompress` 和 `LogShowLine` 都使用了这种模式。
- timeout 相关默认值则依赖 `server.yaml` 明文输出，与布尔默认值机制形成两套语义。

#### Recommendation

把默认值逻辑改成“先构造默认配置，再反序列化覆盖”，或者把需要区分显式 false 的字段改成指针布尔。

#### Suggested Change

- 在 `Load(...)` 中先初始化带默认值的 `Config` 再 `yaml.Unmarshal(...)` 覆盖。
- 或者只对字符串、整数、切片做代码默认值，布尔默认值完全依赖生成出的 `server.yaml`。

#### Verification

- 在生成项目里把 `log_compress: false`、`log_show_line: false` 写入配置。
- 确认加载后的结果保持 false，而不是被代码重新置回 true。

### P3 - Low

### P3-001: 示例业务依赖包级单例，降低了测试性和后续多领域扩展的清晰度

- Severity: P3
- Category: Maintainability
- Location: `generator/assets/packs/preset-medium-v3/internal/app/user/service/user_service.go.tmpl:18`，`generator/assets/packs/preset-medium-v3/internal/app/user/controller/user_controller.go.tmpl:17`
- Status: Open
- Confidence: Medium

#### Problem

当前路由风格已经回到更贴近 Go 日常开发的简洁写法，但代价是 `userService` 和 `UserAPI` 依赖包级单例与隐式初始化顺序。它比集中式 DI 更轻，但也会让测试隔离、并行测试和后续多业务域扩展更依赖约定。

#### Impact

这个问题不会立刻造成线上故障，但会逐步提高“继续加第二个第三个业务域”时的心智负担，也会让单元测试更容易依赖全局状态。

#### Evidence

- `user_service.go.tmpl:18` 使用包级 `userService`。
- `Init(...)` 通过副作用写入全局状态。
- controller 通过 `service.GetUserService()` 隐式读取该状态。

#### Recommendation

在不回退到厚重 DI 风格的前提下，给骨架补一个更清晰的领域初始化约定，例如：

- 每个业务域统一暴露 `Init(...)` 和可测试的 reset helper
- 或者引入更轻的 domain bootstrap 包，集中做显式初始化

#### Suggested Change

- 保留当前 `url.go` 的简洁路由风格。
- 只收口业务初始化方式，避免 controller 直接依赖不可替换的全局服务实例。

#### Verification

- 为生成项目补一组并行测试或领域级单测。
- 确认不同测试之间不会因为全局服务状态互相污染。

## Recommended Fix Plan

1. 先处理 `P1-001`，统一未知错误的对外脱敏策略。
2. 再处理 `P2-001`，把 timeout 语义从“首个 handler 包装”修正为“整条路由链一致生效”。
3. 然后处理 `P2-002` 和 `P2-003`，分别收口缓存一致性和配置默认值语义。
4. 最后处理 `P3-001`，在不破坏当前 Go 风格路由的前提下，改善默认骨架的初始化清晰度和可测试性。

## Verification Suggestions

```bash
go test ./...
go test -race ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

生成骨架回归建议：

```bash
go run ./cmd/fiberx new audit-demo --preset medium
cd audit-demo
go test ./...
```

重点补的黑盒验证：

- 人工构造一个返回底层错误的 DAO，确认客户端不会看到内部错误文本
- 为同一路由挂多个 handler，确认 timeout 对整条 handler 链都生效
- 执行 `Create / Update / Delete` 后再次请求列表，确认缓存不会长期返回陈旧数据
- 显式把布尔配置写成 `false`，确认加载后仍为 `false`

## Notes

- 本次审计重点不是代码风格，而是默认骨架在真实项目里放大后的安全边界、正确性和维护成本。
- 当前结论并不否定现有生成骨架已经具备较好的可读性；问题主要集中在默认行为的边界还不够稳固。
