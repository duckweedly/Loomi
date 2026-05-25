package productdata

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
)

type Service interface {
	CurrentIdentity(context.Context, identity.LocalIdentity) (User, error)
	CreateThread(context.Context, identity.LocalIdentity, CreateThreadInput) (Thread, error)
	ListThreads(context.Context, identity.LocalIdentity, bool) ([]Thread, error)
	GetThread(context.Context, identity.LocalIdentity, string) (Thread, error)
	UpdateThread(context.Context, identity.LocalIdentity, string, UpdateThreadInput) (Thread, error)
	ArchiveThread(context.Context, identity.LocalIdentity, string) (Thread, error)
	CreateMessage(context.Context, identity.LocalIdentity, string, CreateMessageInput) (Message, bool, error)
	AppendAssistantMessage(context.Context, identity.LocalIdentity, string, AppendAssistantMessageInput) (Message, error)
	ListMessages(context.Context, identity.LocalIdentity, string) ([]Message, error)
	StartRun(context.Context, identity.LocalIdentity, string, StartRunInput) (Run, error)
	GetRun(context.Context, identity.LocalIdentity, string) (Run, error)
	GetCurrentRun(context.Context, identity.LocalIdentity, string) (Run, error)
	ListRunEvents(context.Context, identity.LocalIdentity, string, int) ([]RunEvent, error)
	AppendRunEvent(context.Context, identity.LocalIdentity, string, AppendRunEventInput) (RunEvent, error)
	PrepareRunContext(context.Context, identity.LocalIdentity, BackgroundJob) (RunContext, error)
	StopRun(context.Context, identity.LocalIdentity, string) (StopRunOutput, error)
	GetToolCall(context.Context, identity.LocalIdentity, string, string, string) (ToolCall, error)
	RecordToolCallRequest(context.Context, identity.LocalIdentity, string, RecordToolCallRequestInput) (ToolCall, []RunEvent, error)
	ApproveToolCall(context.Context, identity.LocalIdentity, string, string, string) (ToolCall, []RunEvent, error)
	DenyToolCall(context.Context, identity.LocalIdentity, string, string, string) (ToolCall, []RunEvent, error)
	StartToolCallExecution(context.Context, identity.LocalIdentity, string, string, string) (ToolCall, []RunEvent, error)
	CompleteToolCallSuccess(context.Context, identity.LocalIdentity, string, string, string, map[string]any) (ToolCall, []RunEvent, error)
	FailToolCallExecution(context.Context, identity.LocalIdentity, string, string, string, string, string) (ToolCall, []RunEvent, error)
	ClaimBackgroundJob(context.Context, identity.LocalIdentity, ClaimBackgroundJobInput) (BackgroundJob, Run, bool, error)
	RenewBackgroundJobLease(context.Context, identity.LocalIdentity, RenewBackgroundJobLeaseInput) (BackgroundJob, bool, error)
	RecoverBackgroundJobs(context.Context, identity.LocalIdentity, RecoverBackgroundJobsInput) ([]BackgroundJobRecovery, error)
	CompleteBackgroundJob(context.Context, identity.LocalIdentity, CompleteBackgroundJobInput) (BackgroundJob, bool, error)
	FailBackgroundJob(context.Context, identity.LocalIdentity, FailBackgroundJobInput) (BackgroundJob, bool, error)
	WorkerQueueDiagnostics(context.Context, identity.LocalIdentity) (WorkerQueueDiagnostics, error)
}

type SeedService interface {
	Service
	UpsertSeedThread(context.Context, identity.LocalIdentity, SeedThreadInput) (Thread, error)
	UpsertSeedMessage(context.Context, identity.LocalIdentity, SeedMessageInput) (Message, error)
}

type Repository interface {
	SeedService
}

type MemoryService struct {
	mu             sync.Mutex
	now            func() time.Time
	users          map[string]User
	threads        map[string]Thread
	messages       map[string][]Message
	runs           map[string]Run
	runEvents      map[string][]RunEvent
	backgroundJobs map[string]BackgroundJob
	toolCalls      map[string]ToolCall
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		now:            time.Now,
		users:          map[string]User{},
		threads:        map[string]Thread{},
		messages:       map[string][]Message{},
		runs:           map[string]Run{},
		runEvents:      map[string][]RunEvent{},
		backgroundJobs: map[string]BackgroundJob{},
		toolCalls:      map[string]ToolCall{},
	}
}

func (s *MemoryService) CurrentIdentity(_ context.Context, ident identity.LocalIdentity) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.ensureUserLocked(ident), nil
}

func (s *MemoryService) CreateThread(_ context.Context, ident identity.LocalIdentity, input CreateThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	return s.upsertThreadLocked(NewThreadID(), user.ID, title, input.Mode), nil
}

func (s *MemoryService) UpsertSeedThread(_ context.Context, ident identity.LocalIdentity, input SeedThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	return s.upsertThreadLocked(input.ID, user.ID, title, input.Mode), nil
}

