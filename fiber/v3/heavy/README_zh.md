# Fiber v3 Heavy 模板

[English](./README.md)

`heavy` 是这个骨架的重型版本，面向中大型 Go 后端服务。它强调清晰的模块边界、完整的生命周期管理、可复用的基础设施能力，以及一套默认可运行的 HTTP 能力基线，同时仍然保留开箱即用的 demo 体验。

## 模板包含的能力

- 默认使用 SQLite。初次运行不需要安装 MySQL 或 PostgreSQL。
- 内置完整的 `user` CRUD 示例模块，覆盖路由、控制器、服务层、DAO、模型、迁移和集成测试。
- 统一的模块装配模型。一个模块可以在同一个 bundle 中注册路由、数据库模型、迁移、定时任务、启动钩子、关闭钩子和后台服务。
- 完整的生命周期启动链路。应用启动、迁移、基础设施初始化、定时任务、后台服务和优雅关闭都通过 bootstrap 层统一管理。
- 基于 Fiber 官方中间件的 HTTP 基线。已经集成 Request ID、Access Log、Timeout、Health Probes、Security Headers、Compression、ETag 和增强限流。
- 可按需开启的基础设施能力。Redis、Prometheus、Swagger、WAF、Scheduler、嵌入式 UI 都可以通过配置单独开关。

## 项目定位

这个模板有意做得比最小化 Go HTTP 服务更重一些。

- 适合：中型服务、团队内部统一模板、希望模块和生命周期结构稳定的大型项目。
- 不适合：很小的服务、一次性 demo、或者强烈偏好接近标准库风格的轻量布局的团队。

如果你后续要再拆一个更轻的版本，这个 `heavy` 可以作为能力母体。

## 快速开始

默认配置文件位置：

```bash
./config/server.yaml
```

启动 HTTP 服务：

```bash
go run . serve
```

第一次启动时，模板会自动完成：

- 创建 `./data/app.db`
- 在 `database.auto_migrate` 启用时自动建表
- 执行已注册模块迁移
- 暴露内置的 user demo 接口

只执行数据库迁移：

```bash
go run . migrate up
```

查看当前服务版本：

```bash
go run . version
```

通过服务管理集成安装或卸载服务：

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

- `GET /api/v1/users/`
- `POST /api/v1/users/`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

按需启用的接口：

- `GET /csrf/token`，当开启 CSRF 时可用
- `GET /metrics`，当开启 Prometheus 时可用
- `GET /swagger`，当 debug 模式下开启 Swagger 时可用
- `GET /debug/pprof/...`，在 debug 模式下可用

## CRUD 示例请求

创建用户：

```bash
curl -X POST http://127.0.0.1:9999/api/v1/users/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 24,
    "status": "active"
  }'
```

查询用户列表：

```bash
curl "http://127.0.0.1:9999/api/v1/users/?page_num=1&page_size=10&keyword=alice"
```

更新用户：

```bash
curl -X PUT http://127.0.0.1:9999/api/v1/users/1 \
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
curl -X DELETE http://127.0.0.1:9999/api/v1/users/1
```

默认 `user` 模型使用 `users` 表，字段包括：

- `id`
- `name`
- `email`
- `age`
- `status`
- `created_at`
- `updated_at`

## 中间件基线

这个 heavy 模板已经围绕 Fiber 官方中间件接好了一套实用的 HTTP 基线。

默认开启：

- Request ID
- Access Log
- Recover
- CORS
- Security Headers
- Compression
- ETag
- Rate Limiter
- Legacy 和官方 Health Probes

默认关闭但可以直接打开：

- CSRF
- Swagger
- Prometheus
- Redis
- WAF
- Scheduler

### 中间件说明

- Request ID 默认通过 `X-Request-ID` 响应头暴露。
- Access Log 使用 Fiber 官方 `logger`，默认日志格式中已经包含 Request ID。
- Timeout 通过对 `/api` 路由组做一层超时包装来生效，健康检查和 metrics 默认排除。
- ETag 默认开启，客户端可以直接使用条件请求。
- Compression 默认开启，客户端带上 `Accept-Encoding` 后会返回压缩内容。
- Limiter 支持多种策略和 key 来源，均可通过配置切换。

## 健康探针

模板同时支持 Fiber 官方探针接口和兼容旧风格的 JSON 健康检查接口。

- `/livez`：进程级存活检查
- `/readyz`：基于运行时状态和已启用基础设施依赖的就绪检查
- `/startupz`：启动完成状态检查
- `/healthz`：兼容型 JSON 接口，返回 `name`、`version`、`status`、`live`、`ready`、`startup`

