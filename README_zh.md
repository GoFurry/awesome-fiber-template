# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[English](./README.md)

`fiberx` 现在是一个以 CLI 为入口的 Fiber 项目生成器仓库。

`fiberx` 由更早期的 `awesome-fiber-template` 仓库演进而来，后续会以这个名字继续正式维护。

这个仓库不再尝试覆盖多套 Web 框架，而是聚焦在 Fiber 上，并逐步从模板仓库收敛为生成器仓库，同时保留四个稳定的官方起点语义。

当前仓库仍保留 `v3/*` 参考模板工程，也保留独立的 `addons/` 目录作为可选能力层，用来承载可复用的外部服务封装和基础设施能力。

## 文档入口

- [文档索引](./docs/README.md)
- [模板边界](./docs/architecture/template-boundaries.md)
- [Addon 设计规则](./docs/architecture/addon-design-rules.md)
- [模板选择指南](./docs/guides/template-selection.md)
- [Addon 接入指南](./docs/guides/addon-integration.md)
- [路线图归档](./docs/roadmap/roadmap.md)

## 当前参考 Presets

- [`v3/heavy`](./v3/heavy)：能力最完整的版本，保留 Redis、定时任务、service 安装卸载、WAF、Prometheus、Swagger，以及 `pkg/httpkit`、`pkg/abstract` 这类可复用工具能力
- [`v3/medium`](./v3/medium)：偏均衡的 HTTP 服务版本，保留 Redis、WAF、service 管理、embedded UI 和大部分中间件，但去掉定时任务和 Prometheus
- [`v3/light`](./v3/light)：更接近普通 Go 项目的版本，保留常见 API 中间件和可选 embedded UI，移除 Redis、service 管理以及额外工具包
- [`v3/extra-light`](./v3/extra-light)：最极简的版本，使用原生 CLI、仅支持 SQLite、不内置业务示例，并且默认只保留 `recover + healthcheck`

## 如何选择

- 如果你想要最完整的工程基线，选 `heavy`
- 如果你想要一个更适合常规业务服务的均衡模板，选 `medium`
- 如果你希望结构更像普通 Go 项目，选 `light`
- 如果你只想要一个尽可能干净的小起点，选 `extra-light`

## 快速开始

当前仍可以直接进入某个参考 preset 运行，例如：

```bash
cd v3/light
go run . serve
```

每个参考 preset 都是独立的 Go module，拥有各自的 `go.mod`、配置文件、README 和依赖边界。

当前生成器默认产物已经切到 `Fiber v3 + Cobra + Viper`，同时保留 `Fiber v2 + native-cli` 兼容回退，并且会额外生成 `server.yaml / server.dev.yaml / server.prod.yaml` 以及运行手册、返回约定和验证文档。

当前 `medium / heavy / light` 的默认运行时已经切到 `zap + sqlite + stdlib`，并开始支持 Phase 11 运行时参数：

- `--logger zap|slog`
- `--db sqlite|pgsql|mysql`
- `--data-access stdlib|sqlx|sqlc`

## 仓库目标

这个仓库的目标是把当前 preset 语义、规则和验证能力沉淀为生成器系统，同时保持边界清晰：

- 稳定的官方 preset
- 生成器维护的规则与资产
- 实用的启动链路和中间件默认值
- 可验证、可回归的输出结果
- 独立的可选 addon 能力层

## 说明

- `fiberx` 现在的正式定位是生成器仓库，`v3/*` 继续作为参考 preset 快照保留。
- 如果你把当前某个参考 preset 拿去作为自己的项目起点，记得先替换该 preset `go.mod` 里的模块路径。

## License

本项目采用 MIT License，详情见 [LICENSE](./LICENSE)。
