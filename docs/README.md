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
- Current stage: `Phase 15` build and post-generation engineering
- Phase 15 P0: completed
- Phase 15 P2: archive, checksum, dry-run, and parallel build execution completed
- Phase 15 P3: active
- Phase 15 P3 milestone: profiles, build metadata, and release manifest
- Phase 15 deferred P3 items: hooks, upx
- Phase 12 delivery: capability matrix and validation closure completed
- Phase 13 delivery: generated metadata and diff detection completed
- Phase 14 delivery: readonly upgrade planning and compatibility classification completed
- Phase 15 focus: build and post-generation engineering
- Phase 15 delivery target: profiles, hooks, compression, build metadata, and release manifests
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
- Build command baseline: `fiberx build`

## Current Roadmap Position

- Completed state: `State 1`
- Current stage: `State 4 / Phase 15`
- Current progress: Phase 11 completed; Phase 12 completed; Phase 13 completed; Phase 14 completed; Phase 15 active
- External-db runtime progress: the full default-stack runtime matrix passed CI on commit `1a46f0c`
- Capability-matrix progress: Phase 12 full matrix and black-box closure passed local regression
- Metadata progress: generated-project metadata and diff detection are completed
- Upgrade-planning progress: readonly upgrade assessment and compatibility policy are completed
- Build-command progress: Phase 15 P0 and P2 are complete; Phase 15 P3-M1 is in progress for profiles and release-oriented metadata outputs