这样既适合本地开发，也适合 Docker、Kubernetes 或者团队内部已有的旧检查方式。

## CSRF 行为

CSRF 默认关闭，因为很多 API 服务采用 token 鉴权，不一定需要它。

当你开启 `middleware.csrf.enabled` 后：

- 模板会暴露 `GET /csrf/token`
- 接口会返回 token、header 名称和 cookie 名称
- 所有写请求都需要在 CSRF 请求头中回传该 token

示例：

```bash
curl http://127.0.0.1:9999/csrf/token
```

然后在后续写请求里把返回的 token 放进 `X-Csrf-Token`。

## 限流能力

限流通过 `middleware.limiter` 配置。

支持的策略：

- `fixed`
- `sliding`

支持的 key 来源：

- `ip`
- `path`
- `ip_path`
- `header`

如果使用 `header`，需要额外设置 `middleware.limiter.key_header`。

这让模板从本地 demo 平滑过渡到网关、租户或自定义标识限流场景会更容易。

## 配置概览

主配置文件为 `./config/server.yaml`。

重要配置块：

- `server`：应用标识、端口、运行模式和运行时限制
- `database`：SQLite、MySQL、PostgreSQL 配置
- `redis`：可选 Redis 连接
- `prometheus`：指标暴露配置
- `log`：日志级别和输出文件
- `middleware`：所有 HTTP 中间件开关与参数
- `waf`：Coraza 规则配置
- `schedule`：调度器开关
- `proxy`：出站代理配置

默认配置刻意保持为只依赖：

- 已安装 Go
- 不依赖外部数据库
- 不依赖 Redis
- 不依赖 Prometheus

## 迁移机制

heavy 模板支持模块级迁移。

执行所有已注册迁移：

```bash
go run . migrate up
```

当前 demo 仍然保留 `database.auto_migrate: true`，保证首次运行足够顺滑；但长期的表结构演进仍然建议主要依赖显式迁移，而不是只依赖 `AutoMigrate`。

迁移示例：

- `internal/modules/user/migrations/seed.go`

已执行的迁移版本会记录在 `schema_migrations` 表中。

## 模块架构

每个模块都应该暴露一个 `NewBundle()` 工厂函数，并返回 `modules.Bundle`。

Bundle 可以贡献这些扩展点：

- `RouteModules`
- `DatabaseModels`
- `Migrations`
- `StartupHooks`
- `ShutdownHooks`
- `ScheduledJobs`
- `BackgroundServices`

参考实现：

- `internal/modules/module.go`
- `internal/modules/user/module.go`
- `internal/modules/schedule/schedule.go`

新增一个模块的典型步骤：

1. 在 `internal/modules/<module>` 下创建模块目录。
2. 按需要添加 controller、service、dao、models、migrations 等内容。
3. 在模块根目录实现 `NewBundle()`。
4. 在 `internal/bootstrap/application.go` 中通过 `modules.Collect(...)` 注册该模块工厂。

## 目录结构

- `cmd`：CLI 命令，例如 `serve`、`migrate up`、`install`、`uninstall`、`version`
- `config`：配置文件
- `internal/bootstrap`：应用装配和生命周期管理
- `internal/infra`：数据库、日志、指标、缓存、调度器等基础设施
- `internal/modules`：业务模块和模块 bundle
- `internal/transport`：HTTP 路由、中间件和嵌入式 UI
- `pkg`：共享抽象和工具包

## 测试

集成测试位于：

```bash
internal/bootstrap/bootstrap_integration_test.go
```

运行完整测试：

```bash
go test ./...
```

当前集成测试覆盖了：

- 默认 SQLite 配置下的服务启动
- 数据库文件自动创建
- 迁移元数据表创建
- 健康探针
- Request ID 和 Security Headers
- ETag 与 Compression 行为
- user 模块的端到端 CRUD 链路

## 当前的设计取舍

- 模块注册目前仍然集中在 `internal/bootstrap/application.go`。
- Timeout 包装层当前优先覆盖模板里常用的路由注册方式；如果后续引入特殊路由注册模式，需要注意超时覆盖是否仍然生效。
- Request ID 已经进入 access log，但业务日志还没有自动串上请求上下文，这仍然是后续可继续增强的点。

## 作为模板使用前的检查项

在把它替换成你自己的服务之前，通常还需要做这些动作：

- 替换 `go.mod` 里的模块路径
- 更新 `config/server.yaml` 中的应用标识
- 修改鉴权密钥和服务元信息
- 关闭不希望在生产环境保留的 demo 默认项
- 添加自己的业务模块，并在不需要时移除 demo 模块
