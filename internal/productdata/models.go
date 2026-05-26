package productdata

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

type ThreadMode string

type ThreadLifecycleStatus string

type MessageRole string

type RunStatus string

type RunSource string

type RunEventCategory string

type BackgroundJobKind string

type BackgroundJobStatus string

type ToolCallApprovalStatus string

type ToolCallExecutionStatus string

type WorkerQueueStatus string

type WorkerStatus string

type PipelineStepName string

type PersonaSource string

type PersonaResolvedFrom string

type Code string

type MemoryScopeType string

type MemoryEntryStatus string

type MemorySafetyState string

type MemoryWriteStatus string

type ToolCatalogSource string

type ToolCatalogGroup string

type ToolRiskLevel string

type ToolApprovalPolicy string

type ToolExecutionState string

type ModelProviderConfig struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Family  string `json:"family"`
	BaseURL string `json:"base_url"`
	APIKey  string `json:"-"`
	Model   string `json:"model"`
	Enabled bool   `json:"enabled"`
}

type WebSearchConfig struct {
	UserID       string `json:"user_id"`
	TavilyAPIKey string `json:"-"`
	BraveAPIKey  string `json:"-"`
}

type MCPServerConfigRecord struct {
	UserID      string            `json:"user_id"`
	Slug        string            `json:"slug"`
	DisplayName string            `json:"display_name"`
	Enabled     bool              `json:"enabled"`
	Transport   string            `json:"transport"`
	Command     string            `json:"-"`
	Args        []string          `json:"-"`
	Env         map[string]string `json:"-"`
	TimeoutMS   int               `json:"timeout_ms"`
}

func normalizeModelProviderConfig(input ModelProviderConfig) ModelProviderConfig {
	family := strings.TrimSpace(input.Family)
	if family == "" {
		family = "openai_compatible"
	}
	return ModelProviderConfig{
		ID:      strings.TrimSpace(input.ID),
		UserID:  strings.TrimSpace(input.UserID),
		Family:  family,
		BaseURL: strings.TrimSpace(input.BaseURL),
		APIKey:  strings.TrimSpace(input.APIKey),
		Model:   strings.TrimSpace(input.Model),
		Enabled: input.Enabled,
	}
}

func normalizeWebSearchConfig(input WebSearchConfig) WebSearchConfig {
	return WebSearchConfig{
		UserID:       strings.TrimSpace(input.UserID),
		TavilyAPIKey: strings.TrimSpace(input.TavilyAPIKey),
		BraveAPIKey:  strings.TrimSpace(input.BraveAPIKey),
	}
}

func normalizeMCPServerConfigRecord(input MCPServerConfigRecord) MCPServerConfigRecord {
	args := make([]string, 0, len(input.Args))
	for _, arg := range input.Args {
		if trimmed := strings.TrimSpace(arg); trimmed != "" {
			args = append(args, trimmed)
		}
	}
	env := map[string]string{}
	for key, value := range input.Env {
		key = strings.TrimSpace(key)
		if key != "" {
			env[key] = strings.TrimSpace(value)
		}
	}
	timeout := input.TimeoutMS
	if timeout <= 0 {
		timeout = 5000
	}
	return MCPServerConfigRecord{
		UserID:      strings.TrimSpace(input.UserID),
		Slug:        strings.TrimSpace(input.Slug),
		DisplayName: strings.TrimSpace(input.DisplayName),
		Enabled:     input.Enabled,
		Transport:   strings.TrimSpace(input.Transport),
		Command:     strings.TrimSpace(input.Command),
		Args:        args,
		Env:         env,
		TimeoutMS:   timeout,
	}
}

const (
	ThreadModeChat ThreadMode = "chat"
	ThreadModeWork ThreadMode = "work"

	ThreadLifecycleActive   ThreadLifecycleStatus = "active"
	ThreadLifecycleArchived ThreadLifecycleStatus = "archived"

	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"

	RunStatusPending               RunStatus = "pending"
	RunStatusQueued                RunStatus = "queued"
	RunStatusRunning               RunStatus = "running"
	RunStatusRecovering            RunStatus = "recovering"
	RunStatusBlockedOnToolApproval RunStatus = "blocked_on_tool_approval"
	RunStatusCompleted             RunStatus = "completed"
	RunStatusFailed                RunStatus = "failed"
	RunStatusStopped               RunStatus = "stopped"

	RunSourceLocalSimulated RunSource = "local_simulated"
	RunSourceModelGateway   RunSource = "model_gateway"

	RunEventCategoryLifecycle RunEventCategory = "lifecycle"
	RunEventCategoryProgress  RunEventCategory = "progress"
	RunEventCategoryMessage   RunEventCategory = "message"
	RunEventCategoryError     RunEventCategory = "error"
	RunEventCategoryFinal     RunEventCategory = "final"

	BackgroundJobKindRunExecution BackgroundJobKind = "run_execution"

	BackgroundJobStatusQueued    BackgroundJobStatus = "queued"
	BackgroundJobStatusLeased    BackgroundJobStatus = "leased"
	BackgroundJobStatusRetrying  BackgroundJobStatus = "retrying"
	BackgroundJobStatusCompleted BackgroundJobStatus = "completed"
	BackgroundJobStatusFailed    BackgroundJobStatus = "failed"
	BackgroundJobStatusCancelled BackgroundJobStatus = "cancelled"
	BackgroundJobStatusDead      BackgroundJobStatus = "dead"

	ToolCallApprovalNotRequired ToolCallApprovalStatus = "not_required"
	ToolCallApprovalRequired    ToolCallApprovalStatus = "required"
	ToolCallApprovalApproved    ToolCallApprovalStatus = "approved"
	ToolCallApprovalDenied      ToolCallApprovalStatus = "denied"
	ToolCallApprovalCancelled   ToolCallApprovalStatus = "cancelled"

	ToolCallExecutionNotStarted ToolCallExecutionStatus = "not_started"
	ToolCallExecutionBlocked    ToolCallExecutionStatus = "blocked"
	ToolCallExecutionExecuting  ToolCallExecutionStatus = "executing"
	ToolCallExecutionSucceeded  ToolCallExecutionStatus = "succeeded"
	ToolCallExecutionFailed     ToolCallExecutionStatus = "failed"
	ToolCallExecutionCancelled  ToolCallExecutionStatus = "cancelled"

	WorkerQueueStatusReady     WorkerQueueStatus = "ready"
	WorkerQueueStatusPaused    WorkerQueueStatus = "paused"
	WorkerQueueStatusUnhealthy WorkerQueueStatus = "unhealthy"
	WorkerQueueStatusDegraded  WorkerQueueStatus = "degraded"

	WorkerStatusReady     WorkerStatus = "ready"
	WorkerStatusPaused    WorkerStatus = "paused"
	WorkerStatusUnhealthy WorkerStatus = "unhealthy"
	WorkerStatusDegraded  WorkerStatus = "degraded"
	WorkerStatusStopped   WorkerStatus = "stopped"

	PipelineStepEnqueue        PipelineStepName = "enqueue"
	PipelineStepClaim          PipelineStepName = "claim"
	PipelineStepPrepareContext PipelineStepName = "prepare_context"
	PipelineStepResolveTools   PipelineStepName = "resolve_tools"
	PipelineStepInvokeRuntime  PipelineStepName = "invoke_runtime"
	PipelineStepFinalize       PipelineStepName = "finalize"
	PipelineStepRecover        PipelineStepName = "recover"
	PipelineStepFail           PipelineStepName = "fail"

	PersonaSourceBuiltIn PersonaSource = "built_in"

	PersonaResolvedFromRun     PersonaResolvedFrom = "run"
	PersonaResolvedFromThread  PersonaResolvedFrom = "thread"
	PersonaResolvedFromDefault PersonaResolvedFrom = "default"

	EventRunQueued                = "run_queued"
	EventJobClaimed               = "job_claimed"
	EventLeaseRenewed             = "lease_renewed"
	EventPipelineStepStarted      = "pipeline_step_started"
	EventPipelineStepCompleted    = "pipeline_step_completed"
	EventPipelineStepFailed       = "pipeline_step_failed"
	EventJobRecovering            = "job_recovering"
	EventJobRetryScheduled        = "job_retry_scheduled"
	EventStopRequested            = "stop_requested"
	EventJobAttemptFailed         = "job_attempt_failed"
	EventJobRetryExhausted        = "job_retry_exhausted"
	EventToolCallRequested        = "tool_call_requested"
	EventToolCallApprovalRequired = "tool_call_approval_required"
	EventToolCallApproved         = "tool_call_approved"
	EventToolCallDenied           = "tool_call_denied"
	EventToolCallExecuting        = "tool_call_executing"
	EventToolCallSucceeded        = "tool_call_succeeded"
	EventToolCallFailed           = "tool_call_failed"
	EventToolCallCancelled        = "tool_call_cancelled"
	EventRunCompleted             = "run_completed"
	EventRunFailed                = "run_failed"
	EventRunStopped               = "run_stopped"
	EventMemorySnapshotLoaded     = "memory_snapshot_loaded"
	EventMemoryWriteProposed      = "memory_write_proposed"
	EventMemoryWriteApproved      = "memory_write_approved"
	EventMemoryWriteDenied        = "memory_write_denied"
	EventMemoryEntryDeleted       = "memory_entry_deleted"
	EventWorkTodoUpdated          = "work.todo.updated"

	CodeInvalidRequest        Code = "invalid_request"
	CodeThreadNotFound        Code = "thread_not_found"
	CodeRunNotFound           Code = "run_not_found"
	CodeActiveRunExists       Code = "active_run_exists"
	CodeProviderUnavailable   Code = "provider_unavailable"
	CodeProviderMisconfigured Code = "provider_misconfigured"
	CodeMethodNotAllowed      Code = "method_not_allowed"
	CodeMemoryNotFound        Code = "memory_not_found"
	CodeArtifactNotFound      Code = "artifact_not_found"
	CodeInternalError         Code = "internal_error"
)

