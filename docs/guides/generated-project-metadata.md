# Generated Project Metadata

Every generated project now includes a reserved metadata file:

- `.fiberx/manifest.json`

This file records:

- the generator version and commit
- the generation recipe
- the selected asset set
- stable template and rendered-output fingerprints
- the managed file list with SHA256 hashes

## Inspect Metadata

Use `inspect` to read the recorded metadata:

```bash
go run ./cmd/fiberx inspect ./demo
go run ./cmd/fiberx inspect ./demo --json
```

## Compare Against The Current Generator

Use `diff` to compare a generated project against the current generator output:

```bash
go run ./cmd/fiberx diff ./demo
go run ./cmd/fiberx diff ./demo --json
```

Current diff statuses:

- `clean`
- `local_modified`
- `generator_drift`
- `local_and_generator_drift`

`fiberx diff` only compares generator-managed files. It does not compare arbitrary user-added files and it does not write any changes back to the project.
