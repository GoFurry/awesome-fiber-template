# `fiberx build` 路线图

这份文档记录 `fiberx build` 的阶段化目标，用来承接脚手架生成之后的构建、打包与分发能力。

目标不是把 `fiberx` 做成完整 CI 平台，而是让它具备“读取项目级构建描述并稳定产出可分发制品”的工程化能力。

## 目标形态

`fiberx build` 应覆盖以下使用方式：

```bash
fiberx build
fiberx build server
fiberx build server worker
fiberx build --target linux/amd64
fiberx build --clean
fiberx build --dry-run
```

后续阶段再扩展：

```bash
fiberx build --profile prod
fiberx build --all
```

## 当前状态

- 当前阶段：`Phase 15`
- 当前进度：`P0 completed`
- 当前推进：`P2 active`

## 已完成：P0

已交付：

- `fiberx build`
- `fiberx build <target...>`
- `--clean`
- `--target <goos/goarch>`
- 读取项目根目录 `fiberx.yaml`
- 支持多个 Go 入口 target
- 支持 `out_dir`
- 支持 `ldflags`
- 支持 `GOOS / GOARCH`
- 默认生成项目补齐：
  - `fiberx.yaml`
  - `internal/version/version.go`

当前默认输出结构：

- `<out_dir>/<target-name>/<goos>_<goarch>/<binary>`
- Windows 自动追加 `.exe`

## 进行中：P2

当前目标：

- archive：`zip / tar.gz`
- checksums：`sha256`
- `--dry-run`
- 并发构建

当前实现口径：

- `build.parallel`
- `build.targets[].archive.enabled`
- `build.targets[].archive.format`
- `build.targets[].archive.files`
- `build.checksum.enabled`
- `build.checksum.algorithm`
- `fiberx build --dry-run`

固定规则：

- `archive.format`
  - 支持：`auto | zip | tar.gz`
  - `auto`：
    - `windows/*` => `zip`
    - 其他平台 => `tar.gz`
- `checksum.algorithm`
  - 目前只支持 `sha256`
- `--dry-run`
  - 只输出构建计划
  - 不执行 `go build`
  - 不写 archive
  - 不写 checksum
- `parallel=true`
  - 按 `target × platform` 并发执行
  - 最终输出顺序保持稳定

## 后续：P3

后续再进入：

- `profiles`
- `pre/post hooks`
- `UPX`
- `build metadata`
- `release manifest`

## 边界说明

- `UPX` 保持显式 opt-in，不进入默认构建链路
- `build` 优先面向 Go 二进制产物，不扩展到镜像构建或远程发布
- `fiberx.yaml` 的构建段服务于项目级分发，不反向影响 preset/capability 模型
