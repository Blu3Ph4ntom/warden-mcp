package planfile

import (
	"os"
	"time"

	"warden-mcp/internal/domain"
)

func WritePlanFile(path string, plan domain.Plan) (domain.Plan, []domain.ValidationIssue, error) {
	info, err := os.Stat(path)
	mode := os.FileMode(0o644)
	if err == nil {
		mode = info.Mode()
	} else if !os.IsNotExist(err) {
		return domain.Plan{}, nil, err
	}
	content := Render(plan)
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		return domain.Plan{}, nil, err
	}
	return Parse(content, time.Now().UTC())
}
