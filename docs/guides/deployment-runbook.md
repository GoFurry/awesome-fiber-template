# Deployment Runbook Guide

Phase 9 closes the gap between “generated and runnable” and “generated with a usable operating playbook”.

## What Generated Projects Now Include

Every generated project now ships:

- `docs/runbook.md`
- `docs/configuration.md`
- `docs/api-contract.md`
- `docs/verification.md`

These files are meant to stay project-local so teams can evolve them after generation without waiting for a generator-wide feature.

## Expected Runtime Flow

Recommended deployment order:

1. Run `go test ./...`.
2. Start with `config/server.prod.yaml`.
3. Check `/healthz`, `/readyz`, and any enabled optional routes.
4. Confirm writable SQLite and log paths.
5. Keep docs/UI endpoints disabled in public production environments unless they are intentionally exposed.

## Preset Notes

- `heavy`: also verify `/metrics` and scheduler visibility.
- `medium`: verify docs/UI defaults if they remain enabled.
- `light`: verify docs/UI only when explicitly requested.
- `extra-light`: keep the surface minimal and focus on health plus startup.
