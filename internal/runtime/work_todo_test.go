package runtime

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestAppendWorkTodoSnapshotUsesDurableToolEventsForWorkRuns(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Work", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_read", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}

	event, ok := appendWorkTodoSnapshot(context.Background(), svc, run, "runtime")
	if !ok {
		t.Fatal("todo snapshot was not appended")
	}
	if event.Type != productdata.EventWorkTodoUpdated || event.Metadata["updated_by"] != "runtime" {
		t.Fatalf("event = %+v", event)
	}
	items, ok := event.Metadata["todo_items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v", event.Metadata["todo_items"])
	}
	item := items[0].(map[string]any)
	if item["title"] != "Read project file" || item["status"] != "blocked" || item["summary"] != "Waiting for approval" {
		t.Fatalf("item = %+v", item)
	}
}

func TestAppendWorkTodoSnapshotSkipsChatRuns(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Chat", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, productdata.RecordToolCallRequestInput{ToolCallID: "tc_time", ToolName: productdata.ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}

	if _, ok := appendWorkTodoSnapshot(context.Background(), svc, run, "runtime"); ok {
		t.Fatal("chat run appended todo snapshot")
	}
}

func TestAppendProviderWorkTodoSnapshotUsesExplicitTodoItems(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Work", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway})
	if err != nil {
		t.Fatal(err)
	}
	result := map[string]any{
		"todo_items": []any{map[string]any{"id": "todo-1", "title": "Review patch", "status": "running"}},
	}

	event, ok := appendProviderWorkTodoSnapshot(context.Background(), svc, run, result)
	if !ok {
		t.Fatal("provider todo snapshot was not appended")
	}
	if event.Type != productdata.EventWorkTodoUpdated || event.Metadata["updated_by"] != "provider" {
		t.Fatalf("event = %+v", event)
	}
	items := event.Metadata["todo_items"].([]any)
	item := items[0].(map[string]any)
	if item["title"] != "Review patch" || item["status"] != "running" {
		t.Fatalf("item = %+v", item)
	}
}