const (
	MaxThreadTitleLength                                = 120
	MaxClientMessageIDLength                            = 120
	ToolNameCurrentTime                                 = "runtime.get_current_time"
	ToolNameLoadTools                                   = "tool.load_tools"
	ToolNameLoadSkill                                   = "skill.load_skill"
	ToolNameWorkspaceGlob                               = "workspace.glob"
	ToolNameWorkspaceGrep                               = "workspace.grep"
	ToolNameWorkspaceRead                               = "workspace.read"
	ToolNameWorkspaceWriteFile                          = "workspace.write_file"
	ToolNameWorkspaceEdit                               = "workspace.edit"
	ToolNameWorkspacePatchPreview                       = "workspace.patch_preview"
	ToolNameWorkspacePatchApply                         = "workspace.patch_apply"
	ToolNameSandboxExecCommand                          = "sandbox.exec_command"
	ToolNameSandboxStartProcess                         = "sandbox.start_process"
	ToolNameSandboxContinueProcess                      = "sandbox.continue_process"
	ToolNameSandboxTerminateProcess                     = "sandbox.terminate_process"
	ToolNameLSPDiagnostics                              = "lsp.diagnostics"
	ToolNameLSPSymbols                                  = "lsp.symbols"
	ToolNameLSPReferences                               = "lsp.references"
	ToolNameLSPDefinition                               = "lsp.definition"
	ToolNameLSPHover                                    = "lsp.hover"
	ToolNameWebFetch                                    = "web.fetch"
	ToolNameWebSearch                                   = "web.search"
	ToolNameBrowserOpen                                 = "browser.open"
	ToolNameBrowserSnapshot                             = "browser.snapshot"
	ToolNameBrowserClickLink                            = "browser.click_link"
	ToolNameBrowserScreenshot                           = "browser.screenshot"
	ToolNameBrowserType                                 = "browser.type"
	ToolNameBrowserPress                                = "browser.press"
	ToolNameArtifactCreateText                          = "artifact.create_text"
	ToolNameArtifactRead                                = "artifact.read"
	ToolNameArtifactList                                = "artifact.list"
	ToolNameAgentSpawn                                  = "agent.spawn"
	ToolNameAgentList                                   = "agent.list"
	ToolNameAgentComplete                               = "agent.complete"
	ToolNameTodoWrite                                   = "todo.write"
	ToolSourceInternal                                  = "internal"
	ToolSourceMCP                                       = "mcp"
	ToolCatalogSourceBuiltin         ToolCatalogSource  = "builtin"
	ToolCatalogSourceMCP             ToolCatalogSource  = "mcp"
	ToolCatalogGroupRuntime          ToolCatalogGroup   = "runtime"
	ToolCatalogGroupDiscovery        ToolCatalogGroup   = "discovery"
	ToolCatalogGroupMCP              ToolCatalogGroup   = "mcp"
	ToolCatalogGroupWorkspace        ToolCatalogGroup   = "workspace"
	ToolCatalogGroupArtifact         ToolCatalogGroup   = "artifact"
	ToolCatalogGroupSandbox          ToolCatalogGroup   = "sandbox"
	ToolCatalogGroupLSP              ToolCatalogGroup   = "lsp"
	ToolCatalogGroupWeb              ToolCatalogGroup   = "web"
	ToolCatalogGroupBrowser          ToolCatalogGroup   = "browser"
	ToolCatalogGroupAgent            ToolCatalogGroup   = "agent"
	ToolCatalogGroupTodo             ToolCatalogGroup   = "todo"
	ToolRiskLow                      ToolRiskLevel      = "low"
	ToolRiskMedium                   ToolRiskLevel      = "medium"
	ToolRiskHigh                     ToolRiskLevel      = "high"
	ToolApprovalAlwaysRequired       ToolApprovalPolicy = "always_required"
	ToolApprovalReadOnly             ToolApprovalPolicy = "read_only"
	ToolApprovalDisabled             ToolApprovalPolicy = "disabled"
	ToolExecutionStateExecutable     ToolExecutionState = "executable"
	ToolExecutionStateDisabled       ToolExecutionState = "disabled"
	ToolExecutionStateNotDiscovered  ToolExecutionState = "not_discovered"
	ToolExecutionStateNotAllowed     ToolExecutionState = "not_allowed"
	ToolExecutionStateNonExecutable  ToolExecutionState = "non_executable"
	DefaultMaxBoundedToolCallsPerRun                    = 6
	LoopMetadataKeyIndex                                = "loop_index"
	LoopMetadataKeyMax                                  = "loop_max"
	MaxWorkTodoItems                                    = 8
	MaxWorkTodoTitleLength                              = 160
	MaxWorkTodoSummaryLength                            = 240
	MemoryScopeUser                  MemoryScopeType    = "user"
	MemoryScopeThread                MemoryScopeType    = "thread"
	MemoryEntryApproved              MemoryEntryStatus  = "approved"
	MemoryEntryTombstoned            MemoryEntryStatus  = "tombstoned"
	MemoryEntryDisabled              MemoryEntryStatus  = "disabled"
	MemorySafetySafe                 MemorySafetyState  = "safe"
	MemorySafetyRedacted             MemorySafetyState  = "redacted"
	MemorySafetyBlocked              MemorySafetyState  = "blocked"
	MemoryWritePending               MemoryWriteStatus  = "pending"
	MemoryWriteApproved              MemoryWriteStatus  = "approved"
	MemoryWriteDenied                MemoryWriteStatus  = "denied"
)

type ProductError struct {
	Code    Code
	Message string
}

func (e ProductError) Error() string { return e.Message }

func NewError(code Code, message string) error {
	return ProductError{Code: code, Message: message}
}

func ErrorCode(err error) Code {
	var productErr ProductError
	if errors.As(err, &productErr) {
		return productErr.Code
	}
	return CodeInternalError
}

type User struct {
	ID          string    `json:"id"`
	DisplayName string    `json:"display_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Thread struct {
	ID              string                `json:"id"`
	UserID          string                `json:"-"`
	Title           string                `json:"title"`
	Mode            ThreadMode            `json:"mode"`
	PersonaID       string                `json:"persona_id,omitempty"`
	LifecycleStatus ThreadLifecycleStatus `json:"lifecycle_status"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
	ArchivedAt      *time.Time            `json:"archived_at,omitempty"`
}

type Message struct {
	ID              string         `json:"id"`
	ThreadID        string         `json:"thread_id"`
	UserID          string         `json:"-"`
	Role            MessageRole    `json:"role"`
	Content         string         `json:"content"`
	Metadata        map[string]any `json:"metadata"`
	ClientMessageID *string        `json:"client_message_id,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
}

type Run struct {
	ID              string     `json:"id"`
	ThreadID        string     `json:"thread_id"`
	UserID          string     `json:"-"`
	Status          RunStatus  `json:"status"`
	Source          RunSource  `json:"source"`
	Title           string     `json:"title"`
	PersonaID       string     `json:"persona_id,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	StopRequestedAt *time.Time `json:"stop_requested_at,omitempty"`
	ErrorCode       *string    `json:"error_code,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

type ToolCall struct {
	ID                  string                  `json:"id"`
	ThreadID            string                  `json:"thread_id"`
	RunID               string                  `json:"run_id"`
	ToolCallID          string                  `json:"tool_call_id"`
	ToolName            string                  `json:"tool_name"`
	CandidateSchemaHash string                  `json:"candidate_schema_hash,omitempty"`
	ArgumentsSummary    map[string]any          `json:"arguments_summary"`
	ApprovalStatus      ToolCallApprovalStatus  `json:"approval_status"`
	ExecutionStatus     ToolCallExecutionStatus `json:"execution_status"`
	ResultSummary       map[string]any          `json:"result_summary,omitempty"`
	ErrorCode           *string                 `json:"error_code,omitempty"`
	ErrorMessage        *string                 `json:"error_message,omitempty"`
	RequestedAt         time.Time               `json:"requested_at"`
	UpdatedAt           time.Time               `json:"updated_at"`
}

type ToolCatalogEntry struct {
	Name            string             `json:"name"`
	DisplayName     string             `json:"display_name"`
	Description     string             `json:"description"`
	Source          ToolCatalogSource  `json:"source"`
	Group           ToolCatalogGroup   `json:"group"`
	InputSchemaHash string             `json:"input_schema_hash,omitempty"`
	RiskLevel       ToolRiskLevel      `json:"risk_level"`
	ApprovalPolicy  ToolApprovalPolicy `json:"approval_policy"`
	Enabled         bool               `json:"enabled"`
	ExecutionState  ToolExecutionState `json:"execution_state"`
	SafeMetadata    map[string]any     `json:"safe_metadata,omitempty"`
}

type BackgroundJob struct {
	ID               string              `json:"id"`
	RunID            string              `json:"run_id"`
	ThreadID         string              `json:"thread_id"`
	UserID           string              `json:"-"`
	Kind             BackgroundJobKind   `json:"kind"`
	Status           BackgroundJobStatus `json:"status"`
	Priority         int                 `json:"priority"`
	AttemptCount     int                 `json:"attempt_count"`
	MaxAttempts      int                 `json:"max_attempts"`
	ScheduledAt      time.Time           `json:"scheduled_at"`
	LeasedBy         *string             `json:"leased_by,omitempty"`
	LeaseExpiresAt   *time.Time          `json:"lease_expires_at,omitempty"`
	OwnershipVersion int                 `json:"ownership_version"`
	Metadata         map[string]any      `json:"metadata,omitempty"`
	LastErrorCode    *string             `json:"last_error_code,omitempty"`
	LastError        *string             `json:"last_error_message,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
}

