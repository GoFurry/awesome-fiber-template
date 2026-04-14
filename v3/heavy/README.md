# Fiber v3 Heavy Template

[Chinese](./README_zh.md)

`heavy` is now the practical full-featured edition of this scaffold. It keeps the stronger infrastructure and middleware baseline, but avoids the old over-assembled module style: plain `controller`, `dao`, `service`, and `models`, with routes registered directly in `url.go`.

## What This Template Includes

- SQLite-first demo experience with no external database required
- Complete `user` CRUD example
- Plain Go-style business structure under `internal/app`
- Lifecycle bootstrap for DB, Redis, scheduler, logging, and graceful shutdown
- Reusable helper packages such as `pkg/httpkit` and `pkg/abstract` for common outbound HTTP and generic CRUD scenarios
- Official Fiber middleware baseline:
  request ID, access log, timeout, health probes, security headers, compression, ETag, and rate limiting
- Optional infrastructure switches for Redis, Prometheus, Swagger, WAF, scheduler, and embedded UI

## Recommended Use

Use `heavy` when you want the broadest built-in engineering baseline while keeping the business layer plain.

This version is meant for projects that expect richer infrastructure from the start and are willing to keep a wider dependency surface in exchange for convenience.

## Quick Start

Default config file:

```bash
./config/server.yaml
```

Start the service:

```bash
go run . serve
```

On first startup the template will automatically:

- create `./data/app.db`
- auto-migrate registered models when `database.auto_migrate` is enabled
- expose the built-in user demo endpoints

Show the current version:

```bash
go run . version
```

Install or uninstall the service through the service manager integration:

```bash
go run . install
go run . uninstall
```

## Default Endpoints

Health and runtime endpoints:

- `GET /healthz`
- `GET /livez`
- `GET /readyz`
- `GET /startupz`

User CRUD demo:

- `GET /api/v1/user/`
- `POST /api/v1/user/`
- `GET /api/v1/user/:id`
- `PUT /api/v1/user/:id`
- `DELETE /api/v1/user/:id`

Optional endpoints:

- `GET /csrf/token` when CSRF is enabled
- `GET /metrics` when Prometheus is enabled
- `GET /swagger` when Swagger is enabled in debug mode
- `GET /debug/pprof/...` in debug mode

## CRUD Demo

Create a user:

```bash
curl -X POST http://127.0.0.1:9999/api/v1/user/ \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "age": 24,
    "status": "active"
  }'
```

List users:

```bash
curl "http://127.0.0.1:9999/api/v1/user/?page_num=1&page_size=10&keyword=alice"
```

Update a user:

```bash
curl -X PUT http://127.0.0.1:9999/api/v1/user/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Updated",
    "email": "alice.updated@example.com",
    "age": 25,
    "status": "active"
  }'
```

Delete a user:

```bash
curl -X DELETE http://127.0.0.1:9999/api/v1/user/1
```

## Business Structure

The business layer stays intentionally plain:

- business code lives under `internal/app/<domain>`
- keep `controller`, `dao`, `service`, and `models` as normal packages
- route registration stays in `internal/transport/http/router/url.go`
- bootstrap registers runtime data such as database models and scheduled jobs directly

Current references:

- `internal/app/user/controller`
- `internal/app/user/dao`
- `internal/app/user/service`
- `internal/app/user/models`
- `internal/transport/http/router/url.go`
- `internal/bootstrap/lifecycle.go`

To add a new business domain:

1. Create `internal/app/<domain>`.
2. Add `controller`, `dao`, `service`, and `models` packages as needed.
3. Register that domain's routes in `internal/transport/http/router/url.go`.
4. Register its database models or jobs directly in `internal/bootstrap/lifecycle.go` if needed.

## Auto Migration

This template keeps `database.auto_migrate` because it is useful for the SQLite out-of-the-box experience.

That is the only schema bootstrap behavior kept in this `heavy` version.

- no explicit migration command
- no migration directory requirement
- no migration tracking table

If you later need stricter schema management, that can be added in another variant without making the default workflow heavier.

For long-term repository rules and template boundaries, see the docs under the repository root:

- `docs/architecture/template-boundaries.md`
- `docs/architecture/repository-rules.md`

## Middleware Baseline

Enabled by default:

- Request ID
- Access log
- Recover
- CORS
- Security headers
- Compression
- ETag
- Rate limiter
- Health probes

Disabled by default but available:

- CSRF
- Swagger
- Prometheus
- Redis
- WAF
- Scheduler

## Configuration Overview

Main config file:

```bash
./config/server.yaml
```

Important sections:

- `server`
- `database`
- `redis`
- `prometheus`
- `log`
- `middleware`
- `waf`
- `schedule`

The `redis` section supports address, username, password, database index, and pool size.

The default config is intentionally runnable with only Go installed.

## Directory Layout

- `cmd`: CLI commands such as `serve`, `install`, `uninstall`, and `version`
- `config`: configuration files
- `internal/app`: business domains
- `internal/bootstrap`: lifecycle, startup state, and health probes
- `internal/infra`: DB, logging, metrics, cache, and scheduler infrastructure
- `internal/jobs`: scheduled jobs
- `internal/transport`: HTTP router, middleware, and embedded UI
- `pkg`: shared abstractions and utilities

`pkg/httpkit` and `pkg/abstract` are intentionally kept in `heavy` as reusable building blocks, even though the default demo module does not depend on them directly.

## Testing

Run the centralized test suites from `v3/test`:

```bash
cd ../test
go test ./...
```

Current integration coverage includes:

- service bootstrap with SQLite
- automatic database creation
- automatic table creation through `auto_migrate`
- health probes
- request ID and security headers
- ETag and compression behavior
- end-to-end user CRUD flow

## Known Tradeoffs

- Route registration is intentionally centralized in `internal/transport/http/router/url.go`.
- Runtime model and job registration is still centralized in `internal/bootstrap/lifecycle.go`.
- Request ID is already present in access logs, but business logs are not yet automatically enriched with request context.

## Template Checklist

Before turning this into your own service:

- replace the module path in `go.mod`
- update app identity in `config/server.yaml`
- remove the demo user domain if you no longer need it
- add your own business domains under `internal/app`
