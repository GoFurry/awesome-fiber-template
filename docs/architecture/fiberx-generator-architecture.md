# fiberx Generator Architecture

This document defines the maintained architecture for `fiberx` as a CLI-first Fiber project generator.

The current source of truth is the generator mainline itself:

- `cmd/`
- `internal/`
- `generator/`
- the root test matrix
- the maintained docs under `docs/`

Future work should extend those surfaces directly instead of recreating parallel in-repo systems.

## Positioning

`fiberx` is a generator repository, not a template warehouse.

Its maintained outputs are:

- stable preset semantics
- explicit capability policy
- runtime and build configuration support
- generated project metadata
- diff and upgrade inspection
- project-level build engineering

## Public Product Surface

The user-facing model stays intentionally small:

- `preset`
- `capability`
- a limited set of generation parameters
- generated project metadata
- build configuration

Internal assembly concepts such as packs and rules remain implementation details.

## Preset Model

The repository maintains four official presets:

- `heavy`
- `medium`
- `light`
- `extra-light`

Those presets are long-lived product entry points, not temporary implementation phases.

## Generator Model

The generator mainline is composed from:

- `base` assets
- preset packs
- capability packs
- runtime overlays
- replace rules
- injection rules

The execution flow is:

1. manifest loading
2. validation
3. planning
4. rendering
5. writing
6. reporting

## Metadata And Upgrade Model

Generated projects carry generator-owned metadata through `.fiberx/manifest.json`.

That metadata supports:

- project inspection
- diff detection
- readonly upgrade assessment

These are part of the maintained generator surface, not optional side systems.

## Build Engineering Model

Generated projects may also carry `fiberx.yaml` for project-level build automation.

The build flow is:

1. config loading
2. optional profile overlay
3. target and platform expansion
4. build execution
5. optional packaging and checksums
6. build metadata output
7. release manifest output

## Directory Responsibilities

- `cmd/fiberx`
  - CLI entrypoint and command wiring
- `internal/core`
  - top-level generation orchestration
- `internal/manifest`
  - catalog loading and manifest resolution
- `internal/planner`
  - plan selection and asset composition
- `internal/validator`
  - request and catalog validation
- `internal/renderer`
  - template rendering and rule application
- `internal/writer`
  - filesystem writes
- `internal/report`
  - generation summaries
- `internal/metadata`
  - project metadata and diff logic
- `internal/upgrade`
  - readonly upgrade assessment
- `internal/buildconfig`
  - project build configuration parsing
- `internal/build`
  - build, archive, checksum, metadata, and release manifest execution
- `generator`
  - maintained generator assets, rules, and manifest data

## Long-Term Boundaries

- The repository maintains only the generator mainline.
- Optional future engineering concerns should enter through the generator model, runtime options, or build configuration.
- Historical content is preserved in Git history, not through parallel in-repo legacy systems.
