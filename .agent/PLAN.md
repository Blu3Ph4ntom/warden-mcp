---
plan_id: warden-mcp-v1
title: Warden MCP Full Platform Plan
version: 1.0.0
status: active
current_phase: PH11
can_finish: false
completed_tasks: 146
---

# Warden MCP Full Platform Plan

## Phase 1 — Product Definition & Non-Negotiable Rules
- [x] PH01-T01 Define exact product promise: governance platform vs checklist utility
- [x] PH01-T02 Define enforcement boundary: integrated-agent enforcement only
- [x] PH01-T03 Define v1 personas: solo dev, multi-agent operator, tech lead
- [x] PH01-T04 Define success criteria for “cannot finish early”
- [x] PH01-T05 Catalog agent misbehavior and failure patterns
- [x] PH01-T06 Define human collaboration scenarios
- [x] PH01-T07 Define out-of-scope items for v1
- [x] PH01-T08 Write non-negotiable invariants for plans, phases, and tasks
- [x] PH01-T09 Define plan lifecycle states
- [x] PH01-T10 Define compliance vs convenience trade-off policy

## Phase 2 — Domain Model & State Machine
- [x] PH02-T01 Define canonical entities: plan, phase, task, event, finish request
- [x] PH02-T02 Define immutable ID strategy
- [x] PH02-T03 Define legal task state transitions
- [x] PH02-T04 Define legal phase state transitions
- [x] PH02-T05 Define legal plan state transitions
- [x] PH02-T06 Define dependency model
- [x] PH02-T07 Define task evidence model
- [x] PH02-T08 Define notes and comments model
- [x] PH02-T09 Define versioning semantics
- [x] PH02-T10 Define archival semantics
- [x] PH02-T11 Define rollback and reopen semantics
- [x] PH02-T12 Define waived and cancelled governance rules

## Phase 3 — MCP Tool Contract Design
- [x] PH03-T01 Finalize `init_plan` contract
- [x] PH03-T02 Finalize `validate_plan` contract
- [x] PH03-T03 Finalize `get_status` contract
- [x] PH03-T04 Finalize `update_task` contract
- [x] PH03-T05 Finalize `get_next_task` contract
- [x] PH03-T06 Finalize `request_finish` contract
- [x] PH03-T07 Finalize `edit_plan` contract
- [x] PH03-T08 Finalize `reset_task` contract
- [x] PH03-T09 Finalize `prioritize_tasks` contract
- [x] PH03-T10 Finalize `list_plans` contract
- [x] PH03-T11 Finalize `archive_plan` contract
- [x] PH03-T12 Finalize `import_plan` contract
- [x] PH03-T13 Finalize `export_plan` contract
- [x] PH03-T14 Finalize `reconcile_plan` contract
- [x] PH03-T15 Define shared error response schema for all tools

## Phase 4 — Storage & Persistence Architecture
- [x] PH04-T01 Choose SQLite schema strategy
- [x] PH04-T02 Define indexes for plan, task, and status queries
- [x] PH04-T03 Define event log schema
- [x] PH04-T04 Define migration strategy
- [x] PH04-T05 Define transaction boundaries
- [x] PH04-T06 Define concurrency handling policy
- [x] PH04-T07 Define snapshot and export strategy
- [x] PH04-T08 Define backup and restore flow
- [x] PH04-T09 Define archive storage rules
- [x] PH04-T10 Define integrity-check routines

## Phase 5 — Markdown Format & Sync Engine
- [x] PH05-T01 Define strict markdown template
- [x] PH05-T02 Define frontmatter schema
- [x] PH05-T03 Define explicit task ID syntax
- [x] PH05-T04 Define priority and dependency annotation format
- [x] PH05-T05 Define note and evidence block format
- [x] PH05-T06 Define markdown export rules
- [x] PH05-T07 Define markdown import rules
- [x] PH05-T08 Define parser edge cases
- [x] PH05-T09 Define reconciliation algorithm
- [x] PH05-T10 Define conflict classes and resolutions
- [x] PH05-T11 Define malformed-plan repair suggestions
- [x] PH05-T12 Define round-trip fidelity tests

