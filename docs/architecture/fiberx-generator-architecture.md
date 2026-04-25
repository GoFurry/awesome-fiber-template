# fiberx 生成器主设计

本文档定义 `fiberx` 作为 CLI-first Fiber 项目生成器的当前主设计。

它不是对 `fiberx-repo-migration-plan.md` 的简单转写，而是对其中已形成共识的部分做收敛、对关键边界重新定稿，并提供一份可以直接指导后续实现的架构基线。

`fiberx` 由更早期的 `awesome-fiber-template` 仓库演进而来，但旧名字只作为历史说明保留，不再参与当前设计命名。

## 1. 定位

`fiberx` 是一个 Fiber 项目生成器仓库。

它的主要产出是：

- 统一的生成执行流程
- 生成器维护的资产与规则
- 稳定的官方项目起点

它的目标不是继续扩张为一个越来越大的模板目录集合。

## 2. 非目标

`fiberx` v1 不解决以下问题：

- GUI 工作流
- AST 级代码重写
- 新增第五类公开官方 preset
- 在生成器中直接装配 `addons/`
- 将 `/v3/*` 当作生成器输入资产

## 3. 设计原则

### 3.1 CLI-first

CLI 是唯一正式入口。所有生成行为都必须收敛到同一个生成内核，而不是通过零散命令直接写文件。

### 3.2 用户可见概念最小化

用户可见的核心概念只保留：

- `preset`
- `capability`
- `generator`
- `manifest`

内部装配细节不应演化成新的公开模板分类。

### 3.3 内部组合细节隐藏

生成器内部通过以下概念组织资产：

- `base`
- `pack`
- `capability`

用户只需要理解 preset 和 capability，不需要直接理解 pack 的拆分细节。

### 3.4 规则优先于长期手工复制

仓库应优先使用 manifest、replace rules、injection rules 等声明式规则，而不是长期依赖模板之间的人工同步复制。

### 3.5 输出必须可验证

生成结果必须是可检查、可测试、可回归验证的，而不是只追求“一次生成成功”。

## 4. 系统模型

`fiberx` 采用三层语义模型，加上一层执行模型。

### 4.1 用户层：Presets

官方 preset 继续保留：

- `heavy`
- `medium`
- `light`
- `extra-light`

这四个名称及其大体语义视为强兼容边界。内部实现方式可以重构，但它们作为官方起点的含义应保持稳定。

### 4.2 内部资产层：Base / Packs / Capabilities

生成器资产模型以以下三类内部概念组织：

- `base`
- `packs`
- `capabilities`

定义如下：

- `preset` 是官方起点语义
- `capability` 是用户显式叠加的能力
- `pack` 是内部装配单元
- `base` 是内部最小基座

这里的一个关键判断是：

- `/v3` 不是 pack
- `/v3` 不是 capability 来源
- `/v3` 不参与生成器内部资产建模

### 4.3 执行层：Generator Pipeline

生成器执行链路拆分为：

- manifest loading
- planning
- validation
- rendering
- writing
- reporting

## 5. 目录结构与模块职责

建议目标结构如下：

```text
/cmd/fiberx
/internal/core
/internal/manifest
/internal/planner
/internal/validator
/internal/renderer
/internal/writer
/internal/report
/generator
  /assets
    /base
    /packs
    /capabilities
  /presets
  /rules
/docs
/v3
/addons
```

### 5.1 `/cmd/fiberx`

CLI 入口，负责：

- 参数解析
- 命令分发
- 构造统一请求对象
- 调用生成内核

### 5.2 `/internal/core`

总编排入口，负责：

- 组织生成主流程
- 协调 manifest、planner、validator、renderer、writer、report

`core` 负责流程编排，不承载具体资产规则和硬编码装配逻辑。

### 5.3 `/internal/manifest`

声明读取与解析层，负责：

- preset manifests
- capability manifests
- replace rules
- injection rules

它负责读取和校验声明，不直接执行生成。

### 5.4 `/internal/planner`

装配规划层，负责：

- 解析 preset 与 capability 的最终组合
- 确定参与输出的资产集合
- 决定哪些文件被复制、渲染、注入或跳过

`planner` 只产出计划，不直接落盘。

### 5.5 `/internal/validator`

校验层，负责：

- 请求合法性检查
- capability 依赖检查
- preset 与 capability 冲突检查
- 输出结构约束检查

### 5.6 `/internal/renderer`

渲染层，负责：

- placeholder 替换
- 文本模板渲染
- 基于锚点的代码片段注入

`renderer` 负责生成期转换，不负责高层组合决策。

### 5.7 `/internal/writer`

写出层，负责：

- 目录创建
- 文件写出
- 覆盖策略
- 安全落盘行为

### 5.8 `/internal/report`

报告层，负责输出：

- 使用的 preset
- 启用的 capabilities
- 生成摘要
- 警告项与跳过项

## 6. 统一请求模型与 CLI

所有命令最终都必须收敛到统一请求模型。

最小请求示意：

```go
type Request struct {
	ProjectName  string
	ModulePath   string
	Preset       string
	Capabilities []string
	Options      map[string]string
}
```

核心入口示意：

```go
func Generate(req Request) error
```

这里固定的架构约束是：

