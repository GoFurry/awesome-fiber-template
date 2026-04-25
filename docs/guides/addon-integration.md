# Addon Integration Guide

This guide explains how to integrate repository addons into a real project without coupling the project to this repository structure.

## Basic Rule

Do not make your project depend on `addons/` as a permanent runtime requirement.

Instead:

1. pick the addon you need
2. copy it into your project boundary
3. wire its config and lifecycle in your own application

That keeps the repository reference presets clean and keeps your project ownership clear.

## When To Use A Repository Addon

A repository addon is a good fit when:

- the capability is optional
- it is infrastructure-oriented
- you want a small, copy-friendly implementation
- you do not need a large framework around it

Examples:

- `addons/migrate`
- `addons/redis`
- `addons/mongodb`
- `addons/s3`
- `addons/mail`

## When To Prefer A Community Integration

Some capabilities are better served by mature ecosystem integrations than by a repository-maintained addon.

Current community-first choices:

- observability: `github.com/gofiber/contrib/v3/otel`
- JWT auth middleware: `github.com/gofiber/contrib/v3/jwt`

Use those directly in your project unless you have a strong reason to wrap them locally.

## Suggested Integration Steps

### 1. Copy the addon into the project

Suggested destination examples:

- `internal/infra/redis`
- `internal/infra/migrate`
- `internal/infra/mongodb`

### 2. Replace module-local imports if needed

Repository addons are designed to avoid preset coupling, so integration should mostly be a straight copy.

### 3. Define project-owned config

Do not depend on repository config paths. Re-declare only the fields your project wants to own.

Example for Redis:

- `addr`
- `username`
- `password`
- `db`
- `pool_size`

Example for migrate:

- `dialect`
- `dsn`
- `dir`
- `table`

### 4. Wire lifecycle in your bootstrap

Create the service in startup code, expose what your app needs, and close it in shutdown.

### 5. Document the decision in the project README

If a project adopts an addon, note why it was chosen and whether it replaces a lighter default path such as `auto_migrate`.

## `migrate` Integration Notes

`addons/migrate` is intended for projects that want stricter schema management than `auto_migrate`.

Typical migration directory:

```text
internal/db/migrations/
```

Suggested CLI commands in a real project:

- `migrate create`
- `migrate up`
- `migrate down`
- `migrate status`

Recommended coexistence strategy:

- keep `auto_migrate` for SQLite-first demos
- switch to explicit migrations for stricter environments

## `redis` Integration Notes

`addons/redis` is intentionally aligned with the Redis usage style in `heavy` and `medium`.

That means you can:

- keep raw client access when you need advanced commands
- use helpers for high-frequency operations
- standardize config shape across projects

## Final Guideline

Reference presets define the default starting point.

Addons define optional infrastructure capabilities.

Community integrations remain the first choice when the ecosystem already provides a stronger long-term answer.
