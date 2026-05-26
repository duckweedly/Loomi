package productdata

import (
	"context"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestMemoryServiceCreatesReadsAndListsArtifactsInThreadScope(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Artifacts", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}

	artifact, err := svc.CreateArtifact(context.Background(), ident, CreateArtifactInput{ThreadID: thread.ID, RunID: run.ID, Title: " Notes ", Content: "hello artifact", MaxBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(artifact.ID, "art_") || artifact.Title != "Notes" || artifact.ArtifactType != "text" || artifact.ContentBytes != len("hello artifact") {
		t.Fatalf("artifact = %+v", artifact)
	}

	read, err := svc.ReadArtifact(context.Background(), ident, ReadArtifactInput{ThreadID: thread.ID, ArtifactID: artifact.ID, MaxBytes: 5})
	if err != nil {
		t.Fatal(err)
	}
	if read.TextExcerpt != "hello" || !read.Truncated || read.Content != "hello artifact" {
		t.Fatalf("read = %+v", read)
	}

	list, err := svc.ListArtifacts(context.Background(), ident, ListArtifactsInput{ThreadID: thread.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].ID != artifact.ID || list[0].Content != "" {
		t.Fatalf("list = %+v", list)
	}

	otherThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Other", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.ReadArtifact(context.Background(), ident, ReadArtifactInput{ThreadID: otherThread.ID, ArtifactID: artifact.ID}); err == nil || ErrorCode(err) != CodeArtifactNotFound {
		t.Fatalf("cross-thread read err = %v", err)
	}
}

func TestMemoryServiceRejectsInvalidArtifacts(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Artifacts", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}

	for _, input := range []CreateArtifactInput{
		{ThreadID: thread.ID, RunID: run.ID, Title: "", Content: "hello"},
		{ThreadID: thread.ID, RunID: run.ID, Title: "Notes", Content: ""},
		{ThreadID: thread.ID, RunID: run.ID, Title: "Notes", Content: "too large", MaxBytes: 1},
		{ThreadID: "missing", RunID: run.ID, Title: "Notes", Content: "hello"},
		{ThreadID: thread.ID, RunID: "missing", Title: "Notes", Content: "hello"},
	} {
		if _, err := svc.CreateArtifact(context.Background(), ident, input); err == nil {
			t.Fatalf("CreateArtifact(%+v) err = nil", input)
		}
	}
}