func (s *MemoryService) ListThreads(_ context.Context, ident identity.LocalIdentity, includeArchived bool) ([]Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	threads := make([]Thread, 0, len(s.threads))
	for _, thread := range s.threads {
		if thread.UserID != user.ID {
			continue
		}
		if !includeArchived && thread.LifecycleStatus == ThreadLifecycleArchived {
			continue
		}
		threads = append(threads, thread)
	}
	sort.SliceStable(threads, func(i, j int) bool { return threads[i].UpdatedAt.After(threads[j].UpdatedAt) })
	return threads, nil
}

func (s *MemoryService) GetThread(_ context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	return thread, nil
}

func (s *MemoryService) UpdateThread(_ context.Context, ident identity.LocalIdentity, threadID string, input UpdateThreadInput) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if input.Title != nil {
		title, err := NormalizeThreadTitle(*input.Title)
		if err != nil {
			return Thread{}, err
		}
		thread.Title = title
	}
	if input.Mode != nil {
		if err := ValidateThreadMode(*input.Mode); err != nil {
			return Thread{}, err
		}
		thread.Mode = *input.Mode
	}
	thread.UpdatedAt = s.now()
	s.threads[thread.ID] = thread
	return thread, nil
}

func (s *MemoryService) ArchiveThread(_ context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	now := s.now()
	thread.LifecycleStatus = ThreadLifecycleArchived
	thread.ArchivedAt = &now
	thread.UpdatedAt = now
	s.threads[thread.ID] = thread
	return thread, nil
}

func (s *MemoryService) CreateMessage(_ context.Context, ident identity.LocalIdentity, threadID string, input CreateMessageInput) (Message, bool, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, false, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	message, created, err := s.upsertMessageLocked(NewMessageID(), threadID, user.ID, content, clientMessageID)
	return message, created, err
}

func (s *MemoryService) AppendAssistantMessage(_ context.Context, ident identity.LocalIdentity, threadID string, input AppendAssistantMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	metadata := RedactEventMetadata(input.Metadata)
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	if runID, ok := metadata["run_id"].(string); ok && runID != "" {
		for _, message := range s.messages[threadID] {
			if message.Role == MessageRoleAssistant && message.Metadata["run_id"] == runID {
				return Message{}, NewError(CodeInvalidRequest, "Assistant message already exists for run.")
			}
		}
	}
	message, err := s.appendMessageLocked(NewMessageID(), threadID, user.ID, MessageRoleAssistant, content, metadata, nil)
	if err != nil {
		return Message{}, err
	}
	return message, nil
}

func (s *MemoryService) UpsertSeedMessage(_ context.Context, ident identity.LocalIdentity, input SeedMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	message, _, err := s.upsertMessageLocked(input.ID, input.ThreadID, user.ID, content, clientMessageID)
	return message, err
}

func (s *MemoryService) ListMessages(_ context.Context, ident identity.LocalIdentity, threadID string) ([]Message, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID {
		return nil, NewError(CodeThreadNotFound, "Thread not found.")
	}
	messages := append([]Message(nil), s.messages[threadID]...)
	sort.SliceStable(messages, func(i, j int) bool { return messages[i].CreatedAt.Before(messages[j].CreatedAt) })
	return messages, nil
}

func (s *MemoryService) StartRun(_ context.Context, ident identity.LocalIdentity, threadID string, input StartRunInput) (Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != user.ID || thread.LifecycleStatus != ThreadLifecycleActive {
		return Run{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	for _, run := range s.runs {
		if run.ThreadID == threadID && run.UserID == user.ID && IsRunActive(run.Status) {
			return Run{}, NewError(CodeActiveRunExists, "Thread already has an active run.")
		}
	}
	source, err := NormalizeRunSource(input.Source)
	if err != nil {
		return Run{}, err
	}
	now := s.now()
	run := Run{ID: NewRunID(), ThreadID: threadID, UserID: user.ID, Status: RunStatusQueued, Source: source, Title: TitleForRunSource(source), CreatedAt: now, UpdatedAt: now}
	s.runs[run.ID] = run
	jobID := NewBackgroundJobID()
	metadata := map[string]any{"source": string(source), "job_id": jobID}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = input.ProviderID
		metadata["model"] = input.Model
	}
	metadata = RedactEventMetadata(metadata)
	job := BackgroundJob{ID: jobID, RunID: run.ID, ThreadID: threadID, UserID: user.ID, Kind: BackgroundJobKindRunExecution, Status: BackgroundJobStatusQueued, Priority: 100, MaxAttempts: 3, ScheduledAt: now, Metadata: metadata, CreatedAt: now, UpdatedAt: now}
	s.backgroundJobs[job.ID] = job
	s.runEvents[run.ID] = append(s.runEvents[run.ID], RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: threadID, UserID: user.ID, Sequence: 1, Category: RunEventCategoryLifecycle, Type: "run_created", Summary: "Run created", Metadata: metadata, CreatedAt: now})
	s.runEvents[run.ID] = append(s.runEvents[run.ID], RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: threadID, UserID: user.ID, Sequence: 2, Category: RunEventCategoryLifecycle, Type: EventRunQueued, Summary: "Run queued", Metadata: RedactEventMetadata(map[string]any{"job_id": job.ID}), CreatedAt: now})
	thread.UpdatedAt = now
	s.threads[threadID] = thread
	return run, nil
}

