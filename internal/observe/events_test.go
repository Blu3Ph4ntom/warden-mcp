package observe

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONRecorderWritesRedactedStructuredEvent(t *testing.T) {
	buffer := &bytes.Buffer{}
	recorder := NewJSONRecorder(buffer)
	recorder.Record(Event{Kind: "command", Command: "status", Message: "use sk-1234567890abcdef", Accepted: Accepted(true)})
	var event Event
	if err := json.Unmarshal(buffer.Bytes(), &event); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if event.Kind != "command" || event.Command != "status" || event.Accepted == nil || !*event.Accepted {
		t.Fatalf("unexpected event %+v", event)
	}
	if strings.Contains(event.Message, "1234567890abcdef") || !strings.Contains(event.Message, "***redacted***") {
		t.Fatalf("expected redacted message, got %+v", event)
	}
	if event.Timestamp == "" {
		t.Fatal("expected timestamp to be set")
	}
}
