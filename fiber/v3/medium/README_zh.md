# Fiber v3 Medium 模板

[English](./README.md)

`medium` 是介于 `light` 和 `heavy` 之间的版本。它保留了较完整的基础设施和中间件能力，但业务层刻意收得更简单：只保留常规的 `controller`、`dao`、`service`、`models`，业务路由统一直接写在 `url.go` 里。

## 这个版本包含什么

- 默认使用 SQLite，开箱即用
- 内置完整的 `user` CRUD 示例
- 业务代码采用更普通的 Go 风格目录组织
- 保留 DB、Redis、scheduler、logging、graceful shutdown 等生命周期能力
- 保留基于 Fiber 官方中间件的 HTTP 基线：
  request ID、access log、timeout、health probes、security headers、compression、ETag、rate limiting
- Redis、Prometheus、Swagger、WAF、scheduler、embedded UI 仍然可以按配置开关

## 版本定位

这个版本的目标很明确：

- 比 `heavy` 更轻，去掉业务层的模块装配和额外包装
- 比 `light` 更全，保留更多基础设施和默认能力

如果你想要一个更适合日常业务开发的模板，而不是偏平台型的重型结构，这一版就是往这个方向走的。

## 快速开始

默认配置文件：

```bash
./config/server.yaml
```

启动服务：

```bash
go run . serve
```

首次启动时会自动完成：

- 创建 `./data/app.db`
- 在开启 `database.auto_migrate` 时自动建表
- 暴露内置的 user demo 接口

查看版本：

```bash
go run . version
```

安装或卸载服务：

```bash
go run . install
go run . uninstall
```

## 默认接口

健康检查和运行时探针：

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`

User CRUD 示例：

- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

按需启用的接口：

- `GET /csrf/token`
- `GET /metrics`
- `GET /swagger`
- `GET /debug/pprof/...`

## CRUD 示例

创建用户：

```bash
curl -X POST http://127.0.0.1:9999/api/v1/user/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 24,
    "status": "active"
  }'
```

查询列表：

```bash
curl "http://127.0.0.1:9999/api/v1/user/?page_num=1&page_size=10&keyword=alice"
```

更新用户：

```bash
curl -X PUT http://127.0.0.1:9999/api/v1/user/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Updated",
    "email": "alice.updated@example.com",
    "age": 25,
    "status": "active"
  }'
```

删除用户：

```bash
curl -X DELETE http://127.0.0.1:9999/api/v1/user/1
```

## 业务层结构

业务层保持尽量普通、直接：

- 业务代码放在 `internal/app/<domain>`
- 只保留 `controller`、`dao`、`service`、`models` 这些常规目录
- 路由统一放在 `internal/transport/http/router/url.go`
- 数据模型和定时任务这类运行时信息直接在 bootstrap 中登记

当前参考位置：

- `internal/app/user/controller`
- `internal/app/user/dao`
- `internal/app/user/service`
- `internal/app/user/models`
- `internal/transport/http/router/url.go`
- `internal/bootstrap/lifecycle.go`

新增一个业务域的典型步骤：

1. 创建 `internal/app/<domain>`。
2. 按需要添加 `controller`、`dao`、`service`、`models`。
3. 在 `internal/transport/http/router/url.go` 中注册该业务的路由。
4. 如果需要数据库模型或定时任务，再直接到 `internal/bootstrap/lifecycle.go` 中登记。

## Auto Migrate

`medium` 保留 `database.auto_migrate`，因为它对 SQLite 开箱即用体验很有帮助。

除此之外，这一版不再保留 migration 体系：

- 没有 `migrate` 命令
- 不要求 `migrations` 目录
- 没有迁移记录表

如果以后某个版本确实需要更严格的 schema 管理，可以单独在别的版本里加，而不是让默认模板变重。

## 中间件基线

默认开启：

- Request ID
- Access Log
- Recover
- CORS
- Security Headers
- Compression
- ETag
- Rate Limiter
- Health Probes

默认关闭但可直接启用：

- CSRF
- Swagger
- Prometheus
- Redis
- WAF
- Scheduler

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
- `proxy`

默认配置只依赖 Go 即可运行。

## 目录结构

- `cmd`：例如 `serve`、`install`、`uninstall`、`version`
- `config`：配置文件
- `internal/app`：业务域代码
- `internal/bootstrap`：生命周期、启动状态与健康探针
- `internal/infra`：数据库、日志、指标、缓存、调度器等基础设施
- `internal/jobs`：定时任务
- `internal/transport`：HTTP 路由、中间件和嵌入式 UI
- `pkg`：共享抽象和工具

## 测试

运行完整测试：

```bash
go test ./...
```

当前集成测试覆盖：

- SQLite 启动
- 数据库文件自动创建
- `auto_migrate` 自动建表
- 健康探针
- request ID 和 security headers
- ETag 与 compression
- user 业务的端到端 CRUD

## 当前取舍

- 路由集中写在 `internal/transport/http/router/url.go`，这是刻意保留的设计。
- 模型和任务的登记仍然集中在 `internal/bootstrap/lifecycle.go`。
- access log 已经带 request ID，但业务日志还没有自动串联请求上下文。

## 作为模板使用前的检查项

- 替换 `go.mod` 模块路径
- 修改 `config/server.yaml` 中的应用标识
- 更新鉴权密钥和服务元信息
- 不需要 demo 时移除内置 user 示例
- 在 `internal/app` 下添加你自己的业务域