func (s *MemoryService) GetRun(_ context.Context, ident identity.LocalIdentity, runID string) (Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return run, nil
}

func (s *MemoryService) GetCurrentRun(_ context.Context, ident identity.LocalIdentity, threadID string) (Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	var latest *Run
	for _, run := range s.runs {
		if run.ThreadID != threadID || run.UserID != user.ID {
			continue
		}
		candidate := run
		if latest == nil || candidate.UpdatedAt.After(latest.UpdatedAt) {
			latest = &candidate
		}
	}
	if latest == nil {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return *latest, nil
}

func (s *MemoryService) ListRunEvents(_ context.Context, ident identity.LocalIdentity, runID string, afterSequence int) ([]RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID {
		return nil, NewError(CodeRunNotFound, "Run not found.")
	}
	events := make([]RunEvent, 0, len(s.runEvents[runID]))
	for _, event := range s.runEvents[runID] {
		if event.Sequence > afterSequence {
			events = append(events, event)
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].Sequence < events[j].Sequence })
	return events, nil
}

func (s *MemoryService) AppendRunEvent(_ context.Context, ident identity.LocalIdentity, runID string, input AppendRunEventInput) (RunEvent, error) {
	input, err := NormalizeRunEventInput(input)
	if err != nil {
		return RunEvent{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID {
		return RunEvent{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if IsRunTerminal(run.Status) {
		return RunEvent{}, NewError(CodeInvalidRequest, "Terminal run cannot accept new events.")
	}
	now := s.now()
	event := RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: run.ThreadID, UserID: user.ID, Sequence: len(s.runEvents[run.ID]) + 1, Category: input.Category, Type: input.Type, Summary: input.Summary, Content: input.Content, Metadata: input.Metadata, CreatedAt: now}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], event)
	run.UpdatedAt = now
	if input.Category == RunEventCategoryFinal {
		run.Status = statusFromFinalType(input.Type)
		run.CompletedAt = &now
		if input.ErrorCode != "" {
			run.ErrorCode = &input.ErrorCode
		}
		if input.ErrorMessage != "" {
			run.ErrorMessage = &input.ErrorMessage
		}
	}
	s.runs[run.ID] = run
	return event, nil
}

func (s *MemoryService) PrepareRunContext(ctx context.Context, ident identity.LocalIdentity, job BackgroundJob) (RunContext, error) {
	run, err := s.GetRun(ctx, ident, job.RunID)
	if err != nil {
		return RunContext{}, err
	}
	thread, err := s.GetThread(ctx, ident, run.ThreadID)
	if err != nil {
		return RunContext{}, err
	}
	if job.ID == "" || job.RunID != run.ID || job.ThreadID != thread.ID || job.UserID != run.UserID {
		return RunContext{}, NewError(CodeInvalidRequest, "Run context job boundary is invalid.")
	}
	messages, err := s.ListMessages(ctx, ident, thread.ID)
	if err != nil {
		return RunContext{}, err
	}
	events, err := s.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		return RunContext{}, err
	}
	return buildRunContext(run, thread, messages, job, events)
}

func buildRunContext(run Run, thread Thread, messages []Message, job BackgroundJob, events []RunEvent) (RunContext, error) {
	if run.Source == RunSourceModelGateway && len(messages) == 0 {
		return RunContext{}, NewError(CodeInvalidRequest, "Run context message history is required.")
	}
	metadata := job.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	created := runCreatedMetadata(events)
	providerID := firstMetadataString(metadata, created, "provider_id")
	model := firstMetadataString(metadata, created, "model")
	messageID := firstMetadataString(metadata, created, "message_id")
	toolCallID := metadataStringValue(metadata, "tool_call_id")
	if run.Source == RunSourceModelGateway {
		if providerID == "" || messageID == "" {
			return RunContext{}, NewError(CodeInvalidRequest, "Run context provider route is required.")
		}
		if !containsMessage(messages, messageID) {
			return RunContext{}, NewError(CodeInvalidRequest, "Run context message history is incomplete.")
		}
	}
	context := RunContext{
		Run:      run,
		Thread:   thread,
		Messages: append([]Message(nil), messages...),
		Job:      job,
		ProviderRoute: ProviderRoute{
			ProviderID: providerID,
			Model:      model,
			Available:  providerID != "",
		},
		EnabledTools: []ToolResolution{{
			Name:           ToolNameCurrentTime,
			ApprovalPolicy: "always_required",
			ExecutionState: "allowlisted",
		}},
	}
	if toolCallID != "" {
		context.ContinuationProjection = ContinuationProjection{ToolCallID: toolCallID, Available: hasToolResult(events, toolCallID)}
	}
	return context, nil
}

