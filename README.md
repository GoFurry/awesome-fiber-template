# fiberx

![License](https://img.shields.io/badge/License-MIT-6C757D?style=flat&color=3B82F6)
![Release](https://img.shields.io/github/v/release/GoFurry/fiberx?style=flat&color=blue)
![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go&logoColor=white)
[![Go Report Card](https://goreportcard.com/badge/github.com/GoFurry/fiberx)](https://goreportcard.com/report/github.com/GoFurry/fiberx)

![Weekend Project](https://img.shields.io/badge/weekend-project-8B5CF6?style=flat)
![Made with Love](https://img.shields.io/badge/made%20with-%E2%9D%A4-E11D48?style=flat&color=orange)

[中文说明](./README_zh.md)

`fiberx` is a CLI-first Fiber project generator repository.

The repository is intentionally focused on the generator mainline itself: assets, planning rules, validation, rendering, upgrade inspection, build automation, and regression coverage.

## Release

- `v0.1.0`: completed
- `v0.1.1`: completed
- `v0.1.2`: completed
- `v0.1.3`: planned

## Docs

- [Docs index](./docs/README.md)
- [Usage guide](./docs/guides/usage.md)
- [Release process](./docs/guides/release-process.md)
- [Build hook safety](./docs/guides/build-hook-safety.md)
- [Generator architecture](./docs/architecture/fiberx-generator-architecture.md)
- [Template boundaries](./docs/architecture/template-boundaries.md)
- [Repository rules](./docs/architecture/repository-rules.md)
- [Contributing](./CONTRIBUTING.md)
- [Changelog](./CHANGELOG.md)
- [Roadmap](./docs/roadmap/roadmap.md)

## Current Generator Tracks

- `medium`: stable production baseline with Swagger and embedded UI by default
- `heavy`: production-oriented track with Swagger, embedded UI, metrics, scheduler jobs, and optional Redis
- `light`: lightweight HTTP service with SQLite-first CRUD and optional Swagger or embedded UI
- `extra-light`: minimal startable base with SQLite startup, health endpoints, and recover-only middleware
- default stack: `Fiber v3 + Cobra + Viper`
- compatibility stack: `Fiber v2 + native-cli`
- runtime options on `medium / heavy / light`: `--logger`, `--db`, `--data-access`
- generated projects include config profiles, runtime metadata, upgrade inspection, and project-level build automation

## Quick Start

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

## Repository Layout

- `sample/`: reference snapshots and test-facing examples, not the maintained generator mainline
- `output/`: local scratch space for generated artifacts and local binaries; ignored by Git except for `.gitkeep`

## v0.1.2 Release Scope

`v0.1.2` closes the current scaffold-hardening pass:

- shared scaffold uplift for `light`, `medium`, and `heavy`
- common constants, base error model, and response compatibility helpers
- configurable timeout routing for business APIs
- release, contribution, and build-hook safety documentation

## v0.1.3 Preview

The next milestone focuses on CLI UX and safer build workflows:

- generation plan preview and dry-run-style creation feedback
- build safety switches such as `--no-hooks` and explicit confirmation flow
- layered `doctor` output for generator vs generated projects
- `explain matrix` for preset and capability visibility

## Build Hook Safety

- `fiberx build` may execute project-defined hooks.
- Only run hooks in trusted repositories.
- Use `fiberx build --dry-run` to inspect planned commands before execution.

## License

This project is licensed under the MIT License. See [LICENSE](./LICENSE) for details.
