package productdata

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

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
	ListToolCatalog(context.Context, identity.LocalIdentity) ([]ToolCatalogEntry, error)
	ListMCPDiscoveryEvents(context.Context, identity.LocalIdentity) ([]RunEvent, error)
	SyncBuiltInPersonas(context.Context, identity.LocalIdentity, []BuiltInPersonaConfig) (PersonaSyncResult, error)
	ListPersonas(context.Context, identity.LocalIdentity) ([]Persona, error)
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
	CreateMemoryEntry(context.Context, identity.LocalIdentity, CreateMemoryEntryInput) (MemoryEntry, error)
	ListMemoryEntries(context.Context, identity.LocalIdentity, MemorySearchInput) (MemorySearchOutput, error)
	SearchMemory(context.Context, identity.LocalIdentity, MemorySearchInput) (MemorySearchOutput, error)
	GetMemoryEntry(context.Context, identity.LocalIdentity, string, MemoryEntryAccessInput) (MemoryEntry, error)
	DeleteMemoryEntry(context.Context, identity.LocalIdentity, string, DeleteMemoryEntryInput) (MemoryTombstone, error)
	ListMemoryAudit(context.Context, identity.LocalIdentity, MemoryAuditInput) (MemoryAuditOutput, error)
	ListMemoryWriteProposals(context.Context, identity.LocalIdentity, MemoryWriteProposalListInput) (MemoryWriteProposalListOutput, error)
	ProposeMemoryWrite(context.Context, identity.LocalIdentity, ProposeMemoryWriteInput) (MemoryWriteProposal, error)
	UpdateMemoryWriteProposal(context.Context, identity.LocalIdentity, string, MemoryWriteProposalUpdateInput) (MemoryWriteProposal, error)
	ApproveMemoryWrite(context.Context, identity.LocalIdentity, string, MemoryWriteDecisionInput) (MemoryWriteDecision, error)
	DenyMemoryWrite(context.Context, identity.LocalIdentity, string, MemoryWriteDecisionInput) (MemoryWriteDecision, error)
	SaveMemoryProviderConfig(context.Context, identity.LocalIdentity, MemoryProviderConfig) (MemoryProviderConfig, error)
	GetMemoryProviderConfig(context.Context, identity.LocalIdentity) (MemoryProviderConfig, error)
	GetMemoryProviderStatus(context.Context, identity.LocalIdentity) (MemoryProviderStatus, error)
	ListMemoryProviderErrors(context.Context, identity.LocalIdentity, int) ([]MemoryProviderErrorEvent, error)
}

type ArtifactService interface {
	CreateArtifact(context.Context, identity.LocalIdentity, CreateArtifactInput) (Artifact, error)
	ReadArtifact(context.Context, identity.LocalIdentity, ReadArtifactInput) (Artifact, error)
	ListArtifacts(context.Context, identity.LocalIdentity, ListArtifactsInput) ([]Artifact, error)
}

type AgentTaskService interface {
	SpawnAgentTask(context.Context, identity.LocalIdentity, SpawnAgentTaskInput) (AgentTask, error)
	ListAgentTasks(context.Context, identity.LocalIdentity, ListAgentTasksInput) ([]AgentTask, error)
	CompleteAgentTask(context.Context, identity.LocalIdentity, CompleteAgentTaskInput) (AgentTask, error)
}

type MemorySnapshotService interface {
	GetMemoryOverviewSnapshot(context.Context, identity.LocalIdentity) (MemoryOverviewSnapshot, error)
	RebuildMemoryOverviewSnapshot(context.Context, identity.LocalIdentity) (MemoryOverviewSnapshot, error)
	GetMemoryImpressionSnapshot(context.Context, identity.LocalIdentity) (MemoryImpressionSnapshot, error)
	RebuildMemoryImpressionSnapshot(context.Context, identity.LocalIdentity) (MemoryImpressionSnapshot, error)
}

type ModelProviderConfigStore interface {
	SaveModelProviderConfig(context.Context, identity.LocalIdentity, ModelProviderConfig) (ModelProviderConfig, error)
	ListModelProviderConfigs(context.Context, identity.LocalIdentity) ([]ModelProviderConfig, error)
	SaveWebSearchConfig(context.Context, identity.LocalIdentity, WebSearchConfig) (WebSearchConfig, error)
	GetWebSearchConfig(context.Context, identity.LocalIdentity) (WebSearchConfig, error)
	SaveMemoryProviderConfig(context.Context, identity.LocalIdentity, MemoryProviderConfig) (MemoryProviderConfig, error)
	GetMemoryProviderConfig(context.Context, identity.LocalIdentity) (MemoryProviderConfig, error)
	GetMemoryProviderStatus(context.Context, identity.LocalIdentity) (MemoryProviderStatus, error)
	ListMemoryProviderErrors(context.Context, identity.LocalIdentity, int) ([]MemoryProviderErrorEvent, error)
}

type WorkspaceRootConfigStore interface {
	SaveWorkspaceRootConfig(context.Context, identity.LocalIdentity, WorkspaceRootConfig) (WorkspaceRootConfig, error)
	GetWorkspaceRootConfig(context.Context, identity.LocalIdentity) (WorkspaceRootConfig, error)
}

type MCPServerConfigStore interface {
	SaveMCPServerConfig(context.Context, identity.LocalIdentity, MCPServerConfigRecord) (MCPServerConfigRecord, error)
	ListMCPServerConfigs(context.Context, identity.LocalIdentity) ([]MCPServerConfigRecord, error)
	DeleteMCPServerConfig(context.Context, identity.LocalIdentity, string) error
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
	mu                 sync.Mutex
	now                func() time.Time
	users              map[string]User
	threads            map[string]Thread
	messages           map[string][]Message
	runs               map[string]Run
	runEvents          map[string][]RunEvent
	memoryAuditEvents  []RunEvent
	backgroundJobs     map[string]BackgroundJob
	toolCalls          map[string]ToolCall
	personas           map[string]Persona
	personaVersions    map[string]PersonaVersion
	personaSnapshots   map[string]PersonaSnapshot
	memoryEntries      map[string]MemoryEntry
	memoryProposals    map[string]MemoryWriteProposal
	memoryProposalKeys map[string]string
	memoryDecisionKeys map[string]MemoryWriteDecision
	artifacts          map[string]Artifact
	agentTasks         map[string]AgentTask
	modelProviders     map[string]ModelProviderConfig
	webSearchConfigs   map[string]WebSearchConfig
	workspaceRoots     map[string]WorkspaceRootConfig
	memoryProviders    map[string]MemoryProviderConfig
	mcpServerConfigs   map[string]MCPServerConfigRecord
}

func NewMemoryService() *MemoryService {
	return &MemoryService{
		now:                time.Now,
		users:              map[string]User{},
		threads:            map[string]Thread{},
		messages:           map[string][]Message{},
		runs:               map[string]Run{},
		runEvents:          map[string][]RunEvent{},
		memoryAuditEvents:  []RunEvent{},
		backgroundJobs:     map[string]BackgroundJob{},
		toolCalls:          map[string]ToolCall{},
		personas:           map[string]Persona{},
		personaVersions:    map[string]PersonaVersion{},
		personaSnapshots:   map[string]PersonaSnapshot{},
		memoryEntries:      map[string]MemoryEntry{},
		memoryProposals:    map[string]MemoryWriteProposal{},
		memoryProposalKeys: map[string]string{},
		memoryDecisionKeys: map[string]MemoryWriteDecision{},
		artifacts:          map[string]Artifact{},
		agentTasks:         map[string]AgentTask{},
		modelProviders:     map[string]ModelProviderConfig{},
		webSearchConfigs:   map[string]WebSearchConfig{},
		workspaceRoots:     map[string]WorkspaceRootConfig{},
		memoryProviders:    map[string]MemoryProviderConfig{},
		mcpServerConfigs:   map[string]MCPServerConfigRecord{},
	}
}

func (s *MemoryService) SaveModelProviderConfig(_ context.Context, ident identity.LocalIdentity, input ModelProviderConfig) (ModelProviderConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	provider := normalizeModelProviderConfig(input)
	if provider.ID == "" || provider.Model == "" || provider.APIKey == "" {
		return ModelProviderConfig{}, NewError(CodeProviderMisconfigured, "Provider configuration is incomplete.")
	}
	provider.UserID = user.ID
	s.modelProviders[provider.ID] = provider
	return provider, nil
}

func (s *MemoryService) ListModelProviderConfigs(_ context.Context, ident identity.LocalIdentity) ([]ModelProviderConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	providers := make([]ModelProviderConfig, 0, len(s.modelProviders))
	for _, provider := range s.modelProviders {
		if provider.UserID == user.ID {
			providers = append(providers, provider)
		}
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ID < providers[j].ID })
	return providers, nil
}

func (s *MemoryService) SaveWebSearchConfig(_ context.Context, ident identity.LocalIdentity, input WebSearchConfig) (WebSearchConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	existing := s.webSearchConfigs[user.ID]
	next := normalizeWebSearchConfig(input)
	if next.TavilyAPIKey == "" {
		next.TavilyAPIKey = existing.TavilyAPIKey
	}
	if next.BraveAPIKey == "" {
		next.BraveAPIKey = existing.BraveAPIKey
	}
	next.UserID = user.ID
	s.webSearchConfigs[user.ID] = next
	return next, nil
}

func (s *MemoryService) GetWebSearchConfig(_ context.Context, ident identity.LocalIdentity) (WebSearchConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	config := s.webSearchConfigs[user.ID]
	config.UserID = user.ID
	return config, nil
}

func (s *MemoryService) SaveWorkspaceRootConfig(_ context.Context, ident identity.LocalIdentity, input WorkspaceRootConfig) (WorkspaceRootConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	next := normalizeWorkspaceRootConfig(input)
	if next.Path == "" {
		return WorkspaceRootConfig{}, NewError(CodeInvalidRequest, "Workspace folder is required.")
	}
	next.UserID = user.ID
	s.workspaceRoots[user.ID] = next
	return next, nil
}

func (s *MemoryService) GetWorkspaceRootConfig(_ context.Context, ident identity.LocalIdentity) (WorkspaceRootConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	config := s.workspaceRoots[user.ID]
	config.UserID = user.ID
	return config, nil
}

func (s *MemoryService) SaveMemoryProviderConfig(_ context.Context, ident identity.LocalIdentity, input MemoryProviderConfig) (MemoryProviderConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	next := normalizeMemoryProviderConfig(input, s.now())
	next.UserID = user.ID
	s.memoryProviders[user.ID] = next
	return next, nil
}

