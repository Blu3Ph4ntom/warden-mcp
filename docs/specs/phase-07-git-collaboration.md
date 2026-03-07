# Phase 7 Git & Collaboration Model

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH07-T01 through PH07-T10

## Purpose

Warden uses git for traceability and safe collaboration, but git is not the canonical source of governed plan state. Canonical state remains the Warden datastore, while git provides human review history, checkpoint snapshots, and recovery support around markdown projections and code changes.

## When git integration is optional vs required

- optional for solo local evaluation where the operator only needs governed state and local markdown projection
- recommended for all normal repository-backed development
- required when multiple operators or agents may touch the same workspace
- required before archive, release, or cross-machine handoff workflows
- required for any workflow that relies on human review of plan or code diffs

If git is unavailable, Warden may still govern plan execution, but it must surface reduced traceability warnings.

## Commit snapshot strategy

- create small commits around meaningful accepted changes, not after every keystroke
- prefer one logical checkpoint per completed governed task or tightly related task update
- snapshot both code changes and user-facing markdown projections when they changed together
- never treat a git commit as proof of task completion by itself
- keep snapshot creation explicit so operators control repo history and signing policy

## Commit message format

Baseline format:

- one short imperative subject line
- lowercase preferred
- no generated branding, agent signatures, or marketing text
- describe the change, not the tool
- keep subjects narrowly scoped to one logical step

Examples:

- `add gitignore`
- `document git workflow`
- `add commit policy tests`

## Branch and merge safety rules

- default to working on a dedicated feature branch unless the repo is intentionally single-user
- avoid concurrent mutation of the same plan projection by multiple operators without reconciliation
- require branch freshness checks before applying markdown imports or finishing a plan
- never auto-resolve merge conflicts in plan markdown by guessing intent
- when plan and datastore disagree after a merge, Warden must force reconciliation before finish approval

## Markdown human-edit workflow

1. operator refreshes local branch and exports or opens the latest markdown projection
2. operator edits only allowed projection fields and preserves immutable task IDs
3. operator runs Warden validation or reconcile dry run before commit
4. Warden classifies drift as safe projection edits, governed mutations, or conflicts
5. operator commits only after validation output is understood

Human markdown edits are allowed, but they are requests for reconciliation, not authoritative state changes on their own.

## Stale-state detection

Warden should treat state as stale when any of these are true:

- the current branch head changed since the last export or reconcile baseline
- the markdown file mtime or content hash changed outside the last known Warden export
- the canonical plan revision changed since the operator loaded status
- a merge or rebase rewrote plan-adjacent files after the last reconcile check

Stale-state must surface as an explicit warning or conflict, not a silent overwrite.

## Import-after-edit workflow

- read current canonical revision and current git head
- parse and validate the edited markdown projection
- compare revision and git baseline against the current workspace
- in dry run, return changed IDs, warnings, and conflicts without persisting
- in apply mode, persist only if the workspace is not stale or the operator explicitly resolves the conflict path
- regenerate markdown after apply so projection and canonical state converge again

## Conflict resolution workflow

- classify conflicts into structural, stale-state, identity, status, dependency, and metadata-loss-risk classes
- prefer operator-visible denial over automatic merge of conflicting task states
- require exact task or phase IDs in every conflict report
- allow safe auto-resolution only for non-semantic formatting drift
- after conflict resolution, regenerate projection content from canonical state before further edits

## Recovery after bad manual edits

- preserve the last clean exported projection or regenerate it from canonical state on demand
- use git diff plus reconcile dry run to isolate accidental deletions, ID corruption, or reordered sections
- if IDs were damaged, restore from canonical export instead of guessing replacements
- if a bad edit was already committed, revert in git first or import a corrected projection explicitly
- always record the recovery action in audit metadata once the runtime exists

## Team and multi-operator ownership rules

- one operator owns the current mutation window for a plan unless the plan explicitly allows parallel independent tasks
- task ownership should be explicit at the human process level even if v1 stores it only as notes or metadata
- required tasks blocked by another operator's unfinished change must remain finish-blocking
- handoffs should include current branch, canonical revision, open blockers, and changed task IDs
- no operator may waive another operator's required task without explicit authority and recorded rationale

## Next implementation step

After this design phase, the Go implementation should add repository-state inspection, commit-policy validation, and stale-state detection helpers that future MCP tools can call before reconcile, archive, and finish-gate operations.