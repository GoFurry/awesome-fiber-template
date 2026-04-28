# Phase 9 Plan

`Phase 9` 已完成，核心是把默认生成栈切到 `Fiber v3 + Cobra + Viper`，同时把部署说明、分环境配置、返回约定和验证指南一起收口进生成结果与仓库文档。

## 目标

- 默认栈切换到 `Fiber v3 + Cobra + Viper`
- `heavy / medium / light / extra-light` 全部支持 `v2` 与 `native-cli` 兼容生成
- 让栈选择进入 CLI 参数、状态输出、文档与回归矩阵

## 已交付

- `new / init` 新增：
  - `--fiber-version v3|v2`
  - `--cli-style cobra|native`
- planner 按 `cli_style` 选择 base，按 `fiber_version` 选择 preset pack 变体
- 新增 `service-base-cobra`
- 为四个 preset 增加 `-v3` 变体 pack
- 默认生成、兼容生成、黑盒回归与生成后测试全部接入栈矩阵
- 生成项目默认附带：
  - `docs/runbook.md`
  - `docs/configuration.md`
  - `docs/api-contract.md`
  - `docs/verification.md`
- 四个 preset 都补齐 `config/server.dev.yaml` 与 `config/server.prod.yaml`
- `light` 默认配置改为真实反映能力边界，未显式启用时不再默认打开 `swagger` 与 `embedded-ui`
- 黑盒验证补到 404 JSON 错误包络、配置 profile 落盘与生成文档存在性

## 明确不在本阶段处理

- 更大的目录重组
- 配置系统全面重构
- `/v3/*` 参考工程的深度结构对齐
- 交互式向导 UI，本轮仍以 CLI 参数承接栈选择
