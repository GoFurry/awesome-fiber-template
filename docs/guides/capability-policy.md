# Capability Policy Guide

This guide locks the current capability contract for `fiberx`.

## Current Capability Roles

- `swagger`: API docs capability
- `embedded-ui`: bundled UI capability
- `redis`: external cache and infrastructure capability

## Current Preset Boundaries

- `swagger`
  - default on: `heavy`, `medium`
  - optional on: `light`
  - unsupported on: `extra-light`
- `embedded-ui`
  - default on: `heavy`, `medium`
  - optional on: `light`
  - unsupported on: `extra-light`
- `redis`
  - optional on: `heavy`, `medium`
  - unsupported on: `light`, `extra-light`

## Important Notes

- `swagger` and `embedded-ui` are independent capabilities
- enabling `embedded-ui` does not require `swagger`
- `redis` is intentionally limited to the production-oriented presets
- preset choice decides service weight first; capability choice only adjusts the supported optional surface inside that preset
