package productdata

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestMemoryServiceThreadContextSources(t *testing.T) {
	svc := NewMemoryService()
	var _ ContextSourceService = svc
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Sources", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	source, err := svc.CreateContextSource(context.Background(), ident, CreateContextSourceInput{ThreadID: thread.ID, Kind: ContextSourceKindURL, Title: "Docs", Locator: " https://example.com/docs?token=secret "})
	if err != nil {
		t.Fatal(err)
	}
	if source.ID == "" || source.ThreadID != thread.ID || source.Kind != ContextSourceKindURL || source.Title != "Docs" || source.Locator != "https://example.com/docs" || source.Status != ContextSourceStatusRegistered {
		t.Fatalf("source = %+v", source)
	}
	listed, err := svc.ListContextSources(context.Background(), ident, ListContextSourcesInput{ThreadID: thread.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0].ID != source.ID {
		t.Fatalf("listed = %+v", listed)
	}
	if _, err := svc.CreateContextSource(context.Background(), ident, CreateContextSourceInput{ThreadID: thread.ID, Kind: ContextSourceKindURL, Title: "Bad", Locator: "http://127.0.0.1/private"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("private url err = %v", err)
	}
	if _, err := svc.CreateContextSource(context.Background(), ident, CreateContextSourceInput{ThreadID: thread.ID, Kind: ContextSourceKindWorkspacePath, Title: "Secret", Locator: ".env"}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("sensitive path err = %v", err)
	}
	other, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Other", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	otherSources, err := svc.ListContextSources(context.Background(), ident, ListContextSourcesInput{ThreadID: other.ID, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(otherSources) != 0 {
		t.Fatalf("otherSources = %+v", otherSources)
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

func TestPrepareRunContextLoadsThreadContextSources(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Sources", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "Read the linked docs"})
	if err != nil {
		t.Fatal(err)
	}
	source, err := svc.CreateContextSource(context.Background(), ident, CreateContextSourceInput{ThreadID: thread.ID, Kind: ContextSourceKindURL, Title: "Docs", Locator: "https://example.com/docs?token=secret"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker-1", LeaseSeconds: 30})
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if len(ctxData.ContextSources) != 1 || ctxData.ContextSources[0].ID != source.ID || ctxData.ContextSources[0].Locator != "https://example.com/docs" {
		t.Fatalf("context sources = %+v", ctxData.ContextSources)
	}
	summary := ctxData.SafeSummary()
	if summary["context_source_count"] != 1 || summary["context_source_kinds"].([]string)[0] != string(ContextSourceKindURL) {
		t.Fatalf("summary = %+v", summary)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, job.RunID, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, event := range events {
		if event.Type == EventContextSourcesLoaded {
			found = true
			if event.Metadata["context_source_count"] != 1 {
				t.Fatalf("source event metadata = %+v", event.Metadata)
			}
			if strings.Contains(event.Summary, "token") {
				t.Fatalf("unsafe event summary = %q", event.Summary)
			}
		}
	}
	if !found {
		t.Fatalf("missing %s event in %+v", EventContextSourcesLoaded, events)
	}
}

func TestWorkToolResolutionsFollowLatestUserIntent(t *testing.T) {
	all := toolResolutionsForNamesAndEvents(BuiltInPersonas()[0].AllowedToolNames, nil)
	listFiles := workToolResolutionsForLatestIntent(all, []Message{{Role: MessageRoleUser, Content: "列一下当前工作目录里的文件，简单分类。"}})
	if !hasToolResolution(listFiles, ToolNameWorkspaceListDirectory) || !hasToolResolution(listFiles, ToolNameWorkspaceTreeSummary) || !hasToolResolution(listFiles, ToolNameWorkspaceRead) {
		t.Fatalf("file listing tools missing: %+v", listFiles)
	}
	for _, disallowed := range []string{ToolNameSandboxExecCommand, ToolNameAgentSpawn, ToolNameArtifactCreateText, ToolNameBrowserOpen} {
		if hasToolResolution(listFiles, disallowed) {
			t.Fatalf("file listing exposed %s: %+v", disallowed, listFiles)
		}
	}

	hello := workToolResolutionsForLatestIntent(all, []Message{{Role: MessageRoleUser, Content: "你好呀"}})
	for _, disallowed := range []string{ToolNameWorkspaceGlob, ToolNameSandboxExecCommand, ToolNameAgentSpawn, ToolNameArtifactCreateText, ToolNameBrowserOpen, ToolNameWebSearch} {
		if hasToolResolution(hello, disallowed) {
			t.Fatalf("casual greeting exposed %s: %+v", disallowed, hello)
		}
	}

	runTests := workToolResolutionsForLatestIntent(all, []Message{{Role: MessageRoleUser, Content: "帮我运行 go test ./..."}})
	if !hasToolResolution(runTests, ToolNameSandboxExecCommand) {
		t.Fatalf("command intent did not expose sandbox exec: %+v", runTests)
	}
}

func TestPrepareRunContextUsesRunScopedWorkspaceRootSnapshot(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: firstRoot}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Workspace root", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "列一下目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: secondRoot}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_workspace_root", LeaseSeconds: 5})
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

	if ctxData.WorkspaceRoot.Path != firstRoot {
		t.Fatalf("workspace root snapshot = %q, want %q", ctxData.WorkspaceRoot.Path, firstRoot)
	}
}

func TestPrepareRunContextExposesSafeWorkspaceLabel(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	root := filepath.Join(t.TempDir(), "Downloads")
	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: root}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Workspace label", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "看一下下载目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_workspace_label", LeaseSeconds: 5})
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

	if ctxData.WorkspaceRoot.DisplayName != "Downloads" {
		t.Fatalf("workspace label = %q, want Downloads", ctxData.WorkspaceRoot.DisplayName)
	}
	summary := ctxData.SafeSummary()
	if summary["workspace_label"] != "Downloads" {
		t.Fatalf("safe summary workspace label = %+v", summary)
	}
	encoded := fmt.Sprint(summary)
	if strings.Contains(encoded, root) {
		t.Fatalf("safe summary leaked absolute root: %s", encoded)
	}
}

func TestNewThreadUsesCurrentWorkspaceInsteadOfPreviousThreadRoot(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	arkloopRoot := filepath.Join(t.TempDir(), "Arkloop")
	downloadsRoot := filepath.Join(t.TempDir(), "Downloads")
	for _, root := range []string{arkloopRoot, downloadsRoot} {
		if err := os.Mkdir(root, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: arkloopRoot}); err != nil {
		t.Fatal(err)
	}
	arkloopThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Arkloop", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	arkloopMessage, _, err := svc.CreateMessage(context.Background(), ident, arkloopThread.ID, CreateMessageInput{Content: "看这个目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, arkloopThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: arkloopMessage.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: downloadsRoot}); err != nil {
		t.Fatal(err)
	}
	downloadsThread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Downloads", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	downloadsMessage, _, err := svc.CreateMessage(context.Background(), ident, downloadsThread.ID, CreateMessageInput{Content: "列一下刚选目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, downloadsThread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: downloadsMessage.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}

	var downloadsContext RunContext
	for i := 0; i < 2; i++ {
		job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: fmt.Sprintf("worker_history_%d", i), LeaseSeconds: 5})
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("claim ok = false")
		}
		if job.ThreadID == downloadsThread.ID {
			downloadsContext, err = svc.PrepareRunContext(context.Background(), ident, job)
			if err != nil {
				t.Fatal(err)
			}
		}
	}

	if downloadsContext.WorkspaceRoot.Path != downloadsRoot || downloadsContext.WorkspaceRoot.DisplayName != "Downloads" {
		t.Fatalf("downloads context workspace = %+v, want %q", downloadsContext.WorkspaceRoot, downloadsRoot)
	}
	if downloadsContext.WorkspaceRoot.Path == arkloopRoot || downloadsContext.WorkspaceRoot.DisplayName == "Arkloop" {
		t.Fatalf("new thread used previous workspace: %+v", downloadsContext.WorkspaceRoot)
	}
}

