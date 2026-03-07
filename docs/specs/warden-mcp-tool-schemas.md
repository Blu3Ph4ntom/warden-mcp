# Warden MCP Tool Schemas

This document defines the canonical request and response schemas for the public Warden MCP tool surface.

## Schema Conventions

- All tools accept a single JSON object as input.
- All tools return a single JSON object as output.
- All timestamps use RFC 3339 UTC strings.
- All IDs are immutable strings.
- Unless explicitly noted otherwise, every schema uses `additionalProperties: false`.
- Markdown is a human interface; canonical state belongs to the Warden datastore.

## Shared Enums

### `plan_status`
- `draft`
- `active`
- `blocked`
- `completed`
- `archived`

### `phase_status`
- `not_started`
- `in_progress`
- `blocked`
- `completed`

### `task_status`
- `not_started`
- `in_progress`
- `done`
- `blocked`
- `cancelled`
- `waived`

### `priority`
- `P0`
- `P1`
- `P2`
- `P3`

### `actor_type`
- `agent`
- `human`
- `system`

### `export_format`
- `markdown`
- `json`
- `csv`

## Shared Objects

### `ErrorObject`
- `code: string`
- `message: string`
- `retryable: boolean`
- `details: object`

### `ValidationIssue`
- `severity: "error" | "warning"`
- `code: string`
- `message: string`
- `path: string`

### `BlockingReason`
- `code: string`
- `message: string`
- `phase_id?: string`
- `task_id?: string`

### `TaskSummary`
- `task_id: string`
- `phase_id: string`
- `title: string`
- `status: task_status`
- `priority: priority`
- `depends_on: string[]`
- `updated_at: string`

### `PhaseSummary`
- `phase_id: string`
- `title: string`
- `status: phase_status`
- `task_count: integer`
- `completed_task_count: integer`
- `blocked_task_count: integer`

### `PlanSummary`
- `plan_id: string`
- `title: string`
- `status: plan_status`
- `version: string`
- `current_phase_id: string | null`
- `total_tasks: integer`
- `completed_tasks: integer`
- `can_finish: boolean`
- `updated_at: string`

### `ToolResponseEnvelope`
- `ok: boolean`
- `tool: string`
- `timestamp: string`
- `plan_id?: string`
- `warnings: ValidationIssue[]`
- `error?: ErrorObject`
- `data: object`

## Tool Schemas

## `init_plan`

### Input
- `type: object`
- `required: ["title"]`
- `properties:`
  - `title: string`
  - `goal?: string`
  - `source_text?: string`
  - `phases?: array` of phase objects where each phase requires `title` and `tasks`
  - `plan_id?: string`
  - `version?: string`
  - `enforcement_profile?: "strict" | "balanced" | "advisory"`
  - `source_type?: "mcp" | "markdown" | "json" | "human"`
  - `create_markdown_projection?: boolean`

### Output `data`
- `plan: PlanSummary`
- `phases: PhaseSummary[]`
- `tasks: TaskSummary[]`
- `validation_issues: ValidationIssue[]`
- `normalized: boolean`

## `validate_plan`

### Input
- `type: object`
- `required: ["plan"]`
- `properties:`
  - `plan: object`
  - `mode?: "strict" | "import" | "lint"`

### Output `data`
- `valid: boolean`
- `issues: ValidationIssue[]`
- `normalized_counts:`
  - `phase_count: integer`
  - `task_count: integer`

## `get_status`

### Input
- `type: object`
- `required: ["plan_id"]`
- `properties:`
  - `plan_id: string`
  - `include_tasks?: boolean`
  - `include_completed_tasks?: boolean`
  - `include_audit_summary?: boolean`

### Output `data`
- `plan: PlanSummary`
- `phases: PhaseSummary[]`
- `tasks?: TaskSummary[]`
- `blocking_reasons: BlockingReason[]`
- `next_task_id?: string`
- `stalled: boolean`
- `stalled_since?: string`

## `update_task`

### Input
- `type: object`
- `required: ["plan_id", "task_id", "status"]`
- `properties:`
  - `plan_id: string`
  - `task_id: string`
  - `status: task_status`
  - `note?: string`
  - `evidence?: string[]`
  - `reason?: string`
  - `actor_type?: actor_type`

### Output `data`
- `task: TaskSummary`
- `phase: PhaseSummary`
- `plan: PlanSummary`
- `transition_accepted: boolean`

## `get_next_task`

### Input
- `type: object`
- `required: ["plan_id"]`
- `properties:`
  - `plan_id: string`
  - `respect_phase_order?: boolean`
  - `respect_dependencies?: boolean`
  - `priority_bias?: priority`

