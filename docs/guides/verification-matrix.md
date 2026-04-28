# Verification Matrix Guide

Phase 9 extends generator verification beyond stack selection.

## Covered By Regression Tests

- default stack generation for all four presets
- compatibility stack generation for all four presets
- black-box HTTP checks for `medium` and `heavy` across the full stack matrix
- black-box checks for `light` and `extra-light` on default and legacy stacks
- startup smoke checks for the mixed lightweight combinations
- generated docs and config profile existence
- missing-route JSON `404` envelope behavior

## Manual Smoke Suggestions

After generation, the minimum manual pass is:

1. `go test ./...`
2. `go run . services`
3. `go run . serve --config config/server.dev.yaml`
4. `GET /healthz`
5. `GET /docs/openapi.yaml` when docs are expected
6. `GET /ui` when embedded UI is expected
7. `GET /metrics` on `heavy`
