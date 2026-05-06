# 路线图

## v0.1.2

`v0.1.2` 当前聚焦默认骨架易用性和公共工具层收口。

这一轮已经完成：

- `light / medium / heavy` 的 `pkg/common/constant.go`
- `light / medium / heavy` 的 `pkg/common/error.go`
- `pkg/common/response.go` 的双接口兼容收口
- `middleware.timeout` 默认配置接入
- `internal/transport/http/router/timeout_router.go`
- 业务路由默认经过 timeout wrapper，系统路由保持直连

这一轮保持不纳入：

- `extra-light`
- `sample` 中更大范围的公共包搬运
- `pkg/common/*` 之外的通用层重构
- timeout 之外的其它 middleware 对齐

后续可以继续评估 `sample/` 中剩余的公共骨架能力，但 `v0.1.2` 主线先保持“公共常量、基础错误模型、响应兼容层、超时路由”这一小闭环。
