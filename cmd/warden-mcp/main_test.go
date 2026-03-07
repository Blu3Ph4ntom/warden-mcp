package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
)

func TestRunStatusCommandEmitsStatusEnvelope(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH02
---

# Sample Plan

## Phase 1 — Setup
- [x] PH01-T01 create repo
- [x] PH01-T02 add tests

## Phase 2 — Build
- [/] PH02-T01 start implementation
- [ ] PH02-T02 finish implementation
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWD)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"status", "-plan", filepath.Join(".agent", "PLAN.md"), "-include-tasks"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	output := stdout.String()
	for _, fragment := range []string{"\"tool\": \"get_status\"", "\"plan_id\": \"sample-plan\"", "\"next_task_id\": \"PH02-T01\""} {
		if !strings.Contains(output, fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, output)
		}
	}
}

func TestRunHealthCommandEmitsHealthEnvelope(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"health", "-plan", filepath.Join(".agent", "PLAN.md")}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	for _, fragment := range []string{"\"tool\": \"health_check\"", "\"status\": \"ok\"", "\"plan_id\": \"sample-plan\""} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, stdout.String())
		}
	}
}

func TestRunExportCommandEmitsJSONExportEnvelope(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"export", "-plan", filepath.Join(".agent", "PLAN.md"), "-format", "json"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	for _, fragment := range []string{"\"tool\": \"export_plan\"", "\"format\": \"json\"", "sample-plan"} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, stdout.String())
		}
	}
}

func TestRunValidateCommandEmitsInvalidPlanReport(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"validate", "-plan", filepath.Join(".agent", "PLAN.md")}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	for _, fragment := range []string{"\"tool\": \"validate_plan\"", "\"valid\": false", "PLAN_TOO_SHALLOW"} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, stdout.String())
		}
	}
}

func TestRunPrioritizeCommandEmitsEnvelope(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
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
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"prioritize", "-plan", filepath.Join(".agent", "PLAN.md"), "-updates", "PH01-T01=P0,PH02-T01=P1"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	for _, fragment := range []string{"\"tool\": \"prioritize_tasks\"", "PH01-T01", "PH02-T01"} {
		if !strings.Contains(stdout.String(), fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, stdout.String())
		}
	}
}

func TestParsePriorityUpdatesRejectsMalformedInput(t *testing.T) {
	if _, err := parsePriorityUpdates("bad-value"); err == nil {
		t.Fatal("expected malformed updates to fail")
	}
}

func TestRunReconcileRejectsContentFileOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outsidePath := filepath.Join(filepath.Dir(root), "outside-plan.md")
	if err := os.WriteFile(outsidePath, []byte("# outside"), 0o644); err != nil {
		t.Fatalf("write outside file failed: %v", err)
	}
	defer os.Remove(outsidePath)
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"reconcile", "-content-file", "..\\outside-plan.md"}, stdout, stderr)
	if code != 0 || !strings.Contains(stdout.String(), contracts.ErrPlanInvalid) {
		t.Fatalf("expected invalid reconcile path error, code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
}

func TestRunRejectsPlanPathOutsideWorkspace(t *testing.T) {
	root := t.TempDir()
	outside := filepath.Join(t.TempDir(), "PLAN.md")
	if err := os.WriteFile(outside, []byte("---\nplan_id: x\ntitle: x\nversion: 1.0\nstatus: active\ncurrent_phase: PH01\n---\n"), 0o644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWD)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"status", "-plan", outside}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected JSON error envelope exit code 0, got %d stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "\"ok\": false") || !strings.Contains(stdout.String(), "\"code\": \"PLAN_INVALID\"") {
		t.Fatalf("expected plan path denial, got %s", stdout.String())
	}
}

func TestRunUpdateCommandMutatesPlanFile(t *testing.T) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
	- [ ] PH01-T02 add tests

	## Phase 2 — Build
	- [ ] PH02-T01 implement server
	- [ ] PH02-T02 verify server
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() {
		_ = os.Chdir(previousWD)
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"update", "-plan", filepath.Join(".agent", "PLAN.md"), "-task", "PH01-T01", "-status", "in_progress"}, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	updatedContent, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatalf("read plan: %v", err)
	}
	for _, fragment := range []string{"\"tool\": \"update_task\"", "- [/] PH01-T01 create repo"} {
		if fragment == "- [/] PH01-T01 create repo" {
			if !strings.Contains(string(updatedContent), fragment) {
				t.Fatalf("expected updated file to contain %s, got %s", fragment, string(updatedContent))
			}
			continue
		}
		if !strings.Contains(stdout.String(), fragment) {
			t.Fatalf("expected output to contain %s, got %s", fragment, stdout.String())
		}
	}
}

func TestRunServeCommandHandlesInitializeAndStatusCall(t *testing.T) {
	testRunServeLikeCommandHandlesInitializeAndStatusCall(t, []string{"serve"})
}

func TestRunWithoutArgsDefaultsToServeAndHandlesStatusCall(t *testing.T) {
	testRunServeLikeCommandHandlesInitializeAndStatusCall(t, []string{})
}

func testRunServeLikeCommandHandlesInitializeAndStatusCall(t *testing.T, args []string) {
	root := t.TempDir()
	planPath := filepath.Join(root, ".agent", "PLAN.md")
	content := `---
plan_id: sample-plan
title: Sample Plan
version: 1.0
status: active
current_phase: PH01
can_finish: false
completed_tasks: 0
---

# Sample Plan

## Phase 1 — Setup
- [ ] PH01-T01 create repo
`
	if err := os.MkdirAll(filepath.Dir(planPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	previousWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(previousWD) }()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	stdin := &bytes.Buffer{}
	writeFrame(t, stdin, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": "2025-03-26", "clientInfo": map[string]any{"name": "test", "version": "1.0.0"}}})
	writeFrame(t, stdin, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized"})
	writeFrame(t, stdin, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": "get_status", "arguments": map[string]any{"plan_path": filepath.Join(".agent", "PLAN.md")}}})
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := runWithIO(args, stdin, stdout, stderr)
	if code != 0 {
		t.Fatalf("expected zero exit code, got %d stderr=%s", code, stderr.String())
	}
	frames := readFrames(t, stdout.Bytes())
	if len(frames) != 2 {
		t.Fatalf("expected 2 MCP responses, got %d", len(frames))
	}
	if !strings.Contains(string(frames[0]), "initialize") && !strings.Contains(string(frames[0]), "protocolVersion") {
		t.Fatalf("expected initialize response, got %s", frames[0])
	}
	if !strings.Contains(string(frames[1]), "sample-plan") || !strings.Contains(string(frames[1]), "get_status") {
		t.Fatalf("expected status tool response, got %s", frames[1])
	}
}

func writeFrame(t *testing.T, buffer *bytes.Buffer, payload map[string]any) {
	t.Helper()
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if _, err := fmt.Fprintf(buffer, "Content-Length: %d\r\n\r\n%s", len(data), data); err != nil {
		t.Fatalf("write frame failed: %v", err)
	}
}

func readFrames(t *testing.T, data []byte) [][]byte {
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
