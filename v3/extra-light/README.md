# Fiber v3 Extra-Light Template

[Chinese](./README_zh.md)

`extra-light` is the smallest edition in this scaffold family. It keeps only the minimum runtime pieces needed to start a Fiber service cleanly: native CLI, minimal config, SQLite bootstrap, logging, panic recovery, and health probes.

## What This Template Includes

- plain native CLI with `serve` and `version`
- minimal config file with only `server`, `database`, and `log`
- SQLite-only setup with automatic database file creation
- no built-in business demo
- empty `internal/app` directory ready for your own domains
- only two default middlewares:
  - `recover`
  - `healthcheck`
- optional embedded UI support kept in the repo, disabled by default

## Quick Start

Run the service:

```bash
go run . serve
```

Show the version:

```bash
go run . version
```

Default config file:

```bash
./config/server.yaml
```

## Default Endpoints

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`

The `/api/v1` route tree exists but is empty by default.

## Configuration

This version intentionally keeps the config surface small.

`server`

- `app_name`
- `mode`
- `ip_address`
- `port`

`database`

- `path`

`log`

- `log_level`
- `log_path`

## Directory Layout

- `cmd`: native CLI entrypoints
- `config`: YAML config
- `internal/app`: add your own business domains here
- `internal/bootstrap`: startup and health state
- `internal/db`: SQLite bootstrap
- `internal/http`: router and optional embedded UI
- `pkg/common`: minimal response and error helpers

## Embedded UI

Embedded UI support is kept, but it is not mounted by default.

If you want to use it, call `AttachEmbeddedUI(app)` inside `internal/http/router.go` after creating the Fiber app.

## What Is Intentionally Removed

- Cobra
- Redis
- scheduler
- Swagger
- WAF
- Prometheus
- CSRF
- Helmet
- pprof
- helper packages such as `pkg/httpkit` and `pkg/abstract`
- built-in CRUD demos

## How To Start Building

1. Add your domain under `internal/app/<domain>`.
2. Register routes in `internal/http/url.go`.
3. Add your own GORM models and pass them into bootstrap if you need auto-migration later.
4. Replace the module path in `go.mod`.

## Verification

Run:

```bash
go test ./...
go vet ./...
```
