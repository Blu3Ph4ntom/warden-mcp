# Implementation Baseline

Status: active implementation decision

## Runtime stack

Warden MCP implementation work is now Go-first.

Current baseline choices:

- language: Go
- initial dependency policy: standard library first
- packaging baseline: Go module rooted at repository root
- near-term architecture: internal domain, MCP contract, and service packages, with MCP transport to follow

## Why this baseline

This keeps the early platform simple, auditable, and easy to validate while the domain model and storage rules are still being defined.

## Current repository direction

- canonical domain logic lives under `internal/`
- public MCP request/response types now live under `internal/mcp/contracts`
- public-facing executable entrypoints should live under `cmd/`
- docs and specs remain under `docs/specs/`
- markdown remains a projection, not the only source of truth

## Migration note

The earlier Python scaffold was a temporary bootstrap and is being removed in favor of the Go implementation.