## Phase 6 — Enforcement Layer Design
- [x] PH06-T01 Define system prompt template
- [x] PH06-T02 Define agent operating rules
- [x] PH06-T03 Define pre-finish validation flow
- [x] PH06-T04 Define forced next-step behavior after denial
- [x] PH06-T05 Define loop detection rules
- [x] PH06-T06 Define repeated-failure escalation
- [x] PH06-T07 Define max-iteration policy
- [x] PH06-T08 Define blocked-task escalation policy
- [x] PH06-T09 Define human intervention triggers
- [x] PH06-T10 Define compliance scoring
- [x] PH06-T11 Define runtime wrapper responsibilities
- [x] PH06-T12 Define anti-hallucination safeguards

## Phase 7 — Git & Collaboration Model
- [x] PH07-T01 Define when git integration is optional vs required
- [x] PH07-T02 Define commit snapshot strategy
- [x] PH07-T03 Define commit message format
- [x] PH07-T04 Define branch and merge safety rules
- [x] PH07-T05 Define markdown human-edit workflow
- [x] PH07-T06 Define stale-state detection
- [x] PH07-T07 Define import-after-edit workflow
- [x] PH07-T08 Define conflict resolution workflow
- [x] PH07-T09 Define recovery after bad manual edits
- [x] PH07-T10 Define team and multi-operator ownership rules

## Phase 8 — Security & Abuse Resistance
- [x] PH08-T01 Define trust model
- [x] PH08-T02 Define malicious and rogue agent assumptions
- [x] PH08-T03 Define unsafe tool misuse scenarios
- [x] PH08-T04 Define file deletion and tamper detection
- [x] PH08-T05 Define audit immutability goals
- [x] PH08-T06 Define input validation hardening
- [x] PH08-T07 Define path traversal and file targeting protections
- [x] PH08-T08 Define safe export and import constraints
- [x] PH08-T09 Define denial-of-service concerns for giant plans
- [x] PH08-T10 Define sensitive metadata redaction policy
- [x] PH08-T11 Define secret-handling expectations for integrations

## Phase 9 — Metrics, Observability & Operations
- [x] PH09-T01 Define structured logging schema
- [x] PH09-T02 Define event taxonomy
- [x] PH09-T03 Define compliance metrics
- [x] PH09-T04 Define phase velocity metrics
- [x] PH09-T05 Define stuck-plan metrics
- [x] PH09-T06 Define finish-denial metrics
- [x] PH09-T07 Define reconciliation-conflict metrics
- [x] PH09-T08 Define operator dashboard requirements
- [x] PH09-T09 Define health-check strategy
- [x] PH09-T10 Define SLA targets per tool call
- [x] PH09-T11 Define support and debug bundle format

## Phase 10 — Performance & Scalability
- [x] PH10-T01 Set performance target for 100-task plans
  notes:
    - {"actor_type":"agent","text":"Defined explicit 100-task performance budgets in the Phase 10 spec and added a matching 100-task status/read-path benchmark.","created_at":"2026-03-08T08:04:33.3122339Z"}
  evidence:
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md with explicit 100-task parse/load, read-path, and interactive command budgets plus validation commands.","summary":""}
    - {"kind":"ref","ref":"Added BenchmarkStatusEvaluation100Tasks in internal/planfile/loader_bench_test.go.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkParsePlan100Tasks -benchmem =\u003e 78451 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkStatusEvaluation100Tasks -benchmem =\u003e 107681 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkParsePlan1000Tasks -benchmem =\u003e 664489 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkStatusEvaluation1000Tasks -benchmem =\u003e 1149246 ns/op","summary":""}
