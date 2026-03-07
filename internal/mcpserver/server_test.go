package mcpserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"warden-mcp/internal/api"
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
	if !strings.Contains(string(frames[1]), "export_plan") || !strings.Contains(string(frames[1]), "export-plan") || !strings.Contains(string(frames[1]), "PlanID") {
		t.Fatalf("expected export tool response, got %s", frames[1])
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