- CLI 命令负责构造统一请求
- 生成内核负责项目输出行为
- 不允许命令绕过内核直接生成文件

### 6.1 首批命令

第一批 CLI 命令定为：

- `fiberx new`
- `fiberx init`
- `fiberx list presets`
- `fiberx list capabilities`
- `fiberx explain preset <name>`
- `fiberx explain capability <name>`
- `fiberx validate`
- `fiberx doctor`

### 6.2 命令职责

- `new`：非交互生成
- `init`：交互式生成，但仍走同一生成内核
- `list presets`：列出官方起点
- `list capabilities`：列出支持的可选能力
- `explain ...`：解释 preset 或 capability 的含义
- `validate`：校验 manifests、rules 与生成器资产
- `doctor`：检查本地环境和生成器状态

## 7. Manifest 与 Rules 的边界

`fiberx` 采用声明驱动模型，但本文档只固定职责边界，不冻结字段级完整 schema。

### 7.1 Preset Manifest

Preset manifest 描述：

- 官方 preset 身份
- 所属 base
- 关联 packs
- 默认或允许的 capability 范围

它描述的是官方起点，不是任意用户自定义组合。

### 7.2 Capability Manifest

Capability manifest 描述：

- capability 身份
- 依赖关系
- 冲突关系
- 它带来的文件、注入点或规则效果

它描述的是用户显式叠加的可选能力。

### 7.3 Replace Rules

Replace rules 负责：

- 文本占位符替换
- module path 替换
- 文件名或路径名替换

### 7.4 Injection Rules

Injection rules 负责：

- 锚点定义
- 片段插入位置
- 多能力共同作用时的注入顺序

### 7.5 v1 的转换边界

v1 只支持：

- 规则化文本替换
- 基于锚点的片段注入

v1 不支持 AST 级结构改写。

## 8. 对外概念与核心对象

这份设计刻意将主要接口面保持在最小范围内。

公开概念：

- `preset`
- `capability`
- `generator`
- `manifest`

核心对象：

- `Request`
- `Generate(req Request) error`
- `PresetManifest`
- `CapabilityManifest`
- `ReplaceRules`
- `InjectionRules`

本文档不展开这些对象的完整字段表，只给出最小示意与职责边界，避免在还未讨论完的细节上过早冻结。

## 9. 与现有仓库内容的关系

### 9.1 `/v3/*`

`/v3/*` 的角色是模板工程参考快照。

它可以用于：

- 历史对照
- 语义讨论
- 回归参考

它不能被定义为：

- 生成器源资产
- pack 定义来源
- capability 定义来源
- 生成器长期维护中心

### 9.2 `addons/`

`addons/` 继续保持独立可复用层定位。

它的角色仍然是：

- copy-friendly
- 与 `v3/*` 模板内部实现解耦
- 在应用边界自行接入和持有

生成器 v1 不直接装配 addons。

### 9.3 `docs/`

`docs/` 继续承载长期规则、架构说明和设计文档。

### 9.4 `generator/assets/*`

`generator/assets/*` 才是后续生成器资产的主维护区域。

## 10. 分阶段落地顺序

后续实现应按架构驱动顺序推进。

### Phase 1

- 完成命名与文档统一
- 建立主设计文档

### Phase 2

- 定义生成内核目录
- 定义统一请求模型
- 建立初始 CLI 外壳与命令职责边界

### Phase 3

- 建立 presets、capabilities、rules 的声明层
- 建立声明读取与基础校验流程

### Phase 4

- 建立 `base / packs / capabilities` 的生成器资产体系
- 实现初始渲染与写出流程

### Phase 5

- 实现首批 capability
- 建立输出验证链路

### Phase 6

- 扩展 capability 覆盖范围
- 完善更强的回归与验证机制

### 阶段性约束

在任何阶段，都不应把 `/v3/*` 重新定义成生成器输入资产。

后续验证可以参考 `/v3` 的语义，但 `/v3` 不是生成器的 source of truth。

## 11. 验收标准

如果另一位工程师读完本文档后，可以明确得出以下结论，则说明设计达标：

- `fiberx` 是项目生成器，而不是继续膨胀的模板仓库
- `/v3` 不是生成器 source of truth
- `addons/` 不属于生成器 v1 的直接装配层
- 后续实现应围绕统一请求驱动的生成内核展开
- 可以直接据此开始实现目录骨架、请求模型、首批 CLI 和声明层

## 12. 一致性检查

任何基于本文档的实现，都应保持与现有规则文档一致。

至少应检查：

- `docs/architecture/template-boundaries.md`
- `docs/architecture/repository-rules.md`
- `docs/roadmap/roadmap.md`

评审时应特别确认：

- 四个 preset 的语义仍与当前仓库定位一致
- 文档没有把 `/v3` 写成 generator source-of-truth
- 文档没有把 `addons/` 纳入生成器 v1
- CLI、请求模型和模块职责之间没有重叠或越界

## 13. 本文档固定的默认值

- 主文档语言使用中文，保留必要英文术语
- 本文档位于 `docs/architecture/`
- 文件名为 `fiberx-generator-architecture.md`
- 保留一小段从 `awesome-fiber-template` 演进而来的历史说明
- 首批 CLI 命令集合在职责级别固定，但不在这里冻结完整交互细节
