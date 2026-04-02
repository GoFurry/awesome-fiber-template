# Fiber v3 Heavy Template

[English](./README.md)

`heavy` 是这套模板里能力最完整的版本，但业务层仍然保持普通 Go 项目的写法，不再使用过度装配的重型结构。你可以直接按 `controller`、`dao`、`service`、`models` 的方式开发，并把路由集中写在 `url.go` 里。

## 这个版本包含什么

- 默认 SQLite 开箱即用
- 内置完整的 `user` CRUD 示例
- 保留 DB、Redis、scheduler、logging、graceful shutdown 等完整基础设施能力
- 保留 `pkg/httpkit`、`pkg/abstract` 这类高频公共工具能力
- 保留较完整的 Fiber 中间件基线：request ID、access log、recover、CORS、timeout、health probes、security headers、compression、ETag、rate limiting
- 通过开关支持 Redis、Prometheus、Swagger、WAF、scheduler、embedded UI
- 保留 `install` / `uninstall` 能力

## 适用场景

- 想要一套完整的 Fiber 工程模板
- 需要较多基础设施预留
- 希望模板本身就能覆盖中大型服务的常见能力

## 快速开始

```bash
go run . serve
```

查看版本：

```bash
go run . version
```

安装或卸载服务：

```bash
go run . install
go run . uninstall
```

## 默认端点

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`
- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

按需开启：

- `GET /csrf/token`
- `GET /metrics`
- `GET /swagger`
- `GET /debug/pprof/...`

## 业务组织方式

- 业务代码位于 `internal/app/<domain>`
- 正常使用 `controller`、`dao`、`service`、`models`
- 路由统一注册在 `internal/transport/http/router/url.go`
- 数据模型和定时任务等运行时注册集中在 `internal/bootstrap/lifecycle.go`

## 配置概览

主配置文件：

```bash
./config/server.yaml
```

重点配置块：

- `server`
- `database`
- `redis`
- `prometheus`
- `log`
- `middleware`
- `waf`
- `schedule`

## 取舍说明

- 这是最完整的版本，不追求最少依赖
- 业务层保持简单，但基础设施能力更全
- `pkg/httpkit` 和 `pkg/abstract` 会刻意保留在这一版里

## 使用前检查

- 替换 `go.mod` 模块路径
- 修改 `config/server.yaml` 里的应用标识
- 如果不需要 demo，删除内置 `user` 示例
- 在 `internal/app` 下添加自己的业务域