func (s *MemoryService) GetMemoryProviderStatus(_ context.Context, ident identity.LocalIdentity) (MemoryProviderStatus, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	config := s.memoryProviders[user.ID]
	if config.UserID == "" {
		config = defaultMemoryProviderConfig(user.ID, s.now())
	}
	return memoryProviderStatus(config, s.now()), nil
}

func (s *MemoryService) GetMemoryProviderConfig(_ context.Context, ident identity.LocalIdentity) (MemoryProviderConfig, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	config := s.memoryProviders[user.ID]
	if config.UserID == "" {
		config = defaultMemoryProviderConfig(user.ID, s.now())
	}
	return config, nil
}

func (s *MemoryService) ListMemoryProviderErrors(_ context.Context, ident identity.LocalIdentity, limit int) ([]MemoryProviderErrorEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	config := s.memoryProviders[user.ID]
	if config.UserID == "" {
		config = defaultMemoryProviderConfig(user.ID, s.now())
	}
	status := memoryProviderStatus(config, s.now())
	items := []MemoryProviderErrorEvent{}
	if status.Diagnostic.Code != "" && status.Diagnostic.Code != "ok" {
		checkedAt := s.now()
		if status.CheckedAt != nil {
			checkedAt = *status.CheckedAt
		}
		items = append(items, MemoryProviderErrorEvent{Code: status.Diagnostic.Code, Message: status.Diagnostic.Message, Provider: status.Provider, State: status.State, CheckedAt: checkedAt})
	}
	for _, runEvents := range s.runEvents {
		for _, event := range runEvents {
			if event.UserID != user.ID || !isMemoryProviderRuntimeErrorEvent(event.Type) {
				continue
			}
			items = append(items, memoryProviderErrorFromRunEvent(event))
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].CheckedAt.After(items[j].CheckedAt)
	})
	return limitMemoryProviderErrors(items, limit), nil
}

func (s *MemoryService) SaveMCPServerConfig(_ context.Context, ident identity.LocalIdentity, input MCPServerConfigRecord) (MCPServerConfigRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	record := normalizeMCPServerConfigRecord(input)
	if record.Slug == "" {
		return MCPServerConfigRecord{}, NewError(CodeInvalidRequest, "MCP server slug is required.")
	}
	record.UserID = user.ID
	s.mcpServerConfigs[user.ID+":"+record.Slug] = record
	return record, nil
}

func (s *MemoryService) ListMCPServerConfigs(_ context.Context, ident identity.LocalIdentity) ([]MCPServerConfigRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	records := make([]MCPServerConfigRecord, 0, len(s.mcpServerConfigs))
	for _, record := range s.mcpServerConfigs {
		if record.UserID == user.ID {
			records = append(records, record)
		}
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Slug < records[j].Slug })
	return records, nil
}

func (s *MemoryService) DeleteMCPServerConfig(_ context.Context, ident identity.LocalIdentity, slug string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	delete(s.mcpServerConfigs, user.ID+":"+strings.TrimSpace(slug))
	return nil
}

