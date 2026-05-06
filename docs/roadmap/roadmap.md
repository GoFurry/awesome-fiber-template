# 路线图

## v0.1.0

`v0.1.0` 是当前 `fiberx` 生成器主线的第一个可发布版本，已经完成。

已交付内容：

- 四个官方 preset：
  - `heavy`
  - `medium`
  - `light`
  - `extra-light`
- `redis`、`swagger`、`embedded-ui` 的稳定 capability contract
- `medium / heavy / light` 的运行时选项：
  - logger
  - database
  - data access
- 生成项目 metadata、diff 检查、只读 upgrade 检查
- 项目级构建工程化：
  - profiles
  - packaging
  - checksums
  - hooks
  - UPX
  - build metadata
  - release manifest
- 仓库已经收敛为纯 generator 主线

## v0.1.1

`v0.1.1` 已完成。

已交付内容：

- Fiber v3 生命周期 hooks 预留区：
  - `OnPreStartupMessage`
  - `OnPostStartupMessage`
  - `OnPreShutdown`
  - `OnPostShutdown`
- `app.Hooks()` 已接入全部 Fiber v3 preset 的默认生成骨架
- graceful shutdown 已作为默认模板能力接入全部 preset，并覆盖：
  - `Fiber v2`
  - `Fiber v3`
- 默认中间件组合已补齐到 `medium / heavy / light`：
  - `recover`
  - `request id`
  - `logger`
  - `cors`
- `extra-light` 保持最小中间件面，不补：
  - `request id`
  - `logger`
  - `cors`
- 可选 JSON backend 已接入生成参数、模板和 recipe 链路：
  - 参数：`--json-lib`
  - 首轮支持：
    - `stdlib`
    - `sonic`
    - `go-json`
  - 已进入：
    - `new / init`
    - generated metadata
    - `inspect / diff / upgrade` recipe 输出
- build hook 的信任边界提示已进入文档：
  - `fiberx build` 可能执行项目自定义 hooks
  - 只应在信任的仓库中运行
  - 可先使用 `--dry-run` 检查将执行的命令

版本边界：

- Fiber v3 hooks 只生成在 `Fiber v3`
- `Fiber v2` 不补同一套 lifecycle hook skeleton
- JSON backend 保持 optional opt-in，不改变默认标准库 JSON 行为
- build hook 安全边界当前只停留在文档层，不改变现有 CLI 执行策略

参考资料：

- [Fiber v3 Hooks](https://docs.gofiber.io/api/hooks/)
- [Fiber v3 Make Fiber Faster](https://docs.gofiber.io/guide/faster-fiber/)
- [Fiber v2 Make Fiber Faster](https://docs.gofiber.io/v2.x/guide/faster-fiber/)

## v0.1.2

`v0.1.2` 是下一个计划版本，聚焦默认骨架易用性和公共工具层收口。

计划吸收 `sample/` 中的这些内容：

- `pkg/common/constant.go`
  - time formats
  - response status flags
  - common HTTP headers
- `pkg/common/error.go`
- `transport/http/timeout_router.go`

目标方向：

- 提升默认骨架的公共常量、错误表达和超时路由组织能力
- 保持主线 generator 简洁，不把这些内容提前混入 `v0.1.1`