func TestApprovedToolResumePreservesRunScopedWorkspaceRootSnapshot(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: firstRoot}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Workspace approval", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "改文件"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_write", ToolName: ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "notes.txt", "content": "hello\n"}, ArgumentsHash: "hash_write", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: secondRoot}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_write"); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_workspace_resume_root", LeaseSeconds: 5})
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

	if ctxData.WorkspaceRoot.Path != firstRoot {
		t.Fatalf("resume workspace root snapshot = %q, want %q", ctxData.WorkspaceRoot.Path, firstRoot)
	}
}

func TestAgentDelegateReconcilePreservesRunScopedWorkspaceRootSnapshot(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	firstRoot := t.TempDir()
	secondRoot := t.TempDir()
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: firstRoot}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, BuiltInPersonas()); err != nil {
		t.Fatal(err)
	}
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Delegate workspace", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "delegate"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_delegate", ToolName: ToolNameAgentDelegate, ArgumentsSummary: map[string]any{"task_id": "task"}, ArgumentsHash: "hash_delegate", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_delegate"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_delegate"); err != nil {
		t.Fatal(err)
	}
	task, err := svc.SpawnAgentTask(context.Background(), ident, SpawnAgentTaskInput{ThreadID: thread.ID, RunID: run.ID, Role: "reviewer", Goal: "Review workspace resume"})
	if err != nil {
		t.Fatal(err)
	}
	delegated, err := svc.DelegateAgentTask(context.Background(), ident, DelegateAgentTaskInput{ThreadID: thread.ID, TaskID: task.ID, ParentToolCallID: "tc_delegate"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.SaveWorkspaceRootConfig(context.Background(), ident, WorkspaceRootConfig{Path: secondRoot}); err != nil {
		t.Fatal(err)
	}
	svc.mu.Lock()
	childRun := svc.runs[delegated.ChildRunID]
	now := svc.now()
	childRun.Status = RunStatusCompleted
	childRun.CompletedAt = &now
	childRun.UpdatedAt = now
	svc.runs[childRun.ID] = childRun
	svc.messages[childRun.ThreadID] = append(svc.messages[childRun.ThreadID], Message{ID: NewMessageID(), ThreadID: childRun.ThreadID, UserID: childRun.UserID, Role: MessageRoleAssistant, Content: "Child done", CreatedAt: now})
	svc.mu.Unlock()
	if _, err := svc.ReconcileAgentTaskChildRuns(context.Background(), ident, 1); err != nil {
		t.Fatal(err)
	}
	var resumeJob BackgroundJob
	for _, job := range svc.backgroundJobs {
		if job.RunID == run.ID && metadataStringValue(job.Metadata, "resume_reason") == "agent_child_run_completed" {
			resumeJob = job
			break
		}
	}
	if resumeJob.ID == "" {
		t.Fatal("resume job not found")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, resumeJob)
	if err != nil {
		t.Fatal(err)
	}
	if ctxData.WorkspaceRoot.Path != firstRoot {
		t.Fatalf("delegate resume workspace root snapshot = %q, want %q", ctxData.WorkspaceRoot.Path, firstRoot)
	}
}

func hasToolResolution(tools []ToolResolution, name string) bool {
	for _, tool := range tools {
		if tool.Name == name {
			return true
		}
	}
	return false
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
	metadata := RedactEventMetadata(map[string]any{"api_key": "sk-live-123", "nested": map[string]any{"password": "abc123"}, "timezone": "UTC", "workspace_root_path": "/var/tmp/project"})
	if metadata["api_key"] != "[redacted]" {
		t.Fatalf("api_key was not redacted: %+v", metadata)
	}
	if metadata["workspace_root_path"] != "[redacted]" {
		t.Fatalf("workspace root was not redacted: %+v", metadata)
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

func TestSyncBuiltInPersonasUpdatesExistingVersionDefinition(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	config := BuiltInPersonaConfig{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General assistant",
		SystemPrompt:     "prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "gpt-5.5"},
		AllowedToolNames: []string{ToolNameCurrentTime},
		ReasoningMode:    "balanced",
		BudgetSummary:    "default budget",
		Version:          "2026-05-27.2",
		IsDefault:        true,
	}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{config}); err != nil {
		t.Fatal(err)
	}
	config.AllowedToolNames = []string{ToolNameCurrentTime, ToolNameWorkspaceTreeSummary, ToolNameWorkspaceListDirectory}
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{config}); err != nil {
		t.Fatal(err)
	}

	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Work", Mode: ThreadModeWork})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "分类当前目录"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"}); err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_persona_update", LeaseSeconds: 5})
	if err != nil || !ok {
		t.Fatalf("claim ok=%v err=%v", ok, err)
	}
	context, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	if !hasToolResolution(context.EnabledTools, ToolNameWorkspaceTreeSummary) || !hasToolResolution(context.EnabledTools, ToolNameWorkspaceListDirectory) {
		t.Fatalf("enabled tools = %+v", context.EnabledTools)
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

func TestWorkModeScopedToolsOnlyEnabledForWorkModeRunContext(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	if _, err := svc.SyncBuiltInPersonas(context.Background(), ident, []BuiltInPersonaConfig{{
		Slug:             "default",
		Name:             "Default",
		Description:      "Default persona",
		SystemPrompt:     "prompt",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "model"},
		AllowedToolNames: []string{ToolNameCurrentTime, ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead, ToolNameWorkspaceListDirectory, ToolNameWorkspaceTreeSummary, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchPreview, ToolNameWorkspacePatchApply, ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess, ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameLSPDefinition, ToolNameLSPHover, ToolNameWebFetch, ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameBrowserScreenshot, ToolNameBrowserType, ToolNameBrowserPress, ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList, ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentStart, ToolNameAgentDelegate, ToolNameAgentComplete, ToolNameAgentFail, ToolNameTodoWrite},
		ReasoningMode:    "balanced",
		BudgetSummary:    "budget",
		Version:          "1",
		IsDefault:        true,
	}}); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []ThreadMode{ThreadModeChat, ThreadModeWork} {
		thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: string(mode), Mode: mode})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{}); err != nil {
			t.Fatal(err)
		}
		job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_" + string(mode), LeaseSeconds: 5})
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
		hasWorkspaceRead := catalogResolutionByName(ctxData.EnabledTools, ToolNameWorkspaceRead).Name != ""
		hasWorkspaceWrite := catalogResolutionByName(ctxData.EnabledTools, ToolNameWorkspaceWriteFile).Name != ""
		hasWorkspacePatchPreview := catalogResolutionByName(ctxData.EnabledTools, ToolNameWorkspacePatchPreview).Name != ""
		hasWorkspacePatchApply := catalogResolutionByName(ctxData.EnabledTools, ToolNameWorkspacePatchApply).Name != ""
		hasSandboxExec := catalogResolutionByName(ctxData.EnabledTools, ToolNameSandboxExecCommand).Name != ""
		hasLSPSymbols := catalogResolutionByName(ctxData.EnabledTools, ToolNameLSPSymbols).Name != ""
		hasWebFetch := catalogResolutionByName(ctxData.EnabledTools, ToolNameWebFetch).Name != ""
		hasBrowserOpen := catalogResolutionByName(ctxData.EnabledTools, ToolNameBrowserOpen).Name != ""
		hasArtifactCreate := catalogResolutionByName(ctxData.EnabledTools, ToolNameArtifactCreateText).Name != ""
		hasAgentSpawn := catalogResolutionByName(ctxData.EnabledTools, ToolNameAgentSpawn).Name != ""
		hasTodoWrite := catalogResolutionByName(ctxData.EnabledTools, ToolNameTodoWrite).Name != ""
		if mode == ThreadModeChat && (hasWorkspaceRead || hasWorkspaceWrite || hasWorkspacePatchPreview || hasWorkspacePatchApply || hasSandboxExec || hasLSPSymbols || hasBrowserOpen || hasArtifactCreate || hasAgentSpawn || hasTodoWrite) {
			t.Fatalf("chat enabled work-mode tools: %+v", ctxData.EnabledTools)
		}
		if mode == ThreadModeChat && !hasWebFetch {
			t.Fatalf("chat missing public web fetch tool: %+v", ctxData.EnabledTools)
		}
		if mode == ThreadModeWork && (!hasWorkspaceRead || !hasWorkspaceWrite || !hasWorkspacePatchPreview || !hasWorkspacePatchApply || !hasSandboxExec || !hasLSPSymbols || !hasWebFetch || !hasBrowserOpen || !hasArtifactCreate || !hasAgentSpawn || !hasTodoWrite) {
			t.Fatalf("work missing work-mode tools: %+v", ctxData.EnabledTools)
		}
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

