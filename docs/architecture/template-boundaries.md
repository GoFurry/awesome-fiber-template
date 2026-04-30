# Template Boundaries

`fiberx` maintains four official preset semantics. Their boundaries are intentionally fixed so the generator can evolve without letting every optional concern leak into every starting point.

## Boundary Rules

- Presets only keep capabilities that are part of their default or explicitly supported project path.
- Optional capabilities should be expressed through generator capability policy or runtime/build options, not by widening every preset.
- New official preset tiers should not be added to represent one-off combinations.
- Generated output should stay copy-friendly and understandable without introducing framework-style orchestration layers.

## `heavy`

Use `heavy` when you want the most complete engineering baseline.

It may keep:

- Redis
- scheduler
- metrics
- WAF
- Swagger
- embedded UI
- the heaviest middleware and ops-oriented baseline

It should not grow into:

- a platform framework
- a large built-in demo collection
- a catch-all place for speculative infrastructure

## `medium`

Use `medium` when you want a production-oriented HTTP baseline without the heavier runtime burden from scheduler and metrics defaults.

It may keep:

- Redis
- embedded UI
- Swagger
- common web middleware
- production-oriented HTTP lifecycle defaults

It should not keep:

- scheduler defaults
- metrics defaults
- platform-style orchestration layers

## `light`

Use `light` when you want a practical API template that still feels close to a normal Go service.

It may keep:

- SQLite-first setup
- common API middleware
- CRUD demo
- optional embedded UI support
- optional Swagger support

It should not keep:

- Redis
- metrics/jobs
- the heavier ops-oriented middleware set from upper tiers

## `extra-light`

Use `extra-light` when you want the smallest maintainable starting point.

It may keep:

- SQLite startup
- minimal config
- health probes
- `recover`

It should not keep:

- built-in business demos
- expanded runtime options
- docs/UI defaults
- heavier infrastructure integrations

## Capability Placement Defaults

When a new capability is proposed, use this decision order:

1. If it belongs to a preset's default path, keep it in that preset only.
2. If it is optional but broadly reusable, express it through capability policy or runtime/build configuration.
3. If it is mainly a usage pattern or demo concern, prefer docs or generated examples over preset expansion.

## Selection Notes

- Choose `heavy` for the broadest engineering baseline.
- Choose `medium` for balanced production HTTP services.
- Choose `light` for plain Go-style API services.
- Choose `extra-light` for the smallest clean base.
