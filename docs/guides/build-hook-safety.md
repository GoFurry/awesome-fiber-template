# Build Hook Safety

`fiberx build` may execute project-defined hooks.

Only run build hooks in repositories you trust.

## Recommended Practice

- inspect the project before running hooks
- use `fiberx build --dry-run` to review planned commands
- keep hook commands explicit and auditable
- prefer predictable project-local scripts over opaque shell chains
- in interactive shells, review the hook summary before confirming execution
- in non-interactive environments, use `--yes` to approve hooks or `--no-hooks` to skip them

## CI Guidance

In CI, treat hooks as trusted-repository behavior rather than a default assumption.

If a repository is not trusted, do not execute its build hooks. Prefer `fiberx build --no-hooks` unless the repository and its hook commands are explicitly approved.

## Why This Matters

Build hooks run with the same local privileges as the build command itself. They are useful for project-specific preparation and packaging, but they are also part of the repository trust boundary.
