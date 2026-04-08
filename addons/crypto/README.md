# crypto addon

Status: placeholder only.

## Purpose

Provide small reusable crypto helpers such as hashing, encryption, and signing utilities.

## Intended Shape

- `Hash(...)`
- `Encrypt(...)`
- `Decrypt(...)`
- `Sign(...)`
- `Verify(...)`

## Notes

- Keep the surface area minimal.
- Avoid coupling to any particular web framework.
