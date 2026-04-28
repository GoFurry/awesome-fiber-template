# Phase 7 Plan

这份文档把 `State 2 / Phase 7` 从“方向描述”收敛成一版可执行计划，目标是在不破坏当前 `medium` 主线稳定性的前提下，把 `heavy` 推进为第二条可验证的生产级主线。

## 1. 当前判断

当前仓库已经具备：

- `heavy` preset 的 manifest、pack 和基础生成路径
- `medium` 作为已验证的近生产主线
- 根模块生成级测试
- `v3/heavy` 作为参考快照

当前仍然缺少：

- `heavy` 的真实服务启动闭环，而不是仅生成占位结构
- 与 `medium` 对等的行为级黑盒验证
- `heavy` 的 observability 基线
- `scheduler / jobs` 的最小可运行闭环
- 文档、CLI、测试对当前阶段的一致表达

因此，`Phase 7` 不应追求“一次性搬完 v3/heavy”，而应先完成一版 `heavy MVP`。

## 2. Phase 7 目标

`Phase 7` 完成时，应满足以下判断：

- `heavy` 不再只是“可生成 preset”，而是“第二条可运行、可验证的生产级主线”
- 生成后的 `heavy` 项目可以启动、暴露健康检查、完成基础业务闭环
- `heavy` 默认具备比 `medium` 更强的运维导向能力
- `heavy` 的新增能力都有生成后验证，而不是只停留在模板存在

## 3. 范围收敛

本阶段只做：

- `heavy` 服务启动链路补全
- `heavy` 配置模型补全
- `heavy` observability 基线
- `heavy` jobs / scheduler 最小闭环
- `heavy` 生成后黑盒测试
- 文档与 CLI 阶段口径对齐

本阶段不做：

- 新增第五类官方 preset
- 把 `addons/` 并入主生成链路
- AST 级模板改写
- 远程模板源
- 一次性追平 `v3/heavy` 全部历史能力

## 4. 建议拆分

### Workstream A：阶段口径对齐

目标：先让文档、CLI、测试说的是同一件事。

交付：

- `docs/README.md`、`docs/roadmap/roadmap.md` 对 `Phase 7` 的表述统一
- `fiberx doctor` 输出从“State 1 / Phase 6”推进到真实阶段
- `fiberx validate` 输出补充 `heavy` 主线状态
- CLI 测试更新为新的阶段断言

完成标准：

- 用户从 README、docs、`doctor`、`validate` 看到的阶段信息一致

### Workstream B：`heavy MVP` 运行闭环

目标：让 `heavy` 先成为真正可启动的服务型脚手架。

交付：

- `heavy` 配置加载与默认值模型
- `heavy` bootstrap / serve 链路
- `heavy` HTTP 路由与健康检查
- `heavy` 基础业务示例闭环
- `heavy` 默认 infra 装配框架

建议最低能力：

- 配置加载
- 日志初始化
- sqlite 默认接入
- 健康检查：`/healthz`、`/livez`、`/readyz`、`/startupz`
- request id
- recovery
- 安全头
- gzip
- ETag
- `user` CRUD 示例闭环

完成标准：

- `go run . serve` 可启动
- 基础 HTTP 路由可访问
- 生成产物 `go test ./...` 通过

### Workstream C：运维导向能力补全

目标：体现 `heavy` 相比 `medium` 的增量价值，而不是只是另一套目录。

建议首批能力：

- metrics 暴露
- scheduler / jobs 最小闭环
- 更明确的 service 列表与组件状态
- 更完整的配置分区，例如 `ops`、`scheduler`、`metrics`

建议收敛方式：

- metrics 先做 `/metrics` 或等价最小暴露面
- scheduler 先做一个固定示例 job，不引入复杂分布式语义
- 先验证本地单实例行为，不急着做高可用编排

完成标准：

- `heavy` 具备至少一项 `medium` 没有的真实可运行能力
- 这项能力出现在文档、生成结果和黑盒测试中

### Workstream D：回归与冒烟矩阵

目标：把 `heavy` 从“模板存在”推进到“持续可证明”。

交付：

- 根模块生成级测试补上 `heavy` 行为断言
- `heavy` 生成后黑盒测试
- CI 增加根模块 generator 测试
- 输出目录冒烟流程固化为可复用命令

建议最小测试矩阵：

- `heavy`
- `heavy + redis`
- `medium`
- `medium + redis`

完成标准：

- 根模块 `go test ./...` 覆盖 `heavy` 行为
- CI 不再只验证 `v3/*` 与 `addons/`

## 5. 建议执行顺序

建议按以下顺序推进：

1. 先做 Workstream A，避免文档和代码继续漂移。
2. 再做 Workstream B，把 `heavy` 变成真正可运行的脚手架。
3. 接着做 Workstream C，只补最能体现 `heavy` 定位的那一组运维能力。
4. 最后做 Workstream D，把能力固化进测试和 CI。

## 6. 里程碑建议

### M1：阶段对齐

- `doctor` / `validate` / docs 统一更新
- 不引入新功能

### M2：`heavy MVP` 可运行

- 生成后可启动
- 具备基础 API 与健康检查
- 有最小业务闭环

### M3：`heavy` 运维能力成形

- metrics 与 scheduler 至少落地一个真实闭环
- 文档明确 `heavy` 相比 `medium` 的定位差异

### M4：验证闭环

- `heavy` 黑盒测试加入根模块
- CI 加入 generator 主链路校验

## 7. 完成定义

`Phase 7` 可以视为完成，当且仅当：

- `heavy` 生成结果不是占位工程
- `heavy` 有明确强于 `medium` 的默认能力
- `heavy` 的关键能力有生成后验证
- CLI 与文档明确宣称 `heavy` 是第二条生产级主线

在此之前，`heavy` 仍应被表述为“正在推进中的生产主线”，而不是已经完成的生产级基线。
