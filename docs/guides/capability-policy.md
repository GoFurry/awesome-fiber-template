# Capability Policy Guide

This guide locks the current `fiberx` capability contract for `swagger`, `embedded-ui`, and `redis`.

## Capability Roles

- `swagger`: API docs capability
- `embedded-ui`: bundled UI capability
- `redis`: external cache and infrastructure capability

## Preset Boundaries

- `heavy`
  - defaults: `swagger`, `embedded-ui`
  - optional: `redis`
- `medium`
  - defaults: `swagger`, `embedded-ui`
  - optional: `redis`
- `light`
  - defaults: none
  - optional: `swagger`, `embedded-ui`
- `extra-light`
  - defaults: none
  - optional: none

## Combination Matrix

- `heavy`
  - valid: `default`
  - valid: `redis`
  - valid: `swagger`
  - valid: `embedded-ui`
  - valid: `swagger + embedded-ui`
  - valid: `swagger + redis`
  - valid: `embedded-ui + redis`
  - valid: `swagger + embedded-ui + redis`
- `medium`
  - valid: `default`
  - valid: `redis`
  - valid: `swagger`
  - valid: `embedded-ui`
  - valid: `swagger + embedded-ui`
  - valid: `swagger + redis`
  - valid: `embedded-ui + redis`
  - valid: `swagger + embedded-ui + redis`
- `light`
  - valid: `default`
  - valid: `swagger`
  - valid: `embedded-ui`
  - valid: `swagger + embedded-ui`
  - invalid: any combination that includes `redis`
- `extra-light`
  - valid: `default`
  - invalid: any non-empty capability combination

## Behavior Rules

- `swagger` and `embedded-ui` are independent capabilities.
- Enabling `embedded-ui` does not require `swagger`.
- For `medium` and `heavy`, explicitly passing `swagger` or `embedded-ui` does not change the final generated behavior because both are already default capabilities.
- `redis` remains limited to `medium` and `heavy`.
- Phase 12 validates unsupported combinations at request-validation time whenever possible.

## Verification Scope

- Generation tests lock the full capability matrix.
- Black-box tests verify:
  - `light` default, `swagger`, `embedded-ui`, and `swagger + embedded-ui`
  - `medium` default, `redis`, and full explicit capability request
  - `heavy` default, `redis`, and full explicit capability request
- `redis` verification in Phase 12 stays at assembly, startup, health, and service-reporting level.