### Output `data`
- `next_task?: TaskSummary`
- `reason: string`
- `blocked: boolean`
- `blocking_reasons: BlockingReason[]`

## `request_finish`

### Input
- `type: object`
- `required: ["plan_id"]`
- `properties:`
  - `plan_id: string`
  - `actor_type?: actor_type`
  - `summary?: string`

### Output `data`
- `can_finish: boolean`
- `plan: PlanSummary`
- `blocking_reasons: BlockingReason[]`
- `incomplete_task_ids: string[]`
- `next_required_actions: string[]`
- `recommended_next_task_id?: string`

### Mandatory denial conditions
- any required task is non-terminal
- any required phase is incomplete
- any unresolved dependency exists
- any validation error exists
- any sync conflict exists
- any blocked task lacks explicit waiver rationale

## `edit_plan`

### Input
- `type: object`
- `required: ["plan_id", "operation"]`
- `properties:`
  - `plan_id: string`
  - `operation: "add_phase" | "add_task" | "update_task_fields" | "move_task" | "split_task" | "reprioritize_task" | "add_dependency" | "remove_dependency" | "waive_task" | "cancel_task"`
  - `target_id?: string`
  - `payload?: object`
  - `reason?: string`

### Output `data`
- `plan: PlanSummary`
- `changed_ids: string[]`
- `diff_summary: string`

## `reset_task`

### Input
- `type: object`
- `required: ["plan_id", "task_id"]`
- `properties:`
  - `plan_id: string`
  - `task_id: string`
  - `status?: "not_started" | "in_progress"`
  - `reason: string`

### Output `data`
- `task: TaskSummary`
- `plan: PlanSummary`

## `prioritize_tasks`

### Input
- `type: object`
- `required: ["plan_id", "updates"]`
- `properties:`
  - `plan_id: string`
  - `updates: array` of objects with:
    - `task_id: string`
    - `priority: priority`

### Output `data`
- `updated_task_ids: string[]`
- `plan: PlanSummary`

## `list_plans`

### Input
- `type: object`
- `required: []`
- `properties:`
  - `status?: plan_status`
  - `include_archived?: boolean`

### Output `data`
- `plans: PlanSummary[]`

## `archive_plan`

### Input
- `type: object`
- `required: ["plan_id"]`
- `properties:`
  - `plan_id: string`
  - `reason?: string`
  - `create_final_export?: boolean`

### Output `data`
- `archived: boolean`
- `plan: PlanSummary`
- `archive_path?: string`

## `import_plan`

### Input
- `type: object`
- `required: ["format", "content"]`
- `properties:`
  - `format: "markdown" | "json"`
  - `content: string`
  - `plan_id?: string`
  - `mode?: "create" | "merge" | "replace"`

### Output `data`
- `plan: PlanSummary`
- `issues: ValidationIssue[]`
- `conflicts_detected: boolean`

## `export_plan`

### Input
- `type: object`
- `required: ["plan_id", "format"]`
- `properties:`
  - `plan_id: string`
  - `format: export_format`
  - `include_audit_summary?: boolean`

### Output `data`
- `format: export_format`
- `content: string`
- `content_path?: string`

## `reconcile_plan`

### Input
- `type: object`
- `required: ["plan_id", "markdown_content"]`
- `properties:`
  - `plan_id: string`
  - `markdown_content: string`
  - `mode?: "dry_run" | "apply"`

### Output `data`
- `reconciled: boolean`
- `conflicts: array` of objects with:
  - `code: string`
  - `message: string`
  - `target_id?: string`
- `changed_ids: string[]`
- `plan: PlanSummary`

## Error Codes

The following error codes should be reserved:

- `PLAN_NOT_FOUND`
- `PLAN_INVALID`
- `PHASE_INVALID`
- `TASK_NOT_FOUND`
- `TASK_TRANSITION_INVALID`
- `DEPENDENCY_VIOLATION`
- `FINISH_DENIED`
- `SYNC_CONFLICT`
- `IMPORT_INVALID`
- `EXPORT_FAILED`
- `ARCHIVE_DENIED`
- `INTERNAL_ERROR`

## Implementation Notes

- `request_finish` must be authoritative, not advisory.
- `update_task` must update phase and plan rollups atomically.
- `reconcile_plan` must never silently discard task IDs.
- `init_plan` should reject shallow plans unless explicitly allowed by policy.
- Every mutation tool should emit an audit event even if the change is denied.