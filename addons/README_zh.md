# Addons

`addons/` 是这个仓库里的可选增强能力层。

它和 `v3/*` 模板是解耦的：模板负责提供默认可开箱即用的工程骨架，而 addon 负责承载那些并不是所有项目都默认需要、但很适合复用的基础设施能力。

## 当前目录

```text
addons/
  mail/
  migrate/
  mongodb/
  redis/
  s3/
```

## 什么能力适合做成 Addon

满足以下条件的能力，更适合进入 `addons/`：

- 对大多数项目来说是可选项
- 明显属于基础设施或通用服务接入
- 适合复制到项目边界独立使用
- 作为可复用能力块，比直接塞进模板默认主链更合理

## 什么能力不适合做成 Addon

以下内容不应该进入 `addons/`：

- 业务域代码
- 只对某一个模板层级有意义的胶水代码
- 没有清晰边界的工具杂货堆
- 明显应该属于某一层模板默认主路径的能力

## 已实现的 Addon

### `mail/`

可复用的 SMTP 邮件发送 addon，支持：

- 多 SMTP 账号池
- `none`、`round_robin`、`random` 等轮转策略
- 遇到可重试的连接或 SMTP 错误时自动切换账号
- 自定义 HTML 和内置 HTML 模板
- `cc`、`bcc`、`reply-to`、自定义头、附件等常用邮件能力

### `mongodb/`

基于官方 `mongo-driver/v2` 的 MongoDB addon，支持：

- `URI` 优先和结构化配置两种方式
- `Client`、`Database`、`Collection` 入口
- `Ping` 和 `Close`
- 围绕 collection 的薄 CRUD helper
- 需要更高级能力时直接使用原始 driver

### `s3/`

基于 AWS SDK v2 的 S3 兼容对象存储 addon，支持：

- region、endpoint、credentials、bucket、path-style 等显式配置
- 基于 bytes、reader、本地文件的上传
- 按字节或流式下载对象
- `HeadObject` 和幂等 `DeleteObject`
- `GET`、`PUT`、`DELETE` 预签名 URL

### `migrate/`

基于 `pressly/goose/v3` 的数据库迁移 addon，支持：

- 仅支持 SQL migration
- 显式配置 `Dialect`、`DSN`、`Dir` 和 tracking table
- `Up`、`Down`、`Status`、`Version`、`Create` 等常用能力
- 可直接复制进项目边界使用的薄封装

### `redis/`

基于 `go-redis/v9` 的 Redis addon，支持：

- 地址、用户名、密码、数据库编号、连接池大小等显式配置
- `New`、`Ping`、`Close` 和原始客户端出口
- 常用字符串、Hash、前缀扫描、Pipeline 辅助方法
- 与 `heavy`、`medium` 模板内 Redis 用法保持一致的调用风格

## 社区优先能力

并不是所有通用能力都值得在这个仓库里自建 addon。

对于一些社区已经非常成熟、维护成本又较高的能力，当前仓库明确采用“社区优先”策略：

- `otel`：优先采用 [`github.com/gofiber/contrib/v3/otel`](https://github.com/gofiber/contrib/tree/main/v3/otel)
- `auth`：优先采用 [`github.com/gofiber/contrib/v3/jwt`](https://github.com/gofiber/contrib/tree/main/v3/jwt)

像 API key 认证这类更偏业务边界的能力，建议由项目自己在应用层实现，而不是过早做成通用 addon。

## 接入原则

模板默认不依赖 `addons`。

只有当某个项目真的需要这些能力时，才在项目边界接入，并由项目自己管理配置、生命周期和业务封装。

## 说明

- 当前仓库长期维护的 addon 是 `mail/`、`mongodb/`、`s3/`、`migrate/`、`redis/`
- `mail/`、`mongodb/`、`s3/` 仍然是后续 addon 设计风格的参考样板
- `migrate/` 和 `redis/` 是在模板边界规则稳定后，首批完成产品化的 addon
