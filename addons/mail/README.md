# mail addon

Status: placeholder only.

## Purpose

Provide a reusable SMTP / mail-sending helper.

## Intended Shape

- `Config` for host, port, username, password, sender, and TLS
- `New(...)` for transport creation
- helper methods for text and HTML messages

## Notes

- Keep templates and business content out of this package.
