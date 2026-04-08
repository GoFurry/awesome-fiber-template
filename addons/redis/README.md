# redis addon

Status: placeholder only.

## Purpose

Provide a reusable Redis client wrapper for cache, queue, rate-limit, and session scenarios.

## Intended Shape

- `Config` for address, auth, database, and pool settings
- `New(...)` for client creation
- `Close()` for graceful shutdown

## Notes

- Keep the wrapper thin and explicit.
- Let projects decide how to namespace keys.
