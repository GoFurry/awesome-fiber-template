# Usage Guide

This guide explains how to use the current `fiberx` generator from the repository root.

## Current Reality

Today, `fiberx` is a repository-local generator with four generatable presets:

- `heavy`
- `medium`
- `light`
- `extra-light`

Current implemented capabilities:

- `redis`
- `swagger`
- `embedded-ui`

Current stable production baseline:

- `medium`

Current completed production track:

- `heavy`

Current lightweight product lines:

- `light`
- `extra-light`

Current default stack:

- `fiber-v3 + cobra + viper`

Current default runtime options for `medium / heavy / light`:

- logger: `zap`
- database: `sqlite`
- data access: `stdlib`

Important boundary:

- `medium` remains the stable production baseline.
- `heavy` is the completed Phase 7 production track with built-in Swagger, embedded UI, metrics, and jobs defaults.
- `light` and `extra-light` remain the completed lightweight product lines introduced in Phase 8.
- Phase 10 finished the capability contract consolidation.
- Phase 11 finished runtime-option expansion for `medium / heavy / light`.
- Phase 13 completed generated metadata and diff detection.
- Phase 14 completed readonly upgrade planning and compatibility policy.
- Phase 15 is implementing the `fiberx build` P0 baseline.

## Run From Source

From the repository root:

```bash
go run ./cmd/fiberx --help
```

Available commands:

- `fiberx new <name>`
- `fiberx init`
- `fiberx list presets`
- `fiberx list capabilities`
- `fiberx explain preset <name>`
- `fiberx explain capability <name>`
- `fiberx inspect [path]`
- `fiberx diff [path]`
- `fiberx upgrade inspect [path]`
- `fiberx upgrade plan [path]`
- `fiberx build [target...]`
- `fiberx validate`
- `fiberx doctor`

Equivalent source form:

```bash
go run ./cmd/fiberx <command>
```

## Create A New Project

Generate a new project into `<cwd>/<projectName>`:

```bash
go run ./cmd/fiberx new demo --preset medium
```

Examples:

```bash
go run ./cmd/fiberx new demo --preset light
go run ./cmd/fiberx new demo --preset extra-light
go run ./cmd/fiberx new demo --preset medium --with redis
go run ./cmd/fiberx new demo --preset medium --fiber-version v2 --cli-style native
go run ./cmd/fiberx new demo --preset medium --logger slog --db pgsql --data-access sqlx
go run ./cmd/fiberx new demo --preset light --db mysql --data-access sqlc
```

If `--module` is omitted, `fiberx` falls back to:

```text
github.com/example/<project-name>
```

Explicit module example:

```bash
go run ./cmd/fiberx new demo --preset medium --module github.com/your-org/demo
```

## Initialize In The Current Directory

Generate into the current working directory:

```bash
go run ./cmd/fiberx init --preset light
```

With an explicit project name:

```bash
go run ./cmd/fiberx init --name demo --preset medium
```

## Inspect Presets And Capabilities

List presets:

```bash
go run ./cmd/fiberx list presets
```

List capabilities:

```bash
go run ./cmd/fiberx list capabilities
```

Explain a preset:

```bash
go run ./cmd/fiberx explain preset medium
```

Explain a capability:

```bash
go run ./cmd/fiberx explain capability redis
```

## Validate The Generator

Validate manifests and assets:

```bash
go run ./cmd/fiberx validate
```

Show the current generator status:

```bash
go run ./cmd/fiberx doctor
```

Use these before cutting releases or after changing manifests, packs, capabilities, or rules.

## Inspect Generated Metadata

Every generated project now includes:

```text
.fiberx/manifest.json
```

Read the recorded metadata:

```bash
go run ./cmd/fiberx inspect ./demo
go run ./cmd/fiberx inspect ./demo --json
```

This reports:

- generator version and commit
- preset and capability recipe
- selected asset sets
- template and rendered-output fingerprints
- managed file count

## Diff A Generated Project

Compare a generated project against the current generator:

```bash
go run ./cmd/fiberx diff ./demo
go run ./cmd/fiberx diff ./demo --json
```

Current diff statuses:

- `clean`
- `local_modified`
- `generator_drift`
- `local_and_generator_drift`

Current diff scope:

- only generator-managed files are compared
- user-added files are ignored
- no patch or write-back is performed

## Plan A Generator Upgrade

Phase 14 adds readonly upgrade-planning commands on top of metadata and diff:

```bash
go run ./cmd/fiberx upgrade inspect ./demo
go run ./cmd/fiberx upgrade inspect ./demo --json
go run ./cmd/fiberx upgrade plan ./demo
go run ./cmd/fiberx upgrade plan ./demo --json
```

Current upgrade compatibility levels:

- `compatible`
- `manual_review`
- `breaking`

Current upgrade scope:

- evaluates whether the current generator can still work with the recorded project recipe
- classifies compatibility based on metadata, version direction, and managed-file drift
- remains readonly; it does not rewrite project files

## Build Generated Projects

Phase 15 P0 adds a project-level build command:

```bash
go run ./cmd/fiberx build
go run ./cmd/fiberx build server
go run ./cmd/fiberx build --clean
go run ./cmd/fiberx build --target linux/amd64
```

Current build input:

- project root `fiberx.yaml`
- one or more named targets
- `build.version.source=git`

Current build output:

- binaries under `dist/<target>/<goos>_<goarch>/`
- `-trimpath` and configured `ldflags`
- no archive/checksum/dry-run/profile support yet

## Stack Selection

By default, generated projects use:

- `Fiber v3`
- `Cobra`
- `Viper`

By default, `medium / heavy / light` also use:

- `zap`
- `sqlite`
- `database/sql + handwritten SQL`

Compatibility mode is still supported for all four presets:

```bash
go run ./cmd/fiberx new demo --preset medium --fiber-version v2 --cli-style native
```

Phase 11 runtime options are available on `medium`, `heavy`, and `light`:

```bash
go run ./cmd/fiberx new demo --preset medium --logger zap --db sqlite --data-access stdlib
go run ./cmd/fiberx new demo --preset medium --logger slog --db pgsql --data-access sqlx
go run ./cmd/fiberx new demo --preset light --db mysql --data-access sqlc
```

`extra-light` intentionally rejects `--logger`, `--db`, and `--data-access` in this phase.

## Recommended Starting Points

Use `medium` if:

- you want the most complete currently verified service baseline
- you want built-in Swagger and embedded UI support
- you want a service that can already be generated and run with a practical HTTP stack

Use `light` if:

- you want a slimmer but still directly usable HTTP service
- you want SQLite-first CRUD plus common middleware, without docs/UI by default
- you may want to opt into Swagger or embedded UI later

Use `extra-light` if:

- you want the smallest clean base
- you only need startup, SQLite, health checks, and recover
- you want to add almost everything else explicitly

Use `heavy` if:

- you want the broader ops-oriented production track
- you want built-in metrics and a local scheduler job loop
- you want Swagger and embedded UI included by default

## Capability Rules

Current capability notes:

- `medium` includes `swagger` and `embedded-ui` by default
- `heavy` includes `swagger` and `embedded-ui` by default
- `medium` allows `redis`
- `heavy` currently allows `redis`
- `light` allows opt-in `swagger` and `embedded-ui`
- `light` and `extra-light` do not accept `redis`
- `extra-light` does not currently accept `swagger` or `embedded-ui`
- `embedded-ui` is independent from `swagger`; you can enable UI without docs on `light`

## Capability Policy

Use capability selection as a boundary decision, not as a preset substitute.

- `swagger`: default on `heavy` and `medium`, opt-in on `light`, unsupported on `extra-light`
- `embedded-ui`: default on `heavy` and `medium`, opt-in on `light`, unsupported on `extra-light`
- `redis`: opt-in on `heavy` and `medium` only, unsupported on `light` and `extra-light`

Practical rule:

- choose the preset for service weight and runtime shape first
- then use capability flags to add or remove the supported optional surfaces for that preset

Example:

```bash
go run ./cmd/fiberx new demo --preset medium --with redis
```

```bash
go run ./cmd/fiberx new demo --preset light --with swagger
```

If you request an unsupported combination, generation should fail during validation.

## Run A Generated Project

Current best paths for runnable generated services:

```bash
go run ./cmd/fiberx new demo --preset medium
cd demo
go test ./...
go run . serve
```

```bash
go run ./cmd/fiberx new demo-heavy --preset heavy
cd demo-heavy
go test ./...
go run . serve
```

After startup, the generated `medium` service should expose:

- `/healthz`
- `/livez`
- `/readyz`
- `/startupz`
- `/docs`
- `/ui`

Generated `heavy` should additionally expose `/metrics` and report scheduler activity in its health payload.

Generated `light` should expose the same health routes and the demo CRUD API, but not `/docs` or `/ui` unless you explicitly request those capabilities.

Generated `extra-light` should expose only the health routes and minimal startup path, without CRUD, docs, or UI defaults.

If you generate with `--cli-style cobra`, the project command entry uses Cobra and Viper. If you generate with `--cli-style native`, it keeps the lightweight native command parser.

## Generated Project Guides

Every generated project now includes:

- `docs/runbook.md`
- `docs/configuration.md`
- `docs/api-contract.md`
- `docs/verification.md`

Every generated project also includes:

- `config/server.yaml`
- `config/server.dev.yaml`
- `config/server.prod.yaml`

## Built Binary Note

If you build the `fiberx` binary and execute it outside the repository root, set `FIBERX_MANIFEST_ROOT` so the binary can find `generator/`:

```bash
set FIBERX_MANIFEST_ROOT=D:\WorkSpace\Go\fiberx\generator
fiberx list presets
```

Without that environment variable, an out-of-tree binary may fail to locate manifests.

## Related Docs

- [Template Selection Guide](./template-selection.md)
- [Capability Policy Guide](./capability-policy.md)
- [Addon Integration Guide](./addon-integration.md)
- [Deployment Runbook Guide](./deployment-runbook.md)
- [Config Profiles Guide](./config-profiles.md)
- [Response Contract Guide](./response-contract.md)
- [Verification Matrix Guide](./verification-matrix.md)
- [Phase 7 Plan](../roadmap/phase-7-plan.md)
- [Phase 8 Plan](../roadmap/phase-8-plan.md)
- [Phase 9 Plan](../roadmap/phase-9-plan.md)