func runCreatedMetadata(events []RunEvent) map[string]any {
	for _, event := range events {
		if event.Type == "run_created" {
			return event.Metadata
		}
	}
	return map[string]any{}
}

func firstMetadataString(primary map[string]any, fallback map[string]any, key string) string {
	if value := metadataStringValue(primary, key); value != "" {
		return value
	}
	return metadataStringValue(fallback, key)
}

func metadataStringValue(metadata map[string]any, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func containsMessage(messages []Message, messageID string) bool {
	for _, message := range messages {
		if message.ID == messageID {
			return true
		}
	}
	return false
}

func hasToolResult(events []RunEvent, toolCallID string) bool {
	for _, event := range events {
		if event.Type != EventToolCallSucceeded {
			continue
		}
		if toolCallID == "" || metadataStringValue(event.Metadata, "tool_call_id") == toolCallID {
			return true
		}
	}
	return false
}

func (s *MemoryService) GetToolCall(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID || run.ThreadID != threadID {
		return ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	call, ok := s.toolCalls[run.ID+":"+strings.TrimSpace(toolCallID)]
	if !ok {
		return ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return call, nil
}

func (s *MemoryService) RecordToolCallRequest(_ context.Context, ident identity.LocalIdentity, runID string, input RecordToolCallRequestInput) (ToolCall, []RunEvent, error) {
	input, err := ValidateToolCallRequestInput(input)
	if err != nil {
		return ToolCall{}, nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID {
		return ToolCall{}, nil, NewError(CodeRunNotFound, "Run not found.")
	}
	if IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Terminal runs cannot request tools.")
	}
	for _, existing := range s.toolCalls {
		if existing.RunID == run.ID && existing.ToolCallID != input.ToolCallID {
			return ToolCall{}, nil, NewError(CodeInvalidRequest, "Only one tool call is supported per run.")
		}
	}
	key := run.ID + ":" + input.ToolCallID
	if existing, ok := s.toolCalls[key]; ok {
		return existing, nil, nil
	}
	now := s.now()
	arguments := RedactEventMetadata(input.ArgumentsSummary)
	call := ToolCall{ID: NewToolCallID(), ThreadID: run.ThreadID, RunID: run.ID, ToolCallID: input.ToolCallID, ToolName: input.ToolName, ArgumentsSummary: arguments, ApprovalStatus: input.ApprovalStatus, ExecutionStatus: input.ExecutionStatus, RequestedAt: now, UpdatedAt: now}
	s.toolCalls[key] = call
	run.Status = RunStatusBlockedOnToolApproval
	run.UpdatedAt = now
	s.runs[run.ID] = run
	for id, job := range s.backgroundJobs {
		if job.RunID == run.ID && job.UserID == user.ID && (job.Status == BackgroundJobStatusQueued || job.Status == BackgroundJobStatusRetrying) {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
		}
	}
	metadata := map[string]any{"tool_call_id": call.ToolCallID, "tool_name": call.ToolName, "arguments_summary": call.ArgumentsSummary, "approval_status": string(call.ApprovalStatus), "execution_status": string(call.ExecutionStatus)}
	requested := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallRequested, "Tool call requested", nil, metadata, now)
	required := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallApprovalRequired, "Tool approval required", nil, metadata, now)
	return call, []RunEvent{requested, required}, nil
}

func (s *MemoryService) ApproveToolCall(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, call, key, err := s.scopedToolCallLocked(user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalApproved {
		if call.ExecutionStatus == ToolCallExecutionNotStarted || call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed {
			return call, nil, nil
		}
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	if call.ApprovalStatus != ToolCallApprovalRequired || call.ExecutionStatus != ToolCallExecutionBlocked || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	now := s.now()
	call.ApprovalStatus = ToolCallApprovalApproved
	call.ExecutionStatus = ToolCallExecutionNotStarted
	call.UpdatedAt = now
	s.toolCalls[key] = call
	run.Status = RunStatusQueued
	run.UpdatedAt = now
	s.runs[run.ID] = run
	for id, job := range s.backgroundJobs {
		if job.RunID == run.ID && job.UserID == user.ID && !IsBackgroundJobTerminal(job.Status) {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
		}
	}
	jobID := NewBackgroundJobID()
	metadata := RedactEventMetadata(map[string]any{"source": string(run.Source), "job_id": jobID, "tool_call_id": call.ToolCallID, "resume_reason": "tool_call_approved"})
	s.backgroundJobs[jobID] = BackgroundJob{ID: jobID, RunID: run.ID, ThreadID: run.ThreadID, UserID: user.ID, Kind: BackgroundJobKindRunExecution, Status: BackgroundJobStatusQueued, Priority: 50, MaxAttempts: 3, ScheduledAt: now, Metadata: metadata, CreatedAt: now, UpdatedAt: now}
	event := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallApproved, "Tool call approved", nil, toolCallEventMetadata(call), now)
	return call, []RunEvent{event}, nil
}

func (s *MemoryService) DenyToolCall(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, call, key, err := s.scopedToolCallLocked(user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalDenied {
		return call, nil, nil
	}
	if call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed || call.ExecutionStatus == ToolCallExecutionCancelled {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be denied.")
	}
	if IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be denied.")
	}
	now := s.now()
	call.ApprovalStatus = ToolCallApprovalDenied
	call.ExecutionStatus = ToolCallExecutionCancelled
	call.UpdatedAt = now
	s.toolCalls[key] = call
	run.Status = RunStatusStopped
	run.CompletedAt = &now
	run.UpdatedAt = now
	s.runs[run.ID] = run
	for id, job := range s.backgroundJobs {
		if job.RunID == run.ID && job.UserID == user.ID && !IsBackgroundJobTerminal(job.Status) {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
		}
	}
	denied := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallDenied, "Tool call denied by user", nil, toolCallEventMetadata(call), now)
	final := s.appendRunEventLocked(run, RunEventCategoryFinal, EventRunStopped, "Run stopped after tool denial", nil, map[string]any{"tool_call_id": call.ToolCallID, "reason": "tool_call_denied"}, now)
	return call, []RunEvent{denied, final}, nil
}

func (s *MemoryService) StartToolCallExecution(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, call, key, err := s.scopedToolCallLocked(user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionExecuting {
		return call, nil, nil
	}
	if call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed || call.ExecutionStatus == ToolCallExecutionCancelled {
		return call, nil, nil
	}
	if call.ApprovalStatus != ToolCallApprovalApproved || call.ExecutionStatus != ToolCallExecutionNotStarted || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot execute.")
	}
	now := s.now()
	call.ExecutionStatus = ToolCallExecutionExecuting
	call.UpdatedAt = now
	s.toolCalls[key] = call
	event := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallExecuting, "Tool call executing", nil, toolCallEventMetadata(call), now)
	return call, []RunEvent{event}, nil
}

