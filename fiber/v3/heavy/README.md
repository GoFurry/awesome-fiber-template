# Fiber v3 Heavy Template

[Chinese](./README_zh.md)

`heavy` is the full-featured edition of this scaffold. It targets medium and large Go backend services that need a clear module boundary, lifecycle management, reusable infrastructure, and a production-oriented HTTP baseline without giving up a runnable out-of-the-box demo.

## What This Template Includes

- SQLite-first demo experience. The default configuration uses SQLite, so the project can boot without installing MySQL or PostgreSQL.
- Complete `user` CRUD example. The built-in module covers routes, controller, service, DAO, model, migration, and integration tests.
- Unified module bundle model. A module can register routes, database models, migrations, scheduled jobs, startup hooks, shutdown hooks, and background services in one place.
- Lifecycle bootstrap. Startup, migration, infrastructure initialization, scheduled jobs, background services, and graceful shutdown are wired through the bootstrap layer.
- Official Fiber middleware baseline. Request ID, access log, timeout, health probes, security headers, compression, ETag, and advanced rate limiting are already integrated.
- Optional infrastructure switches. Redis, Prometheus, Swagger, WAF, scheduler, and the embedded UI can be enabled only when a service needs them.

## Project Positioning

This template is intentionally heavier than a minimal Go HTTP service.

- Good fit: medium services, internal platform baselines, and larger projects that want a consistent module and lifecycle structure.
- Not ideal: tiny services, throwaway demos, or teams that prefer a nearly standard-library style layout with very little infrastructure wiring.

If you want a lighter starting point later, this `heavy` version can act as the capability baseline.

## Quick Start

The default configuration file is:

```bash
./config/server.yaml
```

Start the HTTP service:

```bash
go run . serve
```

On first startup the template will automatically:

- create `./data/app.db`
- auto-migrate the registered models when `database.auto_migrate` is enabled
- run registered module migrations
- expose the built-in user demo endpoints

Run database migrations only:

```bash
go run . migrate up
```

Show the current service version:

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

- `GET /api/v1/users/`
- `POST /api/v1/users/`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

Optional endpoints:

- `GET /csrf/token` when CSRF protection is enabled
- `GET /metrics` when Prometheus is enabled
- `GET /swagger` when Swagger is enabled in debug mode
- `GET /debug/pprof/...` in debug mode

## CRUD Demo Payloads

Create a user:

```bash
curl -X POST http://127.0.0.1:9999/api/v1/users/ \
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
curl "http://127.0.0.1:9999/api/v1/users/?page_num=1&page_size=10&keyword=alice"
```

Update a user:

```bash
curl -X PUT http://127.0.0.1:9999/api/v1/users/1 \
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
curl -X DELETE http://127.0.0.1:9999/api/v1/users/1
```

The default user model uses the `users` table and the following fields:

- `id`
- `name`
- `email`
- `age`
- `status`
- `created_at`
- `updated_at`

## Middleware Baseline

The heavy template already wires a practical HTTP baseline around official Fiber middleware.

Enabled by default:

- Request ID
- Access log
- Recover
- CORS
- Security headers
- Compression
- ETag
- Rate limiter
- Legacy and official health probes

Disabled by default but ready to enable:

- CSRF
- Swagger
- Prometheus
- Redis
- WAF
- Scheduler

### Middleware Notes

- Request ID is exposed through the configured header, which defaults to `X-Request-ID`.
- Access log uses the official Fiber logger and includes the request ID in the default format.
- Timeout is applied through a timeout-aware router wrapper around `/api` routes. Health and metrics endpoints are excluded by default.
- ETag is enabled by default, so cache-aware clients can reuse conditional requests.
- Compression is enabled by default and will respond with compressed content when the client sends `Accept-Encoding`.
- The limiter supports multiple strategies and key sources through config.

## Health Probes

The template supports both official Fiber probe endpoints and a legacy JSON-style endpoint.

- `/livez`: process-level liveness
- `/readyz`: readiness based on runtime state and enabled infrastructure dependencies
- `/startupz`: startup state
- `/healthz`: compatibility endpoint with JSON payload containing `name`, `version`, `status`, `live`, `ready`, and `startup`

