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

func TestAgentTasksEndpointListsThreadScopedTasks(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Agents", Mode: productdata.ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{ScriptName: "agent_api"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, productdata.SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review implementation"})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/agent-tasks?limit=5", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	for _, expected := range []string{`"id":"` + task.ID + `"`, `"role":"reviewer"`, `"goal":"Review implementation"`, `"status":"spawned"`} {
		if !strings.Contains(res.Body.String(), expected) {
			t.Fatalf("body missing %q: %s", expected, res.Body.String())
		}
	}
}