func (s *MemoryService) CompleteToolCallSuccess(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, resultSummary map[string]any) (ToolCall, []RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, call, key, err := s.scopedToolCallLocked(user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionSucceeded {
		return call, nil, nil
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot succeed.")
	}
	now := s.now()
	call.ExecutionStatus = ToolCallExecutionSucceeded
	call.ResultSummary = RedactEventMetadata(resultSummary)
	call.UpdatedAt = now
	s.toolCalls[key] = call
	run.Status = RunStatusRunning
	run.CompletedAt = nil
	run.UpdatedAt = now
	s.runs[run.ID] = run
	succeeded := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallSucceeded, "Tool call succeeded", nil, toolCallEventMetadata(call), now)
	return call, []RunEvent{succeeded}, nil
}

func (s *MemoryService) FailToolCallExecution(_ context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, errorCode string, errorMessage string) (ToolCall, []RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, call, key, err := s.scopedToolCallLocked(user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionFailed {
		return call, nil, nil
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot fail.")
	}
	now := s.now()
	code := strings.TrimSpace(errorCode)
	if code == "" {
		code = "tool_execution_failed"
	}
	message := RedactEventText(strings.TrimSpace(errorMessage))
	if message == "" {
		message = "Tool execution failed."
	}
	call.ExecutionStatus = ToolCallExecutionFailed
	call.ErrorCode = &code
	call.ErrorMessage = &message
	call.UpdatedAt = now
	s.toolCalls[key] = call
	run.Status = RunStatusFailed
	run.CompletedAt = &now
	run.ErrorCode = &code
	run.ErrorMessage = &message
	run.UpdatedAt = now
	s.runs[run.ID] = run
	failed := s.appendRunEventLocked(run, RunEventCategoryError, EventToolCallFailed, message, nil, toolCallEventMetadata(call), now)
	final := s.appendRunEventLocked(run, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"tool_call_id": call.ToolCallID, "error_code": code}, now)
	return call, []RunEvent{failed, final}, nil
}

