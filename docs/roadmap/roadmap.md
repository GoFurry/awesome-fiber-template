# Roadmap

This roadmap tracks the long-term evolution of `awesome-fiber-template` as a Fiber v3 engineering baseline repository.

## Current State

The repository already has:

- four stable template tiers under `v3/`
- a centralized black-box test module under `v3/test`
- a documented template boundary system
- an `addons/` capability layer with maintained reference implementations

That means the next steps are not about adding more random templates. They are about stabilizing the composition model.

## Product Direction

The repository should evolve along four layers:

1. template layer: `heavy`, `medium`, `light`, `extra-light`
2. addon layer: optional reusable infrastructure capabilities
3. composition layer: future generator and preset support
4. quality layer: CI, smoke tests, race checks, and compatibility verification

## P0 Status

P0 is complete.

Delivered:

- formal template boundary documentation
- addon design rules
- repository evolution rules
- centralized documentation entrypoints
- stronger CI separation for templates, centralized tests, and addons

## P1 Status

P1 focuses on high-value reusable capabilities and clearer product boundaries.

Delivered or being delivered:

- `addons/migrate`: thin migration wrapper based on `pressly/goose/v3`
- `addons/redis`: reusable Redis service aligned with the Redis shape kept in `heavy` and `medium`

Boundary decisions:

- observability: prefer `github.com/gofiber/contrib/v3/otel`
- auth middleware: prefer `github.com/gofiber/contrib/v3/jwt`

These are intentionally community-first choices, not missing repository features.

## P2 Direction

After the addon layer is stable, the next stage should be a minimal composition workflow:

- generator MVP
- preset manifests
- capability manifests
- project bootstrap automation

The goal is to reduce manual template duplication rather than introduce a heavy abstraction framework.

## What We Will Avoid

To keep the repository maintainable, the roadmap explicitly avoids:

- adding more and more template tiers
- turning templates into business-heavy starter apps
- rebuilding mature community integrations without a strong reason
- coupling projects directly to repository internals

## Near-Term Priorities

1. finish productizing `migrate` and `redis`
2. keep template and addon boundaries stable
3. strengthen validation and compatibility checks
4. prepare a small generator MVP only after the addon model feels proven

## Long-Term Goal

The long-term goal is not to become a giant boilerplate collection.

It is to become a practical Fiber v3 engineering baseline system with:

- layered templates
- reusable optional addons
- clear selection rules
- future-ready project generation
