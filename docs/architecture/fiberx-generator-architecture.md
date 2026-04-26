# fiberx 生成器主设计

本文档定义 `fiberx` 作为 CLI-first Fiber 项目生成器的当前主设计。

它不是旧迁移草案的简单改写，而是当前已经达成共识的架构基线。后续实现需要优先遵守这里的边界，而不是从 `/v3/*` 或历史模板目录反推设计。

## 1. 定位

`fiberx` 是一个 Fiber 项目生成器仓库。

它的核心产出是：

- 统一生成流程
- 生成器主维护资产
- 可验证的 preset 与 capability 组合结果

它不是继续扩张的模板仓库。

## 2. 当前 State 1 结论

当前已经固定的结论如下：

- `fiberx` 以 CLI 作为唯一正式入口
- 四个官方 preset 强兼容保留：`heavy`、`medium`、`light`、`extra-light`
- `/v3/*` 只是参考快照，不是生成器输入源
- `addons/` 继续作为独立复用层存在，v1/v2 主线不直接装配
- 生成器内部资产以 `base / packs / capabilities` 建模
- 公开概念仍然只有 `preset` 与 `capability`

当前 `State 1` 支持矩阵为：

- 可真实生成 preset：`heavy`、`medium`、`light`、`extra-light`
- 已实现 capability：`redis`、`swagger`、`embedded-ui`
- 当前近生产主线：`medium`

其中需要特别说明：

- `swagger` 与 `embedded-ui` 仍保留 capability manifest 和 explain/list 能力
- 对 `medium` 来说，它们同时进入了默认体验层
- `heavy / light / extra-light` 目前不宣称与 `medium` 同等厚度

## 3. 非目标

当前阶段明确不做：

- GUI
- AST 级改写
- 第五类官方 preset
- 将 `/v3/*` 作为生成器 source of truth
- 将 `addons/` 直接接入生成器主装配链路
- 远程模板源

## 4. 设计原则

### 4.1 CLI-first

所有用户可见生成行为都必须收敛到同一个生成内核。

### 4.2 用户概念最小化

用户只需要理解：

- `preset`
- `capability`
- `generator`
- `manifest`

`pack` 是内部装配概念，不应演化为新的公开模板层。

### 4.3 声明优先

组合关系优先通过 manifest、replace rules、injection rules 描述，而不是长期依赖手工复制模板目录。

### 4.4 输出必须可验证

生成结果必须支持编译验证、生成级回归和必要的黑盒行为验证。

## 5. 系统模型

### 5.1 用户层

用户层固定为四个官方 preset：

- `heavy`
- `medium`
- `light`
- `extra-light`

### 5.2 内部资产层

内部资产层由三类对象组成：

- `base`
- `packs`
- `capabilities`

语义约束如下：

- `preset` 是官方起点语义
- `capability` 是用户显式叠加或 preset 默认带入的能力
- `pack` 是内部装配单元
- `/v3/*` 既不是 pack，也不是 capability 来源

### 5.3 执行层

生成执行链路固定为：

- manifest loading
- validation
- planning
- rendering
- writing
- reporting

## 6. 目录与模块职责

推荐结构如下：

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
  /capabilities
  /rules
/docs
/v3
/addons
```

职责边界如下：

- `/cmd/fiberx`
  - CLI 参数解析与命令分发
- `/internal/core`
  - 总编排，不承载具体规则
- `/internal/manifest`
  - 声明读取与解析
- `/internal/planner`
  - 组合结果与命中资产选择，不直接写文件
- `/internal/validator`
  - 声明、请求、资产存在性和组合合法性校验
- `/internal/renderer`
  - 模板替换与锚点注入
- `/internal/writer`
  - 真实落盘与覆盖保护
- `/internal/report`
  - 生成摘要、警告与结构化结果

## 7. 统一请求模型与入口

当前公开入口保持固定：

```go
type Request struct {
	ProjectName  string
	ModulePath   string
	Preset       string
	Capabilities []string
	Options      map[string]string
}
```

```go
func Generate(req Request) error
```

CLI 首批正式命令保持不变：

- `fiberx new`
- `fiberx init`
- `fiberx list presets`
- `fiberx list capabilities`
- `fiberx explain preset <name>`
- `fiberx explain capability <name>`
- `fiberx validate`
- `fiberx doctor`

约束如下：

- 所有命令最终都必须收敛到统一请求模型
- CLI 不能绕过内核直接写项目文件
- `new` 写到 `<cwd>/<projectName>`
- `init` 写到当前目录

## 8. Manifest 与规则边界

需要长期存在的声明对象包括：

- `PresetManifest`
- `CapabilityManifest`
- `ReplaceRules`
- `InjectionRules`

职责边界如下：

- preset manifest 描述官方起点组合
- capability manifest 描述依赖、冲突、默认支持范围和内部 pack
- replace rules 负责文本与路径替换
- injection rules 负责锚点式注入
- 当前阶段仍只做规则化文本替换和锚点注入，不做 AST 改写

## 9. 与现有仓库内容的关系

### 9.1 `/v3/*`

`/v3/*` 是模板工程参考快照，仅用于：

- 语义对照
- 回归参考
- 设计讨论

它不是生成器源资产，也不是主维护目录。

### 9.2 `addons/`

`addons/` 继续作为独立可复用层存在。

它可以提供风格参考和未来扩展方向，但当前主线不把它作为 v1/v2 的直接输入资产。

### 9.3 `generator/assets/*`

这里是当前和未来的主维护资产目录。

## 10. 当前生产基线策略

`State 1` 完成后，生成器进入“单条近生产主线”阶段。

当前策略为：

- `medium` 是第一条近生产主线
- `swagger` 与 `embedded-ui` 在 `medium` 中进入默认体验层
- `redis` 在 `medium` 中进入真实业务链路
- `heavy / light / extra-light` 继续保留可生成能力，但后续再分别深化

`medium` 当前主线能力包括：

- 配置加载
- 日志初始化
- sqlite 默认接入
- 健康检查
- `user` CRUD 示例闭环
- request id
- recovery
- 安全头
- gzip
- ETag

## 11. 后续演进

`State 1` 之后，主线依次进入：

- `State 2`：生产能力深化
- `State 3`：能力体系化
- `State 4`：生成后维护与工程化

这意味着后续优先级是继续提升生产能力，而不是先做插件生态或升级工具链。