func TestBuildRunContextFromRunStepStateHydratesApprovedToolResume(t *testing.T) {
	run := Run{ID: "run_1", ThreadID: "thread_1", UserID: "user_1", Source: RunSourceModelGateway}
	thread := Thread{ID: "thread_1", UserID: "user_1", Mode: ThreadModeWork}
	message := Message{ID: "msg_1", ThreadID: thread.ID, Role: MessageRoleUser, Content: "read files"}
	job := BackgroundJob{ID: "job_1", RunID: run.ID, ThreadID: thread.ID, UserID: run.UserID, Metadata: map[string]any{"tool_call_id": "tc_read"}}
	state := RunStepState{
		TriggerMessageID:      message.ID,
		ProviderID:            "custom",
		Model:                 "model-a",
		EnabledToolNames:      []string{ToolNameWorkspaceRead, ToolNameWebFetch},
		LastCompletedSequence: 7,
		CompletedToolResults: []RunStep{{
			ToolCallID: "tc_read",
			ToolName:   ToolNameWorkspaceRead,
			SafeMetadata: map[string]any{
				"tool_call_id":      "tc_read",
				"tool_name":         ToolNameWorkspaceRead,
				"arguments_summary": map[string]any{"path": "notes.txt"},
				"result_summary":    map[string]any{"path": "notes.txt", "content": "hello"},
			},
		}},
	}

	context, err := buildRunContextFromState(run, thread, []Message{message}, job, state)
	if err != nil {
		t.Fatal(err)
	}

	if context.ProviderRoute.ProviderID != "custom" || context.ProviderRoute.Model != "model-a" || !context.ProviderRoute.Available {
		t.Fatalf("provider route = %+v", context.ProviderRoute)
	}
	if !context.ContinuationProjection.Available || context.ContinuationProjection.ToolCallID != "tc_read" {
		t.Fatalf("continuation = %+v", context.ContinuationProjection)
	}
	if len(context.EnabledTools) != 2 || context.EnabledTools[0].Name != ToolNameWorkspaceRead || context.EnabledTools[1].Name != ToolNameWebFetch {
		t.Fatalf("enabled tools = %+v", context.EnabledTools)
	}
}

