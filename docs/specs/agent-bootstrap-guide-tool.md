# Agent Bootstrap Guide Tool Design

Date: 2026-03-07
Status: approved for implementation

## Purpose

Define a single MCP tool that gives coding agents a reliable, end-to-end playbook for using Warden MCP safely.

## Scope

In scope:
- one new MCP tool: `get_agent_guide`
- agent-facing guidance only
- structured output suitable for MCP clients and LLM agents
- optional light live context from the active plan
- tests and public tool-schema documentation

Out of scope:
- human/operator help via MCP tool calls
- dynamic coaching that mutates plans or tasks
- replacing README/CLI docs

## Goals

The tool should help an agent:
1. understand the safest Warden workflow,
2. choose the right next MCP call,
3. avoid claiming completion early,
4. use `plan_id` as an identity guard,
5. follow the finish gate before declaring work complete.

## Tool Contract

### Name
`get_agent_guide`

### Input
- `plan_path?: string`
- `plan_id?: string`
- `detail_level?: "brief" | "full"`

Defaults:
- `plan_path = ".agent/PLAN.md"`
- `detail_level = "full"`

### Output Envelope
Return the normal Warden tool envelope with:
- `ok: true`
- `tool: "get_agent_guide"`
- `timestamp`
- optional `plan_id`
- `warnings: []`
- `data`

### Output Data
- `guide_version: string`
- `summary: string`
- `core_rules: string[]`
- `recommended_sequence: object[]`
- `tool_playbook: object[]`
- `finish_gate_rules: string[]`
- `example_calls: object[]`
- `live_context?: object`

## Live Context

If the active plan can be read, include light live context such as:
- current `plan_id`
- current `current_phase_id`
- active plan status
- suggested next calls

If no plan exists, still return the guide successfully with a warning.

## Detail Levels

### brief
Compact startup guidance for agents that only need the essential workflow.

### full
Includes the full tool playbook, finish-gate rules, and example calls.

## Behavioral Rules

- never fail just because a plan is absent
- do not mutate plan state
- do not hide `plan_id` mismatch information when detectable
- keep the guidance deterministic and auditable
- keep the live context informational, not authoritative over core rules

## Validation

Add tests for:
- `tools/list` includes `get_agent_guide`
- `tools/call` returns the guide in `brief` mode
- `tools/call` returns the guide in `full` mode
- tool succeeds without an active plan
- tool includes live context with a real plan

## Documentation Impact

Update the MCP tool schema reference so the new tool is part of the public contract.

