# Phase 1 Product Foundations

Status: proposed baseline for Phase 1 completion
Plan coverage: PH01-T01 through PH01-T10

## 1. Product Promise

Warden MCP is a governance platform for AI coding agents that makes plan state, task status, dependency rules, and finish approval authoritative.

It is not a lightweight checklist utility. Its core promise is: a compliant integrated agent cannot honestly claim work is complete early, because finish approval is denied until required plan conditions are satisfied.

## 2. Enforcement Boundary

V1 enforcement applies to integrated agents and operator workflows that use the Warden MCP tool surface.

Warden does:

- validate plans and task mutations
- control finish approval authoritatively
- preserve auditable state and denial reasons
- detect sync, dependency, and completion violations

Warden does not:

- physically prevent humans from editing files
- control non-integrated or malicious external tools
- guarantee truth of evidence outside its observable boundary
- replace source control, CI, or human judgment

## 3. V1 Personas

### Solo developer with one coding agent

Needs strong structure, clear next tasks, and reliable finish denial when work is incomplete.

### Multi-agent operator

Needs one canonical plan, dependency-safe parallel execution, and visibility into blocked or conflicting work.

### Tech lead or reviewer

Needs auditability, evidence trails, and confidence that completion claims map to real plan state.

## 4. Success Criteria

Warden succeeds in v1 when all of the following are true:

- finish requests are denied whenever any required task is non-terminal
- finish requests are denied whenever required phases are incomplete
- dependency violations block progress or finish as designed
- denied requests return actionable reasons and next required actions
- task, phase, and plan rollups remain internally consistent after mutations
- task IDs remain stable across edits, imports, exports, and reconciliation
- operators can inspect current status and understand why work is blocked
- audit history shows who attempted what change and whether it was accepted

## 5. Agent Misbehavior and Failure Patterns

Warden must explicitly guard against:

- creating shallow plans that are too vague to govern execution
- marking tasks done without evidence or acceptance criteria
- skipping blocked tasks and claiming global progress anyway
- ignoring dependencies or phase order
- rewriting or deleting task IDs during edits
- collapsing multiple unfinished tasks into a vague summary
- claiming tests passed without validation evidence
- requesting finish after partial implementation only
- silently changing scope without plan updates
- hiding unresolved conflicts between datastore and markdown

## 6. Human Collaboration Scenarios

V1 must support these human-in-the-loop patterns:

- human initializes or imports a plan, agent executes against it
- human edits markdown projection, then reconciliation determines safe application
- human adds or reprioritizes work when scope changes
- human reviews blocked tasks and supplies waiver or recovery rationale
- human audits finish denial reasons and directs next action
- human reopens or resets work after failed validation or regression

## 7. Out of Scope for V1

The following are intentionally out of scope:

- generic project management for non-agent teams
- guaranteed enforcement against malicious humans or rogue non-integrated agents
- automatic proof that external evidence is truthful
- replacement for git hosting, issue tracking, or CI systems
- arbitrary workflow builder for every organizational process
- deep SaaS collaboration features such as live multi-user editing
- full financial, scheduling, or resourcing management

## 8. Non-Negotiable Invariants

- Canonical state lives in a durable datastore, not only in markdown.
- Every plan has immutable plan, phase, and task IDs.
- Every executable plan contains multiple phases and multiple actionable tasks per phase.
- Required tasks must reach a terminal state before finish can be approved.
- Required phases cannot be skipped for finish approval.
- Dependencies reference immutable task IDs only.
- State transitions must be explicit and validated.
- Denied mutations and denied finish requests still produce audit events.
- Markdown is a projection and must not silently override canonical state.
- Reconciliation must never silently discard IDs or evidence.

## 9. Plan Lifecycle States

### Draft

Plan is being authored or imported and is not yet trusted for execution.

### Active

Plan is valid for execution and can accept governed task mutations.

### Blocked

Plan cannot safely progress because of unresolved blockers, invalid state, or required intervention.

### Completed

All required finish-gate conditions passed and finish approval was granted.

### Archived

Plan is retained for history and is not expected to accept normal execution mutations.

## 10. Compliance vs Convenience Policy

When compliance and convenience conflict, Warden chooses compliance.

Operational policy:

- deny ambiguous finish requests rather than guess intent
- deny invalid state transitions rather than coerce silently
- require explicit rationale for waivers, cancellations, and resets
- prefer surfaced warnings and repair suggestions over hidden normalization
- optimize for auditability and correctness before ergonomics

## 11. Immediate Implications for Build Work

The first implementation slices should focus on:

- canonical domain model and legal state transitions
- authoritative finish gating
- auditable mutation handling
- durable local persistence
- markdown import/export only as a projection layer