func TestRunStepStateCanPrepareInitialModelContext(t *testing.T) {
	run := Run{ID: "run_1", Source: RunSourceModelGateway}
	state := RunStepState{TriggerMessageID: "msg_1", ProviderID: "custom", Model: "model-a"}

	if !runStepStateCanPrepareContext(run, state) {
		t.Fatal("initial model context should be prepared from run-step state")
	}
	state.ProviderID = ""
	if runStepStateCanPrepareContext(run, state) {
		t.Fatal("model gateway context without provider route should not prepare")
	}
	state.ProviderID = "custom"
	state.EnabledToolNames = []string{"mcp.local-search.search"}
	if runStepStateCanPrepareContext(run, state) {
		t.Fatal("MCP context without schema hash should not prepare")
	}
	state.MCPToolSchemaHashes = map[string]string{"mcp.local-search.search": "hash_1"}
	if !runStepStateCanPrepareContext(run, state) {
		t.Fatal("MCP context with schema hash should prepare")
	}
}

func TestClaimToolContinuationIsSingleUsePerJobAndRecoverableByNewJob(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	now := time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Continuation claim", Mode: ThreadModeChat})
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
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_1", map[string]any{"iso_time": "2026-05-29T00:00:00Z"}); err != nil {
		t.Fatal(err)
	}

	first, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || first.Type != "model_request_started" || first.Metadata["model_phase"] != "continuation" || first.Metadata["job_id"] != "job_1" {
		t.Fatalf("first claim ok=%v event=%+v", ok, first)
	}
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_1", ProviderID: "custom", Model: "model"}); err != nil || ok {
		t.Fatalf("same job claim ok=%v err=%v", ok, err)
	}
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_2", ProviderID: "custom", Model: "model"}); err != nil || ok {
		t.Fatalf("active claim should block recovery ok=%v err=%v", ok, err)
	}
	now = now.Add(31 * time.Second)
	recovery, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_2", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	if !ok || recovery.Metadata["job_id"] != "job_2" {
		t.Fatalf("recovery claim ok=%v event=%+v", ok, recovery)
	}
	delta := "hello"
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryMessage, Type: "model_output_delta", Summary: "Model output delta", Content: &delta, Metadata: map[string]any{"model_phase": "continuation"}}); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_3", ProviderID: "custom", Model: "model"}); err != nil || ok {
		t.Fatalf("post-output claim ok=%v err=%v", ok, err)
	}
}