- [x] PH10-T02 Set performance target for 1000-task plans
  notes:
    - {"actor_type":"agent","text":"Started defining explicit 1000-task performance budgets and validating benchmark coverage.","created_at":"2026-03-08T08:08:04.103101Z"}
    - {"actor_type":"agent","text":"Defined explicit 1000-task performance budgets and added finish-check benchmark coverage to keep the target auditable.","created_at":"2026-03-08T08:08:09.9069615Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed docs/specs/phase-10-performance-scalability.md and internal/planfile/loader_bench_test.go to identify missing 1000-task finish-check coverage.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md with explicit 1000-task parse/load, read-path, finish-check, and interactive command budgets.","summary":""}
    - {"kind":"ref","ref":"Added BenchmarkFinishEvaluation1000Tasks in internal/planfile/loader_bench_test.go.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkParsePlan1000Tasks -benchmem =\u003e 651396 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkStatusEvaluation1000Tasks -benchmem =\u003e 1160015 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile -bench BenchmarkFinishEvaluation1000Tasks -benchmem =\u003e 1117473 ns/op","summary":""}
- [x] PH10-T03 Define parse and export performance goals
  notes:
    - {"actor_type":"agent","text":"Started defining parse/export performance goals with export-path benchmark coverage.","created_at":"2026-03-08T08:09:55.8803993Z"}
    - {"actor_type":"agent","text":"Defined explicit parse/export performance goals and added end-to-end 1000-task export benchmarks for markdown and JSON export paths.","created_at":"2026-03-08T08:10:00.6176046Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/api/api.go Export path and docs/specs/phase-10-performance-scalability.md to identify missing export benchmark coverage.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md with explicit parse inheritance plus 1000-task markdown and JSON export budgets and validation commands.","summary":""}
    - {"kind":"ref","ref":"Added internal/api/api_bench_test.go with BenchmarkExportMarkdown1000Tasks and BenchmarkExportJSON1000Tasks.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api -bench BenchmarkExportMarkdown1000Tasks -benchmem =\u003e 1975523 ns/op","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api -bench BenchmarkExportJSON1000Tasks -benchmem =\u003e 4122467 ns/op","summary":""}
- [x] PH10-T04 Define concurrent-agent test targets
  notes:
    - {"actor_type":"agent","text":"Started defining explicit concurrent-agent validation targets grounded in stale-state detection, reconcile conflicts, and atomic write behavior.","created_at":"2026-03-08T08:11:15.8035262Z"}
    - {"actor_type":"agent","text":"Defined explicit concurrent-agent validation targets in the Phase 10 spec and aligned them with stale-state detection, reconcile conflict handling, and atomic write behavior.","created_at":"2026-03-08T08:11:21.6487419Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/collab/git_policy.go, internal/service/reconcile.go, internal/planfile/write.go, and internal/fsutil/atomic.go for current concurrency-related behavior.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md concurrent-agent section with explicit detection, bounded latency, and integrity targets, including 10-agent mixed-load and 2-writer conflict suites.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/collab","summary":""}
    - {"kind":"ref","ref":"go test ./internal/service -run TestReconcilePlan","summary":""}
    - {"kind":"ref","ref":"go test ./internal/fsutil","summary":""}
- [x] PH10-T05 Define cache strategy if needed
  notes:
    - {"actor_type":"agent","text":"Started defining a conservative cache strategy based on current uncached plan load and evaluator paths.","created_at":"2026-03-08T08:12:03.7688296Z"}
    - {"actor_type":"agent","text":"Defined a conservative no-cache-by-default strategy with explicit invalidation and benchmark requirements for any future memoization.","created_at":"2026-03-08T08:12:09.3284964Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/planfile/loader.go, internal/service/evaluator.go, internal/api/api.go, and internal/security/guard.go to confirm current hot paths are uncached and to anchor future invalidation guidance in file fingerprinting behavior.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md cache strategy section to prohibit persistent/cross-command caches by default and define file-fingerprint-based invalidation requirements for any future memoization.","summary":""}
    - {"kind":"ref","ref":"Reviewed internal/planfile/loader.go, internal/service/evaluator.go, internal/api/api.go, and internal/security/guard.go to confirm current hot paths are uncached and to anchor future invalidation guidance in file fingerprinting behavior.","summary":""}
