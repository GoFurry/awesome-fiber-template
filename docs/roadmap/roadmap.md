# Roadmap

这份路线图服务于 `fiberx` 作为 CLI-first Fiber 项目生成器的落地实施。

它不是愿景宣言，也不是继续扩张模板仓库的计划说明，而是基于 `docs/architecture/fiberx-generator-architecture.md` 制定的实施路线图。本文档只覆盖从当前基础推进到生成器 v1，再到 v2 少量高价值 capability 扩展的阶段安排。

## 当前状态

当前仓库已经具备以下基础：

- 四个官方 preset：`heavy`、`medium`、`light`、`extra-light`
- 集中化模板验证：`v3/test`
- 模板边界与 addon 边界文档
- 独立维护的 `addons/` 能力层

这意味着后续主线不再是继续稳定模板仓库形态，而是将现有规则、边界和验证能力收敛为可执行的生成器系统。

## 总体目标

`fiberx` 的总体目标是从模板仓库演进为生成器仓库：

- 以 CLI 作为唯一正式入口
- 以统一请求模型驱动生成流程
- 以 `base / packs / capabilities` 作为内部资产体系
- 以 `preset` 和 `capability` 作为用户可见概念
- 以可验证、可回归的输出结果作为质量约束

在这条主线上：

- `/v3/*` 只是模板工程参考快照，不是生成器输入源
- `addons/` 在 v1 中继续保持独立，不进入生成器直装配路径

## 当前阶段

- 已完成：`Phase 1：文档与命名统一`
- 当前阶段：`Phase 2：生成器骨架与统一请求模型`

## Phase 1：文档与命名统一

目标：让仓库定位、文档叙述和主设计基线保持一致。

本阶段交付物：

- 完成仓库命名、CLI 命名和文档主命名统一
- 让 `fiberx-generator-architecture.md` 成为生成器方向的上位设计文档
- 清理仍以“模板仓库扩张”为前提的旧叙事
- 统一 roadmap、README 与架构文档中的核心术语

完成标准：

- 主要文档不再把 `fiberx` 表述为持续扩张的模板集合
- 文档中对 `/v3/*`、`addons/`、preset、capability 的含义不冲突

## Phase 2：生成器骨架与统一请求模型

目标：建立生成器最小骨架，让后续实现围绕统一入口展开。

本阶段交付物：

- 建立 `/cmd/fiberx`
- 建立 `/internal/core`、`/internal/manifest`、`/internal/planner`、`/internal/validator`、`/internal/renderer`、`/internal/writer`、`/internal/report`
- 定义统一请求对象 `Request`
- 明确生成器主入口 `Generate(req Request) error`
- 确定首批 CLI 命令与职责：
  - `fiberx new`
  - `fiberx init`
  - `fiberx list presets`
  - `fiberx list capabilities`
  - `fiberx explain preset <name>`
  - `fiberx explain capability <name>`
  - `fiberx validate`
  - `fiberx doctor`

阶段约束：

- 所有命令最终都必须收敛到统一请求模型
- 不允许命令直接绕过内核写项目文件

## Phase 3：声明层与基础执行链路

目标：把组合规则从文档和人工约定收敛为可读、可校验的声明层。

本阶段交付物：

- 建立 preset manifests
- 建立 capability manifests
- 建立 replace rules
- 建立 injection rules
- 建立 manifest loading 流程
- 建立 validation 与 planning 的基础执行链路

阶段重点：

- preset manifest 描述官方起点组合
- capability manifest 描述依赖、冲突与注入行为
- replace rules 负责规则化文本替换
- injection rules 负责锚点式片段注入

阶段约束：

- 这一阶段只建立声明层与基础解析链路
- 不引入 AST-heavy 改写

## Phase 4：生成器资产体系落地

目标：建立真正由生成器维护的内部资产体系。

本阶段交付物：

- 建立 `generator/assets/base`
- 建立 `generator/assets/packs`
- 建立 `generator/assets/capabilities`
- 建立 `generator/presets`
- 建立 `generator/rules`

阶段重点：

- 以 `base / packs / capabilities` 为第一原则组织资产
- pack 作为内部装配单元，不暴露为用户新模板层
- 生成器资产不从 `/v3/*` 目录结构反推

阶段约束：

- `/v3/*` 只能作为语义和回归参考
- `/v3/*` 不是 source of truth

## Phase 5：v1 可用生成器闭环

目标：形成首个可用、可验证、可解释的生成器闭环。

本阶段交付物：

- 贯通 `new / init / list / explain / validate / doctor`
- 完成请求进入、装配规划、渲染、写出、报告的主链路
- 支持规则化文本替换
- 支持锚点式片段注入
- 建立基础验证方式
- 建立与生成结果相关的基础回归方式

v1 范围内应支持：

- module path 替换
- preset 选择
- capability 叠加
- README 与配置基础生成

v1 明确不做：

- AST 级结构改写
- 远程模板市场
- 插件化平台能力
- 直接装配 `addons/`

完成标准：

- 读者可以通过 CLI 完成最小生成闭环
- 输出结果可以被基础验证和回归检查覆盖

## Phase 6：v2 高价值 capability 扩展

目标：在 v1 闭环稳定后，扩展少量高价值、结构清晰、便于验证的 capability。

本阶段优先方向：

- `redis`
- `swagger`
- `embedded-ui`

本阶段交付物：

- 增加少量高价值 capability 的声明与资产支持
- 完善 capability 依赖与冲突处理
- 提升 capability 注入后的验证能力
- 补强 capability 层面的回归检查

阶段原则：

- 少量优先，不做能力大杂烩
- 优先结构影响清晰、价值高、容易验证的能力
- 扩展 capability，而不是扩张公开 preset 数量

## 当前不做

为避免路线图被误读，当前明确不进入以下方向：

- 不做 GUI
- 不做 AST-heavy 改写
- 不新增第五类官方 preset
- 不把 `/v3/*` 作为生成器输入源
- v1 不直接装配 `addons/`
- 不把这份 roadmap 理解为模板继续膨胀的计划

## 完成标准

当这份路线图对应的实施推进到位时，应满足以下判断：

- `fiberx` 已明确成为生成器仓库，而不是继续增长的模板仓库
- 用户通过 `preset` 和 `capability` 理解系统，而不是通过内部 pack 结构理解系统
- `Request` 与 `Generate(req Request) error` 成为统一生成入口
- `/v3/*` 被稳定地视为参考快照，而不是生成器资产源
- `addons/` 继续作为独立复用层，而不是 v1 生成器直装配层
- v1 提供完整可用闭环
- v2 在不破坏边界的前提下扩展少量高价值 capability
