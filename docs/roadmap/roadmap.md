# Roadmap

这份路线图服务于 `fiberx` 作为 CLI-first Fiber 项目生成器的持续落地，而不是继续扩张模板仓库。

它以 [fiberx-generator-architecture.md](../architecture/fiberx-generator-architecture.md) 为上位设计依据，按 `State -> Phase` 两层结构组织实施节奏。

## 当前状态

当前仓库已经具备以下基础：

- 四个官方 preset：`heavy`、`medium`、`light`、`extra-light`
- 生成器骨架、声明层、资产层、真实写出链路
- 基础回归与生成级测试
- `v3/test`
- 独立维护的 `addons/`

同时需要明确：

- `/v3/*` 只是参考快照，不是生成器输入源
- `addons/` 继续独立，不进入 v1 直装配路径
- 生成器公开概念仍然只有 `preset` 与 `capability`

## 当前阶段

- 已完成：`State 1`
- 当前阶段：`State 2 / Phase 7`

## State 1：生成器基础成立

`State 1` 的目标是让 `fiberx` 从模板仓库完成向生成器仓库的第一阶段转型，并建立一条近生产主线。

### Phase 1：文档与命名统一

已完成。

交付结果：

- 仓库主命名统一到 `fiberx`
- 文档身份统一到“CLI-first 生成器仓库”
- 主设计文档成为上位依据

### Phase 2：生成器骨架与统一请求模型

已完成。

交付结果：

- 建立 `/cmd/fiberx`
- 建立 `/internal/core`、`manifest`、`planner`、`validator`、`renderer`、`writer`、`report`
- 统一请求对象 `Request`
- 统一入口 `Generate(req Request) error`

### Phase 3：声明层与基础执行链路

已完成。

交付结果：

- YAML manifest 落盘
- preset / capability / replace rule / injection rule 统一加载
- `load -> validate -> plan` 基础链路成立

### Phase 4：生成器资产体系落地

已完成。

交付结果：

- 建立 `generator/assets/base`
- 建立 `generator/assets/packs`
- 建立 `generator/assets/capabilities`
- 打通真实渲染与真实写出

### Phase 5：v1 可用生成器闭环

已完成。

交付结果：

- `new / init / list / explain / validate / doctor` 全部可用
- 四个 preset 全部可真实生成
- `redis` 成为首个正式实现的 capability
- 生成级 temp dir 回归建立

### Phase 6：`medium` 核心生产基线

已完成。

交付结果：

- `medium` 从薄脚手架升级为接近可直接投入生产起步阶段的服务型工程
- 默认内置：
  - 配置加载
  - 日志初始化
  - sqlite 默认接入
  - 健康检查：`/healthz`、`/livez`、`/readyz`、`/startupz`
  - `user` CRUD 示例闭环
  - request id
  - recovery
  - 安全头
  - gzip
  - ETag
- `redis` 在 `medium` 上进入真实业务链路
- `swagger` 与 `embedded-ui` 进入 `medium` 默认体验层，同时保留 capability 公开模型

### State 1 完成定义

`State 1` 完成意味着：

- 生成器架构成立
- 四个 preset 都可生成
- `medium` 成为第一条近生产主线
- 生成结果具备基础回归与行为验证

## State 2：生产能力深化

`State 2` 的目标是把“单条近生产主线”扩展为“多档可选的成熟生产基线”。

### Phase 7：`heavy` 生产级主线

当前阶段。

目标：

- 将 `heavy` 升级为第二条生产级主线
- 引入更完整的 infra 组织
- 加入 observability 基线
- 引入 scheduler / jobs 等更偏运维场景的能力
- 建立比 `medium` 更强的回归矩阵

### Phase 8：`light / extra-light` 产品化定位

目标：

- 保持轻量，但不是简单瘦身版 `medium`
- 明确两档 preset 的适用场景、目录结构与默认能力
- 让轻量 preset 也达到各自语义下的成熟可用

### Phase 9：横向生产能力补强

目标：

- 部署说明与运行手册
- 更清晰的分环境配置组织
- 更系统的错误处理与返回约定
- 更完整的生成后验证矩阵

## State 3：能力体系化

`State 3` 的目标是让 capability 从“少量可用”进入“结构稳定、组合清晰、易扩展”的阶段。

### Phase 10：现有 capability 收口

目标：

- 正式收口 `swagger / embedded-ui / redis`
- 明确默认体验与显式装配边界
- 强化依赖与冲突模型

### Phase 11：新增高价值 capability

目标：

- 只引入结构清晰、验证容易的高价值能力
- 保持 capability 体系节制，不做能力大杂烩

### Phase 12：capability 级验证体系

目标：

- 建立组合矩阵
- 建立规则层断言
- 建立 capability 级生成产物行为验证

## State 4：生成后维护与工程化

`State 4` 的目标是让 `fiberx` 从“能生成”推进到“能长期维护”。

### Phase 13：版本升级与差异检测

目标：

- 支持生成器版本演进后的差异识别
- 明确生成产物与资产版本之间的关系

### Phase 14：迁移助手与兼容性策略

目标：

- 提供基础迁移辅助
- 明确向后兼容与破坏性变更策略

### Phase 15：团队工程化工具

目标：

- 增强 `doctor`
- 增强 `validate`
- 提供生成后检查与修复建议

## 当前不做

当前 roadmap 明确不进入以下方向：

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 将 `/v3/*` 作为生成器输入源
- 在 v1/v2 主线上直接装配 `addons/`
- 远程模板源或模板市场

## 完成标准

当 `State 1` 完成后，读者应能从文档、CLI 和测试中清楚看出：

- 哪些 preset 只是可生成
- 哪些 preset 已达到近生产厚度
- 哪些 capability 属于默认体验的一部分
- 哪些 capability 仍属于后续深化内容

当前这个判断已经成立：

- `medium` 是近生产主线
- `heavy / light / extra-light` 目前是可生成 preset，但不宣称与 `medium` 同等厚度