type WorkerQueueDiagnostics struct {
	QueueStatus              WorkerQueueStatus `json:"queue_status"`
	WorkerStatus             WorkerStatus      `json:"worker_status"`
	QueuedCount              int               `json:"queued_count"`
	LeasedCount              int               `json:"leased_count"`
	StaleCount               int               `json:"stale_count"`
	RetryingCount            int               `json:"retrying_count"`
	BlockedToolApprovalCount int               `json:"blocked_tool_approval_count"`
	ResumableToolCallCount   int               `json:"resumable_tool_call_count"`
	DeadCount                int               `json:"dead_count"`
	UpdatedAt                time.Time         `json:"updated_at"`
}

type Persona struct {
	ID            string        `json:"id"`
	Slug          string        `json:"slug"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Source        PersonaSource `json:"source"`
	IsDefault     bool          `json:"is_default"`
	IsActive      bool          `json:"is_active"`
	ActiveVersion string        `json:"active_version"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
}

type PersonaModelRoute struct {
	ProviderID string `json:"provider_id"`
	Model      string `json:"model"`
}

type PersonaVersion struct {
	PersonaID        string            `json:"persona_id"`
	Version          string            `json:"version"`
	SystemPrompt     string            `json:"-"`
	ModelRoute       PersonaModelRoute `json:"model_route"`
	AllowedToolNames []string          `json:"allowed_tool_names"`
	ReasoningMode    string            `json:"reasoning_mode"`
	BudgetSummary    string            `json:"budget_summary"`
	CreatedAt        time.Time         `json:"created_at"`
}

type BuiltInPersonaConfig struct {
	Slug             string
	Name             string
	Description      string
	SystemPrompt     string
	ModelRoute       PersonaModelRoute
	AllowedToolNames []string
	ReasoningMode    string
	BudgetSummary    string
	Version          string
	IsDefault        bool
}

type PersonaSyncResult struct {
	Synced             int    `json:"synced"`
	CreatedPersonas    int    `json:"created_personas"`
	CreatedVersions    int    `json:"created_versions"`
	ActivatedVersions  int    `json:"activated_versions"`
	DefaultPersonaSlug string `json:"default_persona_slug"`
}

type PersonaSnapshot struct {
	ID               string              `json:"id"`
	Slug             string              `json:"slug"`
	Version          string              `json:"version"`
	Name             string              `json:"name"`
	Description      string              `json:"description"`
	SystemPrompt     string              `json:"-"`
	ModelRoute       PersonaModelRoute   `json:"model_route"`
	AllowedToolNames []string            `json:"allowed_tool_names"`
	ReasoningMode    string              `json:"reasoning_mode"`
	BudgetSummary    string              `json:"budget_summary"`
	ResolvedFrom     PersonaResolvedFrom `json:"resolved_from"`
}

func (p PersonaSnapshot) SafeSummary() map[string]any {
	if p.ID == "" {
		return map[string]any{}
	}
	return RedactEventMetadata(map[string]any{
		"persona_id":                 p.ID,
		"persona_slug":               p.Slug,
		"persona_version":            p.Version,
		"persona_name":               p.Name,
		"persona_description":        p.Description,
		"persona_model_provider_id":  p.ModelRoute.ProviderID,
		"persona_model":              p.ModelRoute.Model,
		"persona_allowed_tools":      append([]string(nil), p.AllowedToolNames...),
		"persona_allowed_tool_count": len(p.AllowedToolNames),
		"persona_reasoning_mode":     p.ReasoningMode,
		"persona_budget_summary":     p.BudgetSummary,
		"persona_resolved_from":      string(p.ResolvedFrom),
	})
}

type RunContext struct {
	Run                    Run
	Thread                 Thread
	Messages               []Message
	Job                    BackgroundJob
	ProviderRoute          ProviderRoute
	EnabledTools           []ToolResolution
	MCPAvailability        MCPToolAvailabilitySummary
	ContinuationProjection ContinuationProjection
	Persona                PersonaSnapshot
	MemorySnapshot         MemorySnapshot
}

type ProviderRoute struct {
	ProviderID string
	Model      string
	Available  bool
}

type ToolResolution struct {
	Name            string
	ApprovalPolicy  string
	ExecutionState  string
	Source          string
	Group           string
	InputSchemaHash string
	RiskLevel       string
}

type MCPToolAvailabilitySummary struct {
	ServersConfigured           int
	ServersEnabled              int
	ServersSucceeded            int
	ServersFailed               int
	ServerSummaries             []MCPServerAvailabilitySummary
	CandidateNames              []string
	NonExecutableCandidateNames []string
	ExecutionEnabled            bool
	RedactedErrorCodes          []string
	LastDiscoveredAt            string
}

type MCPServerAvailabilitySummary struct {
	ServerSafeID      string
	ServerSlug        string
	Enabled           bool
	DiscoveryStatus   string
	CandidateCount    int
	CandidateNames    []string
	RedactedErrorCode string
	LastDiscoveredAt  string
}

type ContinuationProjection struct {
	ToolCallID string
	Available  bool
}

type MemoryEntry struct {
	ID             string            `json:"id"`
	UserID         string            `json:"-"`
	ScopeType      MemoryScopeType   `json:"scope_type"`
	ScopeID        string            `json:"scope_id"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	Content        string            `json:"content,omitempty"`
	Status         MemoryEntryStatus `json:"status"`
	SafetyState    MemorySafetyState `json:"safety_state"`
	SourceThreadID string            `json:"source_thread_id,omitempty"`
	SourceRunID    string            `json:"source_run_id,omitempty"`
	SourceEventID  string            `json:"source_event_id,omitempty"`
	ContentHash    string            `json:"content_hash"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      *time.Time        `json:"deleted_at,omitempty"`
	DeletedBy      string            `json:"-"`
	DeleteReason   string            `json:"delete_reason,omitempty"`
}

