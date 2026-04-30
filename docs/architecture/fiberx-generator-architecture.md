# fiberx 生成器主架构

本文档定义 `fiberx` 作为 CLI-first Fiber 项目生成器的当前主架构。

它描述的是当前仓库正式维护的 source of truth：`cmd/`、`internal/`、`generator/` 与相关文档/测试链路。后续实现应优先遵守这里的边界，而不是从历史目录或旧模板仓库反推设计。

## 1. 定位

`fiberx` 是一个 Fiber 项目生成器仓库。

它的核心产出是：

- 统一的生成流程
- 生成器自维护资产
- 可验证的 preset 与 capability 组合结果
- 项目级升级评估与构建工程化能力

它不是一个继续并行维护多套独立模板工程的仓库。

## 2. 当前主线结论

当前已经固定的架构结论如下：

- `fiberx` 以 CLI 作为唯一正式入口
- 四个官方 preset 强兼容保留：`heavy`、`medium`、`light`、`extra-light`
- 生成器内部资产以 `base / packs / capabilities / rules` 建模
- 公开概念保持为：
  - `preset`
  - `capability`
  - 少量生成参数
- 构建、元信息与升级评估都属于生成器主线的一部分

需要特别说明：

- `pack` 是内部装配概念，不是新的公开产品层
- 运行时选项和构建选项都通过 generator-owned 配置与规划链路进入，而不是依赖仓库内平行子体系

## 3. 非目标

当前阶段明确不做：

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 远程模板源
- 模板市场
- 另一套与 generator 并行维护的仓库内历史工程体系

## 4. 设计原则

### 4.1 CLI-first

所有用户可见生成行为都必须收敛到同一个生成内核。

### 4.2 用户概念最小化

用户只需要理解：

- `preset`
- `capability`
- `generator`
- `manifest`
- `build`

### 4.3 声明优先

组合关系优先通过 manifest、replace rules、injection rules 和 build config 描述，而不是长期依赖复制整套工程样板。

### 4.4 输出必须可验证

生成结果必须支持：

- 编译验证
- 生成级回归
- 黑盒行为验证
- 构建与发布结果验证

## 5. 系统模型

### 5.1 用户层

用户层固定为四个官方 preset：

- `heavy`
- `medium`
- `light`
- `extra-light`

### 5.2 内部资产层

内部资产层由这些对象组成：

- `base`
- `packs`
- `capabilities`
- `rules`

语义约束如下：

- `preset` 是官方起点语义
- `capability` 是用户显式叠加或 preset 默认带入的能力
- `pack` 是内部装配单元
- `rules` 负责替换、注入、后处理和构建配置联动

### 5.3 执行层

生成执行链路固定为：

- manifest loading
- validation
- planning
- rendering
- writing
- reporting

构建工程化链路固定为：

- build config loading
- profile merge
- target/platform expansion
- build execution
- packaging/checksum
- metadata/manifest output

## 6. 目录职责

主线目录职责固定如下：

- `/cmd/fiberx`
  - CLI 参数解析与命令分发
- `/internal/core`
  - 总编排，不承载具体规则
- `/internal/manifest`
  - manifest 读取与解析
- `/internal/planner`
  - 组合结果与资产命中选择
- `/internal/validator`
  - 请求、声明和组合合法性校验
- `/internal/renderer`
  - 模板替换、注入与生成结果组装
- `/internal/writer`
  - 落盘与覆盖保护
- `/internal/report`
  - 生成摘要与结构化结果
- `/internal/buildconfig`
  - 项目级构建配置模型
- `/internal/build`
  - 构建、打包、checksum、metadata、release manifest
- `/internal/metadata`
  - 项目元信息与 diff/upgrade 相关模型
- `/generator`
  - 生成资产、manifest 与 rules

## 7. 统一请求模型与入口

当前公开入口保持为：

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

CLI 正式命令包括：

- `fiberx new`
- `fiberx init`
- `fiberx list presets`
- `fiberx list capabilities`
- `fiberx explain preset <name>`
- `fiberx explain capability <name>`
- `fiberx inspect`
- `fiberx diff`
- `fiberx upgrade inspect`
- `fiberx upgrade plan`
- `fiberx build`
- `fiberx validate`
- `fiberx doctor`

## 8. 长期边界

长期边界固定如下：

- 当前仓库只维护 generator 主线
- 可选基础设施能力若未来进入主线，应通过 preset、capability、运行时选项或构建配置扩展进入
- 不再依赖仓库内平行的旧模板体系或 addon 池来承接主线能力演进
