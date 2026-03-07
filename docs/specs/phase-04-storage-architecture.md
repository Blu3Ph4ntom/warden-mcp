# Phase 4 Storage & Persistence Architecture

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH04-T01 through PH04-T10

## Purpose

Warden needs a durable canonical datastore. Markdown remains a projection, while SQLite is the source of truth for plans, tasks, audit history, finish gating, archival metadata, and reconciliation state.

## Canonical SQLite strategy

- use a normalized relational schema in SQLite
- keep one canonical database per Warden workspace
- treat plan/phase/task tables as current state
- treat audit and snapshot tables as append-only history
- use immutable IDs as stable foreign keys everywhere

## Baseline tables

- `plans`: plan identity, title, status, version, current phase, finish flag, timestamps
- `phases`: phase identity, owning plan, ordinal position, title, status, rollup counts
- `tasks`: task identity, owning phase/plan, ordinal position, title, status, priority, required flag, timestamps
- `task_dependencies`: immutable edge list from task to dependency task
- `task_evidence`: evidence rows keyed to task ID with kind/ref/summary
- `task_notes`: task comments keyed to task ID with actor type, text, timestamp
- `audit_events`: append-only mutation and denial history
- `finish_requests`: finish attempts plus decision payload metadata
- `version_snapshots`: plan version, revision, schema version, recorded time
- `archive_records`: archive metadata and final export paths
- `reconcile_runs`: markdown import/reconcile attempts and conflict summary metadata

## Index strategy

Required indexes:

- `plans(plan_id)` unique
- `plans(status, updated_at)` for list/status queries
- `phases(plan_id, phase_id)` unique and `phases(plan_id, position)` unique
- `tasks(plan_id, task_id)` unique and `tasks(phase_id, position)` unique
- `tasks(plan_id, status, required)` for finish and blocked-work scans
- `tasks(plan_id, priority, status)` for next-task selection
- `task_dependencies(task_id, depends_on_task_id)` unique
- `audit_events(plan_id, timestamp desc)` for plan history reads
- `finish_requests(plan_id, requested_at desc)` for gating audits
- `archive_records(plan_id, archived_at desc)` for archive lookups

## Event log schema

`audit_events` should capture:

- `event_id`
- `plan_id`
- optional `phase_id` and `task_id`
- `actor_type`
- `event_type`
- `accepted`
- `message`
- optional structured `details_json`
- `timestamp`

Denied changes and denied finish requests must still produce audit rows.

## Migration strategy

- store a schema version using SQLite `PRAGMA user_version`
- keep forward-only numbered SQL migrations under a dedicated storage package
- run migrations inside a single transaction at process startup
- fail closed if migration state is unknown or partially applied
- reserve destructive migrations for explicit future upgrade steps, not silent startup behavior

## Transaction boundaries

Use one transaction for each logical mutation:

- plan initialization
- task update plus phase/plan rollup updates plus audit event insert
- plan edit plus affected dependency/order changes plus audit event insert
- finish request evaluation plus finish request record plus audit event insert
- archive operation plus snapshot/export metadata plus audit event insert
- reconcile apply plus changed entities plus conflict/audit records

Read-only status queries may use a consistent read transaction when multiple tables are consulted.

## Concurrency policy

- prefer SQLite WAL mode for safe concurrent readers with a single writer
- serialize writes within the process to avoid overlapping rollup mutations
- use optimistic checks on plan `updated_at` or revision when applying reconcile/edit flows
- surface write contention and stale-state conflicts as explicit sync/conflict errors
- never silently retry a mutation that could change operator intent

## Snapshot and export strategy

- create version snapshots whenever a plan version changes or an archive is created
- exports are generated from canonical SQLite state, not markdown caches
- export metadata should record format, path, generation time, and whether audit summary was included
- markdown projection files may be regenerated, but they are never the only recovery source

## Backup and restore flow

- support cold-copy workspace backups of the SQLite file plus generated exports
- before restore, run integrity checks and schema-version validation
- record restore operations in audit metadata when the runtime is available
- treat restore as replacing canonical state; markdown projections must be regenerated afterward

## Archive storage rules

- archived plans remain queryable but immutable to normal mutation tools
- archive records keep plan ID, version, reason, archive time, and final export path
- final exports for archived plans should be stored beside archive metadata, not used as canonical state
- archived plans stay in SQLite unless a later retention policy explicitly moves them

## Integrity checks

The storage layer should provide routines that verify:

- foreign-key integrity across plans, phases, tasks, and dependencies
- phase/task position uniqueness within parent scopes
- dependency edges only reference existing task IDs
- rollup counts match task rows
- current phase points to an existing phase for the plan
- finish request and archive rows reference valid plans
- SQLite `PRAGMA integrity_check` passes before backup/restore and on demand

## Next implementation step

After this design phase, the Go code should introduce storage package boundaries and repository interfaces that preserve these rules without yet requiring markdown sync or transport concerns.