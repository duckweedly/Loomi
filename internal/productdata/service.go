package productdata

import (
	"context"
	"sort"
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
	StopRun(context.Context, identity.LocalIdentity, string) (StopRunOutput, error)
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
	mu        sync.Mutex
	now       func() time.Time
	users     map[string]User
	threads   map[string]Thread
	messages  map[string][]Message
	runs      map[string]Run
	runEvents map[string][]RunEvent
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		now:       time.Now,
		users:     map[string]User{},
		threads:   map[string]Thread{},
		messages:  map[string][]Message{},
		runs:      map[string]Run{},
		runEvents: map[string][]RunEvent{},
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
	run := Run{ID: NewRunID(), ThreadID: threadID, UserID: user.ID, Status: RunStatusRunning, Source: source, Title: TitleForRunSource(source), CreatedAt: now, UpdatedAt: now}
	s.runs[run.ID] = run
	metadata := map[string]any{"source": string(source)}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = input.ProviderID
		metadata["model"] = input.Model
	}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: threadID, UserID: user.ID, Sequence: 1, Category: RunEventCategoryLifecycle, Type: "run_created", Summary: "Run created", Metadata: RedactEventMetadata(metadata), CreatedAt: now})
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
	run.Status = RunStatusStopped
	run.UpdatedAt = now
	run.CompletedAt = &now
	s.runs[run.ID] = run
	lifecycle := RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: run.ThreadID, UserID: user.ID, Sequence: len(s.runEvents[run.ID]) + 1, Category: RunEventCategoryLifecycle, Type: "run_stopped", Summary: "Run stopped", Metadata: map[string]any{}, CreatedAt: now}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], lifecycle)
	final := RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: run.ThreadID, UserID: user.ID, Sequence: len(s.runEvents[run.ID]) + 1, Category: RunEventCategoryFinal, Type: "run_stopped", Summary: "Run stopped", Metadata: map[string]any{}, CreatedAt: now}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], final)
	return StopRunOutput{Run: run, Result: StopRunResultStopped, Events: []RunEvent{lifecycle, final}}, nil
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
