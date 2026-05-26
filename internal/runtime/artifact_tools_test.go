package runtime

import (
	"context"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestArtifactCreateReadAndList(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := artifactTestThreadRun(t, svc)
	executor := ArtifactToolExecutor{Artifacts: svc}

	created, err := executor.Execute(context.Background(), ToolInvocation{
		ThreadID:         thread.ID,
		RunID:            run.ID,
		ToolCallID:       "tc_artifact_create",
		ToolName:         productdata.ToolNameArtifactCreateText,
		ArgumentsSummary: map[string]any{"title": "Notes", "content": "hello artifact"},
		ApprovalStatus:   productdata.ToolCallApprovalApproved,
		ExecutionStatus:  productdata.ToolCallExecutionExecuting,
		Catalog:          productdata.ToolCatalogFromEvents(nil),
		EnabledTools:     ToolResolutionsForPersona([]string{productdata.ToolNameArtifactCreateText, productdata.ToolNameArtifactRead, productdata.ToolNameArtifactList}),
	})
	if err != nil {
		t.Fatal(err)
	}
	artifactID, _ := created["artifact_id"].(string)
	if !strings.HasPrefix(artifactID, "art_") || created["operation"] != "create_text" || created["title"] != "Notes" || created["text_excerpt"] != "hello artifact" {
		t.Fatalf("created = %+v", created)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameArtifactRead, ArgumentsSummary: map[string]any{"artifact_id": artifactID, "max_bytes": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if read["operation"] != "read" || read["text_excerpt"] != "hello" || read["truncated"] != true {
		t.Fatalf("read = %+v", read)
	}

	list, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameArtifactList, ArgumentsSummary: map[string]any{"limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	items, _ := list["artifacts"].([]map[string]any)
	if list["operation"] != "list" || len(items) != 1 || items[0]["artifact_id"] != artifactID || items[0]["content"] != nil {
		t.Fatalf("list = %+v", list)
	}
}

func TestArtifactRejectsUnsafeInputsAndScope(t *testing.T) {
	svc := productdata.NewMemoryService()
	thread, run := artifactTestThreadRun(t, svc)
	executor := ArtifactToolExecutor{Artifacts: svc}

	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameArtifactCreateText, ArgumentsSummary: map[string]any{"title": "Huge", "content": strings.Repeat("x", 33*1024)}}); err == nil {
		t.Fatal("expected oversized artifact to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: productdata.ToolNameArtifactRead, ArgumentsSummary: map[string]any{"artifact_id": "art_missing"}}); err == nil {
		t.Fatal("expected unknown artifact to fail")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ThreadID: thread.ID, RunID: run.ID, ToolName: "artifact.render", ArgumentsSummary: map[string]any{}}); err == nil {
		t.Fatal("expected unsupported artifact tool to fail")
	}
}

func artifactTestThreadRun(t *testing.T, svc *productdata.MemoryService) (productdata.Thread, productdata.Run) {
	t.Helper()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Artifacts", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	return thread, run
}