func TestClaimToolContinuationBlocksDifferentJobWhilePreviousClaimJobIsActive(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	now := time.Date(2026, 5, 29, 1, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Continuation claim active job", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, ClaimBackgroundJobInput{WorkerID: "worker_1", LeaseSeconds: 30})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim job ok = false")
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
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_1", map[string]any{"iso_time": "2026-05-29T01:00:00Z"}); err != nil {
		t.Fatal(err)
	}
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: job.ID, ProviderID: "custom", Model: "model"}); err != nil || !ok {
		t.Fatalf("first claim ok=%v err=%v", ok, err)
	}
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_recovery", ProviderID: "custom", Model: "model"}); err != nil || ok {
		t.Fatalf("active previous job should block recovery claim ok=%v err=%v", ok, err)
	}
	now = now.Add(31 * time.Second)
	if _, ok, err := svc.ClaimToolContinuation(context.Background(), ident, ClaimToolContinuationInput{ThreadID: thread.ID, RunID: run.ID, ToolCallID: "tc_1", JobID: "job_recovery", ProviderID: "custom", Model: "model"}); err != nil || !ok {
		t.Fatalf("expired previous job should allow recovery claim ok=%v err=%v", ok, err)
	}
}

func TestToolCallEventMetadataForStatePreservesLoopIndex(t *testing.T) {
	call := ToolCall{ToolCallID: "tc_2", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}}
	state := RunStepState{SeenToolCallIDs: []string{"tc_1", "tc_2"}}
	metadata := toolCallEventMetadataForState(state, call)
	if metadata[LoopMetadataKeyIndex] != 2 || metadata[LoopMetadataKeyMax] != DefaultMaxBoundedToolCallsPerRun {
		t.Fatalf("metadata = %+v", metadata)
	}

	next := ToolCall{ToolCallID: "tc_3", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}}
	nextMetadata := toolCallEventMetadataForState(state, next)
	if nextMetadata[LoopMetadataKeyIndex] != 3 {
		t.Fatalf("next metadata = %+v", nextMetadata)
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

func TestTerminalRunRejectsLateModelAndToolOverwrite(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeWork})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	late := "late provider output"
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryMessage, Type: "model_output_completed", Summary: "Late model output", Content: &late}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("late model err = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_late", ToolName: ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "notes.txt"}, ArgumentsHash: "hash_late", ApprovalStatus: ToolCallApprovalApproved, ExecutionStatus: ToolCallExecutionNotStarted}); err == nil || ErrorCode(err) != CodeInvalidRequest {
		t.Fatalf("late tool err = %v", err)
	}
	got, err := svc.GetRun(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != RunStatusCompleted {
		t.Fatalf("run overwritten = %+v", got)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	if len(events) != 3 || events[2].Type != EventRunCompleted {
		t.Fatalf("events overwritten = %+v", events)
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

func TestRunEventKeepsAssistantFinalContentWithBenignTokenWords(t *testing.T) {
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
	content := "## Analysis\n\n- Design Tokens are not API tokens.\n- This project has key design ideas."
	event, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryMessage, Type: "model_output_completed", Summary: "Model output completed", Content: &content, Metadata: map[string]any{"provider_id": "custom"}})
	if err != nil {
		t.Fatalf("AppendRunEvent(message) error = %v", err)
	}
	if event.Content == nil || *event.Content != content {
		t.Fatalf("assistant final content redacted = %+v", event)
	}
}

func TestAppendRunEventNormalizesTodoMetadata(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Run", Mode: ThreadModeWork})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	metadata := map[string]any{
		"todo_items": []any{
			map[string]any{"id": "todo-1", "title": "Find files", "status": "completed", "summary": "Safe summary"},
			map[string]any{"id": "todo-secret", "title": "Open /Users/xuean/private.md with sk-secret-token", "status": "running", "summary": "curl https://example.test/private", "command": "bash /tmp/run.sh", "file_path": "/Users/xuean/private.md"},
			map[string]any{"id": "todo-bad-status", "title": "Fallback status", "status": "launched"},
		},
		"updated_by":        "provider",
		"browser_url":       "https://example.test/private",
		"redaction_applied": false,
	}
	event, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventWorkTodoUpdated, Summary: "Todo updated", Metadata: metadata})
	if err != nil {
		t.Fatalf("AppendRunEvent(todo) error = %v", err)
	}
	items, ok := event.Metadata["todo_items"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("todo_items = %#v", event.Metadata["todo_items"])
	}
	first := items[0].(map[string]any)
	if first["title"] != "Find files" || first["status"] != "completed" || first["summary"] != "Safe summary" {
		t.Fatalf("first = %+v", first)
	}
	second := items[1].(map[string]any)
	if second["id"] != "[redacted]" || second["title"] != "[redacted]" || second["summary"] != "[redacted]" || second["status"] != "running" || second["redaction_applied"] != true {
		t.Fatalf("second = %+v", second)
	}
	if _, ok := second["command"]; ok {
		t.Fatalf("unsafe command preserved: %+v", second)
	}
	third := items[2].(map[string]any)
	if third["status"] != "pending" {
		t.Fatalf("third = %+v", third)
	}
	if event.Metadata["updated_by"] != "provider" || event.Metadata["redaction_applied"] != true {
		t.Fatalf("metadata = %+v", event.Metadata)
	}
	if _, ok := event.Metadata["browser_url"]; ok {
		t.Fatalf("unsafe root metadata preserved: %+v", event.Metadata)
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

func TestRunStepLedgerProjectsDurableRunAndToolState(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Step ledger", Mode: ThreadModeChat})
	if err != nil {
		t.Fatalf("CreateThread() error = %v", err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: "msg_1", ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatalf("StartRun() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "model_request_started", Summary: "Model request started", Metadata: map[string]any{"model_phase": "initial", "api_key": "sk-secret"}}); err != nil {
		t.Fatalf("AppendRunEvent(model) error = %v", err)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ArgumentsSummary: map[string]any{"timezone": "UTC"}, ArgumentsHash: "hash_1", ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatalf("RecordToolCallRequest() error = %v", err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatalf("ApproveToolCall() error = %v", err)
	}
	if _, _, err := svc.StartToolCallExecution(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatalf("StartToolCallExecution() error = %v", err)
	}
	if _, _, err := svc.CompleteToolCallSuccess(context.Background(), ident, thread.ID, run.ID, "tc_1", map[string]any{"iso_time": "2026-05-28T00:00:00Z", "api_key": "sk-secret"}); err != nil {
		t.Fatalf("CompleteToolCallSuccess() error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: "model_request_started", Summary: "Model request started", Metadata: map[string]any{"model_phase": "continuation"}}); err != nil {
		t.Fatalf("AppendRunEvent(continuation) error = %v", err)
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunCompleted, Summary: "Run completed"}); err != nil {
		t.Fatalf("AppendRunEvent(final) error = %v", err)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatalf("ListRunEvents() error = %v", err)
	}
	var continuationEvent RunEvent
	for _, event := range events {
		if event.Type == "model_request_started" && event.Metadata["model_phase"] == "continuation" {
			continuationEvent = event
		}
	}
	if continuationEvent.Metadata["run_step_kind"] != string(RunStepKindContinuation) || continuationEvent.Metadata["run_step_status"] != string(RunStepStatusRunning) || continuationEvent.Metadata["run_step_summary"] != "Model request started" {
		t.Fatalf("continuation event step metadata = %+v", continuationEvent.Metadata)
	}

	ledger := BuildRunStepLedger(events)
	got := runStepKindsAndStatuses(ledger)
	want := []string{
		"model_request:running",
		"tool_requested:pending",
		"approval:required",
		"approval:approved",
		"tool_execution:running",
		"tool_execution:succeeded",
		"continuation:running",
		"terminal:completed",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ledger = %v, want %v", got, want)
	}
	for _, step := range ledger {
		if strings.Contains(step.Summary, "sk-secret") {
			t.Fatalf("step summary leaked secret: %+v", step)
		}
		if step.Kind == RunStepKindToolExecution && step.Status == RunStepStatusSucceeded && step.ToolCallID != "tc_1" {
			t.Fatalf("tool execution step = %+v", step)
		}
	}
}

func TestRunStepLedgerSeparatesPendingToolFromCompletedResult(t *testing.T) {
	events := []RunEvent{
		{Sequence: 1, Type: EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_done", "tool_name": ToolNameWorkspaceRead, "arguments_summary": map[string]any{"path": "notes.txt"}}},
		{Sequence: 2, Type: EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_done", "tool_name": ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "notes.txt", "content": "done"}}},
		{Sequence: 3, Type: EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_pending", "tool_name": ToolNameWorkspaceGrep, "arguments_summary": map[string]any{"query": "needle"}}},
	}

	state := RebuildRunStepState(events)

	if len(state.CompletedToolResults) != 1 || state.CompletedToolResults[0].ToolCallID != "tc_done" {
		t.Fatalf("completed = %+v", state.CompletedToolResults)
	}
	if len(state.PendingToolCalls) != 1 || state.PendingToolCalls[0].ToolCallID != "tc_pending" {
		t.Fatalf("pending = %+v", state.PendingToolCalls)
	}
	if state.NextAction != RunStepNextActionWaitForToolApproval {
		t.Fatalf("next action = %q", state.NextAction)
	}
}

func TestAdvanceRunStepStateMatchesFullRebuild(t *testing.T) {
	events := []RunEvent{
		{Sequence: 1, Type: "model_request_started", Summary: "Model request started", Metadata: map[string]any{"model_phase": "initial"}},
		{Sequence: 2, Type: EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameCurrentTime}},
		{Sequence: 3, Type: EventToolCallApproved, Summary: "Tool call approved", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameCurrentTime}},
		{Sequence: 4, Type: EventToolCallExecuting, Summary: "Tool call executing", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameCurrentTime}},
		{Sequence: 5, Type: EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameCurrentTime}},
		{Sequence: 6, Type: "model_request_started", Summary: "Model request started", Metadata: map[string]any{"model_phase": "continuation"}},
		{Sequence: 7, Type: EventRunCompleted, Summary: "Run completed"},
	}
	incremental := RunStepState{}
	for _, event := range events {
		incremental = AdvanceRunStepState(incremental, event)
	}
	rebuilt := RebuildRunStepState(events)

	if incremental.NextAction != rebuilt.NextAction || incremental.LastEventSequence != rebuilt.LastEventSequence || incremental.LastCompletedSequence != rebuilt.LastCompletedSequence || incremental.LastContinuationSequence != rebuilt.LastContinuationSequence {
		t.Fatalf("incremental = %+v, rebuilt = %+v", incremental, rebuilt)
	}
	if strings.Join(runStepKindsAndStatuses(incremental.Steps), ",") != strings.Join(runStepKindsAndStatuses(rebuilt.Steps), ",") {
		t.Fatalf("incremental steps = %+v, rebuilt steps = %+v", incremental.Steps, rebuilt.Steps)
	}
}

func TestRunStepStateAllowsResumeAfterContinuationStartedWithoutOutput(t *testing.T) {
	events := []RunEvent{
		{Sequence: 1, Type: "run_created", Summary: "Run created", Metadata: map[string]any{"message_id": "msg_1", "provider_id": "custom", "model": "model"}},
		{Sequence: 2, Type: EventToolCallRequested, Summary: "Tool call requested", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameWorkspaceRead}},
		{Sequence: 3, Type: EventToolCallApproved, Summary: "Tool call approved", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameWorkspaceRead}},
		{Sequence: 4, Type: EventToolCallExecuting, Summary: "Tool call executing", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameWorkspaceRead}},
		{Sequence: 5, Type: EventToolCallSucceeded, Summary: "Tool call succeeded", Metadata: map[string]any{"tool_call_id": "tc_1", "tool_name": ToolNameWorkspaceRead, "result_summary": map[string]any{"path": "notes.txt"}}},
		{Sequence: 6, Type: "model_request_started", Summary: "Model request started", Metadata: map[string]any{"model_phase": "continuation"}},
	}

	state := RebuildRunStepState(events)
	if state.NextAction != RunStepNextActionContinueModel {
		t.Fatalf("next action after start-only continuation = %q", state.NextAction)
	}

	events = append(events, RunEvent{Sequence: 7, Type: "model_output_delta", Summary: "Model output delta", Metadata: map[string]any{"model_phase": "continuation"}})
	state = RebuildRunStepState(events)
	if state.NextAction != RunStepNextActionNone {
		t.Fatalf("next action after continuation output = %q", state.NextAction)
	}
}

