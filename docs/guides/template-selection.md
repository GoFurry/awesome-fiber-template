# Template Selection Guide

Use this guide when choosing a reference preset as the starting point for a new service.

## Choose `heavy` if

- you want the stronger ops-oriented production track
- you want built-in Swagger, embedded UI, metrics, and a local scheduler job loop by default
- you expect Redis and broader runtime infrastructure needs to matter early

## Choose `medium` if

- you want a practical production-oriented HTTP preset
- you still want Redis, Swagger, and embedded UI support
- you do not want metrics and scheduler defaults in the starting path

## Choose `light` if

- you want a smaller but still directly usable HTTP service
- you want SQLite-first startup, CRUD demo, and the common API middleware baseline
- you may want to opt into Swagger or embedded UI later, but do not want them by default
- you do not want Redis, metrics, or scheduler defaults

## Choose `extra-light` if

- you want the smallest clean base
- you only need SQLite startup, minimal config, and basic health and panic handling
- you do not want CRUD, docs, UI, or the broader middleware stack in the starting point
- you prefer adding everything else yourself

## Practical Defaults

- Most production-style HTTP services should start from `medium`.
- Move to `heavy` when you want the second production track with stronger ops defaults.
- Most simple internal APIs and small services should start from `light`.
- Use `extra-light` when minimalism matters more than convenience.

## Capability Boundaries

- `swagger` and `embedded-ui` are default on `heavy` and `medium`
- `swagger` and `embedded-ui` are opt-in on `light`
- `embedded-ui` does not require `swagger`
- `redis` is only available on `heavy` and `medium`
- `extra-light` intentionally stays outside these three capabilities

## Stack Defaults

- Default generated stack: `Fiber v3 + Cobra + Viper`
- Compatibility stack: `Fiber v2 + native-cli`
- Use compatibility mode only when you must stay aligned with an older runtime or command model
