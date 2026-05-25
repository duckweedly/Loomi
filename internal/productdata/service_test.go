package productdata

import (
	"context"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
)

func TestGeneratedIDsDoNotUseProcessLocalSequence(t *testing.T) {
	first := NewThreadID()
	second := NewThreadID()
	if first == "thr_1" || second == "thr_2" {
		t.Fatalf("ids use process-local sequence: %q %q", first, second)
	}
}

func TestCurrentIdentityEnsuresLocalUser(t *testing.T) {
	svc := NewMemoryService()
	user, err := svc.CurrentIdentity(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("CurrentIdentity() error = %v", err)
	}
	if user.ID != "user_local_dev" || user.DisplayName != "Local Developer" {
		t.Fatalf("user = %+v", user)
	}
	again, err := svc.CurrentIdentity(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		t.Fatalf("CurrentIdentity() second error = %v", err)
	}
	if again.ID != user.ID || !again.CreatedAt.Equal(user.CreatedAt) {
		t.Fatalf("second user = %+v, first = %+v", again, user)
	}
}

func TestThreadLifecycleForCurrentIdentity(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "  First thread  ", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if thread.Title != "First thread" || thread.Mode != ThreadModeChat || thread.LifecycleStatus != ThreadLifecycleActive {
		t.Fatalf("thread = %+v", thread)
	}

	threads, err := svc.ListThreads(context.Background(), ident, false)
	if err != nil {
		t.Fatalf("ListThreads() error = %v", err)
	}
	if len(threads) != 1 || threads[0].ID != thread.ID {
		t.Fatalf("threads = %+v", threads)
	}

	updated, err := svc.UpdateThread(context.Background(), ident, thread.ID, UpdateThreadInput{Title: ptr("Renamed"), Mode: ptr(ThreadModeWork)})
	if err != nil {
		t.Fatalf("UpdateThread() error = %v", err)
	}
	if updated.Title != "Renamed" || updated.Mode != ThreadModeWork {
		t.Fatalf("updated = %+v", updated)
	}

	archived, err := svc.ArchiveThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ArchiveThread() error = %v", err)
	}
	if archived.LifecycleStatus != ThreadLifecycleArchived || archived.ArchivedAt == nil {
		t.Fatalf("archived = %+v", archived)
	}
	active, err := svc.ListThreads(context.Background(), ident, false)
	if err != nil {
		t.Fatalf("ListThreads(active) error = %v", err)
	}
	if len(active) != 0 {
		t.Fatalf("active = %+v", active)
	}
	got, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread(archived) error = %v", err)
	}
	if got.ID != thread.ID {
		t.Fatalf("got = %+v", got)
	}
}

func TestThreadValidation(t *testing.T) {
	svc := NewMemoryService()
	_, err := svc.CreateThread(context.Background(), identity.LocalDevIdentity(), CreateThreadInput{Title: " ", Mode: ThreadModeChat})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty title err = %v", err)
	}
	_, err = svc.CreateThread(context.Background(), identity.LocalDevIdentity(), CreateThreadInput{Title: "Thread", Mode: ThreadMode("run")})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid mode err = %v", err)
	}
}

