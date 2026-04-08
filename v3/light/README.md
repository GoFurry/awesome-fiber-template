# Fiber v3 Light Template

[Chinese](./README_zh.md)

`light` is the plain Go-style edition of this scaffold. It keeps the minimum engineering baseline needed for a practical API service, while dropping extra infrastructure, helper packages, and heavier middleware layers.

## What This Template Includes

- SQLite-first demo experience with no external database required
- Complete `user` CRUD example
- Plain Go-style business structure under `internal/app`
- Lifecycle bootstrap for DB, logging, and graceful shutdown
- Optional embedded UI support, disabled by default
- Official Fiber middleware baseline:
  request ID, access log, recover, health probes, route-level timeout, CORS, compression, ETag, and rate limiting

## Project Positioning

This is the current `light` version.

- Compared with `medium`, it removes Redis, service manager support, helper packages, WAF, Swagger, CSRF, Helmet, and pprof.
- Compared with `extra-light`, it still keeps a practical API baseline, SQLite auto-migrate, and optional embedded UI support.
- It is intentionally close to a normal Go service structure: `controller`, `dao`, `service`, and `models`, with routes registered directly in `url.go`.

If you want a lightweight backend template that still feels complete enough for real work, this is the intended direction.

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
- bootstrap registers runtime data such as database models directly

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
4. Register its database models directly in `internal/bootstrap/lifecycle.go` if needed.

## Auto Migration

This template keeps `database.auto_migrate` because it is useful for the SQLite out-of-the-box experience.

That is the only schema bootstrap behavior kept in this `light` version.

- no explicit migration command
- no migration directory requirement
- no migration tracking table

## Middleware Baseline

Enabled by default:

- Request ID
- Access log
- Recover
- CORS
- Compression
- ETag
- Rate limiter
- Health probes
- Route-level timeout

Not included in `light`:

- Redis
- service install/uninstall
- WAF
- Swagger
- CSRF
- Helmet
- pprof
- `pkg/httpkit`
- `pkg/abstract`

## Configuration Overview

Main config file:

```bash
./config/server.yaml
```

Important sections:

- `server`
- `database`
- `log`
- `middleware`

The default config is intentionally runnable with only Go installed.

## Directory Layout

- `cmd`: CLI commands such as `serve` and `version`
- `config`: configuration files
- `internal/app`: business domains
- `internal/bootstrap`: lifecycle and health probes
- `internal/infra`: DB and logging infrastructure
- `internal/transport`: HTTP router and embedded UI
- `pkg`: shared response and utility helpers

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
- request ID
- ETag and compression behavior
- end-to-end user CRUD flow

## Known Tradeoffs

- Route registration is intentionally centralized in `internal/transport/http/router/url.go`.
- Runtime model registration is still centralized in `internal/bootstrap/lifecycle.go`.
- Embedded UI is kept, but disabled by default through `server.is_full_stack`.

## Template Checklist

Before turning this into your own service:

- replace the module path in `go.mod`
- update app identity in `config/server.yaml`
- remove the demo user domain if you no longer need it
- add your own business domains under `internal/app`
