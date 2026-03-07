package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunStatusCommandEmitsStatusEnvelope(t *testing.T) {
	planPath := filepath.Join(t.TempDir(), "PLAN.md")
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
	if err := os.WriteFile(planPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write plan: %v", err)
	}
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	code := run([]string{"status", "-plan", planPath, "-include-tasks"}, stdout, stderr)
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
