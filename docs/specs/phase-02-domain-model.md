# Phase 2 Domain Model Baseline

Status: initial implementation-aligned draft
Plan coverage: implemented baseline for PH02-T01 through PH02-T12

## Canonical Entities

V1 currently defines these core entity families:

- `Plan`
- `Phase`
- `Task`
- `AuditEvent`
- `FinishRequest`
- `ValidationIssue`
- `EvidenceItem`
- `Note`
- transition rule tables for plan, phase, and task lifecycle changes

The implemented baseline now includes baseline entities for audit events, finish requests, evidence, notes, version snapshots, archive records, reset requests, and closure requests. Full finish-gate evaluation is still a later layer.

## Current Status Enums

### Plan

- `draft`
- `active`
- `blocked`
- `completed`
- `archived`

### Phase

- `not_started`
- `in_progress`
- `blocked`
- `completed`

### Task

- `not_started`
- `in_progress`
- `done`
- `blocked`
- `cancelled`
- `waived`

## Baseline Shape Rules

- plan IDs, phase IDs, and task IDs are required
- plan IDs must be lowercase slug-like strings
- phase IDs must match `PH##`
- task IDs must match `PH##-T##`
- phase titles and task titles are required
- plans with fewer than two phases are invalid
- phases with fewer than two tasks are invalid
- task dependency lists must not contain duplicates
- task dependency lists must not contain self-dependencies
- task dependencies must reference existing task IDs
- dependency cycles are invalid
- terminal task states are `done`, `cancelled`, and `waived`
- evidence items require at least `kind` and `ref`
- notes require valid `actor_type` and non-empty text
- plan versions must look like semantic versions
- `current_phase_id` must reference an existing phase when present

## Transition Baseline

### Task transitions

- `not_started -> in_progress | blocked | cancelled`
- `in_progress -> done | blocked | not_started`
- `blocked -> in_progress | waived | cancelled`
- terminal task states do not transition further in the current baseline

### Phase transitions

- `not_started -> in_progress | blocked`
- `in_progress -> blocked | completed | not_started`
- `blocked -> in_progress`
- `completed` is terminal in the current baseline

### Plan transitions

- `draft -> active | archived`
- `active -> blocked | completed | archived`
- `blocked -> active | archived`
- `completed -> archived`
- `archived` is terminal in the current baseline

## Governance Baseline

### Versioning

- plans carry a version string
- version snapshots record `plan_id`, `plan_version`, `revision`, `schema_version`, and `recorded_at`
- revisions must be positive integers

### Archival

- archive records preserve `plan_id`, `plan_version`, and `archived_at`
- archive reason and final export path are allowed metadata fields

### Reset and reopen

- reset is a governed override flow, separate from normal task state transitions
- reset targets are limited to `not_started` and `in_progress`
- reset requests require a non-empty reason and timestamp

### Waived and cancelled governance

- `waived` and `cancelled` are explicit governance closure states
- closure requests require a valid actor type, non-empty reason, and timestamp
- `done` is not treated as a waiver/cancellation closure state

## Intentional Gaps

The following still need later refinement before they should be considered fully stable:

- immutable ID generation strategy beyond validation and preservation
- finish-request denial evidence and rollup behavior
- deeper archival/export coupling semantics

## Current Code Mapping

The current implementation lives in:

- `internal/domain/model.go`
- `internal/domain/governance.go`
- `internal/domain/transitions.go`
- `internal/domain/model_test.go`

These files are a baseline scaffold, not a completed domain layer.
