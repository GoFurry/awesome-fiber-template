# Fiber v3 Light Template

[English](./README.md)

`light` 是更接近普通 Go 项目风格的版本。它保留了常见 API 服务需要的中间件和 SQLite 开箱即用体验，但移除了 Redis、service 管理、WAF、Swagger、CSRF、pprof 以及额外工具包。

## 这个版本包含什么

- 默认 SQLite 开箱即用
- 内置完整的 `user` CRUD 示例
- `internal/app` 下保持普通 Go 风格的业务目录结构
- 保留 DB、logging、graceful shutdown
- 保留 embedded UI 能力，但默认关闭
- 保留 API 服务常用中间件：request ID、access log、recover、CORS、timeout、health probes、compression、ETag、rate limiting

## 适用场景

- 想要一个更像普通 Go 项目的模板
- 需要一套足够实用的 API 模板，但又不想带太多附加能力
- 适合小型到中小型 API 服务

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
- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

## 业务组织方式

- 业务代码位于 `internal/app/<domain>`
- 正常使用 `controller`、`dao`、`service`、`models`
- 路由统一注册在 `internal/transport/http/router/url.go`
- 数据模型注册集中在 `internal/bootstrap/lifecycle.go`

## 配置概览

主配置文件：

```bash
./config/server.yaml
```

重点配置块：

- `server`
- `database`
- `log`
- `middleware`

## 取舍说明

- 比 `medium` 更轻，去掉了 Redis、service 管理和额外工具包
- 比 `extra-light` 更完整，仍然保留常见 API 中间件和 demo 业务
- 适合作为普通 Go API 项目的直接起点
- 仓库层面的长期边界和演进规则统一放在根目录 `docs/` 下维护，而不是继续扩散到各模板 README

## 使用前检查

- 替换 `go.mod` 模块路径
- 修改 `config/server.yaml` 里的应用标识
- 如果不需要 demo，删除内置 `user` 示例
- 在 `internal/app` 下添加自己的业务域
