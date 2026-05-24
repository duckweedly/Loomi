package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestGatewayPersistsProviderDeltasAndCompletion(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventTextDelta, Text: "hel"}, {Type: ProviderEventTextDelta, Text: "lo"}, {Type: ProviderEventCompleted}}}
	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-1].Type != "run_completed" {
		t.Fatalf("events = %+v", events)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(messages) != 2 || messages[1].Role != productdata.MessageRoleAssistant || messages[1].Content != "hello" {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestGatewayRecordsToolCallsAsBoundaryEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventToolCall, ToolName: "read_file", Metadata: map[string]any{"arguments": "secret"}}, {Type: ProviderEventCompleted, Text: "done"}}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, event := range events {
		if event.Type == "tool_call_blocked" {
			found = true
			if event.Metadata["tool_name"] != "read_file" || event.Metadata["arguments"] != nil {
				t.Fatalf("event = %+v", event)
			}
		}
	}
	if !found {
		t.Fatalf("events = %+v", events)
	}
}

func TestGatewayMapsProviderFailureToRedactedRunFailure(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventRateLimited}}}
	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "provider_rate_limited" {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayLoadsCurrentThreadContextThroughTriggerMessage(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	otherThread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Other", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.CreateMessage(context.Background(), ident, otherThread.ID, productdata.CreateMessageInput{Content: "do not include"}); err != nil {
		t.Fatal(err)
	}
	first, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, productdata.AppendAssistantMessageInput{Content: "hi there"}); err != nil {
		t.Fatal(err)
	}
	current, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "continue"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: current.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: current.ID, ProviderID: "custom"})

	if provider.request.ThreadID != thread.ID || provider.request.MessageID != current.ID {
		t.Fatalf("request = %+v", provider.request)
	}
	want := []ProviderMessage{{Role: "user", Content: first.Content}, {Role: "assistant", Content: "hi there"}, {Role: "user", Content: current.Content}}
	if len(provider.request.Messages) != len(want) {
		t.Fatalf("messages = %+v", provider.request.Messages)
	}
	for i := range want {
		if provider.request.Messages[i] != want[i] {
			t.Fatalf("messages = %+v", provider.request.Messages)
		}
	}
}

type capturingProvider struct {
	config  ProviderConfig
	request ProviderRequest
}

func (p *capturingProvider) Config() ProviderConfig { return p.config }

func (p *capturingProvider) Stream(_ context.Context, request ProviderRequest) (<-chan ProviderEvent, error) {
	p.request = request
	ch := make(chan ProviderEvent, 1)
	ch <- ProviderEvent{Type: ProviderEventCompleted, Text: "ok"}
	close(ch)
	return ch, nil
}

func TestQueuedRunRouterHydratesGatewayInputFromJobMetadata(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "override"})
	if err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := &capturingProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "base", Enabled: true}}

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err != nil {
		t.Fatal(err)
	}

	if provider.request.ThreadID != thread.ID || provider.request.MessageID != message.ID || provider.request.Model != "override" {
		t.Fatalf("request = %+v", provider.request)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestQueuedRunRouterReturnsErrorWhenGatewayRunFails(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_gateway", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventRateLimited}}}

	if err := (QueuedRunRouter{Gateway: NewGateway(svc, nil, []Provider{provider})}).Run(context.Background(), claimedRun, job); err == nil {
		t.Fatal("Run() error = nil")
	}
	if _, changed, err := svc.CompleteBackgroundJob(context.Background(), ident, productdata.CompleteBackgroundJobInput{JobID: job.ID, WorkerID: "worker_gateway", OwnershipVersion: job.OwnershipVersion}); err != nil || !changed {
		t.Fatalf("CompleteBackgroundJob() changed=%v err=%v", changed, err)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayRunAsyncOutlivesRequestContext(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	provider := contextAwareProvider{config: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}}

	NewGateway(svc, nil, []Provider{provider}).RunAsync(ctx, run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got := waitForTerminalRun(t, svc, run.ID)
	if got.Status != productdata.RunStatusCompleted {
		t.Fatalf("run = %+v", got)
	}
}

func TestGatewayFailsWhenAssistantMessageCannotBePersisted(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Gateway", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, productdata.CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{Source: productdata.RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	provider := StaticProvider{ProviderConfig: ProviderConfig{ID: "custom", Family: ProviderFamilyOpenAICompatible, BaseURL: "https://example.test/v1", APIKey: "key", Model: "model", Enabled: true}, Events: []ProviderEvent{{Type: ProviderEventCompleted, Text: "hello"}}}
	wrapped := assistantPersistFailingService{Service: svc}

	NewGateway(wrapped, nil, []Provider{provider}).run(context.Background(), run, GatewayRunInput{ThreadID: thread.ID, MessageID: message.ID, ProviderID: "custom"})

	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != productdata.RunStatusFailed || got.ErrorCode == nil || *got.ErrorCode != "assistant_message_persist_failed" {
		t.Fatalf("run = %+v", got)
	}
}

type contextAwareProvider struct {
	config ProviderConfig
}

func (p contextAwareProvider) Config() ProviderConfig { return p.config }

func (p contextAwareProvider) Stream(ctx context.Context, _ ProviderRequest) (<-chan ProviderEvent, error) {
	ch := make(chan ProviderEvent, 1)
	if ctx.Err() != nil {
		ch <- ProviderEvent{Type: ProviderEventError, ErrorCode: "request_context_canceled", Message: "request context was canceled"}
	} else {
		ch <- ProviderEvent{Type: ProviderEventCompleted, Text: "ok"}
	}
	close(ch)
	return ch, nil
}

type assistantPersistFailingService struct {
	productdata.Service
}

func (s assistantPersistFailingService) AppendAssistantMessage(context.Context, identity.LocalIdentity, string, productdata.AppendAssistantMessageInput) (productdata.Message, error) {
	return productdata.Message{}, productdata.NewError(productdata.CodeInternalError, "assistant message persistence failed")
}

func waitForTerminalRun(t *testing.T, svc productdata.Service, runID string) productdata.Run {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		run, err := svc.GetRun(context.Background(), identity.LocalDevIdentity(), runID)
		if err != nil {
			t.Fatal(err)
		}
		if productdata.IsRunTerminal(run.Status) {
			return run
		}
		time.Sleep(10 * time.Millisecond)
	}
	run, err := svc.GetRun(context.Background(), identity.LocalDevIdentity(), runID)
	if err != nil {
		t.Fatal(err)
	}
	return run
}
