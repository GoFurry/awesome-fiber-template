# Roadmap

这份路线图只保留 `fiberx` 生成器的当前状态、近期目标和后续优先级；更细的拆分放在各个 phase 计划文档中。

## 当前状态

- 当前阶段：`State 4 / Phase 15`
- 当前默认栈：`Fiber v3 + Cobra + Viper`
- 首轮服务 preset 默认运行时：`zap + sqlite + stdlib`
- 当前公开模型：`preset + capability + 少量生成参数`
- `Phase 11` 首轮覆盖：`medium`、`heavy`、`light`
- `extra-light` 继续保持最小化，暂不接入 `logger / db / data-access`

## 已完成摘要

- `State 1`：生成器主链路稳定，`medium` 成为第一条生产基线。  
- `State 2 / Phase 7`：`heavy` 成为第二条生产主线。详见 [phase-7-plan.md](./phase-7-plan.md)
- `State 2 / Phase 8`：`light / extra-light` 完成产品化定位。详见 [phase-8-plan.md](./phase-8-plan.md)
- `State 2 / Phase 9`：默认栈切换到 `Fiber v3 + Cobra + Viper`，并保留兼容回退。详见 [phase-9-plan.md](./phase-9-plan.md)
- `State 3 / Phase 10`：`swagger / embedded-ui / redis` 的 capability contract、CLI 输出、文档和校验边界完成收口。
- `State 3 / Phase 11`：`logger / db / data-access` 生成参数完成首轮接入；默认栈下的运行矩阵已在 CI 闭环。
- `State 3 / Phase 12`：完整 capability matrix 已被请求校验、生成级断言和黑盒回归锁住。

## State 4：生成后维护与工程化

### Phase 13：版本元信息与差异检测

当前状态：`completed`

已交付：

- `.fiberx/manifest.json`
- `fiberx inspect`
- `fiberx diff`
- `clean / local_modified / generator_drift / local_and_generator_drift` 四类只读差异判断

边界：

- 不自动修复
- 不自动迁移
- 不输出 patch

### Phase 14：升级助手与兼容策略

当前状态：`completed`

已交付：

- `fiberx upgrade inspect`
- `fiberx upgrade plan`
- `compatible / manual_review / breaking`
- 基于 metadata + diff + 版本方向的只读升级评估

边界：

- 不自动修改项目文件
- 不输出 patch
- 不支持直接变更 `preset / capability / runtime recipe`
- 不引入 `fiberx migrate`

### Phase 15：`fiberx build` 与生成后工程化

详见 [build-command-plan.md](./build-command-plan.md)

当前状态：`active`

当前阶段进度：

- `P0`：`completed`
- `P2`：`completed`
- `P3`：`active`
- `P3-M1`：`completed`
- `P3-M2`：`completed`

已完成的 `P0` 能力：

- `fiberx build`
- `fiberx build <target...>`
- `--clean`
- `--target <goos/goarch>`
- 读取项目根目录 `fiberx.yaml`
- 支持多 target、`out_dir`、`ldflags` 和 `GOOS / GOARCH`
- 默认生成项目补齐可直接使用的 `fiberx.yaml` 与最小 `internal/version`

已完成的 `P2` 能力：

- `zip / tar.gz` archive
- `sha256` checksums
- `--dry-run`
- 并发构建

`P2` 完成依据：

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

已完成的 `P3-M1` 能力：

- `build.profiles`
- `fiberx build --profile <name>`
- `dist/build-metadata.json`
- `dist/release-manifest.json`

已完成的 `P3-M2` 能力：

- `build.targets[].pre_hooks`
- `build.targets[].post_hooks`
- `build.compress.upx`

`P3-M2` 完成依据：

- 核心实现完成
- 自动回归完成
- 手动冒烟完成
- `--dry-run` 已验证：
  - 正确展示 hooks / UPX / metadata / release manifest 计划
  - 不写 `dist`
  - 不执行 hooks
- archive 已验证：
  - Linux => `.tar.gz`
  - Windows => `.zip`
  - archive 内包含二进制和附加文件
- hook 失败路径已验证
- UPX 缺失与启用后的成功路径已验证
- `build-metadata.json` / `release-manifest.json` 已验证反映 hooks / UPX 结果

当前对 `Phase 15` 的判断：

- 构建工程化主线已经进入可收口状态
- 剩余工作主要是提交收口与决定是否进入新的工程化阶段

## 暂不进入

- GUI
- AST-heavy 改写
- 第五类官方 preset
- 远程模板源或模板市场
- 另一套与 generator 主线并行维护的仓库内历史体系
