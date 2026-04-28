# Phase 8 Plan

`Phase 8` 的目标是把 `light` 与 `extra-light` 从“可生成 preset”收敛为两条语义清晰、可运行、可验证的轻量产品线。

## 目标定位

- `light`：成熟的轻量 HTTP 服务起点
- `extra-light`：最小可启动基础底座

## 交付范围

### 1. 状态口径对齐

- `Phase 7` 标记为已完成
- 当前阶段切换为 `State 2 / Phase 8`
- `doctor` / `validate` / README / usage / template selection 同步更新

### 2. `light` 产品化

- 提供真实可运行的 `serve` 主链路
- 默认具备：
  - sqlite 启动
  - `/healthz`、`/livez`、`/readyz`、`/startupz`
  - request id
  - recovery
  - 安全头
  - gzip
  - ETag
  - `user` CRUD 示例
- 不默认包含：
  - `swagger`
  - `embedded-ui`
  - `redis`
  - metrics / jobs
- 允许显式装配：
  - `swagger`
  - `embedded-ui`

### 3. `extra-light` 产品化

- 提供真实可运行的 `serve` 主链路
- 默认只保留：
  - sqlite 启动
  - `/healthz`、`/livez`、`/readyz`、`/startupz`
  - `recover`
- 不包含 CRUD、docs、UI、redis、metrics、jobs 与扩展中间件

### 4. 验证闭环

- `light` / `extra-light` 从文件断言升级为行为级黑盒验证
- 覆盖生成、编译、启动、健康检查、路由边界与能力组合约束
- 保持根仓库 `go test ./...` 为主回归入口
