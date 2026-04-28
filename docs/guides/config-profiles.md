# Config Profiles Guide

Phase 9 standardizes the shallow environment layout used by generated services.

## Generated Profiles

Every generated project now includes:

- `config/server.yaml`
- `config/server.dev.yaml`
- `config/server.prod.yaml`

## Intent

- `server.yaml`: first-run baseline
- `server.dev.yaml`: local development defaults
- `server.prod.yaml`: safer production-oriented starting point

## Why This Is Intentionally Shallow

This is not a full config-system redesign. The goal is to make environment-specific startup clearer while preserving the current runtime structure and `--config` flow.

## Guidance

- Keep key structure aligned across profiles.
- Use production profiles to disable public docs/UI unless required.
- Treat these files as generated starting points, not frozen policy.