func TestMessageCreationIsIdempotent(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Messages", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	msg, created, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: " hello ", ClientMessageID: "client-1"})
	if err != nil {
		t.Fatalf("CreateMessage() error = %v", err)
	}
	if !created {
		t.Fatal("CreateMessage() created = false, want true")
	}
	if msg.Role != MessageRoleUser || msg.Content != "hello" {
		t.Fatalf("msg = %+v", msg)
	}
	threadAfterFirst, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread() error = %v", err)
	}
	dup, created, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: " hello again ", ClientMessageID: "client-1"})
	if err != nil {
		t.Fatalf("CreateMessage(duplicate) error = %v", err)
	}
	if created {
		t.Fatal("CreateMessage(duplicate) created = true, want false")
	}
	if dup.ID != msg.ID || dup.Content != msg.Content {
		t.Fatalf("dup = %+v, msg = %+v", dup, msg)
	}
	threadAfterDuplicate, err := svc.GetThread(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("GetThread() duplicate error = %v", err)
	}
	if !threadAfterDuplicate.UpdatedAt.Equal(threadAfterFirst.UpdatedAt) {
		t.Fatalf("duplicate changed updated_at: first=%s duplicate=%s", threadAfterFirst.UpdatedAt, threadAfterDuplicate.UpdatedAt)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].ID != msg.ID {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestAppendAssistantMessagePersistsAssistantRole(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Assistant", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	message, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, AppendAssistantMessageInput{Content: "  hello from model  ", Metadata: map[string]any{"api_key": "secret", "run_id": "run_1"}})
	if err != nil {
		t.Fatalf("AppendAssistantMessage() error = %v", err)
	}
	if message.Role != MessageRoleAssistant || message.Content != "hello from model" {
		t.Fatalf("message = %+v", message)
	}
	if message.Metadata["api_key"] != "[redacted]" {
		t.Fatalf("metadata = %+v", message.Metadata)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 || messages[0].Role != MessageRoleAssistant {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestAppendAssistantMessageRejectsDuplicateRunMessage(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Assistant", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	input := AppendAssistantMessageInput{Content: "hello from model", Metadata: map[string]any{"run_id": "run_1"}}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, input); err != nil {
		t.Fatalf("AppendAssistantMessage() error = %v", err)
	}
	if _, err := svc.AppendAssistantMessage(context.Background(), ident, thread.ID, input); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("duplicate err = %v", err)
	}
	messages, err := svc.ListMessages(context.Background(), ident, thread.ID)
	if err != nil {
		t.Fatalf("ListMessages() error = %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("messages = %+v", messages)
	}
}

func TestMessageValidationAndThreadNotFound(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	_, _, err := svc.CreateMessage(context.Background(), ident, "thr_missing", CreateMessageInput{Content: "hello"})
	if err == nil || ErrorCode(err) != CodeThreadNotFound {
		t.Fatalf("missing thread err = %v", err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Messages", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	_, _, err = svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "   "})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("empty message err = %v", err)
	}
}

func TestRedactEventMetadataRedactsSensitiveKeys(t *testing.T) {
	metadata := RedactEventMetadata(map[string]any{"api_key": "sk-live-123", "nested": map[string]any{"password": "abc123"}, "timezone": "UTC"})
	if metadata["api_key"] != "[redacted]" {
		t.Fatalf("api_key was not redacted: %+v", metadata)
	}
	nested := metadata["nested"].(map[string]any)
	if nested["password"] != "[redacted]" {
		t.Fatalf("nested password was not redacted: %+v", metadata)
	}
	if metadata["timezone"] != "UTC" {
		t.Fatalf("safe metadata was changed: %+v", metadata)
	}
}

func TestSyncBuiltInPersonasCreatesDefaultVersionIdempotently(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	configs := []BuiltInPersonaConfig{{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General assistant",
		SystemPrompt:     "never expose this prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "gpt-5.5"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "default budget",
		Version:          "2026-05-25.1",
		IsDefault:        true,
	}}

	first, err := svc.SyncBuiltInPersonas(context.Background(), ident, configs)
	if err != nil {
		t.Fatalf("SyncBuiltInPersonas() error = %v", err)
	}
	second, err := svc.SyncBuiltInPersonas(context.Background(), ident, configs)
	if err != nil {
		t.Fatalf("SyncBuiltInPersonas() second error = %v", err)
	}
	if first.CreatedPersonas != 1 || first.CreatedVersions != 1 || second.CreatedPersonas != 0 || second.CreatedVersions != 0 {
		t.Fatalf("first=%+v second=%+v", first, second)
	}
	personas, err := svc.ListPersonas(context.Background(), ident)
	if err != nil {
		t.Fatalf("ListPersonas() error = %v", err)
	}
	if len(personas) != 1 || personas[0].ActiveVersion != "2026-05-25.1" || !personas[0].IsDefault {
		t.Fatalf("personas = %+v", personas)
	}
}

func TestSyncBuiltInPersonasPreservesOldVersionForRunSnapshots(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	base := BuiltInPersonaConfig{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General assistant",
		SystemPrompt:     "old prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "old-model"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "old budget",
		Version:          "2026-05-25.1",
		IsDefault:        true,
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{base}); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Persona", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "fallback"}); err != nil {
		t.Fatal(err)
	}
	updated := base
	updated.Version = "2026-05-25.2"
	updated.SystemPrompt = "new prompt"
	updated.ModelRoute.Model = "new-model"
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{updated}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_persona", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Persona.Version != "2026-05-25.1" || context.Persona.ModelRoute.Model != "old-model" {
		t.Fatalf("persona snapshot = %+v", context.Persona)
	}
}

func TestPrepareRunContextResolvesRunThreadAndDefaultPersona(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	configs := []BuiltInPersonaConfig{
		{Slug: "default", Name: "Default", Description: "Default persona", SystemPrompt: "default prompt", ModelRoute: PersonaModelRoute{ProviderID: "custom", Model: "default-model"}, AllowedToolNames: []string{ToolNameCurrentTime}, ReasoningMode: "balanced", BudgetSummary: "default budget", Version: "1", IsDefault: true},
		{Slug: "focused", Name: "Focused", Description: "Focused persona", SystemPrompt: "focused prompt", ModelRoute: PersonaModelRoute{ProviderID: "custom", Model: "focused-model"}, AllowedToolNames: []string{}, ReasoningMode: "low", BudgetSummary: "small budget", Version: "1"},
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, configs); err != nil {
		t.Fatal(err)
	}
	personas, err := svc.ListPersonas(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	var focused Persona
	for _, persona := range personas {
		if persona.Slug == "focused" {
			focused = persona
		}
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Persona", Mode: ThreadModeChat, PersonaID: focused.ID})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "fallback"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_persona", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Run.ID != run.ID || context.Persona.ID != focused.ID || context.Persona.ResolvedFrom != PersonaResolvedFromThread {
		t.Fatalf("context persona = %+v", context.Persona)
	}
	if context.ProviderRoute.Model != "focused-model" || len(context.EnabledTools) != 0 {
		t.Fatalf("route/tools = %+v %+v", context.ProviderRoute, context.EnabledTools)
	}
	if summary := context.SafeSummary(); summary["persona_system_prompt"] != nil || summary["persona_version"] != "1" || summary["persona_name"] != "Focused" {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestPrepareRunContextPreservesExplicitLocalProviderOverDefaultPersonaRoute(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default persona",
		SystemPrompt:     "default prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "default-model"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "default budget",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Local Codex", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "local_codex", Model: "gpt-5"}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_local_codex", LeaseSeconds: 5})
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.ProviderRoute.ProviderID != "local_codex" || context.ProviderRoute.Model != "gpt-5" {
		t.Fatalf("provider route = %+v", context.ProviderRoute)
	}
}

func TestSyncBuiltInPersonasRejectsUnsupportedTool(t *testing.T) {
	svc := NewMemoryService()
	_, err := svc.SyncBuiltInPersonas(context.Background(), identity.LocalDevIdentity(), []BuiltInPersonaConfig{{
		Slug:             "bad",
		Name:             "Bad",
		Description:      "Bad persona",
		SystemPrompt:     "prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{"runtime.shell"},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "1",
		IsDefault:        true,
	}})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("err = %v", err)
	}
}

func TestSyncBuiltInPersonasKeepsUndiscoveredMCPOutOfEnabledTools(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default persona",
		SystemPrompt:     "secret prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{ToolNameCurrentTime, "mcp.local-search.search"},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatalf("SyncBuiltInPersonas() error = %v", err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "MCP persona", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_mcp_persona", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if len(ctxData.EnabledTools) != 1 {
		t.Fatalf("enabled tools = %+v", ctxData.EnabledTools)
	}
	if ctxData.EnabledTools[0].Name != ToolNameCurrentTime {
		t.Fatalf("enabled tools = %+v", ctxData.EnabledTools)
	}
}

func TestMCPAvailabilityIncludesDiscoveryFailureSummary(t *testing.T) {
	createdAt := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	summary := mcpAvailabilityForToolResolutions([]ToolResolution{{
		Name:           "mcp.local-search.search",
		ApprovalPolicy: "always_required",
		ExecutionState: "discovered_non_executable",
	}}, []RunEvent{{
		Type:      "mcp_discovery_failed",
		Metadata:  map[string]any{"server_slug": "local-search", "status": "failed", "error_code": "mcp_discovery_timeout"},
		CreatedAt: createdAt,
	}})

	if summary.ServersConfigured != 1 || summary.ServersEnabled != 1 || summary.ServersFailed != 1 {
		t.Fatalf("mcp server counts = %+v", summary)
	}
	if len(summary.RedactedErrorCodes) != 1 || summary.RedactedErrorCodes[0] != "mcp_discovery_timeout" {
		t.Fatalf("mcp error codes = %+v", summary.RedactedErrorCodes)
	}
	if summary.LastDiscoveredAt != createdAt.Format(time.RFC3339Nano) {
		t.Fatalf("mcp last discovered = %q", summary.LastDiscoveredAt)
	}
	if len(summary.ServerSummaries) != 1 {
		t.Fatalf("mcp server summaries = %+v", summary.ServerSummaries)
	}
	server := summary.ServerSummaries[0]
	if server.ServerSafeID != "mcp:local-search" || server.ServerSlug != "local-search" || server.DiscoveryStatus != "failed" || server.RedactedErrorCode != "mcp_discovery_timeout" {
		t.Fatalf("mcp server summary = %+v", server)
	}
}

func TestRunValidation(t *testing.T) {
	if err := ValidateRunStatus(RunStatusRunning); err != nil {
		t.Fatalf("ValidateRunStatus(running) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatusBlockedOnToolApproval); err != nil {
		t.Fatalf("ValidateRunStatus(blocked_on_tool_approval) error = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategoryFinal); err != nil {
		t.Fatalf("ValidateRunEventCategory(final) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatusQueued); err != nil {
		t.Fatalf("ValidateRunStatus(queued) error = %v", err)
	}
	if err := ValidateToolCallApprovalStatus(ToolCallApprovalRequired); err != nil {
		t.Fatalf("ValidateToolCallApprovalStatus(required) error = %v", err)
	}
	if err := ValidateToolCallExecutionStatus(ToolCallExecutionBlocked); err != nil {
		t.Fatalf("ValidateToolCallExecutionStatus(blocked) error = %v", err)
	}
	if err := ValidateRunStatus(RunStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid status err = %v", err)
	}
	if err := ValidateRunEventCategory(RunEventCategory("tool")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid category err = %v", err)
	}
	if err := ValidateToolCallApprovalStatus(ToolCallApprovalStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid approval status err = %v", err)
	}
	if err := ValidateToolCallExecutionStatus(ToolCallExecutionStatus("unknown")); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("invalid execution status err = %v", err)
	}
	if !IsRunActive(RunStatusRunning) || !IsRunActive(RunStatusBlockedOnToolApproval) || IsRunActive(RunStatusCompleted) {
		t.Fatalf("active status helpers returned wrong result")
	}
	if !IsRunTerminal(RunStatusStopped) || IsRunTerminal(RunStatusPending) {
		t.Fatalf("terminal status helpers returned wrong result")
	}
}

func TestRecordToolCallRequestValidatesM7SafetyBoundary(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 tool safety", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	call, events, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest() error = %v", err)
	}
	if call.ArgumentsSummary["timezone"] != "UTC" || events[0].Metadata["arguments_summary"].(map[string]any)["timezone"] != "UTC" {
		t.Fatalf("tool metadata: call=%+v events=%+v", call, events)
	}
	again, againEvents, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked})
	if err != nil {
		t.Fatalf("RecordToolCallRequest(duplicate) error = %v", err)
	}
	if again.ID != call.ID || len(againEvents) != 0 {
		t.Fatalf("duplicate call = %+v events = %+v", again, againEvents)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "Asia/Shanghai"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("non-UTC timezone err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.get_current_time", ArgumentsSummary: map[string]any{"timezone": "UTC", "api_key": "sk-live-123"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unknown argument err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_2", ToolName: "runtime.unknown", ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_2", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("unsupported tool err = %v", err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatalf("WorkerQueueDiagnostics() error = %v", err)
	}
	if diagnostics.BlockedToolApprovalCount != 1 {
		t.Fatalf("BlockedToolApprovalCount = %d, want 1", diagnostics.BlockedToolApprovalCount)
	}
}

func TestToolCallExecutionEventsRedactResultAndErrors(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "M7 result redaction", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}
	call, events, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_1", map[string]any{"timezone": "UTC", "api_key": "sk-live-secret"})
	if err != nil {
		t.Fatal(err)
	}
	if call.ResultSummary["api_key"] == "sk-live-secret" {
		t.Fatalf("result was not redacted: %+v", call.ResultSummary)
	}
	if len(events) != 1 || events[0].Type != EventToolCallSucceeded {
		t.Fatalf("events = %+v", events)
	}
	if result, ok := events[0].Metadata["result_summary"].(map[string]any); ok && result["api_key"] == "sk-live-secret" {
		t.Fatalf("event metadata was not redacted: %+v", events[0].Metadata)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusRunning || got.CompletedAt != nil {
		t.Fatalf("run = %+v", got)
	}
}

func TestStartRunCreatesInitialLifecycleEvent(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{ScriptName: "m4_smoke token"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if run.ThreadID != thread.ID || run.Status != RunStatusQueued || run.Source != RunSourceLocalSimulated {
		t.Fatalf("run = %+v", run)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 2 || events[0].Sequence != 1 || events[0].Type != "run_created" || events[1].Type != EventRunQueued {
		t.Fatalf("events = %+v", events)
	}
	if events[0].Metadata["script_name"] != "[redacted]" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestStartRunSupportsModelGatewaySource(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "gpt-5.5"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if run.Source != RunSourceModelGateway || run.Title != "Model gateway run" {
		t.Fatalf("run = %+v", run)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if events[0].Metadata["provider_id"] != "custom" || events[0].Metadata["model"] != "gpt-5.5" {
		t.Fatalf("metadata = %+v", events[0].Metadata)
	}
}

func TestPrepareRunContextRestoresDurableRunThreadMessagesJobRouteAndTools(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Context", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model-a"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_context", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}

	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if context.Run.ID != run.ID || context.Thread.ID != thread.ID || context.Job.ID != job.ID {
		t.Fatalf("context boundary = %+v", context)
	}
	if len(context.Messages) != 1 || context.Messages[0].ID != message.ID {
		t.Fatalf("messages = %+v", context.Messages)
	}
	if context.ProviderRoute.ProviderID != "custom" || context.ProviderRoute.Model != "model-a" || !context.ProviderRoute.Available {
		t.Fatalf("provider route = %+v", context.ProviderRoute)
	}
	if len(context.EnabledTools) != 1 || context.EnabledTools[0].Name != ToolNameCurrentTime {
		t.Fatalf("tools = %+v", context.EnabledTools)
	}
	summary := context.SafeSummary()
	if summary["message_count"] != 1 || summary["provider_id"] != "custom" || summary["enabled_tool_count"] != 1 {
		t.Fatalf("summary = %+v", summary)
	}
}

func TestPrepareRunContextFailsBeforeRuntimeForMissingProviderRoute(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Context missing", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_context", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}

	_, err = svc.PrepareRunContext(context.Background(), ident, job)
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("PrepareRunContext() err = %v", err)
	}
}

func TestStartRunRejectsSecondActiveRunForThread(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	_, err = svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err == nil || ErrorCode(err) != CodeActiveRunExists {
		t.Fatalf("second active run err = %v", err)
	}
}

func TestStartRunAndJobCreationAreAtomicFromServiceBoundary(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Jobs", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.QueuedCount != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	job, claimedRun, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || job.RunID != run.ID || claimedRun.Status != RunStatusRunning {
		t.Fatalf("job=%+v run=%+v ok=%v", job, claimedRun, ok)
	}
}

func TestFailBackgroundJobRedactsFailureAndTerminalEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Fail", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	failed, changed, err := svc.FailBackgroundJob(context.Background(), ident, FailBackgroundJobInput{JobID: job.ID, WorkerID: "worker_test", OwnershipVersion: job.OwnershipVersion, ErrorCode: "provider_failed", ErrorMessage: "token secret leaked"})
	if err != nil {
		t.Fatal(err)
	}
	if !changed || failed.Status != BackgroundJobStatusFailed || failed.LastError == nil || *failed.LastError != "[redacted]" {
		t.Fatalf("failed = %+v", failed)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusFailed || got.ErrorMessage == nil || *got.ErrorMessage != "[redacted]" {
		t.Fatalf("run = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	if events[len(events)-2].Type != EventJobAttemptFailed || events[len(events)-1].Type != EventRunFailed || events[len(events)-1].Summary != "[redacted]" {
		t.Fatalf("events = %+v", events)
	}
}

func TestRecoverBackgroundJobsReschedulesExpiredLeaseAndRejectsStaleOwner(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Recover", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_stale", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || job.OwnershipVersion != 1 {
		t.Fatalf("job = %+v ok=%v", job, ok)
	}
	base = base.Add(2 * time.Second)
	recoveries, err := svc.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{})
	if err != nil {
		t.Fatal(err)
	}
	if len(recoveries) != 1 || recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusQueued || recoveries[0].Run.Status != RunStatusRecovering {
		t.Fatalf("recoveries = %+v", recoveries)
	}
	if recoveries[0].Events[0].Type != EventJobRecovering || recoveries[0].Events[1].Type != EventJobRetryScheduled {
		t.Fatalf("events = %+v", recoveries[0].Events)
	}
	if !recoveries[0].Job.ScheduledAt.After(base) {
		t.Fatalf("retry was not backed off: scheduled_at=%s base=%s", recoveries[0].Job.ScheduledAt, base)
	}
	if _, changed, err := svc.CompleteBackgroundJob(context.Background(), ident, CompleteBackgroundJobInput{JobID: job.ID, WorkerID: "worker_stale", OwnershipVersion: job.OwnershipVersion}); err != nil || changed {
		t.Fatalf("stale completion changed=%v err=%v", changed, err)
	}
	base = recoveries[0].Job.ScheduledAt
	claimed, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_fresh", LeaseSeconds: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || claimed.OwnershipVersion <= job.OwnershipVersion || claimed.AttemptCount != 2 {
		t.Fatalf("fresh claim = %+v ok=%v", claimed, ok)
	}
}

func TestRecoverBackgroundJobsExhaustsRetriesWithRedactedFailure(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	base := time.Date(2026, 5, 24, 10, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return base }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Recover", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 1; attempt <= 3; attempt++ {
		if _, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_retry", LeaseSeconds: 1}); err != nil || !ok {
			t.Fatalf("claim attempt %d ok=%v err=%v", attempt, ok, err)
		}
		base = base.Add(2 * time.Second)
		recoveries, err := svc.RecoverBackgroundJobs(context.Background(), ident, RecoverBackgroundJobsInput{ErrorMessage: "token secret leaked"})
		if err != nil {
			t.Fatal(err)
		}
		if len(recoveries) != 1 {
			t.Fatalf("attempt %d recoveries = %+v", attempt, recoveries)
		}
		if attempt < 3 && recoveries[0].Exhausted {
			t.Fatalf("attempt %d exhausted early: %+v", attempt, recoveries[0])
		}
		if attempt == 3 {
			if !recoveries[0].Exhausted || recoveries[0].Job.Status != BackgroundJobStatusDead || recoveries[0].Run.Status != RunStatusFailed {
				t.Fatalf("final recovery = %+v", recoveries[0])
			}
			if recoveries[0].Run.ErrorMessage == nil || *recoveries[0].Run.ErrorMessage != "[redacted]" || recoveries[0].Events[0].Summary != "[redacted]" {
				t.Fatalf("final recovery did not redact = %+v", recoveries[0])
			}
		}
		if attempt < 3 {
			base = recoveries[0].Job.ScheduledAt
		}
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != RunStatusFailed || got.CompletedAt == nil {
		t.Fatalf("run = %+v", got)
	}
}

func TestStopRunCancelsQueuedJobAndPreventsClaim(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	stopped, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if stopped.Run.StopRequestedAt == nil || stopped.Run.Status != RunStatusStopped {
		t.Fatalf("stopped = %+v", stopped)
	}
	if _, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_test", LeaseSeconds: 1}); err != nil || ok {
		t.Fatalf("claim after stop ok=%v err=%v", ok, err)
	}
	diagnostics, err := svc.WorkerQueueDiagnostics(context.Background(), ident)
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics.QueuedCount != 0 || diagnostics.LeasedCount != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}

func TestStopRunRecordsStoppedTerminalEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	stopped, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if stopped.Result != StopRunResultStopped || stopped.Run.Status != RunStatusStopped || stopped.Run.CompletedAt == nil {
		t.Fatalf("stopped = %+v", stopped)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 4 || events[2].Type != EventStopRequested || events[3].Category != RunEventCategoryFinal {
		t.Fatalf("events = %+v", events)
	}
}

func TestStopRunReturnsAlreadyTerminalWithoutChangingOutcome(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	output, err := svc.StopRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("StopRun() error = %v", err)
	}
	if output.Result != StopRunResultAlreadyTerminal || output.Run.Status != RunStatusCompleted {
		t.Fatalf("output = %+v", output)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("events = %+v", events)
	}
}

func TestAppendRunEventRejectsTerminalRun(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	_, err = svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "late", Summary: "Late"})
	if err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("late append err = %v", err)
	}
}

