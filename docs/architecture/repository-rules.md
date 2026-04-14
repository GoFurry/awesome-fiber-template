# Repository Rules

These rules keep `awesome-fiber-template` maintainable as a long-lived template collection.

## Template Evolution Rules

- Do not add new official template tiers for one-off combinations.
- Do not move optional capabilities into template defaults unless they are part of that tier's high-frequency path.
- Do not turn business demos into feature-rich example applications inside the templates.

## Addon Evolution Rules

- Prefer `addons/` for optional infrastructure integrations.
- Keep addon boundaries explicit and copy-friendly.
- Treat `mail`, `mongodb`, and `s3` as style references for future addons.

## Documentation Rules

- Root README explains repository positioning and points to deeper docs.
- Template READMEs explain the current template only.
- Long-term architecture and evolution rules live under `docs/`, not in individual template READMEs.

## Quality Rules

- Central template verification stays in `v3/test`.
- Template-local tests should not be reintroduced unless there is a strong reason.
- CI must validate template modules, centralized tests, and addon modules separately.

## Scope Rules

- `heavy` is the upper bound for built-in engineering baseline.
- `extra-light` is the lower bound for built-in engineering baseline.
- Future complexity should usually be added through docs, addons, or generator work, not by widening template responsibilities.
