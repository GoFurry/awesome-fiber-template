# Release Process

This guide describes the lightweight release flow for the `fiberx` generator.

## 1. Confirm Release Scope

- verify the roadmap entry is up to date
- verify README and docs use the correct release wording
- confirm generated scaffold changes are reflected in docs and examples

## 2. Run Core Checks

From the repository root:

```bash
go test ./...
go run ./cmd/fiberx validate
go run ./cmd/fiberx doctor
```

If the release changes generated output, also generate at least one representative project and review the affected files.

## 3. Review Release-Facing Surface

Check these files before tagging:

- `README.md`
- `README_zh.md`
- `docs/README.md`
- `docs/guides/usage.md`
- `docs/roadmap/roadmap.md`
- `CHANGELOG.md`

## 4. Prepare Release Notes

Release notes should stay short and focus on user-visible changes:

- new generator features
- scaffold changes
- build or upgrade behavior changes
- documentation or release-surface changes when relevant

Avoid internal phase history or implementation detail dumps.

## 5. Tag And Publish

Recommended sequence:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```

Then create the GitHub release and use the prepared release notes.
