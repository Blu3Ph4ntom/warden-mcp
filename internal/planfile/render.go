package planfile

import (
	"fmt"
	"strings"

	"warden-mcp/internal/domain"
)

func Render(plan domain.Plan) string {
	var builder strings.Builder
	builder.WriteString("---\n")
	builder.WriteString(fmt.Sprintf("plan_id: %s\n", plan.PlanID))
	builder.WriteString(fmt.Sprintf("title: %s\n", plan.Title))
	builder.WriteString(fmt.Sprintf("version: %s\n", plan.Version))
	builder.WriteString(fmt.Sprintf("status: %s\n", plan.Status))
	builder.WriteString(fmt.Sprintf("current_phase: %s\n", plan.CurrentPhaseID))
	builder.WriteString(fmt.Sprintf("can_finish: %t\n", plan.CanFinish))
	builder.WriteString(fmt.Sprintf("completed_tasks: %d\n", plan.CompletedTaskCount()))
	builder.WriteString("---\n\n")
	builder.WriteString("# " + plan.Title + "\n\n")
	for index, phase := range plan.Phases {
		builder.WriteString(fmt.Sprintf("## Phase %d — %s\n", index+1, phase.Title))
		for _, task := range phase.Tasks {
			builder.WriteString(fmt.Sprintf("- [%s] %s %s%s\n", markerForStatus(task.Status), task.TaskID, task.Title, taskMetadataSuffix(task)))
		}
		builder.WriteString("\n")
	}
	return builder.String()
}

func taskMetadataSuffix(task domain.Task) string {
	segments := make([]string, 0, 3)
	if task.Priority != "" && task.Priority != domain.PriorityP2 {
		segments = append(segments, fmt.Sprintf("priority:%s", task.Priority))
	}
	if len(task.DependsOn) > 0 {
		segments = append(segments, fmt.Sprintf("depends_on:%s", strings.Join(task.DependsOn, ",")))
	}
	if !task.Required {
		segments = append(segments, "required:false")
	}
	if len(segments) == 0 {
		return ""
	}
	return " | " + strings.Join(segments, " | ")
}

func markerForStatus(status domain.TaskStatus) string {
	switch status {
	case domain.TaskDone:
		return "x"
	case domain.TaskInProgress:
		return "/"
	case domain.TaskBlocked, domain.TaskCancelled, domain.TaskWaived:
		return "-"
	default:
		return " "
	}
}