func (s *MemoryService) SpawnAgentTask(_ context.Context, ident identity.LocalIdentity, input SpawnAgentTaskInput) (AgentTask, error) {
	role := strings.TrimSpace(input.Role)
	goal := strings.TrimSpace(input.Goal)
	if !isSupportedAgentRole(role) || goal == "" || len([]rune(goal)) > 4000 {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task role and goal are invalid.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[input.ThreadID]
	if !ok || thread.UserID != user.ID {
		return AgentTask{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	run, ok := s.runs[input.RunID]
	if !ok || run.UserID != user.ID || run.ThreadID != thread.ID {
		return AgentTask{}, NewError(CodeRunNotFound, "Run not found.")
	}
	now := s.now()
	task := AgentTask{
		ID:        NewAgentTaskID(),
		ThreadID:  thread.ID,
		RunID:     run.ID,
		Role:      role,
		Goal:      goal,
		Status:    AgentTaskStatusSpawned,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.agentTasks[task.ID] = task
	return task, nil
}

func (s *MemoryService) ListAgentTasks(_ context.Context, ident identity.LocalIdentity, input ListAgentTasksInput) ([]AgentTask, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[input.ThreadID]
	if !ok || thread.UserID != user.ID {
		return nil, NewError(CodeThreadNotFound, "Thread not found.")
	}
	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	result := make([]AgentTask, 0, limit)
	for _, task := range s.agentTasks {
		if task.ThreadID != thread.ID {
			continue
		}
		result = append(result, task)
		if len(result) >= limit {
			break
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result, nil
}

func (s *MemoryService) CompleteAgentTask(_ context.Context, ident identity.LocalIdentity, input CompleteAgentTaskInput) (AgentTask, error) {
	summary := strings.TrimSpace(input.ResultSummary)
	if summary == "" || len([]rune(summary)) > 4000 {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent result summary is invalid.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	thread, ok := s.threads[input.ThreadID]
	if !ok || thread.UserID != user.ID {
		return AgentTask{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	task, ok := s.agentTasks[strings.TrimSpace(input.TaskID)]
	if !ok || task.ThreadID != thread.ID {
		return AgentTask{}, NewError(CodeInvalidRequest, "Agent task not found.")
	}
	now := s.now()
	task.Status = AgentTaskStatusCompleted
	task.ResultSummary = RedactEventText(summary)
	task.UpdatedAt = now
	s.agentTasks[task.ID] = task
	return task, nil
}

func (s *MemoryService) CreateArtifact(_ context.Context, ident identity.LocalIdentity, input CreateArtifactInput) (Artifact, error) {
	title := strings.TrimSpace(input.Title)
	content := input.Content
	if title == "" || strings.TrimSpace(content) == "" {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact title and content are required.")
	}
	if !utf8.ValidString(content) {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact content must be valid UTF-8.")
	}
	limit := boundedArtifactBytes(input.MaxBytes)
	if len([]byte(content)) > limit {
		return Artifact{}, NewError(CodeInvalidRequest, "Artifact content is too large.")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureUserLocked(ident)
	if _, ok := s.threads[input.ThreadID]; !ok {
		return Artifact{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	if _, ok := s.runs[input.RunID]; !ok {
		return Artifact{}, NewError(CodeRunNotFound, "Run not found.")
	}
	now := s.now()
	artifact := Artifact{
		ID:           NewArtifactID(),
		ThreadID:     input.ThreadID,
		RunID:        input.RunID,
		Title:        title,
		ArtifactType: "text",
		Content:      content,
		ContentBytes: len([]byte(content)),
		TextExcerpt:  artifactExcerpt(content, limit),
		Truncated:    false,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.artifacts[artifact.ID] = artifact
	return artifact, nil
}

func (s *MemoryService) ReadArtifact(_ context.Context, ident identity.LocalIdentity, input ReadArtifactInput) (Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureUserLocked(ident)
	artifact, ok := s.artifacts[strings.TrimSpace(input.ArtifactID)]
	if !ok || artifact.ThreadID != input.ThreadID {
		return Artifact{}, NewError(CodeArtifactNotFound, "Artifact not found.")
	}
	limit := boundedArtifactBytes(input.MaxBytes)
	artifact.TextExcerpt = artifactExcerpt(artifact.Content, limit)
	artifact.Truncated = len([]byte(artifact.Content)) > limit
	return artifact, nil
}

func (s *MemoryService) ListArtifacts(_ context.Context, ident identity.LocalIdentity, input ListArtifactsInput) ([]Artifact, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureUserLocked(ident)
	limit := input.Limit
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	result := make([]Artifact, 0, limit)
	for _, artifact := range s.artifacts {
		if artifact.ThreadID != input.ThreadID {
			continue
		}
		artifact.Content = ""
		artifact.TextExcerpt = artifactExcerpt(artifact.TextExcerpt, 512)
		result = append(result, artifact)
		if len(result) >= limit {
			break
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].CreatedAt.Before(result[j].CreatedAt) })
	return result, nil
}

func boundedArtifactBytes(value int) int {
	if value <= 0 || value > 32*1024 {
		return 32 * 1024
	}
	return value
}

func artifactExcerpt(content string, limit int) string {
	if len([]byte(content)) <= limit {
		return content
	}
	if limit < 0 {
		limit = 0
	}
	for limit > 0 && !utf8.RuneStart(content[limit]) {
		limit--
	}
	return content[:limit]
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
	if input.PersonaID != "" {
		if _, _, err := s.activePersonaVersionLocked(input.PersonaID); err != nil {
			return Thread{}, err
		}
	}
	return s.upsertThreadLocked(NewThreadID(), user.ID, title, input.Mode, input.PersonaID), nil
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
	return s.upsertThreadLocked(input.ID, user.ID, title, input.Mode, ""), nil
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
	if input.PersonaID != nil {
		personaID := strings.TrimSpace(*input.PersonaID)
		if personaID != "" {
			if _, _, err := s.activePersonaVersionLocked(personaID); err != nil {
				return Thread{}, err
			}
		}
		thread.PersonaID = personaID
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
	snapshot, err := s.resolvePersonaSnapshotLocked(thread, input.PersonaID)
	if err != nil {
		return Run{}, err
	}
	if snapshot.ID != "" {
		run.PersonaID = snapshot.ID
		s.personaSnapshots[run.ID] = snapshot
	}
	s.runs[run.ID] = run
	jobID := NewBackgroundJobID()
	metadata := map[string]any{"source": string(source), "job_id": jobID}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = runProviderID(input.ProviderID, snapshot)
		metadata["model"] = runModel(input.ProviderID, input.Model, snapshot)
	}
	if snapshot.ID != "" {
		metadata["persona_id"] = snapshot.ID
		metadata["persona_version"] = snapshot.Version
		metadata["persona_name"] = snapshot.Name
		metadata["persona_resolved_from"] = string(snapshot.ResolvedFrom)
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
	if IsRunTerminal(run.Status) && !isTerminalRunEventAppendAllowed(input.Type) {
		return RunEvent{}, NewError(CodeInvalidRequest, "Terminal run cannot accept new events.")
	}
	now := s.now()
	event := RunEvent{ID: NewRunEventID(), RunID: run.ID, ThreadID: run.ThreadID, UserID: user.ID, Sequence: len(s.runEvents[run.ID]) + 1, Category: input.Category, Type: input.Type, Summary: input.Summary, Content: input.Content, Metadata: input.Metadata, CreatedAt: now}
	s.runEvents[run.ID] = append(s.runEvents[run.ID], event)
	if isMemoryAuditEvent(event.Type) {
		auditEvent := event
		auditEvent.Sequence = len(s.memoryAuditEvents) + 1
		s.memoryAuditEvents = append(s.memoryAuditEvents, auditEvent)
	}
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
	context, err := buildRunContext(run, thread, messages, job, events)
	if err != nil {
		return RunContext{}, err
	}
	s.mu.Lock()
	context.Persona = s.personaSnapshots[run.ID]
	s.mu.Unlock()
	applyPersonaToRunContext(&context, events)
	snapshot := s.buildMemorySnapshot(ctx, ident, run, thread)
	context.MemorySnapshot = snapshot
	context.NotebookSnapshot = s.buildNotebookSnapshot(ctx, ident, run, thread)
	status, err := s.GetMemoryProviderStatus(ctx, ident)
	if err == nil {
		context.MemoryReadiness = status
		context.EnabledTools = FilterMemoryToolResolutionsForProvider(context.EnabledTools, status)
	}
	_, _ = s.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventMemorySnapshotLoaded, Summary: "Memory snapshot loaded", Metadata: memorySnapshotEventMetadata(snapshot)})
	return context, nil
}

func (s *MemoryService) ListToolCatalog(_ context.Context, ident identity.LocalIdentity) ([]ToolCatalogEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	events := make([]RunEvent, 0)
	for _, runEvents := range s.runEvents {
		for _, event := range runEvents {
			if event.UserID == user.ID {
				events = append(events, event)
			}
		}
	}
	config := s.memoryProviders[user.ID]
	if config.UserID == "" {
		config = defaultMemoryProviderConfig(user.ID, s.now())
	}
	return ApplyMemoryToolAvailability(SafeToolCatalogFromEvents(events), memoryProviderStatus(config, s.now())), nil
}

func (s *MemoryService) ListMCPDiscoveryEvents(_ context.Context, ident identity.LocalIdentity) ([]RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	events := make([]RunEvent, 0)
	for _, runEvents := range s.runEvents {
		for _, event := range runEvents {
			if event.UserID == user.ID && (event.Type == "mcp_discovery_succeeded" || event.Type == "mcp_discovery_failed" || event.Type == "mcp_discovery_rejected") {
				events = append(events, event)
			}
		}
	}
	sort.SliceStable(events, func(i, j int) bool {
		if events[i].CreatedAt.Equal(events[j].CreatedAt) {
			return events[i].Sequence < events[j].Sequence
		}
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events, nil
}

func (s *MemoryService) CreateMemoryEntry(_ context.Context, ident identity.LocalIdentity, input CreateMemoryEntryInput) (MemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	entry, err := s.newMemoryEntryLocked(user.ID, input)
	if err != nil {
		return MemoryEntry{}, err
	}
	s.memoryEntries[entry.ID] = entry
	return entry, nil
}

func (s *MemoryService) ListMemoryEntries(ctx context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	return s.SearchMemory(ctx, ident, input)
}

func (s *MemoryService) SearchMemory(_ context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	limit := memoryLimit(input.Limit)
	queryTerms := strings.Fields(strings.ToLower(strings.TrimSpace(input.Query)))
	items := make([]MemorySearchResult, 0, limit)
	excluded := 0
	for _, entry := range s.memoryEntries {
		if !memoryEntryVisibleTo(entry, user.ID, input) {
			excluded++
			continue
		}
		if len(queryTerms) > 0 && !memoryEntryMatches(entry, queryTerms) {
			continue
		}
		items = append(items, memorySearchResult(entry))
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return MemorySearchOutput{Items: items, ExcludedCount: excluded}, nil
}

func (s *MemoryService) GetMemoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryOverviewSnapshot, error) {
	return s.memoryOverviewSnapshot(ctx, ident, false)
}

func (s *MemoryService) RebuildMemoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryOverviewSnapshot, error) {
	return s.memoryOverviewSnapshot(ctx, ident, true)
}

func (s *MemoryService) GetMemoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryImpressionSnapshot, error) {
	return s.memoryImpressionSnapshot(ctx, ident, false)
}

func (s *MemoryService) RebuildMemoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity) (MemoryImpressionSnapshot, error) {
	return s.memoryImpressionSnapshot(ctx, ident, true)
}

func (s *MemoryService) memoryOverviewSnapshot(ctx context.Context, ident identity.LocalIdentity, rebuilt bool) (MemoryOverviewSnapshot, error) {
	output, err := s.SearchMemory(ctx, ident, MemorySearchInput{Limit: 7, Purpose: "snapshot"})
	if err != nil {
		return MemoryOverviewSnapshot{}, err
	}
	return buildMemoryOverviewSnapshot(semanticMemorySnapshotItems(output.Items), s.now(), rebuilt), nil
}

func (s *MemoryService) memoryImpressionSnapshot(ctx context.Context, ident identity.LocalIdentity, rebuilt bool) (MemoryImpressionSnapshot, error) {
	output, err := s.SearchMemory(ctx, ident, MemorySearchInput{Limit: 7, Purpose: "impression"})
	if err != nil {
		return MemoryImpressionSnapshot{}, err
	}
	return buildMemoryImpressionSnapshot(semanticMemorySnapshotItems(output.Items), s.now(), rebuilt), nil
}

func (s *MemoryService) GetMemoryEntry(_ context.Context, ident identity.LocalIdentity, entryID string, input MemoryEntryAccessInput) (MemoryEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	entry, ok := s.memoryEntries[strings.TrimSpace(entryID)]
	if !ok || !memoryEntryReadableTo(entry, user.ID, input) {
		return MemoryEntry{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	entry.Content = ""
	return entry, nil
}

func (s *MemoryService) DeleteMemoryEntry(_ context.Context, ident identity.LocalIdentity, entryID string, input DeleteMemoryEntryInput) (MemoryTombstone, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	entry, ok := s.memoryEntries[strings.TrimSpace(entryID)]
	if !ok || !memoryEntryReadableTo(entry, user.ID, MemoryEntryAccessInput{ScopeType: input.ScopeType, ScopeID: input.ScopeID, SourceThreadID: input.SourceThreadID, SourceRunID: input.SourceRunID}) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if entry.Status == MemoryEntryTombstoned && entry.DeletedAt != nil {
		return MemoryTombstone{EntryID: entry.ID, Status: string(MemoryEntryTombstoned), DeletedAt: *entry.DeletedAt}, nil
	}
	now := s.now()
	entry.Status = MemoryEntryTombstoned
	entry.Content = ""
	entry.Summary = "[deleted]"
	entry.DeletedAt = &now
	entry.DeletedBy = user.ID
	entry.DeleteReason = RedactEventText(strings.TrimSpace(input.Reason))
	entry.UpdatedAt = now
	s.memoryEntries[entry.ID] = entry
	s.appendMemoryAuditEventLocked(user.ID, entry.SourceRunID, EventMemoryEntryDeleted, "Memory entry deleted", memoryEntryAuditMetadata(entry, ""))
	return MemoryTombstone{EntryID: entry.ID, Status: string(MemoryEntryTombstoned), DeletedAt: now}, nil
}

func (s *MemoryService) ListMemoryAudit(_ context.Context, ident identity.LocalIdentity, input MemoryAuditInput) (MemoryAuditOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	limit := memoryLimit(input.Limit)
	items := make([]MemoryAuditItem, 0, limit)
	for _, event := range s.memoryAuditEvents {
		if event.UserID != user.ID || !isMemoryAuditEvent(event.Type) {
			continue
		}
		if input.ThreadID != "" && event.ThreadID != strings.TrimSpace(input.ThreadID) {
			continue
		}
		if input.SourceRunID != "" && event.RunID != strings.TrimSpace(input.SourceRunID) {
			continue
		}
		if input.EventType != "" && memoryAuditEventType(event.Type) != strings.TrimSpace(input.EventType) {
			continue
		}
		items = append(items, memoryAuditItem(event))
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].OccurredAt.After(items[j].OccurredAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return MemoryAuditOutput{Items: items}, nil
}

func (s *MemoryService) ListMemoryWriteProposals(_ context.Context, ident identity.LocalIdentity, input MemoryWriteProposalListInput) (MemoryWriteProposalListOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	limit := memoryLimit(input.Limit)
	status := input.Status
	if status == "" {
		status = MemoryWritePending
	}
	items := make([]MemoryWriteProposal, 0, limit)
	for _, proposal := range s.memoryProposals {
		if proposal.UserID != user.ID || proposal.Status != status {
			continue
		}
		if input.ScopeType != "" && proposal.ScopeType != input.ScopeType {
			continue
		}
		if strings.TrimSpace(input.ScopeID) != "" && proposal.ScopeID != strings.TrimSpace(input.ScopeID) {
			continue
		}
		if strings.TrimSpace(input.SourceRunID) != "" && proposal.SourceRunID != strings.TrimSpace(input.SourceRunID) {
			continue
		}
		items = append(items, safeMemoryProposal(proposal))
	}
	sort.SliceStable(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	if len(items) > limit {
		items = items[:limit]
	}
	return MemoryWriteProposalListOutput{Items: items}, nil
}

func (s *MemoryService) ProposeMemoryWrite(_ context.Context, ident identity.LocalIdentity, input ProposeMemoryWriteInput) (MemoryWriteProposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	key := strings.TrimSpace(input.IdempotencyKey)
	if key != "" {
		if proposalID, ok := s.memoryProposalKeys[user.ID+":"+key]; ok {
			return s.memoryProposals[proposalID], nil
		}
	}
	scopeType, scopeID, err := normalizeMemoryScope(user.ID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	now := s.now()
	proposal := MemoryWriteProposal{ID: NewMemoryProposalID(), UserID: user.ID, ScopeType: scopeType, ScopeID: scopeID, Title: title, Summary: summary, Content: content, Status: MemoryWritePending, SafetyState: safety, SourceThreadID: strings.TrimSpace(input.SourceThreadID), SourceRunID: strings.TrimSpace(input.SourceRunID), SourceEventID: strings.TrimSpace(input.SourceEventID), IdempotencyKey: key, CreatedAt: now}
	if safety == MemorySafetyBlocked {
		proposal.Status = MemoryWriteDenied
	}
	s.memoryProposals[proposal.ID] = proposal
	if key != "" {
		s.memoryProposalKeys[user.ID+":"+key] = proposal.ID
	}
	s.appendMemoryAuditEventLocked(user.ID, proposal.SourceRunID, EventMemoryWriteProposed, "Memory write proposed", memoryProposalAuditMetadata(proposal, ""))
	return proposal, nil
}

func (s *MemoryService) UpdateMemoryWriteProposal(_ context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteProposalUpdateInput) (MemoryWriteProposal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	proposal, ok := s.memoryProposals[strings.TrimSpace(proposalID)]
	if !ok || proposal.UserID != user.ID {
		return MemoryWriteProposal{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if proposal.Status != MemoryWritePending || proposal.SafetyState == MemorySafetyBlocked {
		return MemoryWriteProposal{}, NewError(CodeInvalidRequest, "Only pending memory proposals can be edited.")
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Summary)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	if safety == MemorySafetyBlocked {
		return MemoryWriteProposal{}, NewError(CodeInvalidRequest, "Memory proposal edit contains sensitive content.")
	}
	proposal.Title = title
	proposal.Summary = summary
	proposal.Content = content
	proposal.SafetyState = safety
	s.memoryProposals[proposal.ID] = proposal
	return safeMemoryProposal(proposal), nil
}

func (s *MemoryService) ApproveMemoryWrite(_ context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	if decision, ok := s.memoryDecisionKeys[user.ID+":approve:"+strings.TrimSpace(input.IdempotencyKey)]; ok && input.IdempotencyKey != "" {
		return decision, nil
	}
	proposal, ok := s.memoryProposals[strings.TrimSpace(proposalID)]
	if !ok || proposal.UserID != user.ID {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if proposal.Status == MemoryWriteApproved && proposal.CreatedEntryID != "" {
		entry := s.memoryEntries[proposal.CreatedEntryID]
		return MemoryWriteDecision{Proposal: proposal, Entry: safeMemoryEntry(entry)}, nil
	}
	if proposal.Status != MemoryWritePending || proposal.SafetyState == MemorySafetyBlocked {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Memory write cannot be approved.")
	}
	entry, err := s.newMemoryEntryLocked(user.ID, CreateMemoryEntryInput{ScopeType: proposal.ScopeType, ScopeID: proposal.ScopeID, Title: proposal.Title, Content: proposal.Content, SourceThreadID: proposal.SourceThreadID, SourceRunID: proposal.SourceRunID, SourceEventID: proposal.SourceEventID})
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	now := s.now()
	proposal.Status = MemoryWriteApproved
	proposal.DecidedAt = &now
	proposal.DecidedBy = user.ID
	proposal.DecisionReason = RedactEventText(strings.TrimSpace(input.Reason))
	proposal.CreatedEntryID = entry.ID
	s.memoryEntries[entry.ID] = entry
	s.memoryProposals[proposal.ID] = proposal
	decision := MemoryWriteDecision{Proposal: proposal, Entry: safeMemoryEntry(entry)}
	if input.IdempotencyKey != "" {
		s.memoryDecisionKeys[user.ID+":approve:"+strings.TrimSpace(input.IdempotencyKey)] = decision
	}
	s.appendMemoryAuditEventLocked(user.ID, proposal.SourceRunID, EventMemoryWriteApproved, "Memory write approved", memoryProposalAuditMetadata(proposal, entry.ID))
	return decision, nil
}

func (s *MemoryService) DenyMemoryWrite(_ context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	user := s.ensureUserLocked(ident)
	if decision, ok := s.memoryDecisionKeys[user.ID+":deny:"+strings.TrimSpace(input.IdempotencyKey)]; ok && input.IdempotencyKey != "" {
		return decision, nil
	}
	proposal, ok := s.memoryProposals[strings.TrimSpace(proposalID)]
	if !ok || proposal.UserID != user.ID {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if proposal.Status == MemoryWriteDenied {
		return MemoryWriteDecision{Proposal: proposal}, nil
	}
	if proposal.Status == MemoryWriteApproved {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Approved memory write cannot be denied.")
	}
	now := s.now()
	proposal.Status = MemoryWriteDenied
	proposal.DecidedAt = &now
	proposal.DecidedBy = user.ID
	proposal.DecisionReason = RedactEventText(strings.TrimSpace(input.Reason))
	s.memoryProposals[proposal.ID] = proposal
	decision := MemoryWriteDecision{Proposal: proposal}
	if input.IdempotencyKey != "" {
		s.memoryDecisionKeys[user.ID+":deny:"+strings.TrimSpace(input.IdempotencyKey)] = decision
	}
	s.appendMemoryAuditEventLocked(user.ID, proposal.SourceRunID, EventMemoryWriteDenied, "Memory write denied", memoryProposalAuditMetadata(proposal, ""))
	return decision, nil
}

func (s *MemoryService) SyncBuiltInPersonas(_ context.Context, ident identity.LocalIdentity, configs []BuiltInPersonaConfig) (PersonaSyncResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureUserLocked(ident)
	if err := validateBuiltInPersonaConfigs(configs); err != nil {
		return PersonaSyncResult{}, err
	}
	now := s.now()
	result := PersonaSyncResult{Synced: len(configs)}
	for _, config := range configs {
		slug := strings.TrimSpace(config.Slug)
		var persona Persona
		var exists bool
		for _, candidate := range s.personas {
			if candidate.Slug == slug && candidate.Source == PersonaSourceBuiltIn {
				persona = candidate
				exists = true
				break
			}
		}
		if !exists {
			persona = Persona{ID: NewPersonaID(), Slug: slug, Source: PersonaSourceBuiltIn, CreatedAt: now}
			result.CreatedPersonas++
		}
		persona.Name = strings.TrimSpace(config.Name)
		persona.Description = strings.TrimSpace(config.Description)
		persona.IsDefault = config.IsDefault
		persona.IsActive = true
		persona.ActiveVersion = strings.TrimSpace(config.Version)
		persona.UpdatedAt = now
		if persona.IsDefault {
			result.DefaultPersonaSlug = persona.Slug
			for id, existing := range s.personas {
				if existing.ID != persona.ID && existing.Source == PersonaSourceBuiltIn {
					existing.IsDefault = false
					existing.UpdatedAt = now
					s.personas[id] = existing
				}
			}
		}
		s.personas[persona.ID] = persona
		key := persona.ID + ":" + persona.ActiveVersion
		if _, exists := s.personaVersions[key]; !exists {
			result.CreatedVersions++
			s.personaVersions[key] = PersonaVersion{
				PersonaID:        persona.ID,
				Version:          persona.ActiveVersion,
				SystemPrompt:     strings.TrimSpace(config.SystemPrompt),
				ModelRoute:       config.ModelRoute,
				AllowedToolNames: append([]string(nil), config.AllowedToolNames...),
				ReasoningMode:    strings.TrimSpace(config.ReasoningMode),
				BudgetSummary:    strings.TrimSpace(config.BudgetSummary),
				CreatedAt:        now,
			}
		}
		result.ActivatedVersions++
	}
	if result.DefaultPersonaSlug == "" {
		for _, persona := range s.personas {
			if persona.IsDefault && persona.IsActive {
				result.DefaultPersonaSlug = persona.Slug
				break
			}
		}
	}
	return result, nil
}

func (s *MemoryService) ListPersonas(_ context.Context, ident identity.LocalIdentity) ([]Persona, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureUserLocked(ident)
	personas := make([]Persona, 0, len(s.personas))
	for _, persona := range s.personas {
		if persona.IsActive {
			personas = append(personas, persona)
		}
	}
	sort.SliceStable(personas, func(i, j int) bool {
		if personas[i].IsDefault != personas[j].IsDefault {
			return personas[i].IsDefault
		}
		return personas[i].Name < personas[j].Name
	})
	return personas, nil
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
			Name:            ToolNameCurrentTime,
			ApprovalPolicy:  string(ToolApprovalAlwaysRequired),
			ExecutionState:  string(ToolExecutionStateExecutable),
			Source:          string(ToolCatalogSourceBuiltin),
			Group:           string(ToolCatalogGroupRuntime),
			InputSchemaHash: builtinCurrentTimeCatalogEntry().InputSchemaHash,
			RiskLevel:       string(ToolRiskLow),
		}},
	}
	if toolCallID != "" {
		context.ContinuationProjection = ContinuationProjection{ToolCallID: toolCallID, Available: hasToolResult(events, toolCallID)}
	}
	return context, nil
}

func (s *MemoryService) buildMemorySnapshot(ctx context.Context, ident identity.LocalIdentity, run Run, thread Thread) MemorySnapshot {
	output, err := s.SearchMemory(ctx, ident, MemorySearchInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Limit: 5, Purpose: "run_context"})
	if err != nil {
		return MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}
	}
	status := "loaded"
	if len(output.Items) == 0 {
		status = "empty"
	}
	return MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: output.Items, Limit: 5, TotalCandidates: len(output.Items), LoadStatus: status, RedactionApplied: true}
}

func (s *MemoryService) buildNotebookSnapshot(ctx context.Context, ident identity.LocalIdentity, run Run, thread Thread) MemorySnapshot {
	output, err := s.SearchMemory(ctx, ident, MemorySearchInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, SourceType: "notebook", Limit: 5, Purpose: "run_context_notebook"})
	if err != nil {
		return MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}
	}
	status := "loaded"
	if len(output.Items) == 0 {
		status = "empty"
	}
	return MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: output.Items, Limit: 5, TotalCandidates: len(output.Items), LoadStatus: status, RedactionApplied: true}
}

func defaultMemoryProviderConfig(userID string, now time.Time) MemoryProviderConfig {
	return MemoryProviderConfig{UserID: userID, Enabled: true, Provider: MemoryProviderLocal, UpdatedAt: now}
}

func memoryProviderStatus(config MemoryProviderConfig, now time.Time) MemoryProviderStatus {
	checkedAt := now
	openviking := safeOpenVikingMemoryConfig(config.OpenViking)
	nowledge := safeNowledgeMemoryConfig(config.Nowledge)
	status := MemoryProviderStatus{
		Enabled:        config.Enabled,
		Provider:       config.Provider,
		Label:          memoryProviderLabel(config.Provider),
		Configured:     true,
		CommitAfterRun: config.CommitAfterRun,
		CheckedAt:      &checkedAt,
		OpenViking:     openviking,
		Nowledge:       nowledge,
		Diagnostic:     MemoryProviderDiagnostic{Code: "ok", Message: "Ready."},
	}
	if status.Provider == "" {
		status.Provider = MemoryProviderLocal
		status.Label = memoryProviderLabel(status.Provider)
	}
	if !config.Enabled {
		status.State = MemoryProviderStateDisabled
		status.Configured = false
		status.Diagnostic = MemoryProviderDiagnostic{Code: "memory_disabled", Message: "Memory is disabled."}
		return status
	}
	switch config.Provider {
	case MemoryProviderLocal:
		status.State = MemoryProviderStateAvailable
	case MemoryProviderSemantic:
		status.Configured = strings.TrimSpace(config.SemanticEndpoint) != ""
		if !status.Configured {
			status.State = MemoryProviderStateUnconfigured
			status.Diagnostic = MemoryProviderDiagnostic{Code: "semantic_unconfigured", Message: "Semantic memory provider is not configured."}
			return status
		}
		status.State = MemoryProviderStateHealthy
	case MemoryProviderOpenViking:
		status.Configured = openviking.BaseURL != "" && openviking.RootAPIKeySet && openviking.EmbeddingModel != "" && openviking.VLMModel != ""
		if !status.Configured {
			status.State = MemoryProviderStateUnconfigured
			status.Diagnostic = MemoryProviderDiagnostic{Code: "openviking_unconfigured", Message: "OpenViking memory provider requires a base URL, root key, embedding model, and VLM model."}
			return status
		}
		status.State = MemoryProviderStateHealthy
	case MemoryProviderNowledge:
		status.Configured = nowledge.BaseURL != ""
		if !status.Configured {
			status.State = MemoryProviderStateUnconfigured
			status.Diagnostic = MemoryProviderDiagnostic{Code: "nowledge_unconfigured", Message: "Nowledge memory provider requires a base URL."}
			return status
		}
		status.State = MemoryProviderStateHealthy
	default:
		status.Provider = MemoryProviderLocal
		status.Label = memoryProviderLabel(status.Provider)
		status.State = MemoryProviderStateDegraded
		status.Configured = true
		status.Diagnostic = MemoryProviderDiagnostic{Code: "unknown_provider", Message: "Unknown memory provider was normalized to local memory."}
	}
	if config.Diagnostic != "" && config.Diagnostic != "[redacted]" {
		status.Diagnostic.Message = config.Diagnostic
	}
	return status
}

func memoryProviderLabel(provider MemoryProviderID) string {
	switch provider {
	case MemoryProviderSemantic:
		return "Semantic"
	case MemoryProviderOpenViking:
		return "OpenViking"
	case MemoryProviderNowledge:
		return "Nowledge"
	default:
		return "Local"
	}
}

func safeOpenVikingMemoryConfig(config OpenVikingMemoryConfig) OpenVikingMemoryConfig {
	config.RootAPIKeySet = config.RootAPIKeySet || config.RootAPIKey != ""
	config.EmbeddingAPIKeySet = config.EmbeddingAPIKeySet || config.EmbeddingAPIKey != ""
	config.VLMAPIKeySet = config.VLMAPIKeySet || config.VLMAPIKey != ""
	config.RerankAPIKeySet = config.RerankAPIKeySet || config.RerankAPIKey != ""
	config.RootAPIKey = ""
	config.EmbeddingAPIKey = ""
	config.VLMAPIKey = ""
	config.RerankAPIKey = ""
	return config
}

func safeNowledgeMemoryConfig(config NowledgeMemoryConfig) NowledgeMemoryConfig {
	config.APIKeySet = config.APIKeySet || config.APIKey != ""
	config.APIKey = ""
	return config
}

func (s *MemoryService) newMemoryEntryLocked(userID string, input CreateMemoryEntryInput) (MemoryEntry, error) {
	scopeType, scopeID, err := normalizeMemoryScope(userID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryEntry{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryEntry{}, err
	}
	status := MemoryEntryApproved
	if safety == MemorySafetyBlocked {
		status = MemoryEntryDisabled
	}
	now := s.now()
	return MemoryEntry{ID: NewMemoryEntryID(), UserID: userID, ScopeType: scopeType, ScopeID: scopeID, Title: title, Summary: summary, Content: content, Status: status, SafetyState: safety, SourceThreadID: strings.TrimSpace(input.SourceThreadID), SourceRunID: strings.TrimSpace(input.SourceRunID), SourceEventID: strings.TrimSpace(input.SourceEventID), ContentHash: memoryContentHash(scopeType, scopeID, content), CreatedAt: now, UpdatedAt: now}, nil
}

func normalizeMemoryScope(userID string, scopeType MemoryScopeType, scopeID string) (MemoryScopeType, string, error) {
	if scopeType == "" {
		scopeType = MemoryScopeUser
	}
	scopeID = strings.TrimSpace(scopeID)
	switch scopeType {
	case MemoryScopeUser:
		if scopeID == "" {
			scopeID = userID
		}
	case MemoryScopeThread:
		if scopeID == "" {
			return "", "", NewError(CodeInvalidRequest, "Thread memory scope id is required.")
		}
	default:
		return "", "", NewError(CodeInvalidRequest, "Memory scope is invalid.")
	}
	return scopeType, scopeID, nil
}

func normalizeMemoryContent(title string, content string) (string, string, string, MemorySafetyState, error) {
	title = strings.TrimSpace(RedactEventText(title))
	content = strings.TrimSpace(content)
	if title == "" {
		return "", "", "", "", NewError(CodeInvalidRequest, "Memory title is required.")
	}
	if content == "" {
		return "", "", "", "", NewError(CodeInvalidRequest, "Memory content is required.")
	}
	redacted := RedactEventText(content)
	safety := MemorySafetySafe
	if redacted != content {
		safety = MemorySafetyBlocked
	} else if len([]rune(content)) > 480 {
		redacted = string([]rune(content)[:480])
		safety = MemorySafetyRedacted
	}
	summary := redacted
	if len([]rune(summary)) > 160 {
		summary = string([]rune(summary)[:160])
		safety = MemorySafetyRedacted
	}
	return title, summary, redacted, safety, nil
}

func memoryContentHash(scopeType MemoryScopeType, scopeID string, content string) string {
	sum := sha256.Sum256([]byte(string(scopeType) + ":" + scopeID + ":" + strings.ToLower(strings.TrimSpace(content))))
	return hex.EncodeToString(sum[:])
}

func memoryLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 50 {
		return 50
	}
	return limit
}

func memoryEntryVisibleTo(entry MemoryEntry, userID string, input MemorySearchInput) bool {
	if entry.UserID != userID || entry.SafetyState == MemorySafetyBlocked {
		return false
	}
	if entry.Status != MemoryEntryApproved && !(input.IncludeTombstoned && entry.Status == MemoryEntryTombstoned) {
		return false
	}
	switch input.ScopeType {
	case MemoryScopeThread:
		if !((entry.ScopeType == MemoryScopeUser && entry.ScopeID == userID) || (entry.ScopeType == MemoryScopeThread && entry.ScopeID == strings.TrimSpace(input.ScopeID))) {
			return false
		}
	case MemoryScopeUser, "":
		if !(entry.ScopeType == MemoryScopeUser && entry.ScopeID == userID) {
			return false
		}
	default:
		return false
	}
	if input.SourceRunID != "" && entry.SourceRunID != strings.TrimSpace(input.SourceRunID) {
		return false
	}
	if input.SourceThreadID != "" && entry.SourceThreadID != strings.TrimSpace(input.SourceThreadID) {
		return false
	}
	switch strings.TrimSpace(input.SourceType) {
	case "", "any":
		return true
	case "run":
		return entry.SourceRunID != ""
	case "thread":
		return entry.SourceThreadID != ""
	case "notebook":
		return entry.SourceEventID == "notebook"
	case "manual":
		return entry.SourceRunID == "" && entry.SourceThreadID == "" && entry.SourceEventID != "notebook"
	default:
		return false
	}
}

func memoryEntryReadableTo(entry MemoryEntry, userID string, input MemoryEntryAccessInput) bool {
	if entry.UserID != userID || entry.SafetyState == MemorySafetyBlocked || (entry.Status != MemoryEntryApproved && entry.Status != MemoryEntryTombstoned) {
		return false
	}
	if entry.ScopeType == MemoryScopeUser && entry.ScopeID == userID {
		return true
	}
	if entry.ScopeType != MemoryScopeThread {
		return false
	}
	scopeType := input.ScopeType
	scopeID := strings.TrimSpace(input.ScopeID)
	sourceThreadID := strings.TrimSpace(input.SourceThreadID)
	sourceRunID := strings.TrimSpace(input.SourceRunID)
	return (scopeType == MemoryScopeThread && scopeID != "" && scopeID == entry.ScopeID) ||
		(sourceThreadID != "" && sourceThreadID == entry.SourceThreadID) ||
		(sourceRunID != "" && sourceRunID == entry.SourceRunID)
}

func memoryEntryMatches(entry MemoryEntry, terms []string) bool {
	haystack := strings.ToLower(entry.Title + " " + entry.Summary + " " + entry.Content)
	for _, term := range terms {
		if strings.Contains(haystack, term) {
			return true
		}
	}
	return false
}

func memorySearchResult(entry MemoryEntry) MemorySearchResult {
	return MemorySearchResult{ID: entry.ID, Title: entry.Title, Summary: entry.Summary, ScopeType: entry.ScopeType, ScopeID: entry.ScopeID, Status: string(entry.Status), SafetyState: string(entry.SafetyState), SourceThreadID: entry.SourceThreadID, SourceRunID: entry.SourceRunID, SourceEventID: entry.SourceEventID, SourceType: memorySourceType(entry.SourceThreadID, entry.SourceRunID, entry.SourceEventID), CreatedAt: entry.CreatedAt, UpdatedAt: entry.UpdatedAt, DeletedAt: entry.DeletedAt, RankReason: "text_match", RedactionApplied: entry.SafetyState != MemorySafetySafe || entry.Content != entry.Summary}
}

func safeMemoryEntry(entry MemoryEntry) MemoryEntry {
	entry.Content = ""
	return entry
}

func safeMemoryProposal(proposal MemoryWriteProposal) MemoryWriteProposal {
	proposal.Content = ""
	proposal.IdempotencyKey = ""
	proposal.DecidedBy = ""
	return proposal
}

func memorySnapshotEventMetadata(snapshot MemorySnapshot) map[string]any {
	return map[string]any{"status": snapshot.LoadStatus, "entry_count": len(snapshot.Entries), "limit": snapshot.Limit, "redaction_applied": snapshot.RedactionApplied}
}

func isMemoryAuditEvent(eventType string) bool {
	switch eventType {
	case EventMemorySnapshotLoaded, EventMemoryWriteProposed, EventMemoryWriteApproved, EventMemoryWriteDenied, EventMemoryEntryDeleted:
		return true
	default:
		return false
	}
}

func isTerminalRunEventAppendAllowed(eventType string) bool {
	switch eventType {
	case "memory_provider_commit_completed", "memory_provider_commit_failed":
		return true
	default:
		return false
	}
}

func isMemoryProviderRuntimeErrorEvent(eventType string) bool {
	switch eventType {
	case EventMemoryExternalSnapshotFailed, "memory_provider_commit_failed":
		return true
	default:
		return false
	}
}

func memoryProviderErrorFromRunEvent(event RunEvent) MemoryProviderErrorEvent {
	code := firstEventString(event.Metadata, "error_code")
	if code == "" {
		code = event.Type
	}
	message := event.Summary
	return MemoryProviderErrorEvent{
		Code:      RedactEventText(code),
		Message:   RedactEventText(message),
		Provider:  MemoryProviderID(firstEventString(event.Metadata, "provider")),
		State:     MemoryProviderStateUnhealthy,
		CheckedAt: event.CreatedAt,
		RunID:     event.RunID,
		EventType: event.Type,
	}
}

func firstEventString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	return strings.TrimSpace(RedactEventText(eventValueString(value)))
}

func eventValueString(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case []byte:
		return strings.TrimSpace(string(typed))
	default:
		return ""
	}
}

func limitMemoryProviderErrors(items []MemoryProviderErrorEvent, limit int) []MemoryProviderErrorEvent {
	limit = limitMemoryProviderErrorQueryLimit(limit)
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func limitMemoryProviderErrorQueryLimit(limit int) int {
	if limit <= 0 {
		limit = 10
	}
	if limit > 20 {
		limit = 20
	}
	return limit
}

func memoryAuditEventType(eventType string) string {
	if eventType == EventMemoryEntryDeleted {
		return "memory_deleted"
	}
	return eventType
}

func memoryAuditItem(event RunEvent) MemoryAuditItem {
	return MemoryAuditItem{
		ID:               event.ID,
		EventType:        memoryAuditEventType(event.Type),
		Summary:          RedactEventText(event.Summary),
		ThreadID:         event.ThreadID,
		RunID:            event.RunID,
		MemoryEntryID:    metadataStringValue(event.Metadata, "memory_entry_id"),
		MemoryProposalID: metadataStringValue(event.Metadata, "memory_proposal_id"),
		Status:           firstNonEmpty(metadataStringValue(event.Metadata, "memory_status"), metadataStringValue(event.Metadata, "status")),
		ScopeType:        metadataStringValue(event.Metadata, "memory_scope_type"),
		SourceType:       "run",
		RedactionApplied: true,
		OccurredAt:       event.CreatedAt,
	}
}

func memorySourceType(sourceThreadID string, sourceRunID string, sourceEventID string) string {
	if strings.TrimSpace(sourceEventID) == "notebook" {
		return "notebook"
	}
	if strings.TrimSpace(sourceRunID) != "" {
		return "run"
	}
	if strings.TrimSpace(sourceThreadID) != "" {
		return "thread"
	}
	return "manual"
}

func memoryProposalAuditMetadata(proposal MemoryWriteProposal, entryID string) map[string]any {
	metadata := map[string]any{
		"memory_proposal_id": proposal.ID,
		"memory_status":      proposal.Status,
		"memory_scope_type":  proposal.ScopeType,
		"memory_safety":      proposal.SafetyState,
	}
	if entryID != "" {
		metadata["memory_entry_id"] = entryID
	}
	if proposal.SourceEventID != "" {
		metadata["source_event_id"] = proposal.SourceEventID
	}
	return metadata
}

func memoryEntryAuditMetadata(entry MemoryEntry, action string) map[string]any {
	metadata := map[string]any{
		"memory_entry_id":   entry.ID,
		"memory_status":     entry.Status,
		"memory_scope_type": entry.ScopeType,
		"memory_safety":     entry.SafetyState,
	}
	if action != "" {
		metadata["memory_action"] = action
	}
	return metadata
}

func (s *MemoryService) appendMemoryAuditEventLocked(userID string, runID string, eventType string, summary string, metadata map[string]any) {
	var run Run
	if strings.TrimSpace(runID) != "" {
		run = s.runs[strings.TrimSpace(runID)]
	}
	createdAt := s.now()
	event := RunEvent{ID: NewRunEventID(), RunID: strings.TrimSpace(runID), ThreadID: run.ThreadID, UserID: firstNonEmpty(run.UserID, strings.TrimSpace(userID)), Sequence: len(s.memoryAuditEvents) + 1, Category: RunEventCategoryProgress, Type: eventType, Summary: RedactEventText(summary), Metadata: RedactEventMetadata(metadata), CreatedAt: createdAt}
	if event.UserID != "" {
		s.memoryAuditEvents = append(s.memoryAuditEvents, event)
	}
	if run.ID != "" && !IsRunTerminal(run.Status) {
		s.appendRunEventLocked(run, RunEventCategoryProgress, eventType, summary, nil, metadata, createdAt)
	}
}

func applyPersonaToRunContext(context *RunContext, events []RunEvent) {
	if context == nil {
		return
	}
	if context.Persona.ID != "" {
		if context.Persona.ModelRoute.ProviderID != "" && !isExplicitLocalProviderRoute(context.ProviderRoute.ProviderID) {
			context.ProviderRoute.ProviderID = context.Persona.ModelRoute.ProviderID
			context.ProviderRoute.Available = true
		}
		if context.Persona.ModelRoute.Model != "" && !isExplicitLocalProviderRoute(context.ProviderRoute.ProviderID) {
			context.ProviderRoute.Model = context.Persona.ModelRoute.Model
		}
		context.EnabledTools = toolResolutionsForNamesAndEvents(context.Persona.AllowedToolNames, events)
		if context.Thread.Mode != ThreadModeWork {
			context.EnabledTools = withoutWorkModeToolResolutions(context.EnabledTools)
		} else {
			context.EnabledTools = workToolResolutionsForLatestIntent(context.EnabledTools, context.Messages)
		}
	}
	context.MCPAvailability = mcpAvailabilityForToolResolutions(context.EnabledTools, events)
}

func workToolResolutionsForLatestIntent(tools []ToolResolution, messages []Message) []ToolResolution {
	intent := latestUserIntent(messages)
	if !intent.scoped {
		return tools
	}
	filtered := make([]ToolResolution, 0, len(tools))
	for _, tool := range tools {
		if workToolAllowedForIntent(tool.Name, intent) {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

type workIntent struct {
	scoped         bool
	workspaceRead  bool
	workspaceWrite bool
	command        bool
	lsp            bool
	web            bool
	browser        bool
	artifact       bool
	agent          bool
	memory         bool
	todo           bool
}

func latestUserIntent(messages []Message) workIntent {
	content := ""
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Role == MessageRoleUser {
			content = strings.ToLower(messages[index].Content)
			break
		}
	}
	intent := workIntent{}
	if content == "" {
		return intent
	}
	intent.workspaceRead = containsAny(content, []string{"文件", "目录", "folder", "file", "list", "列", "看下", "看看", "读取", "read", "grep", "搜索代码", "查找", "分类", "下载", "downloads", "desktop", "documents", "项目", "仓库", "repo", "github"})
	intent.workspaceWrite = containsAny(content, []string{"修改", "编辑", "写入", "创建文件", "保存", "删除文件", "改代码", "edit", "write", "create file", "delete file", "patch"})
	intent.command = containsAny(content, []string{"运行命令", "执行命令", "终端", "shell", "命令行", "跑测试", "测试", "验证", "verify", "构建", "启动服务", "go test", "bun test", "npm", "pnpm", "build", "lint"})
	intent.lsp = containsAny(content, []string{"定义", "引用", "符号", "诊断", "lsp", "definition", "references", "symbols", "diagnostics"})
	intent.web = strings.Contains(content, "http://") || strings.Contains(content, "https://") || containsAny(content, []string{"网页", "联网", "搜索", "新闻", "最新", "web", "search", "fetch", "github"})
	intent.browser = containsAny(content, []string{"浏览器", "打开网页", "点击", "截图", "browser", "screenshot", "click"})
	intent.artifact = containsAny(content, []string{"产物", "artifact", "文档", "报告", "生成文件"})
	intent.agent = containsAny(content, []string{"子任务", "并行", "多agent", "multi-agent", "spawn", "reviewer"})
	intent.memory = containsAny(content, []string{"记忆", "memory", "记住", "偏好"})
	intent.todo = containsAny(content, []string{"计划", "步骤", "待办", "todo", "plan"})
	intent.scoped = intent.workspaceRead || intent.workspaceWrite || intent.command || intent.lsp || intent.web || intent.browser || intent.artifact || intent.agent || intent.memory || intent.todo || isChineseCasualGreeting(content)
	return intent
}

func isChineseCasualGreeting(content string) bool {
	trimmed := strings.TrimSpace(content)
	return trimmed == "你好" || trimmed == "你好呀" || trimmed == "你好啊"
}

func workToolAllowedForIntent(name string, intent workIntent) bool {
	switch {
	case name == ToolNameCurrentTime || name == ToolNameLoadTools || name == ToolNameLoadSkill:
		return true
	case IsWorkspaceReadOnlyToolName(name):
		return intent.workspaceRead || intent.workspaceWrite || intent.lsp || intent.todo
	case IsWorkspaceToolName(name):
		return intent.workspaceWrite
	case IsSandboxToolName(name):
		return intent.command
	case IsLSPToolName(name):
		return intent.lsp
	case IsWebToolName(name):
		return intent.web
	case IsBrowserToolName(name):
		return intent.browser
	case IsArtifactToolName(name):
		return intent.artifact
	case IsAgentToolName(name):
		return intent.agent
	case IsMemoryToolName(name):
		return intent.memory
	case IsTodoToolName(name):
		return intent.todo
	default:
		return true
	}
}

func containsAny(value string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(value, needle) {
			return true
		}
	}
	return false
}

func withoutWorkModeToolResolutions(tools []ToolResolution) []ToolResolution {
	filtered := make([]ToolResolution, 0, len(tools))
	for _, tool := range tools {
		if tool.Name == ToolNameWebSearch || tool.Name == ToolNameWebFetch || (!IsWorkspaceToolName(tool.Name) && !IsSandboxToolName(tool.Name) && !IsLSPToolName(tool.Name) && !IsWebToolName(tool.Name) && !IsBrowserToolName(tool.Name) && !IsArtifactToolName(tool.Name) && !IsAgentToolName(tool.Name) && !IsMemoryToolName(tool.Name) && !IsTodoToolName(tool.Name)) {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func runProviderID(inputProviderID string, snapshot PersonaSnapshot) string {
	if isExplicitLocalProviderRoute(inputProviderID) {
		return strings.TrimSpace(inputProviderID)
	}
	return firstNonEmpty(snapshot.ModelRoute.ProviderID, inputProviderID)
}

func runModel(inputProviderID string, inputModel string, snapshot PersonaSnapshot) string {
	if isExplicitLocalProviderRoute(inputProviderID) {
		return strings.TrimSpace(inputModel)
	}
	return firstNonEmpty(snapshot.ModelRoute.Model, inputModel)
}

func isExplicitLocalProviderRoute(providerID string) bool {
	return strings.HasPrefix(strings.TrimSpace(providerID), "local_")
}

func validateBuiltInPersonaConfigs(configs []BuiltInPersonaConfig) error {
	defaults := 0
	for _, config := range configs {
		if strings.TrimSpace(config.Slug) == "" || strings.TrimSpace(config.Name) == "" || strings.TrimSpace(config.SystemPrompt) == "" || strings.TrimSpace(config.Version) == "" {
			return NewError(CodeInvalidRequest, "Built-in persona requires slug, name, prompt, and version.")
		}
		if strings.TrimSpace(config.ModelRoute.ProviderID) == "" || strings.TrimSpace(config.ModelRoute.Model) == "" {
			return NewError(CodeInvalidRequest, "Built-in persona requires a model route.")
		}
		for _, name := range config.AllowedToolNames {
			toolName := strings.TrimSpace(name)
			if toolName != ToolNameCurrentTime && !IsDiscoveryToolName(toolName) && !IsWorkspaceToolName(toolName) && !IsSandboxToolName(toolName) && !IsLSPToolName(toolName) && !IsWebToolName(toolName) && !IsBrowserToolName(toolName) && !IsArtifactToolName(toolName) && !IsAgentToolName(toolName) && !IsMemoryToolName(toolName) && !IsTodoToolName(toolName) && !IsMCPToolName(toolName) {
				return NewError(CodeInvalidRequest, "Built-in persona references an unsupported tool.")
			}
		}
		if config.IsDefault {
			defaults++
		}
	}
	if len(configs) > 0 && defaults != 1 {
		return NewError(CodeInvalidRequest, "Exactly one built-in persona must be default.")
	}
	return nil
}

func (s *MemoryService) resolvePersonaSnapshotLocked(thread Thread, runPersonaID string) (PersonaSnapshot, error) {
	if personaID := strings.TrimSpace(runPersonaID); personaID != "" {
		return s.snapshotForPersonaLocked(personaID, PersonaResolvedFromRun)
	}
	if thread.PersonaID != "" {
		return s.snapshotForPersonaLocked(thread.PersonaID, PersonaResolvedFromThread)
	}
	for _, persona := range s.personas {
		if persona.IsDefault && persona.IsActive {
			return s.snapshotForPersonaLocked(persona.ID, PersonaResolvedFromDefault)
		}
	}
	return PersonaSnapshot{}, nil
}

func (s *MemoryService) snapshotForPersonaLocked(personaID string, resolvedFrom PersonaResolvedFrom) (PersonaSnapshot, error) {
	persona, version, err := s.activePersonaVersionLocked(personaID)
	if err != nil {
		return PersonaSnapshot{}, err
	}
	return PersonaSnapshot{
		ID:               persona.ID,
		Slug:             persona.Slug,
		Version:          version.Version,
		Name:             persona.Name,
		Description:      persona.Description,
		SystemPrompt:     version.SystemPrompt,
		ModelRoute:       version.ModelRoute,
		AllowedToolNames: append([]string(nil), version.AllowedToolNames...),
		ReasoningMode:    version.ReasoningMode,
		BudgetSummary:    version.BudgetSummary,
		ResolvedFrom:     resolvedFrom,
	}, nil
}

func (s *MemoryService) activePersonaVersionLocked(personaID string) (Persona, PersonaVersion, error) {
	persona, ok := s.personas[strings.TrimSpace(personaID)]
	if !ok || !persona.IsActive {
		return Persona{}, PersonaVersion{}, NewError(CodeInvalidRequest, "Persona could not be resolved for this run.")
	}
	version, ok := s.personaVersions[persona.ID+":"+persona.ActiveVersion]
	if !ok {
		return Persona{}, PersonaVersion{}, NewError(CodeInvalidRequest, "Persona could not be resolved for this run.")
	}
	return persona, version, nil
}

func toolResolutionsForNames(names []string) []ToolResolution {
	return toolResolutionsForNamesAndEvents(names, nil)
}

func toolResolutionsForNamesAndEvents(names []string, events []RunEvent) []ToolResolution {
	catalog := ToolCatalogFromEvents(events)
	byName := map[string]ToolCatalogEntry{}
	for _, entry := range catalog {
		byName[entry.Name] = entry
	}
	tools := make([]ToolResolution, 0, len(names))
	for _, name := range names {
		toolName := strings.TrimSpace(name)
		entry, ok := byName[toolName]
		if !ok || !entry.Enabled || entry.ExecutionState != ToolExecutionStateExecutable {
			continue
		}
		tools = append(tools, ToolResolution{Name: entry.Name, ApprovalPolicy: string(entry.ApprovalPolicy), ExecutionState: string(entry.ExecutionState), Source: string(entry.Source), Group: string(entry.Group), InputSchemaHash: entry.InputSchemaHash, RiskLevel: string(entry.RiskLevel)})
	}
	return tools
}

func mcpAvailabilityForToolResolutions(tools []ToolResolution, events []RunEvent) MCPToolAvailabilitySummary {
	names := make([]string, 0)
	executableNames := make([]string, 0)
	byServer := map[string]*MCPServerAvailabilitySummary{}
	for _, tool := range tools {
		if IsMCPToolName(tool.Name) {
			names = append(names, tool.Name)
			if tool.ExecutionState == string(ToolExecutionStateExecutable) {
				executableNames = appendUniqueString(executableNames, tool.Name)
			}
			slug := mcpServerSlugFromToolName(tool.Name)
			server := ensureMCPServerAvailability(byServer, slug)
			server.CandidateNames = appendUniqueString(server.CandidateNames, tool.Name)
			server.CandidateCount = len(server.CandidateNames)
		}
	}
	for _, event := range events {
		applyMCPDiscoveryEvent(byServer, event)
	}
	if len(names) == 0 && len(byServer) == 0 {
		return MCPToolAvailabilitySummary{}
	}
	serverSlugs := make([]string, 0, len(byServer))
	for slug := range byServer {
		serverSlugs = append(serverSlugs, slug)
	}
	sort.Strings(serverSlugs)
	serverSummaries := make([]MCPServerAvailabilitySummary, 0, len(serverSlugs))
	errorCodes := make([]string, 0)
	lastDiscoveredAt := ""
	serversEnabled := 0
	serversSucceeded := 0
	serversFailed := 0
	for _, slug := range serverSlugs {
		server := byServer[slug]
		sort.Strings(server.CandidateNames)
		if len(server.CandidateNames) > 0 {
			server.CandidateCount = len(server.CandidateNames)
			names = append(names, server.CandidateNames...)
		}
		if server.Enabled {
			serversEnabled++
		}
		switch server.DiscoveryStatus {
		case "succeeded":
			serversSucceeded++
		case "failed", "rejected":
			serversFailed++
		}
		if server.RedactedErrorCode != "" {
			errorCodes = appendUniqueString(errorCodes, server.RedactedErrorCode)
		}
		if server.LastDiscoveredAt != "" && server.LastDiscoveredAt > lastDiscoveredAt {
			lastDiscoveredAt = server.LastDiscoveredAt
		}
		serverSummaries = append(serverSummaries, *server)
	}
	names = uniqueSortedStrings(names)
	return MCPToolAvailabilitySummary{
		ServersConfigured:           len(serverSummaries),
		ServersEnabled:              serversEnabled,
		ServersSucceeded:            serversSucceeded,
		ServersFailed:               serversFailed,
		ServerSummaries:             serverSummaries,
		CandidateNames:              names,
		NonExecutableCandidateNames: nonExecutableMCPNames(names, executableNames),
		ExecutionEnabled:            len(executableNames) > 0,
		RedactedErrorCodes:          errorCodes,
		LastDiscoveredAt:            lastDiscoveredAt,
	}
}

func nonExecutableMCPNames(names []string, executable []string) []string {
	enabled := map[string]struct{}{}
	for _, name := range executable {
		enabled[name] = struct{}{}
	}
	result := make([]string, 0)
	for _, name := range names {
		if _, ok := enabled[name]; !ok {
			result = appendUniqueString(result, name)
		}
	}
	return result
}

func ensureMCPServerAvailability(byServer map[string]*MCPServerAvailabilitySummary, slug string) *MCPServerAvailabilitySummary {
	server := byServer[slug]
	if server == nil {
		server = &MCPServerAvailabilitySummary{
			ServerSafeID:    "mcp:" + slug,
			ServerSlug:      slug,
			Enabled:         true,
			DiscoveryStatus: "unavailable",
		}
		byServer[slug] = server
	}
	return server
}

func applyMCPDiscoveryEvent(byServer map[string]*MCPServerAvailabilitySummary, event RunEvent) {
	if event.Type != "mcp_discovery_succeeded" && event.Type != "mcp_discovery_failed" && event.Type != "mcp_discovery_rejected" {
		return
	}
	slug := firstNonEmpty(metadataStringValue(event.Metadata, "server_slug"), metadataStringValue(event.Metadata, "mcp_server_slug"))
	if slug == "" || !isSafeMCPNamePart(slug, true) {
		return
	}
	server := ensureMCPServerAvailability(byServer, slug)
	status := metadataStringValue(event.Metadata, "status")
	if status == "" {
		status = strings.TrimPrefix(event.Type, "mcp_discovery_")
	}
	server.DiscoveryStatus = status
	if status == "disabled" {
		server.Enabled = false
	}
	for _, name := range metadataStringSlice(event.Metadata, "candidate_names") {
		if IsMCPToolName(name) && mcpServerSlugFromToolName(name) == slug {
			server.CandidateNames = appendUniqueString(server.CandidateNames, name)
		}
	}
	if server.CandidateCount == 0 {
		server.CandidateCount = metadataIntValue(event.Metadata, "tool_count")
	}
	if len(server.CandidateNames) > 0 {
		server.CandidateCount = len(server.CandidateNames)
	}
	server.RedactedErrorCode = metadataStringValue(event.Metadata, "error_code")
	server.LastDiscoveredAt = event.CreatedAt.Format(time.RFC3339Nano)
}

func mcpServerSlugFromToolName(name string) string {
	parts := strings.Split(strings.TrimSpace(name), ".")
	if len(parts) != 3 {
		return ""
	}
	return parts[1]
}

func metadataStringSlice(metadata map[string]any, key string) []string {
	switch value := metadata[key].(type) {
	case []string:
		return append([]string(nil), value...)
	case []any:
		items := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok {
				items = append(items, strings.TrimSpace(text))
			}
		}
		return items
	default:
		return nil
	}
}

func metadataIntValue(metadata map[string]any, key string) int {
	switch value := metadata[key].(type) {
	case int:
		return value
	case int64:
		return int(value)
	case float64:
		return int(value)
	default:
		return 0
	}
}

func appendUniqueString(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func uniqueSortedStrings(values []string) []string {
	unique := make([]string, 0, len(values))
	for _, value := range values {
		unique = appendUniqueString(unique, value)
	}
	sort.Strings(unique)
	return unique
}

func IsMCPToolName(name string) bool {
	parts := strings.Split(strings.TrimSpace(name), ".")
	if len(parts) != 3 || parts[0] != "mcp" {
		return false
	}
	return isSafeMCPNamePart(parts[1], true) && isSafeMCPNamePart(parts[2], false)
}

func isSafeMCPNamePart(value string, allowHyphenStart bool) bool {
	if value == "" || len(value) > 64 {
		return false
	}
	for i, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			if i == 0 && r == '-' && !allowHyphenStart {
				return false
			}
			continue
		}
		return false
	}
	return true
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
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
		if existing.RunID == run.ID && existing.ToolCallID != input.ToolCallID && !isToolCallTerminal(existing) {
			return ToolCall{}, nil, NewError(CodeInvalidRequest, "Another tool call is already pending or executing.")
		}
	}
	key := run.ID + ":" + input.ToolCallID
	if existing, ok := s.toolCalls[key]; ok {
		return existing, nil, nil
	}
	now := s.now()
	arguments := RedactEventMetadata(input.ArgumentsSummary)
	call := ToolCall{ID: NewToolCallID(), ThreadID: run.ThreadID, RunID: run.ID, ToolCallID: input.ToolCallID, ToolName: input.ToolName, CandidateSchemaHash: input.CandidateSchemaHash, ArgumentsSummary: arguments, ApprovalStatus: input.ApprovalStatus, ExecutionStatus: input.ExecutionStatus, RequestedAt: now, UpdatedAt: now}
	s.toolCalls[key] = call
	autoApproved := call.ApprovalStatus == ToolCallApprovalApproved && call.ExecutionStatus == ToolCallExecutionNotStarted
	if autoApproved {
		run.Status = RunStatusQueued
	} else {
		run.Status = RunStatusBlockedOnToolApproval
	}
	run.UpdatedAt = now
	s.runs[run.ID] = run
	for id, job := range s.backgroundJobs {
		if job.RunID == run.ID && job.UserID == user.ID && (job.Status == BackgroundJobStatusQueued || job.Status == BackgroundJobStatusRetrying) {
			job.Status = BackgroundJobStatusCancelled
			job.UpdatedAt = now
			s.backgroundJobs[id] = job
		}
	}
	metadata := toolCallEventMetadataForRun(s.runEvents[run.ID], call)
	requested := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallRequested, "Tool call requested", nil, metadata, now)
	if autoApproved {
		approved := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallApproved, "Tool call auto-approved", nil, metadata, now)
		return call, []RunEvent{requested, approved}, nil
	}
	required := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallApprovalRequired, "Tool approval required", nil, metadata, now)
	return call, []RunEvent{requested, required}, nil
}

func isToolCallTerminal(call ToolCall) bool {
	switch call.ExecutionStatus {
	case ToolCallExecutionSucceeded, ToolCallExecutionFailed, ToolCallExecutionCancelled:
		return true
	default:
		return false
	}
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
	event := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallApproved, "Tool call approved", nil, toolCallEventMetadataForRun(s.runEvents[run.ID], call), now)
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
	denied := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallDenied, "Tool call denied by user", nil, toolCallEventMetadataForRun(s.runEvents[run.ID], call), now)
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
	event := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallExecuting, "Tool call executing", nil, toolCallEventMetadataForRun(s.runEvents[run.ID], call), now)
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
	succeeded := s.appendRunEventLocked(run, RunEventCategoryProgress, EventToolCallSucceeded, "Tool call succeeded", nil, toolCallEventMetadataForRun(s.runEvents[run.ID], call), now)
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
	failed := s.appendRunEventLocked(run, RunEventCategoryError, EventToolCallFailed, message, nil, toolCallEventMetadataForRun(s.runEvents[run.ID], call), now)
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
	arguments := call.ArgumentsSummary
	if call.ToolName == ToolNameWorkspaceWriteFile || call.ToolName == ToolNameWorkspaceEdit {
		arguments = safeWorkspaceMutationArguments(call.ToolName, call.ArgumentsSummary)
	} else if call.ToolName == ToolNameSandboxExecCommand {
		arguments = safeSandboxExecArguments(call.ArgumentsSummary)
	}
	metadata := map[string]any{"tool_call_id": call.ToolCallID, "tool_name": call.ToolName, "arguments_summary": arguments, "approval_status": string(call.ApprovalStatus), "execution_status": string(call.ExecutionStatus)}
	if IsMCPToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceMCP)
		metadata["tool_group"] = string(ToolCatalogGroupMCP)
		metadata["server_slug"] = mcpServerSlugFromToolName(call.ToolName)
		metadata["candidate_schema_hash"] = call.CandidateSchemaHash
	} else if IsWorkspaceToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupWorkspace)
	} else if IsSandboxToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupSandbox)
	} else if IsLSPToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupLSP)
	} else if IsWebToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupWeb)
	} else if IsBrowserToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupBrowser)
	} else if IsArtifactToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupArtifact)
	} else if IsAgentToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupAgent)
	} else if IsMemoryToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupMemory)
	} else if IsTodoToolName(call.ToolName) {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupTodo)
	} else {
		metadata["tool_source"] = string(ToolCatalogSourceBuiltin)
		metadata["tool_group"] = string(ToolCatalogGroupRuntime)
	}
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

