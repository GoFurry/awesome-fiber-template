# fiberx Build Command Roadmap

这份文档记录 `fiberx build` 方向的需求收口，用来承接脚手架生成之后的构建、打包与分发能力。

目标不是把 `fiberx` 做成完整 CI 平台，而是让它具备“读取项目级构建描述并稳定产出可分发制品”的工程化能力。

## 目标定位

`fiberx build` 应该覆盖以下使用形态：

```bash
fiberx build
fiberx build server
fiberx build server worker
fiberx build --profile prod
fiberx build --target linux/amd64
fiberx build --all
fiberx build --clean
fiberx build --dry-run
```

## 核心能力

### 构建输入

- 读取项目根目录 `fiberx.yaml`
- 根据配置识别多个 Go 入口
- 支持目标过滤、profile、全量构建与 dry-run

### 构建输出

- 自动输出到 `dist/`
- 支持 `GOOS / GOARCH` 矩阵构建
- 支持 `-ldflags` 注入版本、commit、构建时间
- 支持 `-trimpath`
- 支持 `-s -w`
- 支持 `zip / tar.gz` 打包
- 支持生成 `sha256` checksums
- 可选支持 `UPX`，但默认不启用

## 配置草案

```yaml
project:
  name: fiberx-demo
  module: github.com/GoFurry/fiberx-demo

build:
  out_dir: dist
  clean: true
  parallel: true

  version:
    source: git
    package: github.com/GoFurry/fiberx-demo/internal/version

  defaults:
    cgo: false
    trimpath: true
    ldflags:
      - "-s -w"
      - "-X {{.VersionPackage}}.Version={{.Version}}"
      - "-X {{.VersionPackage}}.Commit={{.Commit}}"
      - "-X {{.VersionPackage}}.BuildTime={{.BuildTime}}"

  targets:
    - name: server
      package: ./cmd/server
      output: fiberx-demo
      platforms:
        - linux/amd64
        - linux/arm64
        - windows/amd64
      archive:
        enabled: true
        format: auto
        files:
          - README.md
          - config.example.yaml

    - name: worker
      package: ./cmd/worker
      output: fiberx-worker
      platforms:
        - linux/amd64
      archive:
        enabled: true

  checksum:
    enabled: true
    algorithm: sha256

  compress:
    upx:
      enabled: false
      level: 5
```

## 分层里程碑

### P0

- `fiberx build`
- `fiberx build <target>`
- 读取 `fiberx.yaml`
- 支持多个 `cmd` target
- 支持 `out_dir`
- 支持 `clean`
- 支持 `ldflags`
- 支持 `GOOS / GOARCH`

### P2

- 自动生成 `dist` 目录结构
- 支持 `zip / tar.gz`
- 支持 `sha256 checksums`
- 支持 `dry-run`
- 支持并发构建

### P3

- 支持 `profiles`
- 支持 `pre/post hooks`
- 支持 `UPX`
- 支持 `build metadata`
- 支持生成 `release manifest`

## 边界说明

- `UPX` 保持显式 opt-in，不进入默认构建链路
- `build` 首先面向 Go 二进制产物，不扩展到镜像构建或远程发布
- `fiberx.yaml` 的构建段应当服务于项目级分发，不反向影响 preset/capability 模型
