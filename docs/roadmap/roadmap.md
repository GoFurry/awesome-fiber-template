# Roadmap

这份路线图只保留 `fiberx` 生成器的当前状态和后续优先级；更细的阶段拆分放在各 phase 计划文档中。

## 当前状态

- 当前阶段：`State 3 / Phase 11`
- 当前进度：`Phase 11 进行中`
- 默认栈：`Fiber v3 + Cobra + Viper`
- 首轮服务 preset 默认运行时：`zap + sqlite + stdlib`
- 当前公开模型：`preset + capability + 少量生成参数`
- `Phase 11` 首轮覆盖：`medium`、`heavy`、`light`
- `extra-light` 继续保持最小化，暂不接入 `logger / db / data-access`

## 已完成摘要

- `State 1`：生成器主链路稳定，`medium` 成为第一条生产基线。
- `State 2 / Phase 7`：`heavy` 成为第二条生产主线。详见 [phase-7-plan.md](./phase-7-plan.md)
- `State 2 / Phase 8`：`light / extra-light` 完成产品化定位。详见 [phase-8-plan.md](./phase-8-plan.md)
- `State 2 / Phase 9`：默认栈切到 `Fiber v3 + Cobra + Viper`，并保留兼容回退。详见 [phase-9-plan.md](./phase-9-plan.md)
- `State 3 / Phase 10`：`swagger / embedded-ui / redis` 的 capability contract、CLI 输出、文档和验证边界完成收口。

## State 3：运行时选项与能力体系化

### Phase 11：日志、数据库与数据访问栈扩展

目标：把日志、数据库和数据访问栈纳入生成参数、模板资产、文档与验证矩阵，而不是设计成 capability。

已完成：

- CLI 已支持：
  - `--logger zap|slog`
  - `--db sqlite|pgsql|mysql`
  - `--data-access stdlib|sqlx|sqlc`
- `medium / heavy / light` 默认值已切到：
  - `logger=zap`
  - `db=sqlite`
  - `data-access=stdlib`
- `extra-light` 会在请求验证阶段拒绝新的 Phase 11 参数。
- `doctor`、`validate`、`explain preset` 已输出 Phase 11 默认值和支持矩阵。
- runtime overlay 已接入：
  - logger：`zap`、`slog`
  - data access：`stdlib`、`sqlx`、`sqlc`
- `medium / heavy / light` 已统一到多数据库配置形状：
  - `sqlite`
  - `postgres`
  - `mysql`
- CLI `pgsql` 到生成配置 `db_type: postgres` 的映射已完成。
- 仓库级回归已覆盖：
  - 参数校验
  - `extra-light` 拒绝 Phase 11 参数
  - 默认 runtime 生成
  - `slog + pgsql + sqlx` 的生成后编译验证
  - `zap + mysql + sqlc` 的生成后编译验证
- 运行级数据库矩阵已接入 root tests：
  - 默认栈下的 `medium / heavy / light`
  - `sqlite / pgsql / mysql`
  - `stdlib / sqlx / sqlc`
  - 各场景都会验证启动、health、CRUD、gzip / ETag 与 preset 默认路由 contract
- CI 已默认提供 `postgres`、`mysql` service，并注入标准 DSN 供 root tests 自动消费。

待完成：

- 在 CI 中把这套 `27` 个运行级场景稳定跑通并观察是否存在脆弱组合。
- 如果出现跨数据库或跨数据访问栈的不稳定行为，继续修复直到可以把 Phase 11 标记为 completed。
- 在 CI 实跑结果稳定之前，`Phase 11` 仍保持 active，不推进到 `Phase 12`。

边界：

- 这些选项仍属于生成参数，不属于 capability。
- 默认数据访问栈仍是 `database/sql + 手写 SQL`。
- `zerolog` 不进入本阶段 committed scope。
- 本阶段不引入更大的目录重组或架构迁移。

### Phase 12：capability 级验证体系

开始条件：`Phase 11` 的跨数据库运行级验证已经稳定收口，不再需要把 runtime matrix 当作当前主工作。

目标：建立 capability 组合矩阵、规则断言和行为级验证。

## State 4：生成后维护与工程化

### Phase 13：版本升级与差异检测

目标：支持生成器演进后的差异识别，并明确生成产物与模板资产版本的关系。

### Phase 14：迁移助手与兼容策略

目标：提供基础迁移辅助，并明确向后兼容与破坏性变更策略。

### Phase 15：`fiberx build` 与生成后工程化

详见 [build-command-plan.md](./build-command-plan.md)

目标：提供 `fiberx build`、多 target、多平台和基础发布能力。

## 暂不进入

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 将 `/v3/*` 直接作为生成器输入
- 在主生成链路中直接装配 `addons/`
- 远程模板源或模板市场
