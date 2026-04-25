# Repository Rules

These rules keep `fiberx` maintainable as a long-lived generator repository with stable preset semantics.

## Preset Evolution Rules

- Do not add new official preset tiers for one-off combinations.
- Do not move optional capabilities into preset defaults unless they are part of that preset's high-frequency path.
- Do not turn business demos into feature-rich example applications inside the reference presets.

## Addon Evolution Rules

- Prefer `addons/` for optional infrastructure integrations that should stay outside generator v1 assembly.
- Keep addon boundaries explicit and copy-friendly.
- Treat `mail`, `mongodb`, and `s3` as style references for future addons.

## Documentation Rules

- Root README explains repository positioning and points to deeper docs.
- The generator architecture document is the top-level design basis for future implementation work.
- Template READMEs explain the current reference preset only.
- Long-term architecture and evolution rules live under `docs/`, not in individual preset READMEs.

## Quality Rules

- Central reference preset verification stays in `v3/test` during the current transition stage.
- Template-local tests should not be reintroduced unless there is a strong reason.
- CI must validate reference preset modules, centralized tests, and addon modules separately.

## Scope Rules

- `heavy` is the upper bound for built-in engineering baseline.
- `extra-light` is the lower bound for built-in engineering baseline.
- Future complexity should usually be added through docs, addons, or generator work, not by widening reference preset responsibilities.
