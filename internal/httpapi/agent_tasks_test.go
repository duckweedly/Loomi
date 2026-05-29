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
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "Review this"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, productdata.SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review implementation"})
	if err != nil {
		t.Fatal(err)
	}
	task, err = svc.DelegateAgentTask(context.Background(), ident, productdata.DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID})
	if err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+thread.ID+"/agent-tasks?limit=5", "")
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	for _, expected := range []string{`"id":"` + task.ID + `"`, `"role":"reviewer"`, `"goal":"Review implementation"`, `"status":"in_progress"`, `"child_thread_id":"` + task.ChildThreadID + `"`, `"child_run_id":"` + task.ChildRunID + `"`} {
		if !strings.Contains(res.Body.String(), expected) {
			t.Fatalf("body missing %q: %s", expected, res.Body.String())
		}
	}
}