type MemorySearchResult struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Summary          string          `json:"summary"`
	Content          string          `json:"content,omitempty"`
	ScopeType        MemoryScopeType `json:"scope_type"`
	ScopeID          string          `json:"scope_id"`
	Status           string          `json:"status"`
	SafetyState      string          `json:"safety_state"`
	SourceThreadID   string          `json:"source_thread_id,omitempty"`
	SourceRunID      string          `json:"source_run_id,omitempty"`
	SourceEventID    string          `json:"source_event_id,omitempty"`
	SourceType       string          `json:"source_type"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        *time.Time      `json:"deleted_at,omitempty"`
	RankReason       string          `json:"rank_reason,omitempty"`
	RedactionApplied bool            `json:"redaction_applied"`
}

type MemorySearchInput struct {
	Query             string
	ScopeType         MemoryScopeType
	ScopeID           string
	SourceThreadID    string
	SourceRunID       string
	SourceType        string
	IncludeTombstoned bool
	Limit             int
	Purpose           string
}

type MemorySearchOutput struct {
	Items         []MemorySearchResult `json:"items"`
	ExcludedCount int                  `json:"excluded_count"`
}

type MemoryAuditInput struct {
	ThreadID    string
	SourceRunID string
	EventType   string
	Limit       int
}

type MemoryEntryAccessInput struct {
	ScopeType      MemoryScopeType `json:"scope_type"`
	ScopeID        string          `json:"scope_id"`
	SourceThreadID string          `json:"source_thread_id"`
	SourceRunID    string          `json:"source_run_id"`
}

type MemoryAuditItem struct {
	ID               string    `json:"id"`
	EventType        string    `json:"event_type"`
	Summary          string    `json:"summary"`
	ThreadID         string    `json:"thread_id,omitempty"`
	RunID            string    `json:"run_id,omitempty"`
	MemoryEntryID    string    `json:"memory_entry_id,omitempty"`
	MemoryProposalID string    `json:"memory_proposal_id,omitempty"`
	Status           string    `json:"status,omitempty"`
	ScopeType        string    `json:"scope_type,omitempty"`
	SourceType       string    `json:"source_type,omitempty"`
	RedactionApplied bool      `json:"redaction_applied"`
	OccurredAt       time.Time `json:"occurred_at"`
}

type MemoryAuditOutput struct {
	Items []MemoryAuditItem `json:"items"`
}

type MemoryWriteProposal struct {
	ID             string            `json:"id"`
	UserID         string            `json:"-"`
	ScopeType      MemoryScopeType   `json:"scope_type"`
	ScopeID        string            `json:"scope_id"`
	Title          string            `json:"title"`
	Summary        string            `json:"summary"`
	Content        string            `json:"content,omitempty"`
	Status         MemoryWriteStatus `json:"status"`
	SafetyState    MemorySafetyState `json:"safety_state"`
	SourceThreadID string            `json:"source_thread_id,omitempty"`
	SourceRunID    string            `json:"source_run_id,omitempty"`
	SourceEventID  string            `json:"source_event_id,omitempty"`
	IdempotencyKey string            `json:"-"`
	CreatedEntryID string            `json:"created_entry_id,omitempty"`
	CreatedAt      time.Time         `json:"created_at"`
	DecidedAt      *time.Time        `json:"decided_at,omitempty"`
	DecidedBy      string            `json:"-"`
	DecisionReason string            `json:"decision_reason,omitempty"`
}

type MemoryWriteDecision struct {
	Proposal MemoryWriteProposal `json:"proposal"`
	Entry    MemoryEntry         `json:"entry,omitempty"`
}

type MemoryTombstone struct {
	EntryID      string    `json:"entry_id"`
	Status       string    `json:"status"`
	DeletedAt    time.Time `json:"deleted_at"`
	AuditEventID string    `json:"audit_event_id,omitempty"`
}

type MemorySnapshot struct {
	RunID            string               `json:"run_id"`
	ThreadID         string               `json:"thread_id"`
	Entries          []MemorySearchResult `json:"entries"`
	Limit            int                  `json:"limit"`
	TotalCandidates  int                  `json:"total_candidates"`
	LoadStatus       string               `json:"load_status"`
	RedactionApplied bool                 `json:"redaction_applied"`
}

func (c RunContext) SafeSummary() map[string]any {
	summary := map[string]any{
		"message_count":               len(c.Messages),
		"has_job_metadata":            len(c.Job.Metadata) > 0,
		"enabled_tool_count":          len(c.EnabledTools),
		"has_continuation_projection": c.ContinuationProjection.Available,
	}
	if c.ProviderRoute.ProviderID != "" {
		summary["provider_id"] = RedactEventText(c.ProviderRoute.ProviderID)
	}
	if c.ProviderRoute.Model != "" {
		summary["model"] = RedactEventText(c.ProviderRoute.Model)
	}
	if c.MemorySnapshot.LoadStatus != "" {
		summary["memory_status"] = c.MemorySnapshot.LoadStatus
		summary["memory_entry_count"] = len(c.MemorySnapshot.Entries)
		summary["memory_redaction_applied"] = c.MemorySnapshot.RedactionApplied
	}
	for key, value := range c.Persona.SafeSummary() {
		summary[key] = value
	}
	for key, value := range c.MCPAvailability.SafeSummary() {
		summary[key] = value
	}
	return RedactEventMetadata(summary)
}

func (m MCPToolAvailabilitySummary) SafeSummary() map[string]any {
	if m.ServersConfigured == 0 && len(m.CandidateNames) == 0 && len(m.RedactedErrorCodes) == 0 {
		return map[string]any{}
	}
	return RedactEventMetadata(map[string]any{
		"mcp_servers_configured":             m.ServersConfigured,
		"mcp_servers_enabled":                m.ServersEnabled,
		"mcp_servers_succeeded":              m.ServersSucceeded,
		"mcp_servers_failed":                 m.ServersFailed,
		"mcp_server_summaries":               m.safeServerSummaries(),
		"mcp_candidate_count":                len(m.CandidateNames),
		"mcp_candidate_names":                append([]string(nil), m.CandidateNames...),
		"mcp_non_executable_candidate_names": append([]string(nil), m.NonExecutableCandidateNames...),
		"mcp_execution_enabled":              m.ExecutionEnabled,
		"mcp_error_codes":                    append([]string(nil), m.RedactedErrorCodes...),
		"mcp_last_discovered_at":             m.LastDiscoveredAt,
	})
}

func (m MCPToolAvailabilitySummary) safeServerSummaries() []any {
	summaries := make([]any, 0, len(m.ServerSummaries))
	for _, server := range m.ServerSummaries {
		summaries = append(summaries, map[string]any{
			"server_safe_id":      server.ServerSafeID,
			"server_slug":         server.ServerSlug,
			"enabled":             server.Enabled,
			"discovery_status":    server.DiscoveryStatus,
			"candidate_count":     server.CandidateCount,
			"candidate_names":     append([]string(nil), server.CandidateNames...),
			"redacted_error_code": server.RedactedErrorCode,
			"last_discovered_at":  server.LastDiscoveredAt,
			"execution_enabled":   false,
		})
	}
	return summaries
}

func (c RunContext) ToolResolutionSummary() map[string]any {
	names := make([]string, 0, len(c.EnabledTools))
	schemaHashes := map[string]string{}
	for _, tool := range c.EnabledTools {
		names = append(names, tool.Name)
		if tool.InputSchemaHash != "" {
			schemaHashes[tool.Name] = tool.InputSchemaHash
		}
	}
	return RedactEventMetadata(map[string]any{
		"enabled_tool_count":          len(c.EnabledTools),
		"enabled_tools":               names,
		"tool_schema_hashes":          schemaHashes,
		"has_continuation_projection": c.ContinuationProjection.Available,
	})
}

type RunEvent struct {
	ID        string           `json:"id"`
	RunID     string           `json:"run_id"`
	ThreadID  string           `json:"thread_id"`
	UserID    string           `json:"-"`
	Sequence  int              `json:"sequence"`
	Category  RunEventCategory `json:"category"`
	Type      string           `json:"type"`
	Summary   string           `json:"summary"`
	Content   *string          `json:"content"`
	Metadata  map[string]any   `json:"metadata"`
	CreatedAt time.Time        `json:"created_at"`
}

type CreateThreadInput struct {
	Title     string
	Mode      ThreadMode
	PersonaID string
}

type UpdateThreadInput struct {
	Title     *string
	Mode      *ThreadMode
	PersonaID *string
}

type CreateMessageInput struct {
	Content         string
	ClientMessageID string
}

type StartRunInput struct {
	ScriptName string
	Source     RunSource
	MessageID  string
	ProviderID string
	Model      string
	PersonaID  string
}

type AppendAssistantMessageInput struct {
	Content  string
	Metadata map[string]any
}

type AppendRunEventInput struct {
	Category     RunEventCategory
	Type         string
	Summary      string
	Content      *string
	Metadata     map[string]any
	ErrorCode    string
	ErrorMessage string
}

type CreateMemoryEntryInput struct {
	ScopeType      MemoryScopeType
	ScopeID        string
	Title          string
	Content        string
	SourceThreadID string
	SourceRunID    string
	SourceEventID  string
}

type ProposeMemoryWriteInput struct {
	ScopeType      MemoryScopeType `json:"scope_type"`
	ScopeID        string          `json:"scope_id"`
	Title          string          `json:"title"`
	Content        string          `json:"content"`
	SourceThreadID string          `json:"source_thread_id"`
	SourceRunID    string          `json:"source_run_id"`
	SourceEventID  string          `json:"source_event_id"`
	IdempotencyKey string          `json:"idempotency_key"`
}

type MemoryWriteDecisionInput struct {
	IdempotencyKey string `json:"idempotency_key"`
	Reason         string `json:"reason"`
}

type DeleteMemoryEntryInput struct {
	Reason         string          `json:"reason"`
	ScopeType      MemoryScopeType `json:"scope_type"`
	ScopeID        string          `json:"scope_id"`
	SourceThreadID string          `json:"source_thread_id"`
	SourceRunID    string          `json:"source_run_id"`
}

type StopRunResult string

const (
	StopRunResultStopped         StopRunResult = "stopped"
	StopRunResultAlreadyTerminal StopRunResult = "already_terminal"
)

type StopRunOutput struct {
	Run    Run
	Result StopRunResult
	Events []RunEvent
}

type ClaimBackgroundJobInput struct {
	WorkerID     string
	LeaseSeconds int
}

type CompleteBackgroundJobInput struct {
	JobID            string
	WorkerID         string
	OwnershipVersion int
}

type FailBackgroundJobInput struct {
	JobID            string
	WorkerID         string
	OwnershipVersion int
	ErrorCode        string
	ErrorMessage     string
}

type RenewBackgroundJobLeaseInput struct {
	JobID            string
	WorkerID         string
	OwnershipVersion int
	LeaseSeconds     int
}

type RecoverBackgroundJobsInput struct {
	Limit        int
	ErrorCode    string
	ErrorMessage string
}

type RecordToolCallRequestInput struct {
	ToolCallID          string
	ToolName            string
	CandidateSchemaHash string
	ArgumentsSummary    map[string]any
	ArgumentsHash       string
	ApprovalStatus      ToolCallApprovalStatus
	ExecutionStatus     ToolCallExecutionStatus
}

type Artifact struct {
	ID           string
	ThreadID     string
	RunID        string
	Title        string
	ArtifactType string
	Content      string
	ContentBytes int
	TextExcerpt  string
	Truncated    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type AgentTaskStatus string

const (
	AgentTaskStatusSpawned   AgentTaskStatus = "spawned"
	AgentTaskStatusCompleted AgentTaskStatus = "completed"
)

type AgentTask struct {
	ID            string
	ThreadID      string
	RunID         string
	Role          string
	Goal          string
	Status        AgentTaskStatus
	ResultSummary string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type SpawnAgentTaskInput struct {
	ThreadID string
	RunID    string
	Role     string
	Goal     string
}

type ListAgentTasksInput struct {
	ThreadID string
	Limit    int
}

type CompleteAgentTaskInput struct {
	ThreadID      string
	TaskID        string
	ResultSummary string
}

type CreateArtifactInput struct {
	ThreadID string
	RunID    string
	Title    string
	Content  string
	MaxBytes int
}

type ReadArtifactInput struct {
	ThreadID   string
	ArtifactID string
	MaxBytes   int
}

type ListArtifactsInput struct {
	ThreadID string
	Limit    int
}

type BackgroundJobRecovery struct {
	Job       BackgroundJob
	Run       Run
	Events    []RunEvent
	Exhausted bool
}

type SeedThreadInput struct {
	ID    string
	Title string
	Mode  ThreadMode
}

type SeedMessageInput struct {
	ID              string
	ThreadID        string
	Content         string
	ClientMessageID string
}

func NewThreadID() string         { return prefixedID("thr") }
func NewMessageID() string        { return prefixedID("msg") }
func NewRunID() string            { return prefixedID("run") }
func NewRunEventID() string       { return prefixedID("evt") }
func NewBackgroundJobID() string  { return prefixedID("job") }
func NewToolCallID() string       { return prefixedID("tool") }
func NewPersonaID() string        { return prefixedID("persona") }
func NewMemoryEntryID() string    { return prefixedID("mem") }
func NewMemoryProposalID() string { return prefixedID("memprop") }
func NewArtifactID() string       { return prefixedID("art") }
func NewAgentTaskID() string      { return prefixedID("agt") }

func prefixedID(prefix string) string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(buf))
}

func ValidateMessageRole(role MessageRole) error {
	switch role {
	case MessageRoleUser, MessageRoleAssistant:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Message role is invalid.")
	}
}

func ValidateRunStatus(status RunStatus) error {
	switch status {
	case RunStatusPending, RunStatusQueued, RunStatusRunning, RunStatusRecovering, RunStatusBlockedOnToolApproval, RunStatusCompleted, RunStatusFailed, RunStatusStopped:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Run status is invalid.")
	}
}

func ValidateToolCallApprovalStatus(status ToolCallApprovalStatus) error {
	switch status {
	case ToolCallApprovalNotRequired, ToolCallApprovalRequired, ToolCallApprovalApproved, ToolCallApprovalDenied, ToolCallApprovalCancelled:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Tool call approval status is invalid.")
	}
}

func ValidateToolCallExecutionStatus(status ToolCallExecutionStatus) error {
	switch status {
	case ToolCallExecutionNotStarted, ToolCallExecutionBlocked, ToolCallExecutionExecuting, ToolCallExecutionSucceeded, ToolCallExecutionFailed, ToolCallExecutionCancelled:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Tool call execution status is invalid.")
	}
}

func ValidateToolCallRequestInput(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	input.ToolCallID = strings.TrimSpace(input.ToolCallID)
	input.ToolName = strings.TrimSpace(input.ToolName)
	input.CandidateSchemaHash = strings.TrimSpace(input.CandidateSchemaHash)
	input.ArgumentsHash = strings.TrimSpace(input.ArgumentsHash)
	if input.ToolCallID == "" || input.ToolName == "" {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call id and name are required.")
	}
	if input.ToolName != ToolNameCurrentTime && !IsDiscoveryToolName(input.ToolName) && !IsWorkspaceToolName(input.ToolName) && !IsSandboxToolName(input.ToolName) && !IsLSPToolName(input.ToolName) && !IsWebToolName(input.ToolName) && !IsBrowserToolName(input.ToolName) && !IsArtifactToolName(input.ToolName) && !IsAgentToolName(input.ToolName) && !IsTodoToolName(input.ToolName) && !IsMCPToolName(input.ToolName) {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool is not supported.")
	}
	if (input.ToolName == ToolNameWebSearch || IsDiscoveryToolName(input.ToolName)) && input.ApprovalStatus == ToolCallApprovalApproved && input.ExecutionStatus == ToolCallExecutionNotStarted {
		// web.search and discovery tools are bounded reads; they may enter the queue without a manual approval row.
	} else if input.ApprovalStatus != ToolCallApprovalRequired || input.ExecutionStatus != ToolCallExecutionBlocked {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call must start blocked on approval.")
	}
	if err := ValidateToolCallApprovalStatus(input.ApprovalStatus); err != nil {
		return RecordToolCallRequestInput{}, err
	}
	if err := ValidateToolCallExecutionStatus(input.ExecutionStatus); err != nil {
		return RecordToolCallRequestInput{}, err
	}
	if input.ArgumentsSummary == nil {
		input.ArgumentsSummary = map[string]any{}
	}
	if IsMCPToolName(input.ToolName) {
		if input.CandidateSchemaHash == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "MCP tool candidate schema hash is required.")
		}
		return input, nil
	}
	if IsWorkspaceToolName(input.ToolName) {
		return validateWorkspaceToolCallArguments(input)
	}
	if IsDiscoveryToolName(input.ToolName) {
		return validateDiscoveryToolCallArguments(input)
	}
	if IsSandboxToolName(input.ToolName) {
		return validateSandboxToolCallArguments(input)
	}
	if IsLSPToolName(input.ToolName) {
		return validateLSPToolCallArguments(input)
	}
	if IsWebToolName(input.ToolName) {
		return validateWebToolCallArguments(input)
	}
	if IsBrowserToolName(input.ToolName) {
		return validateBrowserToolCallArguments(input)
	}
	if IsArtifactToolName(input.ToolName) {
		return validateArtifactToolCallArguments(input)
	}
	if IsAgentToolName(input.ToolName) {
		return validateAgentToolCallArguments(input)
	}
	if IsTodoToolName(input.ToolName) {
		return validateTodoToolCallArguments(input)
	}
	for key := range input.ArgumentsSummary {
		if key != "timezone" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	value, ok := input.ArgumentsSummary["timezone"]
	if !ok || value == nil {
		input.ArgumentsSummary["timezone"] = "UTC"
		return input, nil
	}
	timezone, ok := value.(string)
	if !ok || timezone != "UTC" {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call timezone must be UTC.")
	}
	return input, nil
}

func IsWorkspaceToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchPreview, ToolNameWorkspacePatchApply:
		return true
	default:
		return false
	}
}

func IsDiscoveryToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameLoadTools, ToolNameLoadSkill:
		return true
	default:
		return false
	}
}

func IsSandboxToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess:
		return true
	default:
		return false
	}
}

func IsLSPToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameLSPDefinition, ToolNameLSPHover:
		return true
	default:
		return false
	}
}

func IsWebToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameWebFetch, ToolNameWebSearch:
		return true
	default:
		return false
	}
}

func IsBrowserToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameBrowserScreenshot, ToolNameBrowserType, ToolNameBrowserPress:
		return true
	default:
		return false
	}
}

func IsArtifactToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList:
		return true
	default:
		return false
	}
}

func IsAgentToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentComplete:
		return true
	default:
		return false
	}
}

func IsTodoToolName(name string) bool {
	switch strings.TrimSpace(name) {
	case ToolNameTodoWrite:
		return true
	default:
		return false
	}
}

func validateDiscoveryToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]map[string]struct{}{
		ToolNameLoadTools: {"queries": {}, "names": {}, "limit": {}},
		ToolNameLoadSkill: {"name": {}, "limit": {}},
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[input.ToolName][key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	switch input.ToolName {
	case ToolNameLoadTools:
		if _, ok := input.ArgumentsSummary["queries"]; ok && !safeStringListArgument(input.ArgumentsSummary["queries"], 5) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool lookup queries are invalid.")
		}
		if _, ok := input.ArgumentsSummary["names"]; ok && !safeStringListArgument(input.ArgumentsSummary["names"], 20) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool lookup names are invalid.")
		}
		if _, ok := input.ArgumentsSummary["limit"]; ok && !positiveNumberArgument(input.ArgumentsSummary["limit"]) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool lookup limit is invalid.")
		}
	case ToolNameLoadSkill:
		name, ok := input.ArgumentsSummary["name"].(string)
		if !ok || strings.TrimSpace(name) == "" || len(strings.TrimSpace(name)) > 120 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Skill name is required.")
		}
		if _, ok := input.ArgumentsSummary["limit"]; ok && !positiveNumberArgument(input.ArgumentsSummary["limit"]) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Skill lookup limit is invalid.")
		}
	}
	return input, nil
}

func validateWorkspaceToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]map[string]struct{}{
		ToolNameWorkspaceGlob:         {"pattern": {}, "path": {}, "limit": {}},
		ToolNameWorkspaceGrep:         {"query": {}, "pattern": {}, "path": {}, "include": {}, "case_sensitive": {}, "limit": {}},
		ToolNameWorkspaceRead:         {"path": {}, "offset": {}, "limit": {}, "max_bytes": {}},
		ToolNameWorkspaceWriteFile:    {"path": {}, "content": {}, "max_bytes": {}},
		ToolNameWorkspaceEdit:         {"path": {}, "old_text": {}, "new_text": {}, "max_bytes": {}},
		ToolNameWorkspacePatchPreview: {"path": {}, "old_text": {}, "new_text": {}, "max_bytes": {}},
		ToolNameWorkspacePatchApply:   {"path": {}, "old_text": {}, "new_text": {}, "max_bytes": {}},
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[input.ToolName][key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	switch input.ToolName {
	case ToolNameWorkspaceGlob:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "pattern")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace glob pattern is required.")
		}
	case ToolNameWorkspaceGrep:
		query := workspaceArgumentString(input.ArgumentsSummary, "query")
		if query == "" {
			query = workspaceArgumentString(input.ArgumentsSummary, "pattern")
		}
		if strings.TrimSpace(query) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace grep query is required.")
		}
	case ToolNameWorkspaceRead:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace read path is required.")
		}
	case ToolNameWorkspaceWriteFile:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace write path is required.")
		}
	case ToolNameWorkspaceEdit:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace edit path is required.")
		}
		if workspaceArgumentString(input.ArgumentsSummary, "old_text") == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace edit old text is required.")
		}
	case ToolNameWorkspacePatchPreview:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace patch preview path is required.")
		}
		if workspaceArgumentString(input.ArgumentsSummary, "old_text") == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace patch preview old text is required.")
		}
	case ToolNameWorkspacePatchApply:
		if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace patch apply path is required.")
		}
		if workspaceArgumentString(input.ArgumentsSummary, "old_text") == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace patch apply old text is required.")
		}
	}
	return input, nil
}

func validateSandboxToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]struct{}{"argv": {}, "cwd": {}, "timeout_ms": {}, "max_output_bytes": {}}
	if input.ToolName == ToolNameSandboxStartProcess {
		allowed["stdin"] = struct{}{}
	}
	if input.ToolName == ToolNameSandboxContinueProcess {
		allowed = map[string]struct{}{"process_id": {}, "cursor": {}, "stdin_text": {}, "input_seq": {}, "close_stdin": {}}
	}
	if input.ToolName == ToolNameSandboxTerminateProcess {
		allowed = map[string]struct{}{"process_id": {}}
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	if input.ToolName == ToolNameSandboxContinueProcess || input.ToolName == ToolNameSandboxTerminateProcess {
		processID, ok := input.ArgumentsSummary["process_id"].(string)
		if !ok || strings.TrimSpace(processID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process id is required.")
		}
		input.ArgumentsSummary["process_id"] = strings.TrimSpace(processID)
		if input.ToolName == ToolNameSandboxContinueProcess {
			if _, ok := input.ArgumentsSummary["cursor"]; ok && !nonNegativeNumberArgument(input.ArgumentsSummary["cursor"]) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process cursor is invalid.")
			}
			if _, ok := input.ArgumentsSummary["close_stdin"]; ok && !boolArgument(input.ArgumentsSummary["close_stdin"]) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process close_stdin is invalid.")
			}
			if stdinText, ok := input.ArgumentsSummary["stdin_text"]; ok {
				text, ok := stdinText.(string)
				if !ok || len(text) > 8192 {
					return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process stdin_text is invalid.")
				}
				if !positiveNumberArgument(input.ArgumentsSummary["input_seq"]) {
					return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process input_seq is required.")
				}
			} else if _, ok := input.ArgumentsSummary["input_seq"]; ok {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process input_seq requires stdin_text.")
			}
		}
		return input, nil
	}
	argv, ok := input.ArgumentsSummary["argv"]
	if !ok || !sandboxArgumentStringSliceNonEmpty(argv) {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox exec argv is required.")
	}
	stdinEnabled, ok := input.ArgumentsSummary["stdin"].(bool)
	if _, exists := input.ArgumentsSummary["stdin"]; exists && !ok {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox process stdin is invalid.")
	}
	if !sandboxArgumentUsesBoundedAllowlist(argv) && !(input.ToolName == ToolNameSandboxStartProcess && stdinEnabled && sandboxArgumentAllowsStdinProcess(argv)) {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Sandbox exec command is not allowed.")
	}
	return input, nil
}

func validateLSPToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]struct{}{"path": {}, "query": {}, "line": {}, "column": {}, "include_declaration": {}, "language": {}, "limit": {}}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	if strings.TrimSpace(workspaceArgumentString(input.ArgumentsSummary, "path")) == "" {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "LSP path is required.")
	}
	if input.ToolName == ToolNameLSPReferences || input.ToolName == ToolNameLSPDefinition || input.ToolName == ToolNameLSPHover {
		if !positiveNumberArgument(input.ArgumentsSummary["line"]) || !positiveNumberArgument(input.ArgumentsSummary["column"]) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "LSP line and column are required.")
		}
	}
	return input, nil
}

func validateWebToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	if input.ToolName == ToolNameWebSearch {
		normalizeWebSearchArgumentAliases(input.ArgumentsSummary)
		allowed := map[string]struct{}{"query": {}, "provider": {}, "limit": {}, "timeout_ms": {}}
		for key := range input.ArgumentsSummary {
			if _, ok := allowed[key]; !ok {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
			}
		}
		query, ok := input.ArgumentsSummary["query"].(string)
		if !ok || strings.TrimSpace(query) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Web search query is required.")
		}
		input.ArgumentsSummary["query"] = strings.TrimSpace(query)
		if provider, ok := input.ArgumentsSummary["provider"]; ok && provider != nil {
			value, ok := provider.(string)
			if !ok || !isSupportedWebSearchProvider(value) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Web search provider is not supported.")
			}
			input.ArgumentsSummary["provider"] = strings.TrimSpace(strings.ToLower(value))
		}
		return input, nil
	}
	allowed := map[string]struct{}{"url": {}, "max_bytes": {}, "timeout_ms": {}}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	url, ok := input.ArgumentsSummary["url"].(string)
	if !ok || strings.TrimSpace(url) == "" {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Web fetch URL is required.")
	}
	input.ArgumentsSummary["url"] = strings.TrimSpace(url)
	return input, nil
}

func normalizeWebSearchArgumentAliases(args map[string]any) {
	if args == nil {
		return
	}
	for _, key := range []string{"q", "search_query", "searchQuery"} {
		if _, exists := args["query"]; !exists {
			if value, ok := args[key]; ok {
				args["query"] = value
			}
		}
		delete(args, key)
	}
	for _, key := range []string{"count", "max_results", "maxResults", "num_results", "numResults"} {
		if _, exists := args["limit"]; !exists {
			if value, ok := args[key]; ok {
				args["limit"] = value
			}
		}
		delete(args, key)
	}
}

func isSupportedWebSearchProvider(provider string) bool {
	switch strings.TrimSpace(strings.ToLower(provider)) {
	case "", "tavily", "brave":
		return true
	default:
		return false
	}
}

func validateBrowserToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]map[string]struct{}{
		ToolNameBrowserOpen:       {"url": {}, "max_bytes": {}, "timeout_ms": {}},
		ToolNameBrowserSnapshot:   {"session_id": {}},
		ToolNameBrowserClickLink:  {"session_id": {}, "link_index": {}, "max_bytes": {}, "timeout_ms": {}},
		ToolNameBrowserScreenshot: {"session_id": {}},
		ToolNameBrowserType:       {"session_id": {}, "target": {}, "text": {}},
		ToolNameBrowserPress:      {"session_id": {}, "key": {}},
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[input.ToolName][key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	switch input.ToolName {
	case ToolNameBrowserOpen:
		url, ok := input.ArgumentsSummary["url"].(string)
		if !ok || strings.TrimSpace(url) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser URL is required.")
		}
		input.ArgumentsSummary["url"] = strings.TrimSpace(url)
	case ToolNameBrowserSnapshot, ToolNameBrowserScreenshot:
		sessionID, ok := input.ArgumentsSummary["session_id"].(string)
		if !ok || strings.TrimSpace(sessionID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser session id is required.")
		}
		input.ArgumentsSummary["session_id"] = strings.TrimSpace(sessionID)
	case ToolNameBrowserClickLink:
		sessionID, ok := input.ArgumentsSummary["session_id"].(string)
		if !ok || strings.TrimSpace(sessionID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser session id is required.")
		}
		input.ArgumentsSummary["session_id"] = strings.TrimSpace(sessionID)
		if _, ok := input.ArgumentsSummary["link_index"]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser link index is required.")
		}
	case ToolNameBrowserType:
		sessionID, ok := input.ArgumentsSummary["session_id"].(string)
		if !ok || strings.TrimSpace(sessionID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser session id is required.")
		}
		target, ok := input.ArgumentsSummary["target"].(string)
		if !ok || strings.TrimSpace(target) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser target is required.")
		}
		text, ok := input.ArgumentsSummary["text"].(string)
		if !ok || strings.TrimSpace(text) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser text is required.")
		}
		input.ArgumentsSummary["session_id"] = strings.TrimSpace(sessionID)
		input.ArgumentsSummary["target"] = strings.TrimSpace(target)
	case ToolNameBrowserPress:
		sessionID, ok := input.ArgumentsSummary["session_id"].(string)
		if !ok || strings.TrimSpace(sessionID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser session id is required.")
		}
		key, ok := input.ArgumentsSummary["key"].(string)
		if !ok || strings.TrimSpace(key) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Browser key is required.")
		}
		input.ArgumentsSummary["session_id"] = strings.TrimSpace(sessionID)
		input.ArgumentsSummary["key"] = strings.TrimSpace(key)
	}
	return input, nil
}

func validateArtifactToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]map[string]struct{}{
		ToolNameArtifactCreateText: {"title": {}, "content": {}, "max_bytes": {}},
		ToolNameArtifactRead:       {"artifact_id": {}, "max_bytes": {}},
		ToolNameArtifactList:       {"limit": {}},
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[input.ToolName][key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	switch input.ToolName {
	case ToolNameArtifactCreateText:
		title, ok := input.ArgumentsSummary["title"].(string)
		if !ok || strings.TrimSpace(title) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Artifact title is required.")
		}
		content, ok := input.ArgumentsSummary["content"].(string)
		if !ok || strings.TrimSpace(content) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Artifact content is required.")
		}
		input.ArgumentsSummary["title"] = strings.TrimSpace(title)
	case ToolNameArtifactRead:
		artifactID, ok := input.ArgumentsSummary["artifact_id"].(string)
		if !ok || strings.TrimSpace(artifactID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Artifact id is required.")
		}
		input.ArgumentsSummary["artifact_id"] = strings.TrimSpace(artifactID)
	}
	return input, nil
}

func validateAgentToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	allowed := map[string]map[string]struct{}{
		ToolNameAgentSpawn:    {"role": {}, "goal": {}},
		ToolNameAgentList:     {"limit": {}},
		ToolNameAgentComplete: {"task_id": {}, "result_summary": {}},
	}
	for key := range input.ArgumentsSummary {
		if _, ok := allowed[input.ToolName][key]; !ok {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	switch input.ToolName {
	case ToolNameAgentSpawn:
		role, ok := input.ArgumentsSummary["role"].(string)
		if !ok || strings.TrimSpace(role) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent role is required.")
		}
		goal, ok := input.ArgumentsSummary["goal"].(string)
		if !ok || strings.TrimSpace(goal) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent goal is required.")
		}
		role = strings.TrimSpace(role)
		goal = strings.TrimSpace(goal)
		if !isSupportedAgentRole(role) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent role is not supported.")
		}
		if len([]rune(role)) > 64 || len([]rune(goal)) > 4000 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent task fields are too large.")
		}
		input.ArgumentsSummary["role"] = role
		input.ArgumentsSummary["goal"] = goal
	case ToolNameAgentComplete:
		taskID, ok := input.ArgumentsSummary["task_id"].(string)
		if !ok || strings.TrimSpace(taskID) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent task id is required.")
		}
		summary, ok := input.ArgumentsSummary["result_summary"].(string)
		if !ok || strings.TrimSpace(summary) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent result summary is required.")
		}
		taskID = strings.TrimSpace(taskID)
		summary = strings.TrimSpace(summary)
		if len([]rune(summary)) > 4000 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Agent result summary is too large.")
		}
		input.ArgumentsSummary["task_id"] = taskID
		input.ArgumentsSummary["result_summary"] = summary
	}
	return input, nil
}

func validateTodoToolCallArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	for key := range input.ArgumentsSummary {
		if key != "items" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	items, ok := input.ArgumentsSummary["items"].([]any)
	if !ok || len(items) == 0 {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Todo items are required.")
	}
	metadata := NormalizeWorkTodoMetadata(map[string]any{"todo_items": items, "updated_by": "provider"})
	normalized, _ := metadata["todo_items"].([]any)
	if len(normalized) == 0 {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Todo items are required.")
	}
	input.ArgumentsSummary["items"] = normalized
	return input, nil
}

func isSupportedAgentRole(role string) bool {
	switch strings.TrimSpace(role) {
	case "researcher", "implementer", "reviewer":
		return true
	default:
		return false
	}
}

func positiveNumberArgument(value any) bool {
	switch typed := value.(type) {
	case int:
		return typed > 0
	case int64:
		return typed > 0
	case float64:
		return typed >= 1 && typed == float64(int64(typed))
	default:
		return false
	}
}

func nonNegativeNumberArgument(value any) bool {
	switch typed := value.(type) {
	case int:
		return typed >= 0
	case int64:
		return typed >= 0
	case float64:
		return typed >= 0 && typed == float64(int64(typed))
	default:
		return false
	}
}

func boolArgument(value any) bool {
	_, ok := value.(bool)
	return ok
}

func safeStringListArgument(value any, maxItems int) bool {
	items, ok := value.([]any)
	if !ok || len(items) == 0 || len(items) > maxItems {
		return false
	}
	for _, item := range items {
		text, ok := item.(string)
		if !ok {
			return false
		}
		text = strings.TrimSpace(text)
		if text == "" || len(text) > 160 {
			return false
		}
	}
	return true
}

func sandboxArgumentStringSliceNonEmpty(value any) bool {
	switch typed := value.(type) {
	case []string:
		if len(typed) == 0 {
			return false
		}
		for _, item := range typed {
			if strings.TrimSpace(item) == "" {
				return false
			}
		}
		return true
	case []any:
		if len(typed) == 0 {
			return false
		}
		for _, item := range typed {
			text, ok := item.(string)
			if !ok || strings.TrimSpace(text) == "" {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func sandboxArgumentStringsForValidation(value any) []string {
	switch typed := value.(type) {
	case []string:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			out = append(out, strings.TrimSpace(item))
		}
		return out
	case []any:
		out := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil
			}
			out = append(out, strings.TrimSpace(text))
		}
		return out
	default:
		return nil
	}
}

func sandboxArgumentUsesBoundedAllowlist(value any) bool {
	argv := sandboxArgumentStringsForValidation(value)
	if len(argv) == 0 || strings.ContainsAny(argv[0], `/\`) {
		return false
	}
	switch strings.ToLower(argv[0]) {
	case "pwd":
		return len(argv) == 1
	case "ls":
		return len(argv) == 1 || (len(argv) == 2 && sandboxArgumentPathAllowed(argv[1]))
	case "cat", "wc":
		return len(argv) >= 2 && sandboxArgumentPathsAllowed(argv[1:])
	case "head", "tail":
		return sandboxArgumentHeadTailAllowed(argv[1:])
	case "sed":
		return len(argv) == 4 && argv[1] == "-n" && sandboxArgumentSedRangeAllowed(argv[2]) && sandboxArgumentPathAllowed(argv[3])
	case "rg":
		return sandboxArgumentRGAllowed(argv[1:])
	case "git":
		return sandboxArgumentGitAllowed(argv[1:])
	case "go":
		return len(argv) >= 2 && argv[1] == "test" && sandboxArgumentValidationArgsAllowed(argv[2:])
	case "bun", "npm", "pnpm":
		return sandboxArgumentPackageValidationAllowed(argv)
	default:
		return false
	}
}

func sandboxArgumentAllowsStdinProcess(value any) bool {
	argv := sandboxArgumentStringsForValidation(value)
	return len(argv) == 1 && !strings.ContainsAny(argv[0], `/\`) && strings.ToLower(argv[0]) == "cat"
}

func sandboxArgumentPathsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		if !sandboxArgumentPathAllowed(arg) {
			return false
		}
	}
	return true
}

func sandboxArgumentPathAllowed(arg string) bool {
	text := strings.TrimSpace(arg)
	if text == "" || strings.HasPrefix(text, "-") || strings.HasPrefix(text, "/") || strings.Contains(text, "..") || strings.Contains(text, "\\") {
		return false
	}
	lower := strings.ToLower(text)
	return lower != ".env" && lower != "secrets" && !strings.HasPrefix(lower, ".env.") && !strings.HasPrefix(lower, "secrets/") && !strings.Contains(lower, "/secrets/") && !strings.Contains(lower, ".ssh")
}

func sandboxArgumentHeadTailAllowed(args []string) bool {
	if len(args) == 1 {
		return sandboxArgumentPathAllowed(args[0])
	}
	if len(args) == 3 && args[0] == "-n" && positiveIntegerString(args[1]) {
		return sandboxArgumentPathAllowed(args[2])
	}
	return false
}

func sandboxArgumentSedRangeAllowed(arg string) bool {
	arg = strings.TrimSpace(arg)
	if !strings.HasSuffix(arg, "p") || strings.ContainsAny(arg, ";|&`$") {
		return false
	}
	body := strings.TrimSuffix(arg, "p")
	if body == "" {
		return false
	}
	for _, part := range strings.Split(body, ",") {
		if !positiveIntegerString(part) {
			return false
		}
	}
	return true
}

func sandboxArgumentRGAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" || strings.HasPrefix(arg, "--hidden") || strings.HasPrefix(arg, "--files-with-matches") {
			return false
		}
		if strings.Contains(arg, ".env") || strings.Contains(arg, "secrets") || strings.Contains(arg, ".ssh") {
			return false
		}
	}
	return true
}

func sandboxArgumentGitAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "status":
		return len(args) == 1 || (len(args) == 2 && args[1] == "--short")
	case "diff", "log", "show":
		return len(args) <= 4
	default:
		return false
	}
}

func sandboxArgumentValidationArgsAllowed(args []string) bool {
	for _, arg := range args {
		if strings.TrimSpace(arg) == "" || strings.Contains(arg, "..") || strings.HasPrefix(arg, "/") || strings.Contains(arg, ".env") || strings.Contains(arg, "secrets") {
			return false
		}
	}
	return true
}

func sandboxArgumentPackageValidationAllowed(argv []string) bool {
	if len(argv) < 2 {
		return false
	}
	if argv[1] == "test" {
		return len(argv) == 2 || sandboxArgumentValidationArgsAllowed(argv[2:])
	}
	return len(argv) == 4 && argv[1] == "run" && argv[2] == "build" && sandboxArgumentValidationArgsAllowed(argv[3:])
}

