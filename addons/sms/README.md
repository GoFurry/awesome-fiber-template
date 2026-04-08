# sms addon

Status: placeholder only.

## Purpose

Provide a reusable SMS provider wrapper for common verification and notification flows.

## Intended Shape

- `Config` for provider credentials and region settings
- `New(...)` for client creation
- helper methods for sending verification and alert messages

## Notes

- Treat provider-specific APIs as implementation details.
