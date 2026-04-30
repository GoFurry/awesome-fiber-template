# `fiberx build` 路线图

这份文档记录 `fiberx build` 的阶段化目标，用来承接脚手架生成之后的构建、打包与分发能力。

## 当前状态

- 当前阶段：`Phase 15`
- `P0`：`completed`
- `P2`：`completed`
- `P3-M1`：`completed`
- `P3-M2`：`active`

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

## 已完成：P2

已交付：

- archive：`zip / tar.gz`
- checksums：`sha256`
- `--dry-run`
- 并发构建

完成依据：

- 自动测试通过
- `validate / doctor` 状态正常
- 手动冒烟覆盖：
  - `--dry-run`
  - archive
  - `SHA256SUMS`
  - 并发输出顺序稳定

## 已完成：P3-M1

已交付：

- `build.profiles`
- `fiberx build --profile <name>`
- `dist/build-metadata.json`
- `dist/release-manifest.json`

固定边界：

- profile 只是 base `build` 的 overlay
- profile 不创建新 target
- profile 不改写 `project.*`
- profile 不改写 `version.source / version.package`

## 推进中：P3-M2

当前范围：

- `build.targets[].pre_hooks`
- `build.targets[].post_hooks`
- `build.compress.upx`

当前进度判断：

- 核心实现：`completed`
- 自动回归：`completed`
- 手动冒烟：`completed`
- 提交收口：`pending`

当前规则：

- hooks 只支持 target 层
- hooks 使用 argv 数组，不通过 shell
- 任一 hook 非 0 退出即整体失败
- `--dry-run` 只展示 hooks / UPX 计划，不执行
- UPX 只做显式 opt-in，启用后找不到 `upx` 直接失败

执行顺序固定为：

1. `pre_hooks`
2. `go build`
3. 可选 `UPX`
4. `post_hooks`
5. 可选 archive
6. 可选 checksum
7. `build-metadata.json`
8. `release-manifest.json`

收口依据：

- `--dry-run` 已验证不会写 `dist`，且会展示 hooks / UPX / metadata / manifest 路径
- archive 已验证：
  - Linux 产出 `.tar.gz`
  - Windows 产出 `.zip`
- hook 失败路径已验证会中断构建
- UPX 缺失与启用后的成功路径都已验证
- metadata / release manifest 已验证包含 hooks / UPX 结果

收口前还缺什么：

- 提交当前工作区改动
- 再决定是否将整个 `Phase 15` 标记为 completed

## 当前明确不支持

- `build.pre_hooks`
- `build.post_hooks`
- `build.profiles.*.pre_hooks`
- `build.profiles.*.post_hooks`
- profile 级 UPX 覆盖
- shell-style hook 字符串
- 自动跳过缺失的 `upx`
