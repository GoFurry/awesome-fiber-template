# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[English](./README.md)

`fiberx` 是一个以 CLI 为入口的 Fiber 项目生成器仓库。

仓库现在只聚焦生成器主线本身：生成资产、规划规则、校验、渲染、构建工程化与回归验证。不再把旧参考模板或仓库内 addon 池作为当前维护主线的一部分。

## 文档入口

- [文档索引](./docs/README.md)
- [使用指南](./docs/guides/usage.md)
- [生成器架构](./docs/architecture/fiberx-generator-architecture.md)
- [模板边界](./docs/architecture/template-boundaries.md)
- [仓库规则](./docs/architecture/repository-rules.md)
- [模板选择指南](./docs/guides/template-selection.md)
- [路线图](./docs/roadmap/roadmap.md)

## 当前生成器 Tracks

- `medium`：稳定生产基线，默认带 Swagger 和 embedded UI
- `heavy`：已完成的第二条生产主线，默认带 Swagger、embedded UI，并支持 metrics、scheduler 与可选 Redis
- `light`：成熟的轻量 HTTP 服务，保留 SQLite-first CRUD、常见中间件，以及可选 Swagger / embedded UI
- `extra-light`：最小可启动底座，保留 SQLite 启动、健康检查与 recover-only 中间件
- 默认栈：`Fiber v3 + Cobra + Viper`
- `medium / heavy / light` 默认运行时：`zap + sqlite + stdlib`
- 兼容栈：`Fiber v2 + native-cli`
- `medium / heavy / light` 当前支持的运行时参数：
  - `--logger zap|slog`
  - `--db sqlite|pgsql|mysql`
  - `--data-access stdlib|sqlx|sqlc`
- 生成项目当前已支持配置 profiles、运行元信息、升级评估和项目级构建自动化

## 如何选择

- 如果你想要最强的运维和工程基线，选 `heavy`
- 如果你想要均衡的生产级 HTTP 基线，选 `medium`
- 如果你想要更小但仍可直接使用的 HTTP 服务，选 `light`
- 如果你只想要最干净的启动点，选 `extra-light`

## 快速开始

直接从仓库根目录生成一个可运行项目：

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go run . serve
```

兼容栈示例：

```bash
go run ./cmd/fiberx new demo-legacy --preset medium --fiber-version v2 --cli-style native
```

运行时参数示例：

```bash
go run ./cmd/fiberx new demo-data --preset medium --logger slog --db pgsql --data-access sqlx
```

构建工程化示例：

```bash
go run ./cmd/fiberx build
go run ./cmd/fiberx build --dry-run
go run ./cmd/fiberx build --profile prod
```

## 仓库目标

这个仓库的目标是把 `fiberx` 本身维护成一个干净、长期可演进的生成器系统：

- 稳定的官方 preset 语义
- 生成器自维护的资产与规则
- 可验证的输出结果
- 明确的运行时与 capability 策略
- 项目级构建、元信息与升级工具链

## 说明

- 当前仓库里只有生成器主线是正式维护的 source of truth。
- 历史内容通过 Git 历史保留，而不是继续保留仓库内 legacy 目录。

## License

本项目采用 MIT License，详情见 [LICENSE](./LICENSE)。
