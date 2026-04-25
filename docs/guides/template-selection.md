# Template Selection Guide

Use this guide when choosing a reference preset as the starting point for a new service.

## Choose `heavy` if

- you want the richest engineering baseline
- you expect Redis, scheduler, Prometheus, WAF, Swagger, and service manager support to be relevant early
- you want reusable helper packages such as `pkg/httpkit` and `pkg/abstract` available from day one

## Choose `medium` if

- you want a practical production-oriented HTTP preset
- you still want Redis, WAF, Swagger, service manager support, and embedded UI support
- you do not want scheduler and Prometheus complexity in the default path

## Choose `light` if

- you want a preset that feels closer to a plain Go project
- you want SQLite-first startup and common API middleware
- you do not want Redis, service manager support, or extra helper packages by default

## Choose `extra-light` if

- you want the smallest clean base
- you only need SQLite, minimal config, native CLI, and basic health and panic handling
- you prefer adding everything else yourself

## Practical Defaults

- Most production-style HTTP services should start from `medium`.
- Most simple internal APIs and small services should start from `light`.
- Use `heavy` only when you know you want the broader engineering baseline.
- Use `extra-light` when minimalism matters more than convenience.
