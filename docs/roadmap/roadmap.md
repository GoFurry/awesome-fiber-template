# 路线图

## 定位

`fiberx` 的目标不是变成大而全的平台，而是成为一个面向 GoFiber 项目的可组合生成器，以及一个轻量、可验证的构建辅助工具。

## 当前状态

- `v0.1.0`：已完成
- `v0.1.1`：已完成
- `v0.1.2`：已完成

当前主线已经具备这些稳定能力：

- 项目生成：`new`、`init`
- 结构理解：`list`、`explain`
- 项目检查：`inspect`、`diff`
- 升级评估：`upgrade inspect`、`upgrade plan`
- 构建辅助：`build`
- 生成器自检：`validate`、`doctor`

## v0.1.2

`v0.1.2` 已完成，重点是默认骨架的一致性和可用性收口：

- 为 `light`、`medium`、`heavy` 补齐 `pkg/common/constant.go`
- 为 `light`、`medium`、`heavy` 补齐基础错误模型 `pkg/common/error.go`
- 扩展 `pkg/common/response.go`，兼容简单 helper 和响应包装器两套用法
- 为业务路由接入默认启用、可配置的 timeout wrapper
- 为生成项目补齐 `middleware.timeout` 默认配置
- 保持系统路由不进入 timeout 包装
- 保持 `extra-light` 继续作为最小骨架

## v0.1.3

`v0.1.3` 聚焦 CLI 体验和构建安全边界：

- 生成前预览：`--print-plan`、`--dry-run`
- 构建安全开关：`--no-hooks`、显式确认流程
- `doctor` 分层：区分生成器环境与生成项目诊断
- `explain matrix`：直接展示 preset 与 capability 支持矩阵

## v0.2.0

`v0.2.0` 的目标是把 `fiberx` 从一次性生成器推进为可持续升级工具：

- 在 `inspect / diff / upgrade plan` 基础上推进保守升级执行
- 支持 `upgrade apply --dry-run`
- 优先只更新未被用户修改的 managed files
- 为冲突场景生成 review 文件或 patch

## 后续能力方向

优先考虑这些维护成本适中、与主线契合的能力：

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
