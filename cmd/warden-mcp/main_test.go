package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
