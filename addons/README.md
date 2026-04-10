# Addons

`addons` is the reusable capability area of this repository.

Unlike the `v3/*` templates, addons are intentionally optional and self-contained. Each addon is meant to be copied into a real project only when that capability is needed.

## Current Layout

```text
addons/
  mail/
  mongodb/
  s3/
```

## Design Rules

- keep runtime code small and easy to copy into an app
- keep addons decoupled from `v3/*` templates by default
- prefer one addon for one clear infrastructure capability
- document boundaries and usage before adding more abstraction

## Implemented Addons

### `mail/`

Reusable SMTP mail addon with:

- multi-account SMTP pool
- rotation strategies such as `none`, `round_robin`, and `random`
- failover on retryable SMTP and connection errors
- custom HTML and built-in HTML templates
- common mail fields such as `cc`, `bcc`, `reply-to`, headers, and attachments

### `mongodb/`

Reusable MongoDB addon based on the official `mongo-driver/v2`, with:

- `URI`-first and structured configuration support
- `Client`, `Database`, and `Collection` access
- `Ping` and `Close`
- thin CRUD helpers around a collection wrapper
- direct access to the raw driver for advanced usage

### `s3/`

Reusable S3-compatible object storage addon based on AWS SDK v2, with:

- explicit config for region, endpoint, credentials, bucket, and path-style mode
- upload helpers for bytes, readers, and local files
- object download as bytes or stream
- `HeadObject` and idempotent `DeleteObject`
- pre-signed `GET`, `PUT`, and `DELETE` URLs

## Integration Rule

Templates should not depend on `addons` by default.

When a project needs one of these capabilities, copy the addon into the application boundary and wire it through that project's own config and lifecycle.

## Notes

- `mail/`, `mongodb/`, and `s3/` are the current maintained addons in this repository.
- The old MinIO-specific direction has been consolidated into the more general `s3/` addon.
- If more addons are added later, they should follow the same "optional, single-purpose, easy to copy" rule.
