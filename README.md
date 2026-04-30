# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[中文说明](./README_zh.md)

`fiberx` is a CLI-first Fiber project generator repository.

The repository is now intentionally focused on the generator itself: generator assets, planning rules, validation, rendering, build automation, and regression coverage. It no longer treats legacy reference templates or repository-local addon pools as part of the maintained mainline.

## Docs

- [Docs index](./docs/README.md)
- [Usage guide](./docs/guides/usage.md)
- [Generator architecture](./docs/architecture/fiberx-generator-architecture.md)
- [Template boundaries](./docs/architecture/template-boundaries.md)
- [Repository rules](./docs/architecture/repository-rules.md)
- [Template selection guide](./docs/guides/template-selection.md)
- [Roadmap](./docs/roadmap/roadmap.md)

## Current Generator Tracks

- `medium`: stable production baseline with Swagger and embedded UI by default
- `heavy`: completed second production track with Swagger, embedded UI, metrics, scheduler jobs, and optional Redis
- `light`: mature lightweight HTTP service with SQLite-first CRUD, common middleware, and opt-in Swagger or embedded UI
- `extra-light`: minimal startable base with SQLite startup, health endpoints, and recover-only middleware
- default stack: `Fiber v3 + Cobra + Viper`
- default runtime on `medium / heavy / light`: `zap + sqlite + stdlib`
- compatibility stack: `Fiber v2 + native-cli`
- Phase 11 runtime options on `medium / heavy / light`: `--logger zap|slog`, `--db sqlite|pgsql|mysql`, `--data-access stdlib|sqlx|sqlc`
- generated projects include config profiles, runtime metadata, upgrade inspection, and project-level build automation

## How To Choose

- Choose `heavy` if you want the stronger ops-oriented production track with metrics and scheduler defaults.
- Choose `medium` if you want the stable production-oriented HTTP baseline without scheduler and metrics defaults.
- Choose `light` if you want a smaller but still directly usable HTTP service with CRUD and common middleware.
- Choose `extra-light` if you want the smallest clean starting point with only startup and health basics.

## Quick Start

Generate a runnable project directly from the repository root:

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go run . serve
```

Compatibility example:

```bash
go run ./cmd/fiberx new demo-legacy --preset medium --fiber-version v2 --cli-style native
```

Runtime options example:

```bash
go run ./cmd/fiberx new demo-data --preset medium --logger slog --db pgsql --data-access sqlx
```

Build automation example:

```bash
go run ./cmd/fiberx build
go run ./cmd/fiberx build --dry-run
go run ./cmd/fiberx build --profile prod
```

## Repository Goal

This repository exists to keep `fiberx` itself clean and maintainable as a long-lived generator system:

- stable official preset semantics
- generator-owned assets and rules
- verifiable output
- explicit runtime and capability policy
- project-level build, metadata, and upgrade tooling

## Notes

- The generator mainline is the only maintained source of truth in this repository.
- Historical content is preserved through Git history rather than repository-local legacy directories.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
