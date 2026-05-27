package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestArtifactReadOnlyEndpointsUseThreadScope(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Artifacts", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{ScriptName: "artifact_api"})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := svc.CreateArtifact(context.Background(), ident, productdata.CreateArtifactInput{ThreadID: thread.ID, RunID: run.ID, Title: " Notes ", Content: "hello artifact endpoint", MaxBytes: 100})
	if err != nil {
		t.Fatal(err)
	}
	other, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Other", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	list := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/artifacts?limit=10", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", list.Code, list.Body.String())
	}
	for _, expected := range []string{`"id":"` + artifact.ID + `"`, `"title":"Notes"`, `"text_excerpt":"hello artifact endpoint"`} {
		if !strings.Contains(list.Body.String(), expected) {
			t.Fatalf("list body missing %q: %s", expected, list.Body.String())
		}
	}
	if strings.Contains(list.Body.String(), `"content"`) {
		t.Fatalf("list leaked raw content field: %s", list.Body.String())
	}

	read := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/artifacts/"+artifact.ID+"?max_bytes=5", "")
	if read.Code != http.StatusOK {
		t.Fatalf("read status = %d body=%s", read.Code, read.Body.String())
	}
	if body := read.Body.String(); !strings.Contains(body, `"text_excerpt":"hello"`) || !strings.Contains(body, `"truncated":true`) {
		t.Fatalf("read body = %s", body)
	}

	crossThread := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+other.ID+"/artifacts/"+artifact.ID, "")
	if crossThread.Code != http.StatusNotFound {
		t.Fatalf("cross-thread status = %d body=%s", crossThread.Code, crossThread.Body.String())
	}
}