func positiveIntegerString(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return strings.TrimLeft(value, "0") != ""
}

func workspaceArgumentString(arguments map[string]any, key string) string {
	value, ok := arguments[key]
	if !ok || value == nil {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return text
}

func NormalizeRunSource(source RunSource) (RunSource, error) {
	if source == "" {
		return RunSourceLocalSimulated, nil
	}
	switch source {
	case RunSourceLocalSimulated, RunSourceModelGateway:
		return source, nil
	default:
		return "", NewError(CodeInvalidRequest, "Run source is invalid.")
	}
}

func TitleForRunSource(source RunSource) string {
	switch source {
	case RunSourceModelGateway:
		return "Model gateway run"
	default:
		return "Local simulated run"
	}
}

func ValidateRunEventCategory(category RunEventCategory) error {
	switch category {
	case RunEventCategoryLifecycle, RunEventCategoryProgress, RunEventCategoryMessage, RunEventCategoryError, RunEventCategoryFinal:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Run event category is invalid.")
	}
}

func IsRunTerminal(status RunStatus) bool {
	switch status {
	case RunStatusCompleted, RunStatusFailed, RunStatusStopped:
		return true
	default:
		return false
	}
}

func IsRunActive(status RunStatus) bool {
	switch status {
	case RunStatusPending, RunStatusQueued, RunStatusRunning, RunStatusRecovering, RunStatusBlockedOnToolApproval:
		return true
	default:
		return false
	}
}

func IsBackgroundJobTerminal(status BackgroundJobStatus) bool {
	switch status {
	case BackgroundJobStatusCompleted, BackgroundJobStatusFailed, BackgroundJobStatusCancelled, BackgroundJobStatusDead:
		return true
	default:
		return false
	}
}

func NormalizeThreadTitle(title string) (string, error) {
	trimmed := strings.TrimSpace(title)
	if trimmed == "" {
		return "", NewError(CodeInvalidRequest, "Thread title is required.")
	}
	if len([]rune(trimmed)) > MaxThreadTitleLength {
		return "", NewError(CodeInvalidRequest, "Thread title is too long.")
	}
	return trimmed, nil
}

func ValidateThreadMode(mode ThreadMode) error {
	switch mode {
	case ThreadModeChat, ThreadModeWork:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Thread mode must be chat or work.")
	}
}

func NormalizeMessageContent(content string) (string, error) {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return "", NewError(CodeInvalidRequest, "Message content is required.")
	}
	return trimmed, nil
}

