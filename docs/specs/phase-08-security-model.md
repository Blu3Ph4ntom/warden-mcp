# Phase 8 Security & Abuse Resistance

Status: implementation-aligned design draft
Plan coverage: documented baseline for PH08-T01 through PH08-T11

## Purpose

Warden must assume that operators, agents, and surrounding tools can make unsafe or adversarial requests. The system should prefer explicit denial, narrow file access, auditable behavior, and bounded work over convenience.

## Trust model

- the local operator is trusted to choose the workspace, but not trusted to always provide safe paths or safe content
- integrated agents are semi-trusted and must be constrained by policy instead of intent assumptions
- imported markdown, CLI flags, and file paths are all untrusted inputs
- canonical plan state and audit records are more trusted than markdown projections

## Malicious and rogue agent assumptions

Warden should assume an agent may:

- try to mark work complete early
- request paths outside the workspace
- attempt to hide unsafe changes in large diffs
- feed malformed or overlong plan content
- include secrets or sensitive metadata in notes, evidence, or exports

## Unsafe tool misuse scenarios

- reading or targeting files outside the intended workspace
- overwriting projections from stale state
- exporting or importing files with ambiguous ownership or location
- invoking write paths without enough validation context
- using excessive input size to degrade responsiveness

## File deletion and tamper detection

- treat missing expected plan files as explicit errors, not auto-recreate events
- record file fingerprints or equivalent metadata before sensitive import/export actions once persistence exists
- compare size, timestamp, or content hash when reconciling edited projections
- never silently accept ID loss, major truncation, or structural section deletion

## Audit immutability goals

- audit records should be append-oriented
- denial and conflict events must be preserved even when the operator retries
- audit mutations should be attributable to actor type and request path
- exported markdown should never be the only record of governance decisions

## Input validation hardening

- validate all IDs, statuses, actor types, and enum-like inputs against strict allow-lists
- reject empty required fields before deeper processing
- normalize line endings and trim surrounding whitespace where safe
- bound payload sizes before parsing or hashing

## Path traversal and file targeting protections

- resolve candidate paths against a configured workspace root
- reject any path that escapes the workspace after cleaning and absolute resolution
- require explicit allowed file types for plan-facing flows in v1
- avoid implicit writes to derived sibling paths without explicit policy

## Safe export and import constraints

- exports should only target approved workspace-relative locations
- imports should require known-safe file types and size limits
- dry-run validation should be available before any apply path
- imported content should be parsed before any destructive action occurs

## Denial-of-service concerns for giant plans

- set bounded file-size and line-count thresholds for v1
- avoid repeated full-file hashing or reparsing when a cheaper guard can fail early
- surface “input too large” as an explicit denial instead of risking unstable runtime behavior

## Sensitive metadata redaction policy

- avoid echoing full secret-like tokens in logs or tool output
- redact known credential shapes in warnings, errors, and debug bundles
- show only enough token context for operators to identify the source safely

## Secret-handling expectations for integrations

- Warden should not persist raw secrets in plan state
- integrations must pass secret references or environment lookups rather than embedding credentials into plan artifacts
- debugging output should default to redacted values whenever secret-like text is detected

## Next implementation step

After this design phase, the runtime should enforce workspace-bounded path resolution for plan files, bound file reads, and provide secret-redaction helpers for future import/export and audit paths.