func TestRunStepStateCapturesMCPDiscoverySchemaHashes(t *testing.T) {
	state := RebuildRunStepState([]RunEvent{
		{Sequence: 1, Type: "mcp_discovery_succeeded", Summary: "MCP discovery succeeded", Metadata: map[string]any{
			"status":                  "succeeded",
			"server_slug":             "local-search",
			"candidate_names":         []any{"mcp.local-search.search"},
			"candidate_schema_hashes": map[string]any{"mcp.local-search.search": "sha256:search"},
		}},
		{Sequence: 2, Type: EventPipelineStepCompleted, Summary: "Pipeline step completed", Metadata: map[string]any{
			"enabled_tools": []any{"mcp.local-search.search"},
		}},
	})

	if state.MCPToolSchemaHashes["mcp.local-search.search"] != "sha256:search" {
		t.Fatalf("mcp hashes = %+v", state.MCPToolSchemaHashes)
	}
}

func TestBuildRunContextFromRunStepStateHydratesMCPTools(t *testing.T) {
	run := Run{ID: "run_1", ThreadID: "thread_1", UserID: "user_1", Source: RunSourceModelGateway}
	thread := Thread{ID: "thread_1", UserID: "user_1", Mode: ThreadModeChat}
	message := Message{ID: "msg_1", ThreadID: thread.ID, Role: MessageRoleUser, Content: "search"}
	job := BackgroundJob{ID: "job_1", RunID: run.ID, ThreadID: thread.ID, UserID: run.UserID, Metadata: map[string]any{"tool_call_id": "tc_mcp"}}
	state := RunStepState{
		TriggerMessageID:    message.ID,
		ProviderID:          "custom",
		Model:               "model-a",
		EnabledToolNames:    []string{"mcp.local-search.search"},
		MCPToolSchemaHashes: map[string]string{"mcp.local-search.search": "sha256:search"},
		CompletedToolResults: []RunStep{{
			ToolCallID: "tc_mcp",
			ToolName:   "mcp.local-search.search",
			SafeMetadata: map[string]any{
				"tool_call_id":   "tc_mcp",
				"tool_name":      "mcp.local-search.search",
				"result_summary": map[string]any{"items": []any{"one"}},
			},
		}},
	}

	context, err := buildRunContextFromState(run, thread, []Message{message}, job, state)
	if err != nil {
		t.Fatal(err)
	}

	if len(context.EnabledTools) != 1 || context.EnabledTools[0].Name != "mcp.local-search.search" || context.EnabledTools[0].InputSchemaHash != "sha256:search" {
		t.Fatalf("enabled tools = %+v", context.EnabledTools)
	}
	if !context.MCPAvailability.ExecutionEnabled || context.MCPAvailability.ServersSucceeded != 1 || len(context.MCPAvailability.CandidateNames) != 1 {
		t.Fatalf("mcp availability = %+v", context.MCPAvailability)
	}
}

