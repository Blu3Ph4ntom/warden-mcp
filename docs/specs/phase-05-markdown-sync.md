# Phase 5 Markdown Format & Sync Engine

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH05-T01 through PH05-T12

## Purpose

Markdown is the human-readable projection of canonical Warden state. The projection must be strict enough to round-trip safely, while reconciliation remains conservative and never silently discards IDs, evidence, or status history.

## Strict markdown template

Each exported plan uses this structure:

1. YAML frontmatter
2. one H1 title line
3. fixed top-level sections in order:
   - `## Global Rules`
   - `## Execution Protocol`
   - `## Task Annotation Convention`
   - `## Finish Gate`
   - one `## Phase N — ...` section per phase
4. task lines under each phase using checkbox syntax with immutable task IDs

Projection order is deterministic and must match canonical phase/task position order.

## Frontmatter schema

Required frontmatter keys:

- `plan_id`
- `title`
- `version`
- `status`
- `created_at`
- `can_finish`
- `current_phase`
- `total_phases`
- `total_tasks`
- `completed_tasks`
- `task_id_format`
- `phase_progression`
- `required_finish_checks`

Optional keys:

- `source`
- `owner`
- `updated_at`
- `archived_at`
- `projection_version`

Unknown frontmatter keys are preserved during export/import only if namespaced under `x_`; other unknown keys produce validation warnings.

## Task ID syntax

- phase IDs must appear as `PH##`
- task IDs must appear as `PH##-T##`
- task list entries use `- [ ] PH##-T## Description`
- checkbox state maps to status projection but does not replace canonical validation
- IDs are immutable and must be parsed from the beginning of each task line

## Priority and dependency annotations

Current implementation writes optional per-task metadata inline on the task line using pipe-delimited key/value segments:

- `- [ ] PH01-T01 Description | priority:P1`
- `- [ ] PH01-T02 Description | depends_on:PH01-T01,PH01-T03 | required:false`

Supported keys today:

- `priority: P0|P1|P2|P3`
- `depends_on: PH##-T##,PH##-T##`
- `required: true|false`
- `status: cancelled|waived` when the `[-]` checkbox would otherwise be ambiguous

If no metadata is needed, the task line remains the plain `- [ ] PH##-T## Description` form.

`blocked`, `cancelled`, and `waived` all project through the `[-]` checkbox marker; `status:` is emitted only for `cancelled` and `waived` so those terminal states round-trip exactly.

## Note and evidence blocks

Optional task metadata blocks use indented section headers below the task entry:

- `notes:` followed by compact JSON bullets representing `{ "actor_type", "text", "created_at" }`
- `evidence:` followed by compact JSON bullets representing `{ "kind", "ref", "summary" }`

Exact rendered shape:

- task line metadata stays inline, for example `- [ ] PH01-T01 Description | priority:P1 | required:false`
- note/evidence blocks render as:
  - `  notes:`
  - `    - {"actor_type":"agent","text":"captured note","created_at":"2026-03-07T00:00:00Z"}`
  - `  evidence:`
  - `    - {"kind":"log","ref":"build-log","summary":"smoke run output"}`

These blocks are projections of canonical metadata and are never the sole source of truth, but they must round-trip without discarding audit fields.

## Markdown export rules

- export from canonical SQLite state only
- preserve immutable IDs exactly
- normalize line endings to `\n` in generated content
- emit sections in canonical order only
- omit empty optional metadata blocks
- render archived plans read-only with archive metadata in frontmatter
- include audit summary only in explicit export modes, not in the default plan projection

## Markdown import rules

- parse frontmatter before body validation
- reject plans without valid plan/phase/task IDs in strict mode
- allow whitespace normalization and heading punctuation normalization
- preserve existing canonical IDs during merge/apply flows
- create mode may generate missing IDs only from a dedicated init/import path, never during reconcile of an existing plan
- unknown tasks or phases require explicit conflict reporting rather than silent insertion/deletion

## Parser edge cases

The parser must explicitly handle:

- Windows or Unix line endings
- trailing whitespace and blank lines
- tabs converted to spaces for indentation checks
- wrapped task descriptions
- fenced code blocks containing task-like text that must be ignored
- duplicate task IDs
- duplicate phase IDs
- malformed checkboxes or bullets
- invalid metadata lines or duplicate metadata keys

## Reconciliation algorithm

1. parse markdown into a structured projection AST
2. validate frontmatter, section order, IDs, and metadata syntax
3. compare parsed entities to canonical state by immutable ID
4. classify differences as safe projection drift, canonical mutation request, or conflict
5. in `dry_run`, return conflicts and proposed changed IDs only
6. in `apply`, persist only non-conflicting changes inside one transaction and emit audit events
7. regenerate markdown from canonical state after apply

Reconciliation is ID-first; title text and ordering alone must never be used as identity.

## Conflict classes and resolutions

Baseline conflict classes:

- `PARSE_ERROR`
- `FRONTMATTER_INVALID`
- `DUPLICATE_ID`
- `UNKNOWN_TASK_ID`
- `UNKNOWN_PHASE_ID`
- `DEPENDENCY_TARGET_MISSING`
- `STRUCTURE_MISMATCH`
- `STATUS_CONFLICT`
- `STALE_PROJECTION`
- `METADATA_LOSS_RISK`

Each conflict must return a code, message, optional target ID, and a repair suggestion. Apply mode must abort on unresolved conflicts.

## Malformed-plan repair suggestions

Representative repair guidance:

- suggest restoring missing IDs from canonical export
- suggest splitting merged task lines into one line per immutable task ID
- suggest moving free-form notes into the `notes:` block
- suggest converting prose dependency mentions into `depends_on:` metadata
- suggest re-exporting from canonical state if section order or frontmatter is irreparably damaged

## Round-trip fidelity tests

The future sync layer should prove:

- export → import(dry run) produces no conflicts for untouched projections
- export → import(apply) → export is byte-stable except for normalized whitespace
- notes, evidence, priorities, required flags, and dependencies survive round-trip
- duplicate IDs and malformed metadata are surfaced as explicit conflicts
- reconciliation never deletes canonical IDs silently

## Next implementation step

After this design phase, the Go code should introduce markdown projection and reconciliation package boundaries that operate on canonical plan entities and the Phase 3 contract types.