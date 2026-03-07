package observe

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/Blu3Ph4ntom/warden-mcp/internal/security"
)

type Event struct {
	Kind       string         `json:"kind"`
	Timestamp  string         `json:"timestamp"`
	Command    string         `json:"command,omitempty"`
	Method     string         `json:"method,omitempty"`
	PlanID     string         `json:"plan_id,omitempty"`
	PhaseID    string         `json:"phase_id,omitempty"`
	TaskID     string         `json:"task_id,omitempty"`
	ActorType  string         `json:"actor_type,omitempty"`
	Accepted   *bool          `json:"accepted,omitempty"`
	DurationMS int64          `json:"duration_ms,omitempty"`
	Message    string         `json:"message,omitempty"`
	ErrorCode  string         `json:"error_code,omitempty"`
	Fields     map[string]any `json:"fields,omitempty"`
}

type Recorder interface {
	Record(Event)
}

type JSONRecorder struct {
	writer io.Writer
	mu     sync.Mutex
}

func NewJSONRecorder(writer io.Writer) *JSONRecorder {
	return &JSONRecorder{writer: writer}
}

func (r *JSONRecorder) Record(event Event) {
	if r == nil || r.writer == nil {
		return
	}
	event.Message = security.RedactSecretLikeText(event.Message)
	if event.Timestamp == "" {
		event.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_ = json.NewEncoder(r.writer).Encode(event)
}

func Accepted(value bool) *bool {
	return &value
}

func Since(start time.Time) int64 {
	if start.IsZero() {
		return 0
	}
	return time.Since(start).Milliseconds()
}
