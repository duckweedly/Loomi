package main

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestRunSeedIsIdempotent(t *testing.T) {
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

func ptr[T any](v T) *T { return &v }
