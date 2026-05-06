# Changelog

## v0.1.2

- Added shared scaffold constants for `light`, `medium`, and `heavy`
- Added a base application error model and response compatibility layer
- Added configurable request timeout support for business routes
- Added default `middleware.timeout` config to generated `light`, `medium`, and `heavy` projects
- Kept system routes outside timeout wrapping
- Kept `extra-light` as the minimal scaffold

## v0.1.1

- Added Fiber v3 app hooks skeleton with `app.Hooks()` integration
- Added default graceful shutdown wiring for generated Fiber v2 and v3 projects
- Added stronger default middleware setup for `medium`, `heavy`, and `light`
- Added optional JSON backend selection with `--json-lib stdlib|sonic|go-json`
- Added `json_lib` to generated project metadata and inspection flows
- Documented the trust boundary for build hooks

## v0.1.0

- Added four official presets: `heavy`, `medium`, `light`, `extra-light`
- Added stable capability support for `redis`, `swagger`, and `embedded-ui`
- Added runtime selection for logger, database, and data access options
- Added generated project metadata, diff inspection, and readonly upgrade inspection
- Added project-level build automation with profiles, packaging, checksums, hooks, UPX, build metadata, and release manifest output
- Reduced the repository to the generator mainline only