- [x] PH10-T06 Define database contention mitigation
  notes:
    - {"actor_type":"agent","text":"Started defining future datastore contention expectations without introducing unsupported v1 markdown-projection hacks.","created_at":"2026-03-08T08:13:23.4033621Z"}
    - {"actor_type":"agent","text":"Defined future datastore contention mitigation expectations as explicit design guardrails without overloading the v1 markdown path.","created_at":"2026-03-08T08:13:28.9074939Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/api/api.go command flow plus docs/specs/phase-09-observability.md to anchor contention guidance in current request execution and event-recording behavior.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md database contention mitigation section with optimistic-concurrency, bounded-retry, and hot-path isolation guidance.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
    - {"kind":"ref","ref":"go test ./internal/observe","summary":""}
- [x] PH10-T07 Define large audit-log handling
  notes:
    - {"actor_type":"agent","text":"Started defining large audit-log handling targets around append-only observation and out-of-band compaction.","created_at":"2026-03-08T08:13:33.8264537Z"}
    - {"actor_type":"agent","text":"Defined large audit-log handling targets that keep historical audit processing off the request hot path.","created_at":"2026-03-08T08:13:38.0749261Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/observe package behavior and current command-observation usage to ground the audit-log section in existing append-oriented event recording.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md large audit-log handling section with append-only, summary-first, and out-of-band compaction guidance.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/observe","summary":""}
- [x] PH10-T08 Define archival compaction policy
  notes:
    - {"actor_type":"agent","text":"Started defining archival compaction policy so historical snapshots stay readable without penalizing active-plan operations.","created_at":"2026-03-08T08:13:42.5416084Z"}
    - {"actor_type":"agent","text":"Defined archival compaction policy to preserve readable snapshots while keeping archive scans off the active-path hot loop.","created_at":"2026-03-08T08:13:50.7824007Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed archive/export/list behavior in internal/api to ground archive-policy guidance in current snapshot and include_archived flows.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md archival compaction policy section with active-vs-archive access rules and bounded snapshot/sidecar guidance.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH10-T09 Define worst-case benchmark scenarios
  notes:
    - {"actor_type":"agent","text":"Started defining worst-case benchmark scenarios that map the documented budgets to concrete stress workloads.","created_at":"2026-03-08T08:13:56.7072339Z"}
    - {"actor_type":"agent","text":"Defined explicit worst-case benchmark scenarios spanning deep ordering, finish denial, near-limit loads, and export stress.","created_at":"2026-03-08T08:14:01.7643696Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed the current benchmark matrix in internal/planfile and internal/api together with loader size limits to define scenario coverage gaps and explicit future stress cases.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md worst-case benchmark scenarios section with concrete scenario definitions tied to loader limits and existing benchmark families.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH10-T10 Define regression guardrails
  notes:
    - {"actor_type":"agent","text":"Started defining regression guardrails tying benchmark reruns and documented budgets to future changes in hot-path code.","created_at":"2026-03-08T08:14:07.3650495Z"}
    - {"actor_type":"agent","text":"Defined regression guardrails that require benchmark reruns and explicit evidence before hot-path scope or limits can expand.","created_at":"2026-03-08T08:14:12.0093269Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed the current benchmark matrix and loader-bound constants to turn the guardrail section into explicit change-management rules.","summary":""}
    - {"kind":"ref","ref":"Updated docs/specs/phase-10-performance-scalability.md regression guardrails section with bound-change, rerun, and slowdown-blocking rules.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
    - {"kind":"ref","ref":"go test ./internal/observe","summary":""}

## Phase 11 — Testing Strategy
- [x] PH11-T01 Unit tests for entity validation
  notes:
    - {"actor_type":"agent","text":"Started locating entity validation code paths and existing unit test coverage for plans, phases, and tasks.","created_at":"2026-03-08T08:14:40.0819519Z"}
    - {"actor_type":"agent","text":"Expanded domain entity-validation unit coverage for plan, phase, task, and audit-event validation branches.","created_at":"2026-03-08T08:15:41.8385777Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/domain validation functions and current *_test.go coverage before adding new tests.","summary":""}
    - {"kind":"ref","ref":"Updated internal/domain/model_test.go with table-driven entity-validation tests covering missing titles, invalid statuses, duplicate IDs, phase mismatch, dependency duplication/self-reference, evidence/note validation, and audit-event validation.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/domain","summary":""}
- [x] PH11-T02 Unit tests for state transitions
  notes:
    - {"actor_type":"agent","text":"Started upgrading state-transition tests from spot checks to explicit task/phase/plan transition matrices.","created_at":"2026-03-08T08:16:15.3616687Z"}
    - {"actor_type":"agent","text":"Upgraded state-transition coverage to full task/phase/plan transition matrices.","created_at":"2026-03-08T08:16:25.760722Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed internal/domain/transitions.go and existing transition tests in internal/domain/model_test.go to identify missing matrix coverage.","summary":""}
    - {"kind":"ref","ref":"Updated internal/domain/model_test.go to replace spot transition checks with explicit task, phase, and plan transition matrix tests.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/domain","summary":""}
- [x] PH11-T03 Unit tests for finish gating
  notes:
    - {"actor_type":"agent","text":"Started reviewing finish-gating logic and current evaluator tests before adding missing unit coverage.","created_at":"2026-03-08T08:16:34.2985703Z"}
    - {"actor_type":"agent","text":"Expanded finish-gating unit coverage to include plan-validation denials, blocked-work fallback behavior, and explicit next-action recommendations.","created_at":"2026-03-08T08:20:18.6766027Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/service/evaluator.go and internal/service/evaluator_test.go to identify uncovered finish-denial and acceptance branches.","summary":""}
    - {"kind":"ref","ref":"Updated internal/service/evaluator_test.go with tests for plan-invalid finish denial, blocked required work fallback to 'Resolve blocking reasons', and explicit 'Continue PH02-T01' next-action guidance.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/service","summary":""}
- [x] PH11-T04 Unit tests for dependency enforcement
  notes:
    - {"actor_type":"agent","text":"Started reviewing dependency-enforcement evaluator behavior and current tests before adding missing unit coverage.","created_at":"2026-03-08T08:20:35.4732103Z"}
    - {"actor_type":"agent","text":"Added dependency-enforcement evaluator coverage for blocked phases, disabled dependency checks, and priority-bias filtering.","created_at":"2026-03-08T08:21:26.9716858Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/service/evaluator.go and evaluator_test.go for dependency-blocking, phase-order, and next-task selection branches.","summary":""}
    - {"kind":"ref","ref":"Updated internal/service/evaluator_test.go with tests for dependency-blocked current phases, bypass behavior when RespectDependencies is false, and priority bias that still honors dependency filtering.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/service","summary":""}
- [x] PH11-T05 Unit tests for markdown parsing
  notes:
    - {"actor_type":"agent","text":"Started reviewing markdown parsing coverage to add missing unit tests for edge cases and invalid input handling.","created_at":"2026-03-08T08:21:45.2613597Z"}
    - {"actor_type":"agent","text":"Expanded markdown parsing unit coverage for invalid metadata warnings, malformed blocks, CRLF normalization, and implicit phase parsing.","created_at":"2026-03-08T08:22:40.3424073Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/planfile parser tests and loader behavior before adding targeted markdown parsing coverage.","summary":""}
    - {"kind":"ref","ref":"Updated internal/planfile/loader_test.go with parser edge-case tests covering invalid metadata/block warnings and CRLF plus implicit-phase parsing.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile","summary":""}
- [x] PH11-T06 Unit tests for markdown export
  notes:
    - {"actor_type":"agent","text":"Started reviewing markdown render/export coverage to add missing unit tests for projection stability and metadata formatting.","created_at":"2026-03-08T08:22:40.3616253Z"}
    - {"actor_type":"agent","text":"Expanded markdown export unit coverage for default-metadata omission and stable frontmatter/metadata rendering.","created_at":"2026-03-08T08:23:44.5740666Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/planfile/render.go and render_test.go before adding targeted export tests.","summary":""}
    - {"kind":"ref","ref":"Updated internal/planfile/render_test.go with export tests for omission of default metadata and stable frontmatter plus metadata ordering.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/planfile","summary":""}
- [x] PH11-T07 Unit tests for reconciliation
  notes:
    - {"actor_type":"agent","text":"Started reviewing reconcile-plan unit coverage while publication remains prepared but not yet externally executed.","created_at":"2026-03-08T08:33:50.34104Z"}
    - {"actor_type":"agent","text":"Expanded reconciliation unit coverage for default dry-run behavior, plan ID mismatch, current-phase drift, and structural phase/task conflicts.","created_at":"2026-03-08T08:34:42.1868121Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/service/reconcile.go and reconcile tests after finishing v0.1.6 publish-readiness prep.","summary":""}
    - {"kind":"ref","ref":"Updated internal/service/reconcile_test.go with dry-run default-mode coverage plus plan_id mismatch, current_phase drift, task removal, and phase addition conflict tests.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/service","summary":""}
- [x] PH11-T08 Integration tests for MCP tools
  notes:
    - {"actor_type":"agent","text":"Started reviewing API-layer tests for health/status/export/archive flows after closing reconciliation coverage.","created_at":"2026-03-08T08:34:42.2042464Z"}
    - {"actor_type":"agent","text":"Added API integration-style tests for export behavior and archived-plan listing behavior.","created_at":"2026-03-08T08:35:46.4930757Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/api/api_test.go and API handlers to identify missing unit-level contract coverage.","summary":""}
    - {"kind":"ref","ref":"Updated internal/api/api_test.go with tests covering default markdown export, JSON export writes, unsupported export formats, and archived-plan listing behavior.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH11-T09 Integration tests for database and markdown sync
  notes:
    - {"actor_type":"agent","text":"Started reviewing markdown-projection sync paths to add integration tests for persisted state after API mutations.","created_at":"2026-03-08T08:35:56.08537Z"}
    - {"actor_type":"agent","text":"Added markdown-projection sync integration tests proving persisted state survives reload after update/reset/prioritize/edit API mutations.","created_at":"2026-03-08T08:37:51.1611475Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting update/reset/prioritize/import/archive write paths and existing tests to identify missing projection-sync coverage.","summary":""}
    - {"kind":"ref","ref":"Updated internal/api/api_test.go with sync tests covering update -\u003e reload -\u003e reset and prioritize/edit -\u003e reload flows.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH11-T10 Integration tests for git snapshot path
  notes:
    - {"actor_type":"agent","text":"Started reviewing git snapshot path integration coverage after closing markdown sync tests.","created_at":"2026-03-08T08:37:51.2419028Z"}
    - {"actor_type":"agent","text":"Expanded git-path integration coverage for release/handoff git requirements and detached-head repository inspection.","created_at":"2026-03-08T08:38:29.6856122Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/collab git-policy code and current tests for workspace snapshot and archive/release workflow behavior.","summary":""}
    - {"kind":"ref","ref":"Updated internal/collab/git_policy_test.go with tests covering required git for release/handoff workflows and detached-head repo detection.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/collab","summary":""}
- [x] PH11-T11 Integration tests for denial flows
  notes:
    - {"actor_type":"agent","text":"Started reviewing denial-flow integration behavior for finish, archive, and task mutation APIs.","created_at":"2026-03-08T08:38:43.3402224Z"}
    - {"actor_type":"agent","text":"Added denial-flow integration tests for structured finish denial, archive rejection before finish readiness, and invalid update transitions.","created_at":"2026-03-08T08:39:38.7396301Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting internal/api/api.go and current API tests to identify structured denial vs error-path coverage gaps.","summary":""}
    - {"kind":"ref","ref":"Updated internal/api/api_test.go with denial-flow tests covering request_finish denial, archive denial, and invalid task-transition rejection.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH11-T12 Integration tests for reopen and reset flows
  notes:
    - {"actor_type":"agent","text":"Started reviewing reopen/reset flow coverage after closing denial-flow integration tests.","created_at":"2026-03-08T08:39:38.822548Z"}
    - {"actor_type":"agent","text":"Added reopen/reset integration coverage for successful reopen-to-in-progress flows and denial of non-terminal task resets.","created_at":"2026-03-08T08:40:32.3764093Z"}
  evidence:
    - {"kind":"ref","ref":"Inspecting reset-related service and API tests to identify missing terminal-task reopen flow coverage.","summary":""}
    - {"kind":"ref","ref":"Updated internal/api/api_test.go with reset integration tests covering done-task reopen to in_progress, persisted reset_reason, and denial of non-terminal reset attempts.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api","summary":""}
- [x] PH11-T13 Load tests for 10 concurrent agents
  notes:
    - {"actor_type":"agent","text":"Started scoping concurrent-agent load testing after closing the denial and reset integration tasks.","created_at":"2026-03-08T08:40:32.484693Z"}
    - {"actor_type":"agent","text":"Added concurrent mixed-traffic load coverage and fixed Windows atomic-replace overlap failures exposed by the new test.","created_at":"2026-03-08T08:44:04.6378962Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewing Phase 10 concurrency targets, existing benchmarks, and API write/read paths to define the smallest credible PH11-T13 load harness.","summary":""}
    - {"kind":"ref","ref":"Updated internal/api/api_test.go with TestConcurrentMixedTrafficPreservesParseablePlan covering 10 concurrent agents against a 100-task plan with mixed read/write traffic.","summary":""}
    - {"kind":"ref","ref":"Hardened internal/fsutil/atomic.go Windows fallback swap path with bounded retries for transient sharing violations.","summary":""}
    - {"kind":"ref","ref":"Added bounded transient plan-load retries in internal/api/api.go for overlapping replace windows.","summary":""}
    - {"kind":"ref","ref":"go test ./internal/api -run TestConcurrentMixedTrafficPreservesParseablePlan -count=5","summary":""}
    - {"kind":"ref","ref":"go test ./...","summary":""}
    - {"kind":"ref","ref":"npm test","summary":""}
- [ ] PH11-T14 Load tests for large plans
- [ ] PH11-T15 End-to-end tests from plan creation to finish approval
- [ ] PH11-T16 Adversarial tests for rogue and malformed edits
- [ ] PH11-T17 Recovery tests for crash during write
- [ ] PH11-T18 Snapshot tests for exported plans
- [ ] PH11-T19 Compatibility tests for common agent workflows
- [ ] PH11-T20 Performance regression suite
- [x] PH11-T21 Integration tests for agent bootstrap MCP guidance tool

## Phase 12 — Developer Experience & CLI
- [x] PH12-T01 Define local setup flow
  notes:
    - {"actor_type":"agent","text":"Started simplifying end-user local setup flow with emphasis on Claude Code, Auggie-compatible setup, and generic MCP client onboarding.","created_at":"2026-03-08T09:01:33.4828827Z"}
    - {"actor_type":"agent","text":"Simplified the end-user local setup flow with npm-first onboarding, Claude Code copy/paste config, Auggie-compatible guidance, and generic MCP client instructions.","created_at":"2026-03-08T09:02:55.189686Z"}
  evidence:
    - {"kind":"ref","ref":"Reviewed current README onboarding sections and client examples to identify where the fast add/install path is still too implicit.","summary":""}
    - {"kind":"ref","ref":"Updated README.md with a top-level fastest-setup path for Claude Code, Auggie, and other coding agents, plus clearer MCP config snippets and workspace-root guidance.","summary":""}
    - {"kind":"ref","ref":"Updated package.json description and keywords to improve discovery for Claude Code/Auggie/coding-agent users.","summary":""}
    - {"kind":"ref","ref":"npm test","summary":""}
    - {"kind":"ref","ref":"npm run pack:check","summary":""}
- [ ] PH12-T02 Define CLI admin commands
- [ ] PH12-T03 Define inspect and debug commands
- [ ] PH12-T04 Define import and export commands
- [ ] PH12-T05 Define archive and restore commands
- [ ] PH12-T06 Define dry-run validation commands
- [ ] PH12-T07 Define example prompts for agents
- [ ] PH12-T08 Define sample plans
- [ ] PH12-T09 Define bootstrap templates by project type
- [ ] PH12-T10 Define operator troubleshooting commands
- [x] PH12-T11 Define agent bootstrap guidance flow for MCP clients
- [x] PH12-T12 Default no-arg CLI launch to MCP serve for client compatibility
- [x] PH12-T13 Resolve unsafe MCP launch cwd via project-local workspace or user-home fallback

## Phase 13 — Documentation
- [ ] PH13-T01 Write product overview
- [ ] PH13-T02 Write quickstart
- [ ] PH13-T03 Write MCP tool reference
- [ ] PH13-T04 Write plan format spec
- [ ] PH13-T05 Write sync and reconciliation guide
- [ ] PH13-T06 Write enforcement guide
- [ ] PH13-T07 Write agent prompt template guide
- [ ] PH13-T08 Write operator troubleshooting guide
- [ ] PH13-T09 Write git integration guide
- [ ] PH13-T10 Write security model
- [ ] PH13-T11 Write FAQ
- [ ] PH13-T12 Write migration and versioning notes
- [x] PH13-T13 Write agent bootstrap MCP guide reference
- [x] PH13-T15 Document workspace-root override and user-home fallback behavior
- [x] PH13-T14 Document no-arg MCP server launch behavior for local client configs

## Phase 14 — Packaging & Release
- [ ] PH14-T01 Define repository structure
- [ ] PH14-T02 Define versioning and release policy
- [ ] PH14-T03 Define Docker packaging
- [ ] PH14-T04 Define local dev container strategy
- [ ] PH14-T05 Define hosted deployment recipe
- [ ] PH14-T06 Define release checklist
- [ ] PH14-T07 Define example configs
- [ ] PH14-T08 Define beta feedback workflow
- [ ] PH14-T09 Define v1 acceptance criteria
- [ ] PH14-T10 Define post-v1 roadmap
- [x] PH14-T11 Align Go module path with public GitHub install path
- [x] PH14-T12 Publish end-user setup guide for major MCP clients
- [x] PH14-T13 Package and publish thin npm launcher for local install guidance
- [ ] PH14-T14 Push release-ready main branch and publish public package artifacts
- [x] PH14-T15 Define platform binary release matrix for npm-native installation
- [x] PH14-T16 Add reproducible cross-platform build packaging for release binaries
- [x] PH14-T17 Implement npm install-time native binary downloader and launcher flow
- [x] PH14-T18 Add integrity and version validation for downloaded npm-installed binaries
- [x] PH14-T19 Document dual install paths: Go native and npm-native binary install
- [x] PH14-T20 Publish release notes for npm-native install support and follow-up release
- [x] PH14-T21 Add repository license file and align published package metadata
- [x] PH14-T22 Validate clean reinstall flow after removing existing local Warden binaries
- [ ] PH14-T23 Publish MCP startup compatibility patch release for Codex-style clients
- [ ] PH14-T24 Publish unsafe-launch workspace fallback patch release
- [x] PH14-T25 Harden plan_path handling against unsafe absolute client-expanded defaults
- [ ] PH14-T26 Publish plan_path hardening patch release
- [x] PH14-T27 Package Claude Code git-marketplace plugin manifests for repo-root install
- [x] PH14-T28 Document Claude plugin marketplace install flow alongside existing MCP setup guidance
- [x] PH14-T29 Validate Claude plugin manifests, README instructions, and package metadata consistency
- [x] PH14-T30 Fix Claude marketplace manifest schema compliance for owner metadata and repo-root plugin source paths
- [x] PH14-T31 Align Claude plugin MCP config layout with repo-root plugin conventions and validate install expectations
- [x] PH14-T32 Clarify client UX expectations so Auggie users do not expect slash commands from the Warden MCP tool server