func (s *MemoryService) StopRun(_ context.Context, ident identity.LocalIdentity, runID string) (StopRunOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	run, ok := s.runs[runID]
	if !ok || run.UserID != user.ID {
		return StopRunOutput{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if IsRunTerminal(run.Status) {
		return StopRunOutput{Run: run, Result: StopRunResultAlreadyTerminal}, nil
	}
	now := s.now()
	run.StopRequestedAt = &now
	run.Status = RunStatusStopped
	run.UpdatedAt = now
	run.CompletedAt = &now
	s.runs[run.ID] = run
	for id, job := range s.backgroundJobs {
		if job.RunID == run.ID && job.UserID == user.ID && !IsBackgroundJobTerminal(job.Status) {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
		}
	}
	lifecycle := s.appendRunEventLocked(run, RunEventCategoryProgress, EventStopRequested, "Stop requested", nil, map[string]any{}, now)
	final := s.appendRunEventLocked(run, RunEventCategoryFinal, EventRunStopped, "Run stopped", nil, map[string]any{}, now)
	return StopRunOutput{Run: run, Result: StopRunResultStopped, Events: []RunEvent{lifecycle, final}}, nil
}

func (s *MemoryService) scopedToolCallLocked(userID string, threadID string, runID string, toolCallID string) (Run, ToolCall, string, error) {
	run, ok := s.runs[runID]
	if !ok || run.UserID != userID || run.ThreadID != threadID {
		return Run{}, ToolCall{}, "", NewError(CodeRunNotFound, "Run not found.")
	}
	key := run.ID + ":" + strings.TrimSpace(toolCallID)
	call, ok := s.toolCalls[key]
	if !ok {
		return Run{}, ToolCall{}, "", NewError(CodeRunNotFound, "Run not found.")
	}
	return run, call, key, nil
}

func toolCallEventMetadata(call ToolCall) map[string]any {
	metadata := map[string]any{"tool_call_id": call.ToolCallID, "tool_name": call.ToolName, "arguments_summary": call.ArgumentsSummary, "approval_status": string(call.ApprovalStatus), "execution_status": string(call.ExecutionStatus)}
	if call.ResultSummary != nil {
		metadata["result_summary"] = call.ResultSummary
	}
	if call.ErrorCode != nil {
		metadata["error_code"] = *call.ErrorCode
	}
	if call.ErrorMessage != nil {
		metadata["error_message"] = *call.ErrorMessage
	}
	return RedactEventMetadata(metadata)
}

func (s *MemoryService) ClaimBackgroundJob(_ context.Context, ident identity.LocalIdentity, input ClaimBackgroundJobInput) (BackgroundJob, Run, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	workerID := strings.TrimSpace(input.WorkerID)
	if workerID == "" {
		return BackgroundJob{}, Run{}, false, NewError(CodeInvalidRequest, "Worker id is required.")
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	now := s.now()
	for id, job := range s.backgroundJobs {
		if job.UserID != user.ID || job.Status != BackgroundJobStatusQueued || job.ScheduledAt.After(now) {
			continue
		}
		run, ok := s.runs[job.RunID]
		if !ok || run.UserID != user.ID || IsRunTerminal(run.Status) {
			continue
		}
		if run.StopRequestedAt != nil {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
			continue
		}
		leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
		job.Status = BackgroundJobStatusLeased
		job.LeasedBy = &workerID
		job.LeaseExpiresAt = &leaseExpiresAt
		job.AttemptCount++
		job.OwnershipVersion++
		job.UpdatedAt = now
		s.backgroundJobs[id] = job
		run.Status = RunStatusRunning
		run.UpdatedAt = now
		s.runs[run.ID] = run
		s.appendRunEventLocked(run, RunEventCategoryProgress, EventJobClaimed, "Job claimed", nil, map[string]any{"job_id": job.ID, "worker_id": workerID, "attempt": job.AttemptCount}, now)
		return job, run, true, nil
	}
	return BackgroundJob{}, Run{}, false, nil
}

func (s *MemoryService) RenewBackgroundJobLease(_ context.Context, ident identity.LocalIdentity, input RenewBackgroundJobLeaseInput) (BackgroundJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	job, ok := s.backgroundJobs[input.JobID]
	if !ok || job.UserID != user.ID {
		return BackgroundJob{}, false, NewError(CodeInvalidRequest, "Background job not found.")
	}
	if !jobOwnedBy(job, input.WorkerID, input.OwnershipVersion) || IsBackgroundJobTerminal(job.Status) {
		return job, false, nil
	}
	run, ok := s.runs[job.RunID]
	if !ok || IsRunTerminal(run.Status) {
		return job, false, nil
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	now := s.now()
	leaseExpiresAt := now.Add(time.Duration(leaseSeconds) * time.Second)
	job.LeaseExpiresAt = &leaseExpiresAt
	job.UpdatedAt = now
	s.backgroundJobs[job.ID] = job
	s.appendRunEventLocked(run, RunEventCategoryProgress, EventLeaseRenewed, "Lease renewed", nil, map[string]any{"job_id": job.ID, "worker_id": input.WorkerID, "ownership_version": input.OwnershipVersion}, now)
	return job, true, nil
}

func (s *MemoryService) RecoverBackgroundJobs(_ context.Context, ident identity.LocalIdentity, input RecoverBackgroundJobsInput) ([]BackgroundJobRecovery, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	now := s.now()
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	code := strings.TrimSpace(input.ErrorCode)
	if code == "" {
		code = "worker_lease_expired"
	}
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	if message == "" {
		message = "Worker lease expired."
	}
	recoveries := []BackgroundJobRecovery{}
	for id, job := range s.backgroundJobs {
		if len(recoveries) >= limit || job.UserID != user.ID || IsBackgroundJobTerminal(job.Status) || job.LeaseExpiresAt == nil || job.LeaseExpiresAt.After(now) {
			continue
		}
		run, ok := s.runs[job.RunID]
		if !ok || IsRunTerminal(run.Status) {
			continue
		}
		previousWorkerID := ""
		if job.LeasedBy != nil {
			previousWorkerID = *job.LeasedBy
		}
		job.LastErrorCode = &code
		job.LastError = &message
		job.UpdatedAt = now
		run.UpdatedAt = now
		if job.AttemptCount >= job.MaxAttempts {
			job.Status = BackgroundJobStatusDead
			job.LeasedBy = nil
			job.LeaseExpiresAt = nil
			run.Status = RunStatusFailed
			run.CompletedAt = &now
			run.ErrorCode = &code
			run.ErrorMessage = &message
			event := s.appendRunEventLocked(run, RunEventCategoryError, EventJobRetryExhausted, message, nil, map[string]any{"job_id": job.ID, "attempt_count": job.AttemptCount, "error_code": code}, now)
			final := s.appendRunEventLocked(run, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"job_id": job.ID, "error_code": code}, now)
			s.runs[run.ID] = run
			s.backgroundJobs[id] = job
			recoveries = append(recoveries, BackgroundJobRecovery{Job: job, Run: run, Events: []RunEvent{event, final}, Exhausted: true})
			continue
		}
		job.Status = BackgroundJobStatusQueued
		job.LeasedBy = nil
		job.LeaseExpiresAt = nil
		job.ScheduledAt = now.Add(retryBackoffDuration(job.AttemptCount))
		run.Status = RunStatusRecovering
		recovering := s.appendRunEventLocked(run, RunEventCategoryProgress, EventJobRecovering, "Job recovering", nil, map[string]any{"job_id": job.ID, "previous_worker_id": previousWorkerID, "attempt": job.AttemptCount}, now)
		retry := s.appendRunEventLocked(run, RunEventCategoryProgress, EventJobRetryScheduled, "Job retry scheduled", nil, map[string]any{"job_id": job.ID, "next_attempt": job.AttemptCount + 1, "scheduled_at": job.ScheduledAt}, now)
		s.runs[run.ID] = run
		s.backgroundJobs[id] = job
		recoveries = append(recoveries, BackgroundJobRecovery{Job: job, Run: run, Events: []RunEvent{recovering, retry}})
	}
	return recoveries, nil
}

func retryBackoffDuration(attemptCount int) time.Duration {
	if attemptCount <= 1 {
		return time.Second
	}
	seconds := 1 << min(attemptCount-1, 3)
	return time.Duration(seconds) * time.Second
}

func (s *MemoryService) CompleteBackgroundJob(_ context.Context, ident identity.LocalIdentity, input CompleteBackgroundJobInput) (BackgroundJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	job, ok := s.backgroundJobs[input.JobID]
	if !ok || job.UserID != user.ID {
		return BackgroundJob{}, false, NewError(CodeInvalidRequest, "Background job not found.")
	}
	if IsBackgroundJobTerminal(job.Status) {
		return job, false, nil
	}
	if !jobOwnedBy(job, input.WorkerID, input.OwnershipVersion) {
		return job, false, nil
	}
	now := s.now()
	job.Status = BackgroundJobStatusCompleted
	job.UpdatedAt = now
	s.backgroundJobs[job.ID] = job
	return job, true, nil
}

func (s *MemoryService) FailBackgroundJob(_ context.Context, ident identity.LocalIdentity, input FailBackgroundJobInput) (BackgroundJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	job, ok := s.backgroundJobs[input.JobID]
	if !ok || job.UserID != user.ID {
		return BackgroundJob{}, false, NewError(CodeInvalidRequest, "Background job not found.")
	}
	if IsBackgroundJobTerminal(job.Status) {
		return job, false, nil
	}
	if !jobOwnedBy(job, input.WorkerID, input.OwnershipVersion) {
		return job, false, nil
	}
	now := s.now()
	code := strings.TrimSpace(input.ErrorCode)
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	job.Status = BackgroundJobStatusFailed
	job.LastErrorCode = &code
	job.LastError = &message
	job.UpdatedAt = now
	s.backgroundJobs[job.ID] = job
	if run, ok := s.runs[job.RunID]; ok && !IsRunTerminal(run.Status) {
		run.Status = RunStatusFailed
		run.CompletedAt = &now
		run.UpdatedAt = now
		run.ErrorCode = &code
		run.ErrorMessage = &message
		s.runs[run.ID] = run
		s.appendRunEventLocked(run, RunEventCategoryError, EventJobAttemptFailed, message, nil, map[string]any{"job_id": job.ID, "attempt": job.AttemptCount, "error_code": code}, now)
		s.appendRunEventLocked(run, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"job_id": job.ID, "error_code": code}, now)
	}
	return job, true, nil
}

func (s *MemoryService) WorkerQueueDiagnostics(_ context.Context, ident identity.LocalIdentity) (WorkerQueueDiagnostics, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	now := s.now()
	diagnostics := WorkerQueueDiagnostics{QueueStatus: WorkerQueueStatusReady, WorkerStatus: WorkerStatusReady, UpdatedAt: now}
	for _, call := range s.toolCalls {
		run, ok := s.runs[call.RunID]
		if !ok || run.UserID != user.ID {
			continue
		}
		if call.ApprovalStatus == ToolCallApprovalRequired && call.ExecutionStatus == ToolCallExecutionBlocked {
			diagnostics.BlockedToolApprovalCount++
		}
		if call.ApprovalStatus == ToolCallApprovalApproved && call.ExecutionStatus == ToolCallExecutionNotStarted {
			diagnostics.ResumableToolCallCount++
		}
	}
	for _, job := range s.backgroundJobs {
		if job.UserID != user.ID {
			continue
		}
		switch job.Status {
		case BackgroundJobStatusQueued:
			diagnostics.QueuedCount++
		case BackgroundJobStatusLeased:
			diagnostics.LeasedCount++
			if job.LeaseExpiresAt != nil && job.LeaseExpiresAt.Before(now) {
				diagnostics.StaleCount++
			}
		case BackgroundJobStatusRetrying:
			diagnostics.RetryingCount++
		case BackgroundJobStatusDead:
			diagnostics.DeadCount++
		}
	}
	if diagnostics.StaleCount > 0 || diagnostics.RetryingCount > 0 || diagnostics.DeadCount > 0 {
		diagnostics.QueueStatus = WorkerQueueStatusDegraded
		diagnostics.WorkerStatus = WorkerStatusDegraded
	}
	return diagnostics, nil
}

func (s *MemoryService) appendRunEventLocked(run Run, category RunEventCategory, eventType string, summary string, content *string, metadata map[string]any, createdAt time.Time) RunEvent {
	event := RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: run.ThreadID, UserID: run.UserID, Sequence: len(s.runEvents[run.ID]) + 1, Category: category, Type: eventType, Summary: RedactEventText(summary), Content: content, Metadata: RedactEventMetadata(metadata), CreatedAt: createdAt}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], event)
	return event
}

