# Phase 3 Tool Contracts

Status: implementation-aligned draft
Plan coverage: implemented baseline for PH03-T01 through PH03-T15

## Purpose

Phase 3 turns the public Warden MCP schema spec into strongly typed Go request and response contracts.

## Implemented package

- `internal/mcp/contracts/types.go`
- `internal/mcp/contracts/types_test.go`

## Coverage summary

The Go contract layer now includes typed request and response data for:

- `init_plan`
- `validate_plan`
- `get_status`
- `update_task`
- `get_next_task`
- `request_finish`
- `edit_plan`
- `reset_task`
- `prioritize_tasks`
- `list_plans`
- `archive_plan`
- `import_plan`
- `export_plan`
- `reconcile_plan`

It also includes:

- shared enums for export, import, enforcement, source, validation, edit, and reconcile modes
- shared summary objects for plans, phases, tasks, blocking reasons, and error payloads
- a generic `ToolResponseEnvelope[T]`
- reserved error-code constants matching the public schema spec

## Current implementation notes

- contract status fields reuse canonical domain enums from `internal/domain`
- JSON field names mirror `docs/specs/warden-mcp-tool-schemas.md`
- timestamp fields remain RFC 3339 strings at the contract boundary
- raw object payloads use `json.RawMessage` where the schema intentionally permits open-ended shapes

## Validation performed

- `gofmt -w internal/mcp/contracts/*.go`
- `go test ./...`

## Next phase dependency

These contracts are intended to be consumed by the upcoming storage and service layers, which must preserve the same field names and error-code semantics.