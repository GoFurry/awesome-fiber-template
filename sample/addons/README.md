# Addons

`addons/` is the optional capability layer of this repository.

Unlike the `v3/*` templates, addons are intentionally self-contained and copy-friendly. A template should stay focused on its default path, while infrastructure capabilities that are useful only for some projects should live here.

## Current Layout

```text
addons/
  mail/
  migrate/
  mongodb/
  redis/
  s3/
```

## What Should Become An Addon

An addon is a good fit when it is:

- optional for most projects
- infrastructure-oriented
- easy to copy into an application boundary
- more useful as a reusable capability than as a template default

## What Should Not Become An Addon

Do not use `addons/` for:

- business-domain code
- template-specific glue
- generic dump-bin utilities
- features that clearly belong in one template tier's default path

## Implemented Addons

### `mail/`

Reusable SMTP mail addon with:

- multi-account SMTP pool
- rotation strategies such as `none`, `round_robin`, and `random`
- failover on retryable SMTP and connection errors
- custom HTML and built-in HTML templates
- common mail fields such as `cc`, `bcc`, `reply-to`, headers, and attachments

### `mongodb/`

Reusable MongoDB addon based on the official `mongo-driver/v2`, with:

- `URI`-first and structured configuration support
- `Client`, `Database`, and `Collection` access
- `Ping` and `Close`
- thin CRUD helpers around a collection wrapper
- direct access to the raw driver for advanced usage

### `s3/`

Reusable S3-compatible object storage addon based on AWS SDK v2, with:

- explicit config for region, endpoint, credentials, bucket, and path-style mode
- upload helpers for bytes, readers, and local files
- object download as bytes or stream
- `HeadObject` and idempotent `DeleteObject`
- pre-signed `GET`, `PUT`, and `DELETE` URLs

### `migrate/`

Reusable schema migration addon based on `pressly/goose/v3`, with:

- SQL migration support only
- explicit `Dialect`, `DSN`, `Dir`, and tracking table configuration
- `Up`, `Down`, `Status`, `Version`, and `Create` helpers
- a copy-friendly service wrapper that works outside the template tree

### `redis/`

Reusable Redis addon based on `go-redis/v9`, with:

- explicit config for address, username, password, database index, and pool size
- `New`, `Ping`, `Close`, and raw client access
- common string, hash, prefix-scan, and pipeline helpers
- the same usage style now aligned with the Redis support kept in `heavy` and `medium`

## Community-First Capabilities

Not every useful capability should become a repository-maintained addon.

For some areas, the better long-term choice is to rely on mature community integrations directly:

- `otel`: prefer [`github.com/gofiber/contrib/v3/otel`](https://github.com/gofiber/contrib/tree/main/v3/otel)
- `auth`: prefer [`github.com/gofiber/contrib/v3/jwt`](https://github.com/gofiber/contrib/tree/main/v3/jwt)

For API key authentication and other business-specific auth strategies, implement them in the project boundary instead of forcing a generic addon too early.

## Integration Rule

Templates should not depend on `addons` by default.

When a project needs one of these capabilities, copy the addon into the application boundary and wire it through that project's own config and lifecycle.

## Notes

- `mail/`, `mongodb/`, `s3/`, `migrate/`, and `redis/` are the currently maintained addons in this repository.
- `mail/`, `mongodb/`, and `s3/` remain the original reference style for future addon work.
- `migrate/` and `redis/` are the first addons productized after the template boundary rules were formalized.
