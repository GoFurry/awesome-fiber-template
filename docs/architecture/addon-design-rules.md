# Addon Design Rules

`addons/` is the optional capability layer of this repository. It exists so templates can stay focused while infrastructure helpers remain reusable.

## What Belongs In `addons/`

An addon is appropriate when all of these are true:

- it provides one clear infrastructure capability
- it is optional for most projects
- it can be copied into an application boundary without relying on repository internals
- it is more useful as a reusable building block than as a default template feature

Typical examples:

- mail delivery
- MongoDB integration
- S3-compatible object storage
- migration and cache integrations

## What Should Not Become An Addon

Do not use `addons/` for:

- generic dumping-ground utilities
- business-domain code
- helpers that only make sense inside one template tier
- wrappers whose main job is to hide repository structure instead of providing real infrastructure value

## Required Structure

Every maintained addon must have:

- a standalone `README.md`
- a clear config type or initialization contract
- a minimal usage example
- lifecycle hooks when applicable, usually `New`, `Close`, and `Ping`
- basic tests
- an independent Go module boundary when the addon is maintained as a standalone reusable package

## Coupling Rules

- Addons must not depend on `v3/*` template internals.
- Templates must not depend on addons by default.
- Projects should copy or integrate addons at the application boundary and own their config and lifecycle wiring there.

## Documentation Rules

Each addon README should state:

- what problem it solves
- when it should be used
- why it is not a default template feature
- whether it commonly pairs with other addons
- what remains intentionally out of scope

## Current Reference Addons

The current reference implementations are:

- `mail`
- `mongodb`
- `s3`

Future addons should match their overall style:

- single-purpose
- copy-friendly
- explicit config
- no hidden template coupling

## Community-First Rule

Do not build an addon just because a capability is useful.

If the Fiber ecosystem already has a stable, well-maintained, and widely adopted integration, prefer documenting that community solution instead of maintaining a repository-specific addon.

Current community-first decisions:

- observability: prefer `github.com/gofiber/contrib/v3/otel`
- auth middleware: prefer `github.com/gofiber/contrib/v3/jwt`

Only build and maintain a repository addon when it gives clearer value than reusing the existing ecosystem standard.
