# Fiber v3 Extra-Light Template

[English](./README.md)

`extra-light` 是这一组模板里最极简的版本。它只保留启动一个 Fiber 服务所必需的最小能力：原生 CLI、极简配置、SQLite、日志、`recover` 和健康探针。

## 这个版本包含什么

- 原生 CLI，只保留 `serve` 和 `version`
- 极简配置，只保留 `server`、`database`、`log`
- 仅支持 SQLite
- 不内置任何业务 demo
- 预留空的 `internal/app` 目录供你自己扩展
- 默认只启用两个中间件：`recover` 和 `healthcheck`
- 保留 embedded UI 能力，但默认不挂载

## 适用场景

- 想要一个尽可能干净的小起点
- 希望自己逐步往上加能力，而不是先带一堆依赖
- 适合小工具、小服务、实验项目或快速原型

## 快速开始

```bash
go run . serve
```

查看版本：

```bash
go run . version
```

## 默认端点

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`

默认会创建 `/api/v1` 路由树，但里面不包含具体业务接口。

## 配置概览

这个版本刻意把配置面压到最小：

`server`

- `app_name`
- `mode`
- `ip_address`
- `port`

`database`

- `path`

`log`

- `log_level`
- `log_path`

## 目录结构

- `cmd`：原生 CLI 入口
- `config`：YAML 配置
- `internal/app`：放自己的业务代码
- `internal/bootstrap`：启动和健康状态
- `internal/db`：SQLite 启动逻辑
- `internal/http`：路由和可选 embedded UI
- `pkg/common`：最小响应与错误辅助

## Embedded UI

仓库里保留了 embedded UI 能力，但默认不会自动挂载。

如果你需要它，可以在 `internal/http/router.go` 中创建 Fiber app 后手动调用 `AttachEmbeddedUI(app)`。

## 刻意移除的内容

- Cobra
- Redis
- WAF
- CSRF
- `pkg/httpkit`
- `pkg/abstract`
- 内置 CRUD 示例业务

## 开始扩展

1. 在 `internal/app/<domain>` 下创建你的业务域。
2. 在 `internal/http/url.go` 中注册路由。
3. 如果后续需要自动建表，再把自己的模型接进启动流程。
4. 把 `go.mod` 里的模块路径替换成你自己的。