func NormalizeClientMessageID(value string) (*string, error) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, nil
	}
	if len([]rune(trimmed)) > MaxClientMessageIDLength {
		return nil, NewError(CodeInvalidRequest, "Client message id is too long.")
	}
	return &trimmed, nil
}

func NormalizeScriptName(scriptName string) string {
	trimmed := strings.TrimSpace(scriptName)
	if trimmed == "" {
		return "m4_smoke"
	}
	return trimmed
}

func NormalizeRunEventInput(input AppendRunEventInput) (AppendRunEventInput, error) {
	if err := ValidateRunEventCategory(input.Category); err != nil {
		return AppendRunEventInput{}, err
	}
	input.Type = strings.TrimSpace(input.Type)
	input.Summary = RedactEventText(strings.TrimSpace(input.Summary))
	if input.Type == "" {
		return AppendRunEventInput{}, NewError(CodeInvalidRequest, "Run event type is required.")
	}
	if input.Summary == "" {
		return AppendRunEventInput{}, NewError(CodeInvalidRequest, "Run event summary is required.")
	}
	if input.Content != nil {
		content := RedactEventText(strings.TrimSpace(*input.Content))
		input.Content = &content
	}
	input.Metadata = RedactEventMetadata(input.Metadata)
	if input.Type == EventWorkTodoUpdated {
		input.Metadata = NormalizeWorkTodoMetadata(input.Metadata)
	}
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	input.ErrorCode = strings.TrimSpace(input.ErrorCode)
	input.ErrorMessage = RedactEventText(strings.TrimSpace(input.ErrorMessage))
	return input, nil
}

