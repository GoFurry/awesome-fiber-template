# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/GoFurry/fiberx)](https://goreportcard.com/report/github.com/GoFurry/fiberx)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[English](./README.md)

`fiberx` 是一个以 CLI 为入口的 Fiber 项目生成器仓库。

仓库现在只维护生成器主线本身：生成资产、规划规则、校验、渲染、构建工程化和回归验证。不再把旧参考模板或 addon 池作为当前主线的一部分。

## 版本

- `v0.1.0`：已完成
- `v0.1.1`：已规划
- `v0.1.2`：已规划

## 文档入口

- [文档索引](./docs/README.md)
- [使用指南](./docs/guides/usage.md)
- [生成器架构](./docs/architecture/fiberx-generator-architecture.md)
- [模板边界](./docs/architecture/template-boundaries.md)
- [仓库规则](./docs/architecture/repository-rules.md)
- [模板选择指南](./docs/guides/template-selection.md)
- [路线图](./docs/roadmap/roadmap.md)

## 当前生成器能力

- `medium`：稳定生产基线，默认带 Swagger 和 embedded UI
- `heavy`：更完整的生产向轨道，默认带 Swagger、embedded UI、metrics、scheduler，并支持可选 Redis
- `light`：轻量 HTTP 服务，保留 SQLite-first CRUD、常见中间件，以及可选 Swagger / embedded UI
- `extra-light`：最小可启动底座，保留 SQLite 启动、健康检查与 recover-only 中间件
- 默认栈：`Fiber v3 + Cobra + Viper`
- `medium / heavy / light` 默认运行时：`zap + sqlite + stdlib`
- 兼容栈：`Fiber v2 + native-cli`
- `medium / heavy / light` 当前支持运行时参数：
  - `--logger zap|slog`
  - `--db sqlite|pgsql|mysql`
  - `--data-access stdlib|sqlx|sqlc`
- 生成项目当前支持配置 profiles、运行时元信息、升级评估和项目级构建自动化

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

## v0.1.1 预告

下一版将聚焦 Fiber v3 默认应用骨架和可选性能增强：

- Fiber v3 生命周期 hook 预留区和 `app.Hooks()` 生成
- graceful shutdown 默认模板，以及更完整的默认中间件组合：`recover`、`request id`、`logger`、`cors`
- 可选第三方 JSON backend 支持：计划参数 `--json-lib stdlib|sonic|go-json`

## Build Hook 安全提示

- `fiberx build` 可能执行项目自定义的 hooks。
- 只应在你信任的仓库中运行这些 hooks。
- 可以先使用 `fiberx build --dry-run` 查看将要执行的命令。

## 说明

- 当前仓库里只有生成器主线是正式维护的 source of truth。
- 历史内容通过 Git 历史保留，而不是继续保留仓库内 legacy 目录。

## License

本项目采用 MIT License，详见 [LICENSE](./LICENSE)。
