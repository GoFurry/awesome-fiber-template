# awesome-fiber-template

![License](https://img.shields.io/badge/License-MIT-blue.svg)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

[English](./README.md)

`awesome-fiber-template` 现在是一个专注于 Fiber v3 的 Go 后端模板仓库。

这个仓库不再尝试覆盖多套 Web 框架，而是聚焦在 Fiber 上，并提供四个不同重量级的模板版本，方便你按项目规模和偏好直接选择。

## 模板版本

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

选择一个版本直接进入即可，例如：

```bash
cd v3/light
go run . serve
```

每个版本都是独立的 Go module，拥有各自的 `go.mod`、配置文件、README 和依赖边界。

## 仓库目标

这个仓库的目标是帮你跳过重复的骨架搭建工作，同时保持模板边界清晰：

- 结构清楚、容易理解的项目组织方式
- 实用的启动链路和中间件默认值
- 在合适的版本里提供 SQLite 开箱即用体验
- 用不同重量级版本覆盖不同规模项目

## 说明

- 现在的仓库名已经和实际维护范围保持一致：只维护 Fiber 模板。
- 如果你把某个版本拿去作为自己的项目起点，记得先替换该版本 `go.mod` 里的模块路径。

## License

本项目采用 MIT License，详情见 [LICENSE](./LICENSE)。
