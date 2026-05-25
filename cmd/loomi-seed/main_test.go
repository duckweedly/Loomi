package main

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestRunSeedIsIdempotent(t *testing.T) {
	t.Setenv("LOOMI_SEED_SCENARIO", "")
	svc := productdata.NewMemoryService()
	result, err := runSeed(context.Background(), svc, identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("runSeed() error = %v", err)
	}
	if result.ThreadID != seedThreadID || result.MessageID != seedMessageID {
		t.Fatalf("result = %+v", result)
	}
	second, err := runSeed(context.Background(), svc, identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("runSeed() second error = %v", err)
	}
	if second.ThreadID != result.ThreadID || second.MessageID != result.MessageID {
		t.Fatalf("second = %+v, first = %+v", second, result)
	}
	if _, err := svc.UpdateThread(context.Background(), identity.LocalDevIdentity(), result.ThreadID, productdata.UpdateThreadInput{Title: ptr("Renamed by user")}); err != nil {
		t.Fatalf("UpdateThread() error = %v", err)
	}
	afterRename, err := runSeed(context.Background(), svc, identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("runSeed() after rename error = %v", err)
	}
	if afterRename.ThreadID != result.ThreadID || afterRename.MessageID != result.MessageID {
		t.Fatalf("afterRename = %+v, first = %+v", afterRename, result)
	}
	messages, err := svc.ListMessages(context.Background(), identity.LocalDevIdentity(), result.ThreadID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestRunM17WorkArtifactSeedCreatesRepeatableEvidence(t *testing.T) {
	t.Setenv("LOOMI_SEED_SCENARIO", m17SeedScenario)
	ctx := context.Background()
	ident := identity.LocalDevIdentity()
	svc := productdata.NewMemoryService()

	result, err := runSeed(ctx, svc, ident)
	if err != nil {
		t.Fatalf("runSeed() error = %v", err)
	}
	if result.ThreadID != m17SeedThreadID || result.MessageID != m17SeedMessageID || result.RunID == "" || result.EventID == "" {
		t.Fatalf("result = %+v", result)
	}
	thread, err := svc.GetThread(ctx, ident, result.ThreadID)
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	if thread.Mode != productdata.ThreadModeWork {
		t.Fatalf("thread mode = %q", thread.Mode)
	}
	events, err := svc.ListRunEvents(ctx, ident, result.RunID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	workEvents := m17WorkEvents(events)
	if len(workEvents) != 1 {
		t.Fatalf("workEvents = %+v", workEvents)
	}
	event := workEvents[0]
	if event.Metadata["work_goal"] == "" || event.Metadata["work_steps"] == nil || event.Metadata["work_artifacts"] == nil {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
	artifacts, ok := event.Metadata["work_artifacts"].([]any)
	if !ok || len(artifacts) != 1 {
		t.Fatalf("work_artifacts = %#v", event.Metadata["work_artifacts"])
	}
	artifact, ok := artifacts[0].(map[string]any)
	if !ok {
		t.Fatalf("artifact = %#v", artifacts[0])
	}
	if artifact["source_thread_id"] != result.ThreadID || artifact["source_run_id"] != result.RunID || artifact["redaction_applied"] != true {
		t.Fatalf("artifact metadata = %+v", artifact)
	}
	for _, key := range []string{"command", "private_path", "authorization", "tool_output"} {
		if artifact[key] != "[redacted]" {
			t.Fatalf("artifact[%s] = %#v, want redacted metadata; artifact=%+v", key, artifact[key], artifact)
		}
	}

	second, err := runSeed(ctx, svc, ident)
	if err != nil {
		t.Fatalf("runSeed() second error = %v", err)
	}
	if second.ThreadID != result.ThreadID || second.MessageID != result.MessageID || second.RunID != result.RunID || second.EventID != result.EventID {
		t.Fatalf("second = %+v, first = %+v", second, result)
	}
	events, err = svc.ListRunEvents(ctx, ident, result.RunID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() after second error = %v", err)
	}
	if workEvents = m17WorkEvents(events); len(workEvents) != 1 {
		t.Fatalf("duplicate workEvents = %+v", workEvents)
	}
}

func m17WorkEvents(events []productdata.RunEvent) []productdata.RunEvent {
	var matches []productdata.RunEvent
	for _, event := range events {
		if event.Type == m17SeedEventType && event.Metadata["m17_seed"] == m17SeedScenario {
			matches = append(matches, event)
		}
	}
	return matches
}

func ptr[T any](v T) *T { return &v }
