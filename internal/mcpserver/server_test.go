package mcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/api"
)

func TestServerHandlesInitializeListAndStatusToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH02
can_finish: false
completed_tasks: 2
---

# Sample Plan

## Phase 1 — Setup
- [x] PH01-T01 create repo
- [x] PH01-T02 add tests

## Phase 2 — Build
- [/] PH02-T01 start implementation
- [ ] PH02-T02 finish implementation
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion, "clientInfo": map[string]any{"name": "test", "version": "1.0.0"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/list"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{"name": "health_check", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md")}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": map[string]any{"name": "get_status", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "include_tasks": true}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 4 {
		t.Fatalf("expected 4 responses, got %d", len(frames))
	}
	if !strings.Contains(string(frames[1]), "\"tools\"") || !strings.Contains(string(frames[1]), "health_check") {
		t.Fatalf("expected tools/list payload, got %s", frames[1])
	}
	if !strings.Contains(string(frames[2]), "health_check") || !strings.Contains(string(frames[2]), "plan parsed successfully") {
		t.Fatalf("expected health tool response, got %s", frames[2])
	}
	if !strings.Contains(string(frames[3]), "sample-plan") || !strings.Contains(string(frames[3]), "get_status") {
		t.Fatalf("expected status tool response, got %s", frames[3])
	}
}

func TestServerHandlesExportToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: export-plan
title: Export Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Export Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "export_plan", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "format": "json"}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	if !strings.Contains(string(frames[1]), "export_plan") || !strings.Contains(string(frames[1]), "export-plan") || !strings.Contains(string(frames[1]), "\"plan_id\":\"export-plan\"") {
		t.Fatalf("expected export tool response, got %s", frames[1])
	}
}

func TestServerHandlesValidateToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: validate-plan
title: Validate Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Validate Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "validate_plan", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md")}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	if !strings.Contains(string(frames[1]), "validate_plan") || !strings.Contains(string(frames[1]), "PLAN_TOO_SHALLOW") {
		t.Fatalf("expected validate tool response, got %s", frames[1])
	}
}

func TestServerAgentLifecycleRun(t *testing.T) {
	root := t.TempDir()
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion, "clientInfo": map[string]any{"name": "agent-runner", "version": "1.0.0"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "init_plan", "arguments": map[string]any{"title": "Agent Run Plan", "version": "1.0", "phases": []map[string]any{{"title": "Design", "tasks": []map[string]any{{"title": "define scope"}, {"title": "review design"}}}, {"title": "Build", "tasks": []map[string]any{{"title": "implement work"}, {"title": "verify work"}}}}}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": map[string]any{"name": "list_plans", "arguments": map[string]any{"include_archived": true}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 4, "method": "tools/call", "params": map[string]any{"name": "get_status", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "include_tasks": true}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 5, "method": "tools/call", "params": map[string]any{"name": "request_finish", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "actor_type": "agent", "summary": "checking finish gate too early"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 6, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH01-T01", "status": "in_progress", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 7, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH01-T01", "status": "done", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 8, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH01-T02", "status": "in_progress", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 9, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH01-T02", "status": "done", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 10, "method": "tools/call", "params": map[string]any{"name": "get_next_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "respect_phase_order": true, "respect_dependencies": true}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 11, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH02-T01", "status": "in_progress", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 12, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH02-T01", "status": "done", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 13, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH02-T02", "status": "in_progress", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 14, "method": "tools/call", "params": map[string]any{"name": "update_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH02-T02", "status": "done", "actor_type": "agent"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 15, "method": "tools/call", "params": map[string]any{"name": "validate_plan", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md")}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 16, "method": "tools/call", "params": map[string]any{"name": "request_finish", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "actor_type": "agent", "summary": "all work complete"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 17, "method": "tools/call", "params": map[string]any{"name": "export_plan", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "format": "json"}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 18, "method": "tools/call", "params": map[string]any{"name": "archive_plan", "arguments": map[string]any{"plan_id": "agent-run-plan", "create_final_export": true}}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 19, "method": "tools/call", "params": map[string]any{"name": "list_plans", "arguments": map[string]any{"include_archived": true}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 19 {
		t.Fatalf("expected 19 responses, got %d", len(frames))
	}
	assertFrameContains(t, frames[1], "init_plan", "agent-run-plan")
	assertFrameContains(t, frames[2], "list_plans", "agent-run-plan")
	assertFrameContains(t, frames[3], "get_status", "PH01-T01")
	assertFrameContains(t, frames[4], "FINISH_DENIED")
	assertFrameContains(t, frames[9], "get_next_task", "PH02-T01")
	assertFrameContains(t, frames[14], "validate_plan", "\"valid\":true")
	assertFrameContains(t, frames[15], "request_finish", "\"can_finish\":true")
	assertFrameContains(t, frames[16], "export_plan", "\"plan_id\":\"agent-run-plan\"")
	assertFrameContains(t, frames[17], "archive_plan", "\"archived\":true")
	assertFrameContains(t, frames[18], "list_plans", "archived")
	if _, err := os.Stat(filepath.Join(root, ".agent", "PLAN.md")); !os.IsNotExist(err) {
		t.Fatalf("expected active plan to be removed after archive, got %v", err)
	}
	matches, err := filepath.Glob(filepath.Join(root, ".agent", "archive", "agent-run-plan-*.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("expected exactly one archived markdown file, matches=%v err=%v", matches, err)
	}
	jsonMatches, err := filepath.Glob(filepath.Join(root, ".agent", "archive", "agent-run-plan-*.json"))
	if err != nil || len(jsonMatches) != 1 {
		t.Fatalf("expected exactly one archived json file, matches=%v err=%v", jsonMatches, err)
	}
}

func TestServerHandlesResetToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: reset-plan
title: Reset Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 1
---

# Reset Plan

## Phase 1 — Design
- [x] PH01-T01 done work
- [ ] PH01-T02 pending work

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "reset_task", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "task_id": "PH01-T01", "status": "in_progress", "reason": "reopen after failed verification"}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	assertFrameContains(t, frames[1], "reset_task", "in_progress")
}

func TestServerHandlesPrioritizeToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: prioritize-plan
title: Prioritize Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Prioritize Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "prioritize_tasks", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md"), "updates": []map[string]any{{"task_id": "PH01-T01", "priority": "P0"}, {"task_id": "PH02-T01", "priority": "P1"}}}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	assertFrameContains(t, frames[1], "prioritize_tasks", "PH01-T01", "PH02-T01")
}

func TestServerHandlesReconcileToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: reconcile-plan
title: Reconcile Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Reconcile Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	candidate := strings.ReplaceAll(content, "PH01-T01 first task", "PH01-T01 first task | priority:P0")
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "reconcile_plan", "arguments": map[string]any{"plan_id": "reconcile-plan", "markdown_content": candidate, "mode": "apply"}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	assertFrameContains(t, frames[1], "reconcile_plan", "PH01-T01")
}

func TestServerHandlesEditToolCall(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	content := `---
plan_id: edit-plan
title: Edit Plan
version: 1.0.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Edit Plan

## Phase 1 — Design
- [ ] PH01-T01 first task
- [ ] PH01-T02 second task

## Phase 2 — Build
- [ ] PH02-T01 implement
- [ ] PH02-T02 verify
`
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	input := bytes.NewBuffer(nil)
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": ProtocolVersion}})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeTestFrame(t, input, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "edit_plan", "arguments": map[string]any{"plan_id": "edit-plan", "target_id": "PH01-T01", "operation": "reprioritize_task", "payload": map[string]any{"priority": "P0"}}}})
	output := &bytes.Buffer{}
	server := &Server{API: api.New(root, nil)}
	if err := server.Serve(input, output); err != nil {
		t.Fatalf("serve failed: %v", err)
	}
	frames := readTestFrames(t, output.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 responses, got %d", len(frames))
	}
	assertFrameContains(t, frames[1], "edit_plan", "PH01-T01")
}

func assertFrameContains(t *testing.T, frame []byte, fragments ...string) {
	t.Helper()
	text := string(frame)
	for _, fragment := range fragments {
		if !strings.Contains(text, fragment) {
			t.Fatalf("expected frame to contain %q, got %s", fragment, text)
		}
	}
}

func writeTestFrame(t *testing.T, buffer *bytes.Buffer, payload map[string]any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if _, err := fmt.Fprintf(buffer, "Content-Length: %d\r\n\r\n%s", len(data), data); err != nil {
		t.Fatalf("write frame failed: %v", err)
	}
}

func readTestFrames(t *testing.T, data []byte) [][]byte {
	t.Helper()
	frames := make([][]byte, 0)
	for len(data) > 0 {
		sep := bytes.Index(data, []byte("\r\n\r\n"))
		if sep < 0 {
			t.Fatalf("missing frame separator in %q", data)
		}
		headers := string(data[:sep])
		var length int
		if _, err := fmt.Sscanf(headers, "Content-Length: %d", &length); err != nil {
			t.Fatalf("parse headers failed: %v", err)
		}
		start := sep + 4
		end := start + length
		frames = append(frames, append([]byte(nil), data[start:end]...))
		data = data[end:]
	}
	return frames
}
