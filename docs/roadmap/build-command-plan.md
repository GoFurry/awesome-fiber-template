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
fiberx build --profile prod
```

后续阶段再扩展：

```bash
fiberx build --all
```

## 当前状态

- 当前阶段：`Phase 15`
- 当前进度：`P0 completed`
- 当前推进：`P2 completed`
- 当前里程碑：`P3-M1 active`
- 后续里程碑：`P3-M2 deferred`

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

## 已完成：P2

已交付：

- archive：`zip / tar.gz`
- checksums：`sha256`
- `--dry-run`
- 并发构建

已实现口径：

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
  - 当前只支持 `sha256`
- `--dry-run`
  - 只输出构建计划
  - 不执行 `go build`
  - 不写 archive
  - 不写 checksum
- `parallel=true`
  - 按 `target × platform` 并发执行
  - 最终输出顺序保持稳定

完成依据：

- 自动测试通过：
  - `go test ./...`
  - `buildconfig / build / cmd` 相关回归
- CLI 状态正常：
  - `validate`
  - `doctor`
- 手动冒烟已覆盖：
  - `--dry-run`
  - `archive`
  - `SHA256SUMS`
  - `parallel` 下输出顺序稳定
- archive 验证通过：
  - Linux => `.tar.gz`
  - Windows => `.zip`
  - archive 内包含二进制和附加文件
- checksum 验证通过：
  - `dist/SHA256SUMS`
  - 指向最终 distributable artifacts

## 推进中：P3-M1

当前范围固定为：

- `profiles`
- `build metadata`
- `release manifest`

当前公开接口：

- `fiberx build --profile <name>`
- `build.profiles`
- `dist/build-metadata.json`
- `dist/release-manifest.json`

定位：

- `profiles`
  - 在 `fiberx.yaml` 中支持按环境或场景切换构建参数
- `build metadata`
  - 输出单次构建上下文的元信息文件
- `release manifest`
  - 输出面向交付的制品清单

固定边界：

- profile 只是对 base `build` 的 overlay，不是第二套完整 build config
- profile 只能覆盖：
  - `out_dir`
  - `clean`
  - `parallel`
  - `defaults.cgo`
  - `defaults.trimpath`
  - `defaults.ldflags`
  - `checksum.enabled`
  - `checksum.algorithm`
  - 同名 target 的 `output / platforms / archive.*`
- profile 不允许：
  - 创建新 target
  - 改写 `project.*`
  - 改写 `version.source`
  - 改写 `version.package`

## 已定义、暂不推进：P3-M2

后续范围固定为：

- `pre/post hooks`
- `UPX`

定位：

- `pre/post hooks`
  - 在 build target 前后执行显式配置的脚本
- `UPX`
  - 保持显式 opt-in，不默认启用

当前明确不推进：

- 不新增 hooks 相关 CLI
- 不扩展 target 生命周期命令
- 不接入 `build.compress.upx` 的实际执行逻辑

## 边界说明

- `UPX` 保持显式 opt-in，不进入默认构建链路
- `build` 优先面向 Go 二进制产物，不扩展到镜像构建或远程发布
- `fiberx.yaml` 的构建段服务于项目级分发，不反向影响 preset/capability 模型
