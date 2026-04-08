# jwt addon

Status: placeholder only.

## Purpose

Provide a reusable JWT helper package for signing, parsing, and claims handling.

## Intended Shape

- `Config` for secret / public key / issuer / expiry
- token creation helpers
- token parsing and validation helpers

## Notes

- Keep authentication policy in the application layer.
- Use the package as a cryptographic helper, not as a full auth system.