func jobOwnedBy(job BackgroundJob, workerID string, ownershipVersion int) bool {
	if job.LeasedBy == nil || *job.LeasedBy != strings.TrimSpace(workerID) {
		return false
	}
	return job.OwnershipVersion == ownershipVersion
}

func (s *MemoryService) ensureUserLocked(ident identity.LocalIdentity) User {
	if user, ok := s.users[ident.UserID]; ok {
		return user
	}
	now := s.now()
	user := User{ID: ident.UserID, DisplayName: ident.DisplayName, CreatedAt: now, UpdatedAt: now}
	s.users[user.ID] = user
	return user
}

func (s *MemoryService) upsertThreadLocked(id string, userID string, title string, mode ThreadMode) Thread {
	now := s.now()
	thread, ok := s.threads[id]
	if !ok {
		thread = Thread{ID: id, UserID: userID, CreatedAt: now}
	}
	thread.Title = title
	thread.Mode = mode
	thread.LifecycleStatus = ThreadLifecycleActive
	thread.ArchivedAt = nil
	thread.UpdatedAt = now
	s.threads[thread.ID] = thread
	return thread
}

func statusFromFinalType(eventType string) RunStatus {
	switch eventType {
	case "run_failed":
		return RunStatusFailed
	case "run_stopped":
		return RunStatusStopped
	default:
		return RunStatusCompleted
	}
}