func NormalizeWorkTodoMetadata(metadata map[string]any) map[string]any {
	normalized := map[string]any{}
	items, _ := metadata["todo_items"].([]any)
	if len(items) == 0 {
		items, _ = metadata["todoItems"].([]any)
	}
	if len(items) > MaxWorkTodoItems {
		items = items[:MaxWorkTodoItems]
		normalized["redaction_applied"] = true
	}
	todos := make([]any, 0, len(items))
	for index, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			normalized["redaction_applied"] = true
			continue
		}
		todo, redacted := normalizeWorkTodoItem(item, index)
		if redacted {
			normalized["redaction_applied"] = true
		}
		todos = append(todos, todo)
	}
	normalized["todo_items"] = todos
	if updatedBy := normalizeTodoUpdatedBy(metadata["updated_by"]); updatedBy != "" {
		normalized["updated_by"] = updatedBy
	}
	if redactionFlag(metadata["redaction_applied"]) {
		normalized["redaction_applied"] = true
	}
	for key := range metadata {
		switch key {
		case "todo_items", "todoItems", "updated_by", "updatedBy", "redaction_applied", "redactionApplied":
		default:
			normalized["redaction_applied"] = true
		}
	}
	if _, ok := normalized["redaction_applied"]; !ok {
		normalized["redaction_applied"] = false
	}
	return normalized
}

func normalizeWorkTodoItem(item map[string]any, index int) (map[string]any, bool) {
	normalized := map[string]any{}
	redacted := false
	normalized["id"], redacted = safeTodoString(item["id"], "todo_"+fmt.Sprint(index+1), MaxClientMessageIDLength, redacted)
	normalized["title"], redacted = safeTodoString(item["title"], "Todo "+fmt.Sprint(index+1), MaxWorkTodoTitleLength, redacted)
	normalized["status"] = normalizeTodoStatus(item["status"])
	if summary, nextRedacted := safeTodoString(item["summary"], "", MaxWorkTodoSummaryLength, redacted); summary != "" {
		normalized["summary"] = summary
		redacted = nextRedacted
	} else {
		redacted = nextRedacted
	}
	for key := range item {
		switch key {
		case "id", "title", "status", "summary", "redaction_applied", "redactionApplied":
		default:
			redacted = true
		}
	}
	if redactionFlag(item["redaction_applied"]) || redactionFlag(item["redactionApplied"]) || redacted {
		normalized["redaction_applied"] = true
		redacted = true
	}
	return normalized, redacted
}

func safeTodoString(value any, fallback string, maxLength int, redacted bool) (string, bool) {
	text := strings.TrimSpace(fmt.Sprint(value))
	if value == nil || text == "" {
		text = fallback
	}
	if len(text) > maxLength {
		text = text[:maxLength]
		redacted = true
	}
	safe := RedactEventText(text)
	if safe != text || isUnsafeTodoText(text) {
		return "[redacted]", true
	}
	return safe, redacted
}

func isUnsafeTodoText(value string) bool {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "http://") || strings.Contains(lower, "https://") || strings.Contains(lower, "file://") || strings.Contains(lower, "/tmp/") || strings.Contains(lower, "../") || strings.Contains(lower, "~/") {
		return true
	}
	for _, marker := range []string{"curl ", "wget ", "bash ", "zsh ", "python ", "node ", "open ", "rm "} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func normalizeTodoStatus(value any) string {
	switch strings.TrimSpace(fmt.Sprint(value)) {
	case "pending", "running", "completed", "blocked", "failed":
		return strings.TrimSpace(fmt.Sprint(value))
	default:
		return "pending"
	}
}

func normalizeTodoUpdatedBy(value any) string {
	switch strings.TrimSpace(fmt.Sprint(value)) {
	case "provider", "runtime", "user":
		return strings.TrimSpace(fmt.Sprint(value))
	default:
		return ""
	}
}

func redactionFlag(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.TrimSpace(strings.ToLower(typed)) == "true"
	default:
		return false
	}
}

func RedactEventMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return map[string]any{}
	}
	redacted := make(map[string]any, len(metadata))
	for key, value := range metadata {
		if isSensitiveMetadataKey(key) {
			redacted[key] = "[redacted]"
			continue
		}
		redacted[key] = redactMetadataValue(value)
	}
	return redacted
}

func isSensitiveMetadataKey(key string) bool {
	lower := strings.ToLower(key)
	for _, marker := range []string{"api_key", "authorization", "password", "secret", "token", "credential"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func redactMetadataValue(value any) any {
	switch typed := value.(type) {
	case string:
		return RedactEventText(typed)
	case []string:
		redacted := make([]string, len(typed))
		for i, item := range typed {
			redacted[i] = RedactEventText(item)
		}
		return redacted
	case map[string]any:
		return RedactEventMetadata(typed)
	case []any:
		redacted := make([]any, len(typed))
		for i, item := range typed {
			redacted[i] = redactMetadataValue(item)
		}
		return redacted
	default:
		return value
	}
}

func RedactEventText(value string) string {
	lower := strings.ToLower(value)
	for _, marker := range []string{"postgres://", "postgresql://", "password=", "api_key", " key=", "_key=", "bearer ", "secret", "token", "credential", "authorization", "sk-", ".ssh", "id_ed25519", "id_rsa", ".env", "env=", "stdout", "stderr", "tool output", "tool_output", "provider trace", "provider_trace"} {
		if strings.Contains(lower, marker) {
			return "[redacted]"
		}
	}
	if strings.Contains(value, "/Users/") || strings.Contains(value, "/home/") || strings.Contains(value, "\\Users\\") || strings.Contains(value, ":\\") {
		return "[redacted]"
	}
	return value
}
