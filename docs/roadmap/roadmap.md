# Roadmap

这份路线图只保留 `fiberx` 生成器的当前状态和后续优先级；更细的阶段拆分放在各 phase 计划文档中。

## 当前状态

- 当前阶段：`State 3 / Phase 11`
- 默认栈：`Fiber v3 + Cobra + Viper`
- 首轮服务 preset 默认日志：`zap`
- 默认数据库：`sqlite`
- 默认数据访问栈：`stdlib`
- 当前公开模型：`preset`、`capability`，以及少量生成参数
- `Phase 11` 首轮覆盖：`medium`、`heavy`、`light`
- `extra-light` 保持最小化，暂不接入 `logger / db / data-access`

## 已完成摘要

- `State 1`：生成器主链路、manifest、planner、renderer、writer、report 完成，`medium` 成为第一条稳定生产基线。
- `State 2 / Phase 7`：`heavy` 成为第二条生产主线。详见 [phase-7-plan.md](./phase-7-plan.md)
- `State 2 / Phase 8`：`light / extra-light` 完成产品化定位。详见 [phase-8-plan.md](./phase-8-plan.md)
- `State 2 / Phase 9`：默认栈切到 `Fiber v3 + Cobra + Viper`，并补齐运行文档、配置 profile 和验证矩阵。详见 [phase-9-plan.md](./phase-9-plan.md)
- `State 3 / Phase 10`：`swagger / embedded-ui / redis` 的 capability contract、CLI 输出、文档和验证边界收口完成。

## 后续规划

### State 3：运行时选项与能力体系化

#### Phase 11：日志、数据库与数据访问栈扩展

目标：把日志、数据库和数据访问栈纳入生成参数、模板资产、文档和验证矩阵，而不是设计成 capability。

- 日志参数：`--logger zap|slog`
  - 默认：`zap`
  - 可选：`slog`
  - `zerolog` 仅保留为后续候选
- 数据库参数：`--db sqlite|pgsql|mysql`
  - 默认：`sqlite`
  - 生成配置统一写 `postgres`
- 数据访问栈参数：`--data-access stdlib|sqlx|sqlc`
  - 默认：`stdlib`
  - 可选：`sqlx`、`sqlc`
- 首轮范围：`medium`、`heavy`、`light`
- 暂缓范围：`extra-light`

边界：

- 这三类选择属于生成参数，不属于 capability
- 默认数据访问栈仍是 `database/sql + 手写 SQL`
- 不在本阶段引入更大的目录重组或架构迁移

#### Phase 12：capability 级验证体系

目标：建立 capability 组合矩阵、规则断言和行为级验证。

### State 4：生成后维护与工程化

#### Phase 13：版本升级与差异检测

目标：支持生成器演进后的差异识别，并明确生成产物与模板资产版本的关系。

#### Phase 14：迁移助手与兼容策略

目标：提供基础迁移辅助，并明确向后兼容与破坏性变更策略。

#### Phase 15：`fiberx build` 与生成后工程化

详见 [build-command-plan.md](./build-command-plan.md)

目标：提供 `fiberx build`、多 target、多平台和基础发布能力。

## 暂不进入

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 将 `/v3/*` 作为生成器输入源
- 在主生成链路中直接装配 `addons/`
- 远程模板源或模板市场
