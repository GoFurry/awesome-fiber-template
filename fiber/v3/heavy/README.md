# Fiber v3 Heavy Template

`heavy` is the full-featured edition of this scaffold. It is designed to support medium and large Go backend services with a clear module boundary, lifecycle management, and reusable infrastructure.

## Features

- Uses SQLite by default for an out-of-the-box demo experience.
- Includes a complete `user` CRUD example module.
- Lets each module register routes, models, migrations, scheduled jobs, startup hooks, shutdown hooks, and background services.
- Provides a `migrate up` command for module-level versioned migrations.
- Ships with lifecycle bootstrap, metrics, scheduler, Redis, WAF, and configuration wiring that can be enabled as needed.

## Quick Start

The default configuration lives in `./config/server.yaml`.

Start the HTTP service:

```bash
go run . serve
```

Available endpoints after startup:

- `GET /healthz`
- `GET /api/v1/users/`
- `POST /api/v1/users/`
- `GET /api/v1/users/:id`
- `PUT /api/v1/users/:id`
- `DELETE /api/v1/users/:id`

On the first run the template will automatically:

- create `./data/app.db`
- create the `demo_users` table
- run registered module migrations

## Migrations

Run all registered module migrations:

```bash
go run . migrate up
```

The template still keeps `database.auto_migrate` enabled by default to make the demo frictionless. For long-term schema evolution, prefer explicit module migrations instead of relying only on `AutoMigrate`.

## Directory Layout

- `./cmd`: CLI entry points
- `./config`: configuration files
- `./internal/bootstrap`: application assembly and lifecycle
- `./internal/modules`: business modules
- `./internal/infra`: infrastructure components
- `./pkg`: shared abstractions and utilities

## Creating a Module

Each module should expose a `NewBundle()` function and return a unified module bundle.

Minimum steps:

1. Create `controller`, `service`, `dao`, `models`, and `migrations` under `internal/modules/<module>`.
2. Implement `NewBundle()` in the module root and attach routes, models, migrations, and scheduled jobs there.
3. Register the module's `NewBundle` in `internal/bootstrap/application.go` through `modules.Collect(...)`.

A module can provide these extension points:

- `RouteModules`
- `DatabaseModels`
- `Migrations`
- `ScheduledJobs`
- `StartupHooks`
- `ShutdownHooks`
- `BackgroundServices`

Reference implementations:

- `internal/modules/user/module.go`
- `internal/modules/schedule/schedule.go`
- `internal/modules/module.go`

## CRUD Demo

The built-in `user` module includes:

- route registration: `internal/modules/user/module.go`
- controller: `internal/modules/user/controller/userController.go`
- service layer: `internal/modules/user/service/userService.go`
- data access: `internal/modules/user/dao/userDao.go`
- database model: `internal/modules/user/models/user.go`
- migration example: `internal/modules/user/migrations/seed.go`

## Tests

The integration test lives in `internal/bootstrap/bootstrap_integration_test.go`.

Run the full test suite:

```bash
go test ./...
```