func TestRunEventRedactsSecretText(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	content := "postgres://loomi:secret@localhost/db"
	event, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryError, Type: "run_failed", Summary: "token leaked", Content: &content, Metadata: map[string]any{"database_url": "postgresql://user:password=secret@localhost/db", "nested": map[string]any{"bearer": "Bearer abc"}}})
	if err != nil {
		t.Fatalf("AppendRunEvent(error) error = %v", err)
	}
	if event.Summary != "[redacted]" || event.Content == nil || *event.Content != "[redacted]" {
		t.Fatalf("event = %+v", event)
	}
	if event.Metadata["database_url"] != "[redacted]" {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
}

func TestAppendRunEventOrdersPersistedEvents(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded"}); err != nil {
		t.Fatalf("AppendRunEvent(progress) error = %v", err)
	}
	final, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed"})
	if err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	if final.Sequence != 4 {
		t.Fatalf("final sequence = %d", final.Sequence)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 1)
	if err != nil {
		t.Fatalf("ListRunEvents(after=1) error = %v", err)
	}
	if len(events) != 3 || events[0].Sequence != 2 || events[1].Sequence != 3 || events[2].Sequence != 4 {
		t.Fatalf("events = %+v", events)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != RunStatusCompleted || got.CompletedAt == nil {
		t.Fatalf("run after final = %+v", got)
	}
}

func ptr[T any](v T) *T { return &v }