func TestMemoryServiceMaintainsRunStepStateProjection(t *testing.T) {
	svc := NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, CreateThreadInput{Title: "Projection", Mode: ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	message, _, err := svc.CreateMessage(context.Background(), ident, thread.ID, CreateMessageInput{Content: "hello"})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, StartRunInput{Source: RunSourceModelGateway, MessageID: message.ID, ProviderID: "custom", Model: "model"})
	if err != nil {
		t.Fatal(err)
	}
	initial, err := svc.GetRunStepState(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if initial.LastEventSequence != 2 || initial.TriggerMessageID != message.ID || initial.ProviderID != "custom" || initial.Model != "model" {
		t.Fatalf("initial state = %+v", initial)
	}
	if _, _, err := svc.RecordToolCallRequest(context.Background(), ident, run.ID, RecordToolCallRequestInput{ToolCallID: "tc_1", ToolName: ToolNameCurrentTime, ApprovalStatus: ToolCallApprovalRequired, ExecutionStatus: ToolCallExecutionBlocked}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := svc.ApproveToolCall(context.Background(), ident, thread.ID, run.ID, "tc_1"); err != nil {
		t.Fatal(err)
	}

	state, err := svc.GetRunStepState(context.Background(), ident, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if state.NextAction != RunStepNextActionExecuteTool || len(state.PendingToolCalls) != 1 || state.PendingToolCalls[0].ToolCallID != "tc_1" {
		t.Fatalf("state = %+v", state)
	}
}

func runStepKindsAndStatuses(steps []RunStep) []string {
	result := make([]string, 0, len(steps))
	for _, step := range steps {
		result = append(result, string(step.Kind)+":"+string(step.Status))
	}
	return result
}

func ptr[T any](v T) *T { return &v }
