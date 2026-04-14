# Roadmap

This roadmap now tracks only the next unfinished stages of `awesome-fiber-template`.

Implemented work such as template tiering, centralized tests, template boundary rules, and the first productized addons is intentionally left out so this document stays focused on what comes next.

## Current Position

The repository already has:

- four stable template tiers under `v3/`
- centralized black-box tests under `v3/test`
- formal template and addon boundary documentation
- productized `migrate` and `redis` addons
- community-first decisions for observability and auth integration

That means the next milestone is no longer about stabilizing the repository shape. It is about making composition easier.

## P2: Composition Layer

P2 is the next active milestone.

### Goal

Reduce manual template duplication and make project generation more structured without turning the repository into a heavy framework.

### Planned Deliverables

- a minimal `aft` generator MVP
- preset manifests for official starting points
- capability manifests for optional composition rules
- documented generation rules for template selection and addon injection
- validation for generated project output

### Generator MVP Scope

The first generator should stay intentionally small.

Expected commands:

- `aft new <name> --preset heavy`
- `aft new <name> --preset medium`
- `aft new <name> --preset light`
- `aft new <name> --preset extra-light`
- optional `--with` flags for supported capabilities

The MVP should handle:

- module path replacement
- preset selection
- optional capability injection
- basic README and config bootstrapping

The MVP should not try to solve:

- AST-heavy code rewriting
- remote template marketplaces
- plugin systems
- large internal DSLs

## P3: Data Access Expansion

After the composition layer is stable, the next priority is expanding data-access choices without multiplying template tiers.

### Goal

Keep the current GORM-first path, while offering a cleaner SQL-first alternative.

### Candidate Deliverables

- `medium-sql` as a preset or capability combination
- SQL-first query tooling exploration
- migration-aware database presets
- clearer guidance for choosing ORM vs SQL-first paths

## P4: Production Capability Packs

Once generator and preset composition are stable, the next stage is deeper production-oriented capability packs.

### Candidate Areas

- stronger observability presets
- audit logging
- webhook tooling
- queue integration
- security-oriented capability packs

These should prefer preset and addon composition instead of creating more template tiers.

## Long-Term Direction

The long-term goal is not to become a giant boilerplate collection.

It is to become a practical Fiber v3 engineering baseline system with:

- layered templates
- reusable optional addons
- clear composition rules
- future-ready project generation

## What We Will Avoid

To keep the repository maintainable, the roadmap explicitly avoids:

- adding more and more template tiers
- rebuilding mature community integrations without a strong reason
- turning templates into business-heavy starter apps
- coupling projects directly to repository internals