func safeSandboxExecArguments(args map[string]any) map[string]any {
	safe := map[string]any{}
	if argv := sandboxArgumentStrings(args["argv"]); len(argv) > 0 {
		safe["argv"] = argv
	}
	if cwd, ok := args["cwd"].(string); ok {
		safe["cwd"] = strings.TrimSpace(cwd)
	}
	if timeoutMS, ok := args["timeout_ms"]; ok {
		safe["timeout_ms"] = timeoutMS
	}
	if maxOutputBytes, ok := args["max_output_bytes"]; ok {
		safe["max_output_bytes"] = maxOutputBytes
	}
	return safe
}

func sandboxArgumentStrings(value any) []string {
	switch typed := value.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok {
				items = append(items, text)
			}
		}
		return items
	default:
		return nil
	}
}

func safeWorkspaceMutationArguments(toolName string, args map[string]any) map[string]any {
	safe := map[string]any{}
	if path, ok := args["path"].(string); ok {
		safe["path"] = strings.TrimSpace(path)
	}
	if maxBytes, ok := args["max_bytes"]; ok {
		safe["max_bytes"] = maxBytes
	}
	switch toolName {
	case ToolNameWorkspaceWriteFile:
		if content, ok := args["content"].(string); ok {
			safe["content_bytes"] = len([]byte(content))
			safe["content_provided"] = content != ""
		}
	case ToolNameWorkspaceEdit:
		if oldText, ok := args["old_text"].(string); ok {
			safe["old_text_bytes"] = len([]byte(oldText))
			safe["old_text_provided"] = oldText != ""
		}
		if newText, ok := args["new_text"].(string); ok {
			safe["new_text_bytes"] = len([]byte(newText))
			safe["new_text_provided"] = newText != ""
		}
	}
	return safe
}

