# Addons

`addons` 是这个仓库中的可复用能力区。

它和 `v3/*` 模板是分离的：模板本身保持轻量，只有在项目真的需要某项基础设施能力时，才把对应 addon 复制进项目里使用。

## 当前目录

```text
addons/
  mail/
  mongodb/
  s3/
```

## 设计原则

- 运行时代码尽量小，方便直接复制进项目
- 默认不和 `v3/*` 模板耦合
- 一个 addon 只解决一类明确的基础设施问题
- 先说明边界和用法，再增加实现细节

## 已实现 Addon

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

## 接入原则

模板默认不依赖 `addons`。

只有在某个项目真的需要这些能力时，才在项目边界接入，并由项目自己管理配置、生命周期和业务封装。

## 说明

- 当前仓库里长期维护的 addon 是 `mail/`、`mongodb/`、`s3/`。
- 原先偏向 MinIO 的方向已经收敛到更通用的 `s3/` addon。
- 如果后续继续增加 addon，也应保持“可选、单一职责、易复制”的风格。
