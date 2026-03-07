package contracts

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/domain"
)

func TestRequestFinishEnvelopeMarshalsCanonicalFields(t *testing.T) {
	envelope := ToolResponseEnvelope[RequestFinishData]{
		OK:        true,
		Tool:      "request_finish",
		Timestamp: "2026-03-07T00:00:00Z",
		PlanID:    "warden-plan",
		Warnings:  []domain.ValidationIssue{},
		Data: RequestFinishData{
			CanFinish:           false,
			Plan:                PlanSummary{PlanID: "warden-plan", Title: "Warden", Status: domain.PlanActive, Version: "0.2.0", TotalTasks: 4, CompletedTasks: 2, CanFinish: false, UpdatedAt: "2026-03-07T00:00:00Z"},
			BlockingReasons:     []BlockingReason{{Code: ErrFinishDenied, Message: "tasks remain"}},
			IncompleteTaskIDs:   []string{"PH03-T01"},
			NextRequiredActions: []string{"Complete PH03-T01"},
		},
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	jsonText := string(data)
	for _, fragment := range []string{"\"tool\":\"request_finish\"", "\"blocking_reasons\"", "\"incomplete_task_ids\"", "\"next_required_actions\""} {
		if !strings.Contains(jsonText, fragment) {
			t.Fatalf("expected JSON to contain %s, got %s", fragment, jsonText)
		}
	}
}

func TestUpdateTaskRequestRoundTripUsesSchemaFieldNames(t *testing.T) {
	input := []byte(`{"plan_id":"warden-plan","task_id":"PH03-T04","status":"in_progress","note":"started","evidence":["test log"],"reason":"work began","actor_type":"agent"}`)
	var req UpdateTaskRequest
	if err := json.Unmarshal(input, &req); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if req.PlanID != "warden-plan" || req.TaskID != "PH03-T04" || req.Status != domain.TaskInProgress {
		t.Fatalf("unexpected request contents: %+v", req)
	}
	output, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(output), "\"actor_type\":\"agent\"") {
		t.Fatalf("expected canonical actor_type field, got %s", string(output))
	}
}

func TestReservedErrorCodesRemainStable(t *testing.T) {
	codes := []string{ErrPlanNotFound, ErrPlanInvalid, ErrPhaseInvalid, ErrTaskNotFound, ErrTaskTransitionInvalid, ErrDependencyViolation, ErrFinishDenied, ErrSyncConflict, ErrImportInvalid, ErrExportFailed, ErrArchiveDenied, ErrInternal}
	if len(codes) != 12 {
		t.Fatalf("expected 12 reserved codes, got %d", len(codes))
	}
	seen := map[string]struct{}{}
	for _, code := range codes {
		if _, exists := seen[code]; exists {
			t.Fatalf("duplicate reserved code %s", code)
		}
		seen[code] = struct{}{}
	}
}

func TestValidationIssueUsesSchemaFieldNames(t *testing.T) {
	envelope := ToolResponseEnvelope[struct{}]{
		OK:        true,
		Tool:      "get_status",
		Timestamp: "2026-03-07T00:00:00Z",
		Warnings:  []domain.ValidationIssue{{Severity: "warning", Code: "PLAN_VERSION_NORMALIZED", Message: "normalized", Path: "version"}},
		Data:      struct{}{},
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	jsonText := string(data)
	for _, fragment := range []string{"\"severity\":\"warning\"", "\"code\":\"PLAN_VERSION_NORMALIZED\"", "\"message\":\"normalized\"", "\"path\":\"version\""} {
		if !strings.Contains(jsonText, fragment) {
			t.Fatalf("expected JSON to contain %s, got %s", fragment, jsonText)
		}
	}
}