func (s *MemoryService) upsertMessageLocked(id string, threadID string, userID string, content string, clientMessageID *string) (Message, bool, error) {
	for _, message := range s.messages[threadID] {
		if message.UserID == userID && message.ClientMessageID != nil && clientMessageID != nil && *message.ClientMessageID == *clientMessageID {
			return message, false, nil
		}
	}
	message, err := s.appendMessageLocked(id, threadID, userID, MessageRoleUser, content, map[string]any{}, clientMessageID)
	if err != nil {
		return Message{}, false, err
	}
	return message, true, nil
}

func (s *MemoryService) appendMessageLocked(id string, threadID string, userID string, role MessageRole, content string, metadata map[string]any, clientMessageID *string) (Message, error) {
	thread, ok := s.threads[threadID]
	if !ok || thread.UserID != userID {
		return Message{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if err := ValidateMessageRole(role); err != nil {
		return Message{}, err
	}
	now := s.now()
	message := Message{ID: id, ThreadID: threadID, UserID: userID, Role: role, Content: content, Metadata: RedactEventMetadata(metadata), ClientMessageID: clientMessageID, CreatedAt: now}
	s.messages[threadID] = append(s.messages[threadID], message)
	thread.UpdatedAt = now
	s.threads[threadID] = thread
	return message, nil
}
