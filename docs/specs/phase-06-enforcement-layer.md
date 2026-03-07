# Phase 6 Enforcement Layer Design

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH06-T01 through PH06-T12

## Purpose

The enforcement layer defines how a compliant integrated agent is constrained to plan first, update governed state during execution, and stop claiming completion early. It is the runtime policy bridge between the MCP contracts and actual agent behavior.

## System prompt template

The runtime prompt should always communicate:

- Warden state is authoritative for progress and finish approval
- the agent must read the active plan before implementation work
- the agent must work one scoped task at a time unless the plan explicitly allows parallel work
- the agent must record progress after meaningful task completion
- denied finish requests require continued execution, not rhetorical justification
- ambiguity requires clarification or a plan update before proceeding

## Agent operating rules

- never claim `done` without matching Warden state
- never skip validation when the current task requires executable changes
- never silently rewrite task IDs, dependency edges, or phase boundaries
- prefer smallest safe next action over broad speculative edits
- if blocked, record blocker, proposed resolution, and next viable task

## Pre-finish validation flow

Before any finish claim, the runtime wrapper should force:

1. `get_status`
2. optional `validate_plan` if drift or import activity occurred
3. `request_finish`
4. denial handling if `can_finish` is false

The user-visible finish verdict must come from `request_finish`, not from the agent's own confidence.

## Forced next-step behavior after denial

When finish is denied, the runtime must require the agent to do one of these next actions:

- continue the recommended next task
- resolve a listed blocking reason
- request clarification from a human
- update the plan through governed tools if scope changed

The agent must not respond with a vague completion summary as a substitute for one of those actions.

## Loop detection rules

The runtime should detect ineffective loops such as:

- repeated finish attempts without meaningful state change
- repeated retries of the same denied mutation
- repeated status queries without acting on returned blockers
- cycling between blocked tasks without escalation

Loop detection keys off repeated tool patterns, unchanged blocking reasons, and unchanged changed-ID sets.

## Repeated-failure escalation

- after a small number of repeated denials on the same blocker, force a summary of the blocker and proposed fix
- if the blocker persists, require a plan edit, reset, waiver, or human clarification
- if the agent cannot propose a concrete next action, route to human review

## Max-iteration policy

- impose per-task iteration ceilings for repeated failed attempts
- impose per-plan ceilings for repeated finish-denial cycles
- after the ceiling is reached, require escalation instead of silent continuation

The exact numeric thresholds can remain configurable, but the behavior must be strict.

## Blocked-task escalation policy

- blocked required tasks must prevent finish approval
- blocked tasks need explicit rationale recorded in canonical state
- if recovery is possible, the runtime should prefer reset/reopen guidance
- if recovery is not possible, the runtime should require cancellation or waiver with reason and actor attribution

## Human intervention triggers

Escalate to a human when:

- reconciliation reports unresolved conflicts
- a required task is blocked and the agent lacks authority to waive or cancel it
- repeated validation failures persist after retries
- a plan edit would materially change scope or acceptance criteria
- evidence is missing for a task the agent wants to mark done
- the agent detects ambiguity it cannot resolve from canonical state

## Compliance scoring

The runtime may compute a compliance score for observability, based on:

- plan-first adherence
- successful use of governed mutation tools
- validation-before-finish behavior
- responsiveness to denial reasons
- rate of repeated invalid actions
- frequency of required human escalations

This score is advisory telemetry only; finish gating remains hard policy.

## Runtime wrapper responsibilities

The wrapper around the model/tool loop is responsible for:

- injecting the current Warden operating rules into the prompt
- deciding when to force status/finish checks
- refusing to accept free-form completion claims that bypass Warden
- logging denial and escalation events
- passing structured next-step instructions back to the agent after tool results
- preventing tool-free path changes that would desynchronize canonical state

## Anti-hallucination safeguards

- require tool evidence for any claimed task completion or finish readiness
- prefer denial on missing data instead of optimistic assumptions
- require exact blocker and task references in finish-denial follow-up
- force the agent to distinguish facts from proposals when state is uncertain
- surface stale-state or sync-conflict errors explicitly
- require re-checks after external edits or reconcile attempts

## Next implementation step

After this design phase, the project can move into collaboration and git-traceability design while keeping the enforcement wrapper assumptions stable for future runtime code.