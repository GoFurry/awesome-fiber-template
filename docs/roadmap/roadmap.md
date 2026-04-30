# Roadmap

## v0.1.0

`v0.1.0` marks the first release-ready milestone of the current `fiberx` generator mainline.

It includes:

- four official presets: `heavy`, `medium`, `light`, `extra-light`
- a stable capability contract for `redis`, `swagger`, and `embedded-ui`
- runtime options for logger, database, and data access selection on `medium / heavy / light`
- generated project metadata, diff detection, and readonly upgrade inspection
- project-level build engineering with profiles, packaging, checksums, hooks, UPX, build metadata, and release manifest output
- a repository layout fully reduced to the generator mainline only

## v0.1.1

`v0.1.1` is the next planned milestone.

Planned items:

- Fiber v3 lifecycle hook skeleton points in the default generated app structure
  - source: [Fiber v3 Hooks](https://docs.gofiber.io/api/hooks/)
  - `OnPreStartupMessage`
  - `OnPostStartupMessage`
  - `OnPreShutdown`
  - `OnPostShutdown`
  - this is planned only for `Fiber v3`; `Fiber v2` will not receive the same hook skeleton
  - goal: make graceful startup and graceful shutdown customization easier without injecting business logic by default
- optional JSON backend selection for generated Fiber projects
  - sources:
    - [Fiber v3 Make Fiber Faster](https://docs.gofiber.io/guide/faster-fiber/)
    - [Fiber v2 Make Fiber Faster](https://docs.gofiber.io/v2.x/guide/faster-fiber/)
  - planned generator parameter: `--json-lib`
  - default: `stdlib`
  - first-round planned values:
    - `stdlib`
    - `go-json`
    - `sonic`
    - `segmentio`
    - `simdjson`
  - not planned for the first round:
    - `easyjson`
    - `jettison`
  - goal: support optional `JSONEncoder / JSONDecoder` integration without changing the standard library default