func toolCallEventMetadataForRun(events []RunEvent, call ToolCall) map[string]any {
	metadata := toolCallEventMetadata(call)
	loopIndex := toolCallLoopIndex(events, call.ToolCallID)
	if loopIndex <= 0 {
		loopIndex = nextToolCallLoopIndex(events)
	}
	return withToolLoopMetadata(metadata, loopIndex)
}

func withToolLoopMetadata(metadata map[string]any, loopIndex int) map[string]any {
	if metadata == nil {
		metadata = map[string]any{}
	}
	if loopIndex > 0 {
		metadata[LoopMetadataKeyIndex] = loopIndex
		metadata[LoopMetadataKeyMax] = DefaultMaxBoundedToolCallsPerRun
	}
	return metadata
}

func toolCallLoopIndex(events []RunEvent, toolCallID string) int {
	seen := map[string]bool{}
	index := 0
	for _, event := range events {
		if event.Type != EventToolCallRequested {
			continue
		}
		id, _ := event.Metadata["tool_call_id"].(string)
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		index++
		seen[id] = true
		if id == toolCallID {
			return index
		}
	}
	return 0
}

func nextToolCallLoopIndex(events []RunEvent) int {
	seen := map[string]bool{}
	for _, event := range events {
		if event.Type != EventToolCallRequested {
			continue
		}
		id, _ := event.Metadata["tool_call_id"].(string)
		id = strings.TrimSpace(id)
		if id != "" {
			seen[id] = true
		}
	}
	return len(seen) + 1
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

func (s *MemoryService) upsertThreadLocked(id string, userID string, title string, mode ThreadMode, personaID string) Thread {
	now := s.now()
	thread, ok := s.threads[id]
	if !ok {
		thread = Thread{ID: id, UserID: userID, CreatedAt: now}
	}
	thread.Title = title
	thread.Mode = mode
	thread.PersonaID = strings.TrimSpace(personaID)
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

func buildMemoryOverviewSnapshot(items []MemorySearchResult, now time.Time, rebuilt bool) MemoryOverviewSnapshot {
	hits := make([]MemorySnapshotHit, 0, len(items))
	lines := []string{}
	for _, item := range items {
		abstract := strings.TrimSpace(item.Summary)
		if abstract == "" {
			abstract = strings.TrimSpace(item.Title)
		}
		hits = append(hits, MemorySnapshotHit{URI: "memory://" + item.ID, EntryID: item.ID, Title: item.Title, Abstract: abstract, IsLeaf: true, UpdatedAt: item.UpdatedAt})
		lines = append(lines, "- "+item.Title+": "+abstract)
	}
	block := strings.Join(lines, "\n")
	if block == "" {
		block = "No approved memories yet."
	}
	return MemoryOverviewSnapshot{MemoryBlock: block, Hits: hits, UpdatedAt: now, Rebuilt: rebuilt}
}

func semanticMemorySnapshotItems(items []MemorySearchResult) []MemorySearchResult {
	filtered := make([]MemorySearchResult, 0, len(items))
	for _, item := range items {
		if item.SourceType == "notebook" {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func buildMemoryImpressionSnapshot(items []MemorySearchResult, now time.Time, rebuilt bool) MemoryImpressionSnapshot {
	lines := []string{}
	for _, item := range items {
		summary := strings.TrimSpace(item.Summary)
		if summary == "" {
			summary = strings.TrimSpace(item.Title)
		}
		lines = append(lines, "- "+summary)
	}
	impression := strings.Join(lines, "\n")
	if impression == "" {
		impression = "No approved memories have been saved yet."
	}
	return MemoryImpressionSnapshot{Impression: impression, UpdatedAt: now, Rebuilt: rebuilt}
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
