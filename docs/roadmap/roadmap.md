# Roadmap

这份路线图只保留 `fiberx` 生成器的当前状态、近期目标和后续优先级；更细的拆分放在各个 phase 计划文档中。

## 当前状态

- 当前阶段：`State 4 / Phase 14`
- 当前进度：`Phase 11` 已完成，`Phase 12` 已完成，`Phase 13` 已完成，`Phase 14` 进行中
- 默认栈：`Fiber v3 + Cobra + Viper`
- 首轮服务 preset 默认运行时：`zap + sqlite + stdlib`
- 当前公开模型：`preset + capability + 少量生成参数`
- `Phase 11` 首轮覆盖：`medium`、`heavy`、`light`
- `extra-light` 继续保持最小化，暂不接入 `logger / db / data-access`

## 已完成摘要

- `State 1`：生成器主链路稳定，`medium` 成为第一条生产基线。
- `State 2 / Phase 7`：`heavy` 成为第二条生产主线。详见 [phase-7-plan.md](./phase-7-plan.md)
- `State 2 / Phase 8`：`light / extra-light` 完成产品化定位。详见 [phase-8-plan.md](./phase-8-plan.md)
- `State 2 / Phase 9`：默认栈切换到 `Fiber v3 + Cobra + Viper`，并保留兼容回退。详见 [phase-9-plan.md](./phase-9-plan.md)
- `State 3 / Phase 10`：`swagger / embedded-ui / redis` 的 capability contract、CLI 输出、文档和校验边界完成收口。
- `State 3 / Phase 11`：`logger / db / data-access` 生成参数完成首轮接入；默认栈下的 `medium / heavy / light × sqlite / pgsql / mysql × stdlib / sqlx / sqlc` 运行矩阵已在提交 `1a46f0c` 的 CI 中通过。
- `State 3 / Phase 12`：完整 capability matrix 已经被请求校验、生成级断言和黑盒回归锁住，可以正式视为完成。

## State 4：生成后维护与工程化

### Phase 13：版本升级与差异检测

当前状态：`completed`

目标：支持生成器演进后的差异识别，并明确生成产物与模板资产版本的关系。

本阶段重点：

- 为生成产物补充稳定元信息文件 `.fiberx/manifest.json`
- 记录生成器版本、提交指纹、generation recipe、资产集合和受管文件哈希
- 新增 `fiberx inspect` 用于查看产物元信息
- 新增 `fiberx diff` 用于比较：
  - manifest 记录的历史受管文件
  - 当前工作目录里的受管文件
  - 当前生成器重新生成的受管文件结果
- 差异分类保持只读：
  - `clean`
  - `local_modified`
  - `generator_drift`
  - `local_and_generator_drift`

边界：

- 本阶段不做自动修复
- 本阶段不做自动迁移
- 本阶段不输出 patch
- 本阶段不把升级建议策略引入主命令链路

### Phase 14：迁移助手与兼容策略

当前状态：`active`

目标：提供只读升级辅助，并明确向后兼容、人工复核与破坏性变更策略。

本阶段重点：

- 新增 `fiberx upgrade inspect` 用于输出升级评估摘要
- 新增 `fiberx upgrade plan` 用于输出只读升级步骤建议
- 基于 `.fiberx/manifest.json`、当前 generator 版本和 `fiberx diff` 结果给出兼容等级：
  - `compatible`
  - `manual_review`
  - `breaking`
- 首轮只覆盖“已有生成项目能否被当前 generator 升级”的问题

边界：

- 本阶段不自动修改项目文件
- 本阶段不输出 patch
- 本阶段不支持直接变更 preset / capability / runtime recipe
- 本阶段不引入 `fiberx migrate`
- 本阶段不做 addon 层迁移或数据库 schema 迁移编排

### Phase 15：`fiberx build` 与生成后工程化

详见 [build-command-plan.md](./build-command-plan.md)

目标：提供 `fiberx build`、多 target、多平台和基础发布能力。

## 暂不进入

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 直接把 `/v3/*` 作为生成器输入
- 在主生成链路里直接装配 `addons/`
- 远程模板源或模板市场
