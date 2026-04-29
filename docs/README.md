# Docs

This directory contains the long-term rules, architecture notes, and implementation roadmap for `fiberx`.

## Contents

- [`architecture/fiberx-generator-architecture.md`](./architecture/fiberx-generator-architecture.md)
- [`architecture/template-boundaries.md`](./architecture/template-boundaries.md)
- [`architecture/addon-design-rules.md`](./architecture/addon-design-rules.md)
- [`architecture/repository-rules.md`](./architecture/repository-rules.md)
- [`guides/usage.md`](./guides/usage.md)
- [`guides/template-selection.md`](./guides/template-selection.md)
- [`guides/capability-policy.md`](./guides/capability-policy.md)
- [`guides/addon-integration.md`](./guides/addon-integration.md)
- [`guides/deployment-runbook.md`](./guides/deployment-runbook.md)
- [`guides/config-profiles.md`](./guides/config-profiles.md)
- [`guides/response-contract.md`](./guides/response-contract.md)
- [`guides/verification-matrix.md`](./guides/verification-matrix.md)
- [`guides/generated-project-metadata.md`](./guides/generated-project-metadata.md)
- [`roadmap/roadmap.md`](./roadmap/roadmap.md)
- [`roadmap/phase-7-plan.md`](./roadmap/phase-7-plan.md)
- [`roadmap/phase-8-plan.md`](./roadmap/phase-8-plan.md)
- [`roadmap/phase-9-plan.md`](./roadmap/phase-9-plan.md)
- [`roadmap/build-command-plan.md`](./roadmap/build-command-plan.md)

Use the root README for a quick repository overview, then start from the generator architecture document when making structural decisions.

## Current Support Matrix

- Generatable presets: `heavy`, `medium`, `light`, `extra-light`
- Implemented capabilities: `redis`, `swagger`, `embedded-ui`
- Stable production baseline: `medium`
- Completed production track: `heavy`
- Current stage: `Phase 14` upgrade planning and compatibility policy
- Phase 12 delivery: capability matrix and validation closure completed
- Phase 13 delivery: generated metadata and diff detection completed
- Phase 14 focus: readonly upgrade planning and compatibility policy
- Phase 14 delivery target: readonly upgrade planning and compatibility classification
- Default medium experience: `swagger`, `embedded-ui`
- Default heavy experience: `swagger`, `embedded-ui`
- Light optional experience: `swagger`, `embedded-ui`
- Extra-light optional experience: `none`
- Capability policy:
  - `swagger`: default on `heavy,medium`, optional on `light`
  - `embedded-ui`: default on `heavy,medium`, optional on `light`
  - `redis`: optional on `heavy,medium` only
- Default stack: `fiber-v3 + cobra + viper`
- Default logger: `zap`
- Default database: `sqlite`
- Default data access: `stdlib`
- Supported fiber versions: `v3`, `v2`
- Supported CLI styles: `cobra`, `native`
- Supported loggers: `zap`, `slog`
- Supported databases: `sqlite`, `pgsql`, `mysql`
- Supported data access stacks: `stdlib`, `sqlx`, `sqlc`
- Generated project metadata: `.fiberx/manifest.json`
- New inspection commands: `fiberx inspect`, `fiberx diff`
- New upgrade commands: `fiberx upgrade inspect`, `fiberx upgrade plan`

## Current Roadmap Position

- Completed state: `State 1`
- Current stage: `State 4 / Phase 14`
- Current progress: Phase 11 completed; Phase 12 completed; Phase 13 completed; Phase 14 active
- External-db runtime progress: the full default-stack runtime matrix passed CI on commit `1a46f0c`
- Capability-matrix progress: Phase 12 full matrix and black-box closure passed local regression
- Metadata progress: generated-project metadata and diff detection are completed
- Upgrade-planning progress: readonly upgrade assessment and compatibility policy are now active implementation work
- `fiberx build` is tracked later under `State 4 / Phase 15`
