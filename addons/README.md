# Optional Addons

`addons` is a placeholder capability pool for reusable, opt-in service adapters and utility packages.

It is intentionally kept outside `v3/` so the four template tiers stay focused on their own engineering baseline.

## Design Goals

- Keep each addon small and easy to copy or import.
- Make every addon optional, with no hard dependency on any template tier.
- Prefer one capability per package.
- Document the intended interface before adding implementation details.

## Suggested Layout

```text
addons/
  mongodb/
  redis/
  minio/
  kafka/
  rabbitmq/
  mail/
  sms/
  jwt/
  id/
  crypto/
  retry/
  paginator/
  observability/
```

## Integration Rule

Templates should not depend on `addons` by default.
Use an addon only when a project explicitly needs it, then wire it in at the application boundary.

## Current Status

All addon folders are placeholders only.
Each folder contains documentation that describes the intended scope and future shape.
