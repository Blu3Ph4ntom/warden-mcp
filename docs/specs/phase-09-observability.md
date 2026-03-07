# Phase 9 Metrics, Observability & Operations

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH09-T01 through PH09-T11

## Purpose

Warden should emit structured, low-ambiguity runtime events so operators can understand what command or MCP action ran, whether it was accepted, what plan and task it targeted, and why it failed when denied.

## Structured logging schema

Baseline runtime events should include:

- event kind
- timestamp
- request path or command name
- plan ID when available
- task ID or phase ID when available
- actor type when available
- accepted or denied outcome
- duration in milliseconds
- compact message and error code when applicable

## Event taxonomy

- lifecycle events: startup, shutdown, initialize, initialized
- command events: status, next, finish, update
- tool events: tools/list, tools/call
- denial events: invalid input, unsafe path, illegal transition, finish denied
- audit-adjacent events: task state accepted, plan read, plan write

## Compliance metrics

- finish denial count
- invalid transition count
- unsafe path denial count
- successful governed mutation count

## Phase velocity metrics

- completed task count over time
- current phase duration once persistence exists
- average time between accepted task transitions

## Stuck-plan metrics

- time since last accepted task transition
- time spent in current phase with no completed tasks
- blocked task count per phase

## Finish-denial metrics

- denial reason counts by code
- repeated finish-request attempts before acceptance
- number of incomplete required tasks at denial time

## Reconciliation-conflict metrics

- stale-state detection count
- import conflict count
- manual repair count once recovery flows exist

## Operator dashboard requirements

- latest command and tool events
- current plan summary
- current next-task recommendation
- recent denials and error codes

## Health-check strategy

- process starts successfully
- plan path resolves within workspace rules
- plan loads and validates without fatal parse errors
- MCP initialize and tools/list succeed on stdio transport

## SLA targets per tool call

- status and next should remain low latency on local plans
- finish should stay bounded by plan validation cost
- updates should remain bounded by read-modify-write of a single plan projection

## Support and debug bundle format

- recent structured events
- current plan summary
- version and protocol info
- redacted errors only

## Next implementation step

Add a small structured event emitter for CLI and MCP requests, then use it to log initialize, tools/list, tools/call, command execution, denials, and plan mutations.