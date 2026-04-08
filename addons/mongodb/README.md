# mongodb addon

Status: placeholder only.

## Purpose

Provide a small MongoDB client wrapper that can be reused across templates or copied into a project.

## Intended Shape

- `Config` for connection settings
- `New(...)` for client creation
- `Close()` for graceful shutdown

## Notes

- Keep the package independent from template-specific config models.
- Prefer a thin wrapper over a large abstraction layer.
