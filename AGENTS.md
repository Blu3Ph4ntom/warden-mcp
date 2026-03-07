# AGENTS.md

Public agent instructions for contributors working in this repository.

## Purpose

This repository contains the design and implementation work for **Warden MCP**: a plan-governance and completion-gating system for AI coding agents.

If you are an autonomous or semi-autonomous coding agent, you are expected to behave conservatively, plan first, and avoid claiming completion early.

## Core Rule

Do **not** start implementation work until there is a concrete, multi-phase plan with multiple tasks per phase.

In this repository, the operational plan lives at:

- `.agent/PLAN.md`

That file is the working execution plan. This `AGENTS.md` file is public guidance for contributors and external agents.

## Required Workflow

1. Read `.agent/PLAN.md` before making changes.
2. Confirm the current phase and task you are working on.
3. If the requested work is not represented in the plan, update the plan before coding.
4. Work on one clearly scoped task at a time unless the plan explicitly allows parallel work.
5. Record progress after each meaningful task completion.
6. Before claiming the work is complete, run the equivalent of a finish gate check.

## Non-Negotiable Rules

- Do not claim "done" based on intuition.
- Do not skip planning, research, or design steps just because the implementation feels obvious.
- Do not silently rewrite or renumber task IDs.
- Do not collapse multiple unfinished tasks into a vague summary.
- Do not fake test results, tool outputs, or validation outcomes.
- Do not mark a task complete without evidence that its acceptance criteria were satisfied.
- Do not bypass plan updates when task scope changes.

## Planning Expectations

Plans in this repository should:

- have multiple phases,
- have multiple tasks per phase,
- use stable IDs,
- define clear entry and exit criteria,
- distinguish required work from optional work,
- be detailed enough to prevent premature completion claims.

If a request is underspecified, expand the plan before implementation.

## Implementation Expectations

When implementation begins, prefer the following architectural assumptions unless the plan says otherwise:

- Canonical state should be stored in a durable datastore.
- Markdown should be treated as a human-readable projection, not the sole source of truth.
- Completion gating must be strict.
- Public tool contracts should be explicit and stable.
- Auditability is more important than cleverness.

## Editing Rules

- Make the smallest safe change that satisfies the current task.
- Preserve existing structure unless the current task requires structural change.
- Keep docs aligned with behavior.
- When modifying plan-related files, preserve IDs and phase boundaries.
- Avoid introducing hidden coupling or undocumented conventions.

## Testing and Validation

If you change executable code:

- run the smallest relevant tests first,
- fix failures before widening scope,
- summarize what you validated,
- do not report success without actual validation.

If you only change planning or documentation artifacts, validate for internal consistency.

## Public vs Private Files

### Public
- `AGENTS.md`
- `docs/`
- source code and tests

### Operational / internal-facing
- `.agent/PLAN.md`

Do not put secrets or sensitive internal notes into public files.

## Preferred Contribution Style

- Be explicit.
- Be auditable.
- Be incremental.
- Be honest about uncertainty.
- Prefer reliable systems over optimistic assumptions.

## Completion Gate

Before declaring any task or feature complete, ensure all of the following are true:

- the relevant plan tasks are actually complete,
- validation has been performed where applicable,
- documentation is updated if behavior changed,
- no known blocking issues remain hidden,
- the result matches the planned scope.

If any of these are false, you are not done.

## When in Doubt

If there is ambiguity:

1. stop,
2. clarify scope,
3. update the plan,
4. then continue.

Warden exists to prevent agents from hallucinating completion. Act accordingly.