# Template Boundaries

This repository currently preserves four official Fiber v3 reference presets. Their boundaries are intentionally fixed so `fiberx` can evolve toward a generator repository without losing the semantics of its official starting points.

## Boundary Rules

- Reference presets only keep capabilities that are part of their default, high-frequency project path.
- Capabilities that are optional, niche, or infrastructure-specific should prefer `addons/`.
- New official preset tiers should not be added to represent one-off capability combinations.
- Reference presets should stay copy-friendly: a user should be able to pick one preset, replace the module path, and start building.

## `heavy`

Use `heavy` when you want the most complete engineering baseline.

It may keep:

- Redis
- scheduler
- Prometheus
- WAF
- Swagger
- service install and uninstall support
- `pkg/httpkit`
- `pkg/abstract`
- the full middleware baseline already present in the template

It should not grow into:

- a platform framework with mandatory module assembly
- a large built-in business demo collection
- a place for speculative infrastructure that is not part of the default full-featured path

## `medium`

Use `medium` when you want a production-oriented HTTP service template without the heavier runtime burden from scheduler and Prometheus.

It may keep:

- Redis
- embedded UI support
- service install and uninstall support
- WAF
- Swagger
- common web middleware and request lifecycle support

It should not keep:

- scheduler
- Prometheus
- platform-style orchestration layers

## `light`

Use `light` when you want a practical API template that still feels close to a normal Go service.

It may keep:

- SQLite-first setup
- optional embedded UI support
- common API middleware
- plain `controller`, `dao`, `service`, and `models` business structure

It should not keep:

- Redis
- service install and uninstall support
- `pkg/httpkit`
- `pkg/abstract`
- heavier optional middleware such as WAF, Swagger, CSRF, Helmet, or pprof

## `extra-light`

Use `extra-light` when you want the smallest maintainable starting point.

It may keep:

- SQLite only
- minimal config
- native CLI
- `recover`
- health probes

It should not keep:

- built-in business demos
- enhanced infrastructure integrations
- helper packages beyond the smallest common response and error helpers
- optional middleware layers from heavier tiers

## Capability Placement Defaults

When a new capability is proposed, use this decision order:

1. If it is required by one template tier's default path, keep it in that tier only.
2. If it is optional or integration-specific, add it to `addons/`.
3. If it is mainly a usage pattern or a business example, prefer docs or future examples over template expansion.

## Selection Notes

- Choose `heavy` for the broadest engineering baseline.
- Choose `medium` for balanced production HTTP services.
- Choose `light` for plain Go-style API services.
- Choose `extra-light` for the smallest clean base.