This makes the template usable in local development, Docker, Kubernetes, and older internal tooling at the same time.

## CSRF Behavior

CSRF is disabled by default because many API services are token-based and do not need it.

When you enable `middleware.csrf.enabled`:

- the template exposes `GET /csrf/token`
- the response returns the token, header name, and cookie name
- write requests must send the token back in the CSRF header

Example:

```bash
curl http://127.0.0.1:9999/csrf/token
```

Then send the returned token in `X-Csrf-Token` for state-changing requests.

## Rate Limiting

The rate limiter is configurable through `middleware.limiter`.

Supported strategies:

- `fixed`
- `sliding`

Supported key sources:

- `ip`
- `path`
- `ip_path`
- `header`

For `header`, set `middleware.limiter.key_header` to the header name you want to use.

This is useful when moving from a local demo to gateway-aware or tenant-aware rate limiting.

## Configuration Overview

The main configuration file is `./config/server.yaml`.

Important sections:

- `server`: application identity, port, mode, and runtime limits
- `database`: SQLite, MySQL, or PostgreSQL configuration
- `redis`: optional Redis connection
- `prometheus`: metrics exposure
- `log`: log level and output file
- `middleware`: all HTTP middleware toggles and options
- `waf`: Coraza rules
- `schedule`: scheduler switch
- `proxy`: outbound proxy settings

The default configuration is intentionally runnable with only:

- Go installed
- no external database
- no Redis
- no Prometheus

## Migrations

The heavy template supports module-level migrations.

Run all registered migrations:

```bash
go run . migrate up
```

The current demo keeps `database.auto_migrate: true` for a frictionless first run, but long-term schema evolution should rely on explicit migrations instead of `AutoMigrate` alone.

Migration examples:

- `internal/modules/user/migrations/seed.go`

Applied migration versions are tracked in the `schema_migrations` table.

## Module Architecture

Each module should expose a `NewBundle()` factory and return a `modules.Bundle`.

The bundle can contribute:

- `RouteModules`
- `DatabaseModels`
- `Migrations`
- `StartupHooks`
- `ShutdownHooks`
- `ScheduledJobs`
- `BackgroundServices`

Reference files:

- `internal/modules/module.go`
- `internal/modules/user/module.go`
- `internal/modules/schedule/schedule.go`

To add a new module:

1. Create a new folder under `internal/modules/<module>`.
2. Add the controller, service, DAO, models, and migrations that the module needs.
3. Implement `NewBundle()` in the module root.
4. Register that factory in `internal/bootstrap/application.go` through `modules.Collect(...)`.

## Directory Layout

- `cmd`: CLI commands such as `serve`, `migrate up`, `install`, `uninstall`, and `version`
- `config`: configuration files
- `internal/bootstrap`: application assembly and lifecycle
- `internal/infra`: infrastructure components such as DB, logging, metrics, cache, and scheduler
- `internal/modules`: business modules and module bundles
- `internal/transport`: HTTP router, middleware, and embedded UI
- `pkg`: shared abstractions and utilities

## Testing

The integration test lives in:

```bash
internal/bootstrap/bootstrap_integration_test.go
```

Run the full test suite:

```bash
go test ./...
```

The integration coverage currently validates:

- service bootstrap with the default SQLite setup
- automatic database creation
- migration metadata table creation
- health probes
- request ID and security headers
- ETag and compression behavior
- end-to-end user CRUD flow

## Known Design Tradeoffs

- Module registration is still centralized in `internal/bootstrap/application.go`.
- The timeout wrapper covers the common route registration methods used by the template. If you introduce unusual routing patterns, keep timeout coverage in mind.
- Request ID is already present in access logs, but business log enrichment with request context is still a follow-up improvement.

## Template Checklist

Before turning this into your own service, the usual next steps are:

- replace the module path in `go.mod`
- update the app identity in `config/server.yaml`
- adjust auth secrets and service metadata
- disable demo-only defaults you do not want in production
- add your own modules and remove the demo module if it is no longer needed
