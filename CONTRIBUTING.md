# Contributing

## Scope

`fiberx` is maintained as a generator-first repository. Changes should improve one of these areas:

- generator assets
- planning and validation rules
- rendering and metadata flows
- upgrade inspection
- build automation
- regression coverage
- release-facing documentation

## Workflow

1. Make focused changes.
2. Keep user-facing contracts stable unless the change intentionally updates them.
3. Update docs when the CLI, generated scaffold, or release surface changes.
4. Run the relevant tests before submitting work.

## Local Checks

Minimum checks for most changes:

```bash
go test ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

If the change affects generated output, also generate a sample project and verify the touched files.

## Repository Notes

- `sample/` is reference material and test-facing comparison content.
- `output/` is local scratch space and should stay out of version control except for `.gitkeep`.
- Generated release binaries should not be committed.

## Style

- Prefer small, explicit templates over broad string replacement.
- Keep generated code readable for everyday Go developers.
- Avoid introducing framework-like abstractions unless they clearly improve the generator mainline.
- Preserve compatibility rules for presets, capabilities, and runtime options.

## Release-Facing Changes

When a change affects release behavior, update the relevant documents:

- `README.md`
- `README_zh.md`
- `docs/README.md`
- `docs/guides/usage.md`
- `docs/roadmap/roadmap.md`
- `CHANGELOG.md`
