package planfile_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/mcp/contracts"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/planfile"
	"github.com/Blu3Ph4ntom/warden-mcp/internal/service"
)

func BenchmarkParsePlan100Tasks(b *testing.B) {
	content := generatedPlanMarkdown(10, 10)
	updatedAt := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := planfile.Parse(content, updatedAt); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkParsePlan1000Tasks(b *testing.B) {
	content := generatedPlanMarkdown(20, 50)
	updatedAt := time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, _, err := planfile.Parse(content, updatedAt); err != nil {
			b.Fatalf("parse failed: %v", err)
		}
	}
}

func BenchmarkStatusEvaluation1000Tasks(b *testing.B) {
	content := generatedPlanMarkdown(20, 50)
	plan, _, err := planfile.Parse(content, time.Date(2026, 3, 7, 0, 0, 0, 0, time.UTC))
	if err != nil {
		b.Fatalf("parse failed: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.GetStatus(plan, false)
		_ = service.GetNextTask(plan, contracts.GetNextTaskRequest{PlanID: plan.PlanID, RespectPhaseOrder: true, RespectDependencies: true})
	}
}

func generatedPlanMarkdown(phaseCount, tasksPerPhase int) string {
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString("plan_id: benchmark-plan\n")
	builder.WriteString("title: Benchmark Plan\n")
	builder.WriteString("version: 1.0\n")
	builder.WriteString("status: active\n")
	builder.WriteString("current_phase: PH01\n")
	builder.WriteString("can_finish: false\n")
	builder.WriteString("completed_tasks: 0\n")
	builder.WriteString("---\n\n")
	builder.WriteString("# Benchmark Plan\n\n")
	for phaseIndex := 1; phaseIndex <= phaseCount; phaseIndex++ {
		builder.WriteString(fmt.Sprintf("## Phase %d — Phase %02d\n", phaseIndex, phaseIndex))
		for taskIndex := 1; taskIndex <= tasksPerPhase; taskIndex++ {
			builder.WriteString(fmt.Sprintf("- [ ] PH%02d-T%02d Task %02d.%02d\n", phaseIndex, taskIndex, phaseIndex, taskIndex))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}
