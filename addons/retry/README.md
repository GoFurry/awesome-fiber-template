# retry addon

Status: placeholder only.

## Purpose

Provide a reusable retry helper for outbound calls.

## Intended Shape

- retry policy configuration
- backoff strategy helpers
- optional context-aware execution wrapper

## Notes

- Keep retry policy explicit and easy to reason about.
- Do not hide retries inside unrelated packages.
