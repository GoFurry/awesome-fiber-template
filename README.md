# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[中文说明](./README_zh.md)

`fiberx` is a CLI-first Fiber project generator repository.

`fiberx` evolved from the earlier `awesome-fiber-template` repository and is the formalized home for the project going forward.

Instead of trying to cover many frameworks, this repository now stays focused on Fiber and is evolving toward a generator-centered workflow while preserving four stable official starting points.

It currently preserves `v3/*` reference templates for the existing engineering baselines, and also includes an `addons/` directory as an independent optional capability pool for reusable service wrappers and utility packages.

## Docs

- [Docs index](./docs/README.md)
- [Usage guide](./docs/guides/usage.md)
- [Template boundaries](./docs/architecture/template-boundaries.md)
- [Addon design rules](./docs/architecture/addon-design-rules.md)
- [Template selection guide](./docs/guides/template-selection.md)
- [Addon integration guide](./docs/guides/addon-integration.md)
- [Roadmap archive](./docs/roadmap/roadmap.md)

## Current Reference Presets

- [`v3/heavy`](./v3/heavy): full-featured edition with Redis, scheduler, service install/uninstall, WAF, Prometheus, Swagger, reusable helper packages, and a stronger infrastructure baseline
- [`v3/medium`](./v3/medium): balanced HTTP service edition that keeps Redis, WAF, service manager support, embedded UI, and most middleware, but removes scheduler and Prometheus complexity
- [`v3/light`](./v3/light): plain Go-style service edition that keeps the common API middleware baseline and optional embedded UI, while removing Redis, service manager support, and extra helper packages
- [`v3/extra-light`](./v3/extra-light): minimal edition with native CLI, SQLite-only setup, no built-in business demo, and only `recover + healthcheck`

## Current Generator Tracks

- `medium`: stable production baseline with Swagger and embedded UI by default
- `heavy`: completed second production track with Swagger, embedded UI, metrics, scheduler jobs, and optional Redis
- `light`: mature lightweight HTTP service with SQLite-first CRUD, common middleware, and opt-in Swagger or embedded UI
- `extra-light`: minimal startable base with SQLite startup, health endpoints, and recover-only middleware
- default stack: `Fiber v3 + Cobra + Viper`
- default runtime on `medium / heavy / light`: `zap + sqlite + stdlib`
- compatibility stack: `Fiber v2 + native-cli`
- Phase 11 runtime options on `medium / heavy / light`: `--logger zap|slog`, `--db sqlite|pgsql|mysql`, `--data-access stdlib|sqlx|sqlc`
- generated projects now include `server.yaml`, `server.dev.yaml`, `server.prod.yaml`, plus runbook and verification docs

## How To Choose

- Choose `heavy` if you want the stronger ops-oriented production track with metrics and scheduler defaults.
- Choose `medium` if you want the stable production-oriented HTTP baseline without scheduler and metrics defaults.
- Choose `light` if you want a smaller but still directly usable HTTP service with CRUD and common middleware.
- Choose `extra-light` if you want the smallest clean starting point with only startup and health basics.

## Quick Start

You can generate a runnable service directly from the repository root. For example:

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go run . serve
```

Compatibility example:

```bash
go run ./cmd/fiberx new demo-legacy --preset medium --fiber-version v2 --cli-style native
```

Phase 11 runtime example:

```bash
go run ./cmd/fiberx new demo-data --preset medium --logger slog --db pgsql --data-access sqlx
```

The historical `v3/*` reference presets are still available and runnable directly. For example:

```bash
cd v3/light
go run . serve
```

Each reference preset is a standalone Go module with its own `go.mod`, config, README, and dependency boundary.

## Repository Goal

This repository is meant to turn the current preset semantics and repository rules into a generator-ready system while keeping the boundaries clear:

- stable official presets
- generator-owned rules and assets
- practical middleware and bootstrap defaults
- verifiable output and regression-friendly structure
- independent addon packages for reusable external services and utilities

## Notes

- `fiberx` is now positioned as a generator repository, while `v3/*` remains available as reference preset snapshots.
- If you use one of the current reference presets in your own project, replace the module path inside that preset's `go.mod`.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
