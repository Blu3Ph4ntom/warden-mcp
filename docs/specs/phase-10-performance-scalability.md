# Phase 10 Performance & Scalability

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH10-T01 through PH10-T10

## Purpose

Warden should stay responsive for normal local development plans, fail early on obviously oversized inputs, and keep parsing and evaluation costs bounded enough that governance checks do not become a workflow bottleneck.

## Performance targets for 100-task plans

- status and next-task evaluation should feel effectively immediate on local hardware
- single-plan load plus parse should remain comfortably below interactive latency budgets
- task updates should remain bounded by one read-modify-write cycle of a small markdown plan

## Performance targets for 1000-task plans

- status, next-task, and finish checks should remain practical for local use
- parsing should remain linear in plan size with no hidden quadratic scans across sections
- large but valid plans should still be denied or processed predictably rather than timing out unpredictably

## Parse and export performance goals

- reject obviously oversized plan files before expensive parsing work
- normalize line endings once per load
- avoid repeated whole-file transformations where a single pass is enough
- keep markdown rewrite scope narrow for single-task status changes

## Concurrent-agent targets

- concurrent operators should fail fast on stale state instead of causing long retries
- read-only status calls should remain cheap even while another operator is preparing an update

## Cache strategy

- do not add persistent caches until profiling shows a real bottleneck
- prefer bounded in-memory memoization only after correctness rules are stable

## Database contention mitigation

- canonical datastore contention will matter later; v1 markdown projection work should avoid pretending to solve it with local hacks

## Large audit-log handling

- append-oriented logs should rotate or compact outside hot request paths once persistence exists

## Archival compaction policy

- archived plans should remain readable, but active-path tooling should not pay full historical scan cost on every request

## Worst-case benchmark scenarios

- 100-task plan with many completed phases
- 1000-task plan with deep phase ordering
- large plan near size-limit threshold
- finish-denial evaluation with many incomplete required tasks

## Regression guardrails

- keep generated large-plan benchmarks in the repo
- treat file-size and line-count limits as explicit safety/perf boundaries
- expand scope only when tests show current limits are insufficient

## Next implementation step

Enforce bounded plan load limits, then keep generated 100-task and 1000-task parse/evaluation benchmarks so future changes can be checked against the target workload sizes.