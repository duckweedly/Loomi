package httpapi

import "testing"

func TestSplitResourcePath(t *testing.T) {
	id, suffix := splitResourcePath("/v1/threads/thread-1/messages", "/v1/threads/")
	if id != "thread-1" || suffix != "messages" {
		t.Fatalf("splitResourcePath returned id=%q suffix=%q", id, suffix)
	}

	id, suffix = splitResourcePath("/v1/runs/run-1", "/v1/runs/")
	if id != "run-1" || suffix != "" {
		t.Fatalf("splitResourcePath returned id=%q suffix=%q", id, suffix)
	}
}
