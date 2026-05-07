# 路线图

## 定位

`fiberx` 的目标不是变成一个大而全的平台，而是成为一个面向 GoFiber 项目的可组合生成器，以及一个轻量、可验证的构建辅助工具。

## 当前状态

- `v0.1.0`：已完成
- `v0.1.1`：已完成
- `v0.1.2`：已完成
- `v0.1.3`：进行中

当前主线已经具备这些稳定能力：

- 项目生成：`new`、`init`
- 结构说明：`list`、`explain`
- 项目检查：`inspect`、`diff`
- 升级评估：`upgrade inspect`、`upgrade plan`
- 构建辅助：`build`
- 生成器自检：`validate`、`doctor`

## v0.1.3

`v0.1.3` 聚焦 CLI 体验、构建安全边界，以及一批高优先级骨架修补：

当前状态：核心实现已完成，待手动验收与发布收口。

- 生成前预览：`new/init --print-plan [--json]`
- 构建安全开关：`build --no-hooks`、`build --yes`
- `doctor` 自动区分 generator / project / standalone
- `explain matrix` 输出 preset 与 capability 支持矩阵
- 生成骨架默认错误响应脱敏，避免把内部错误直接暴露给客户端
- `timeout` 对多 handler 链完整覆盖

暂不放入这一版：

- 缓存一致性修补
- 布尔默认值重构
- 全局单例可测试性改造
- `upgrade apply`

## v0.1.4

`v0.1.4` 聚焦默认骨架的可维护性和长期可演进性：

- 为示例业务缓存补齐失效策略或明确 TTL 语义
- 收口布尔配置默认值机制，解决显式 `false` 被吞掉的问题
- 在保留当前 Go 风格路由组织的前提下，降低业务包级单例带来的测试和扩展成本
- 继续打磨默认骨架中的公共错误模型与响应层

## v0.2.0

`v0.2.0` 的目标是把 `fiberx` 从一次性生成器推进为可持续升级工具：

- 在 `inspect / diff / upgrade plan` 基础上推进保守升级执行
- 支持 `upgrade apply --dry-run`
- 优先只更新未被用户修改的 managed files
- 为冲突场景生成 review 文件或 patch

## 后续能力方向

优先考虑这些与主线契合、维护成本适中的能力：

- `pprof`
- `rate-limit`
- `cors-profile`
- `sentry`
- `otel-lite`

暂不推进这些高耦合或高维护项：

- GORM
- 完整 JWT auth
- 多租户 / RBAC
- Kubernetes
- 复杂 CI/CD
- 完整前后端后台

## 模板系统后续打磨

后续会继续减少脆弱的字符串替换，逐步收口为更稳定的模板体系：

- `*.tmpl`：模板渲染
- `*.snippet`：仅注入片段，不直接输出文件
- 普通文件：原样复制

注入点也会继续收敛为显式标记，例如：

```go
// fiberx:inject imports
// fiberx:inject bootstrap
// fiberx:inject routes
```
