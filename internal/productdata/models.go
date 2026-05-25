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

	CodeInvalidRequest        Code = "invalid_request"
	CodeThreadNotFound        Code = "thread_not_found"
	CodeRunNotFound           Code = "run_not_found"
	CodeActiveRunExists       Code = "active_run_exists"
	CodeProviderUnavailable   Code = "provider_unavailable"
	CodeProviderMisconfigured Code = "provider_misconfigured"
	CodeMethodNotAllowed      Code = "method_not_allowed"
	CodeMemoryNotFound        Code = "memory_not_found"
	CodeInternalError         Code = "internal_error"
)

const (
	MaxThreadTitleLength                       = 120
	MaxClientMessageIDLength                   = 120
	ToolNameCurrentTime                        = "runtime.get_current_time"
	ToolSourceInternal                         = "internal"
	ToolSourceMCP                              = "mcp"
	MemoryScopeUser          MemoryScopeType   = "user"
	MemoryScopeThread        MemoryScopeType   = "thread"
	MemoryEntryApproved      MemoryEntryStatus = "approved"
	MemoryEntryTombstoned    MemoryEntryStatus = "tombstoned"
	MemoryEntryDisabled      MemoryEntryStatus = "disabled"
	MemorySafetySafe         MemorySafetyState = "safe"
	MemorySafetyRedacted     MemorySafetyState = "redacted"
	MemorySafetyBlocked      MemorySafetyState = "blocked"
	MemoryWritePending       MemoryWriteStatus = "pending"
	MemoryWriteApproved      MemoryWriteStatus = "approved"
	MemoryWriteDenied        MemoryWriteStatus = "denied"
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
	Name           string
	ApprovalPolicy string
	ExecutionState string
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
	SourceThreadID   string          `json:"source_thread_id,omitempty"`
	SourceRunID      string          `json:"source_run_id,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	RankReason       string          `json:"rank_reason,omitempty"`
	RedactionApplied bool            `json:"redaction_applied"`
}

type MemorySearchInput struct {
	Query     string
	ScopeType MemoryScopeType
	ScopeID   string
	Limit     int
	Purpose   string
}

type MemorySearchOutput struct {
	Items         []MemorySearchResult `json:"items"`
	ExcludedCount int                  `json:"excluded_count"`
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
		summary["provider_id"] = c.ProviderRoute.ProviderID
	}
	if c.ProviderRoute.Model != "" {
		summary["model"] = c.ProviderRoute.Model
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
	for _, tool := range c.EnabledTools {
		names = append(names, tool.Name)
	}
	return RedactEventMetadata(map[string]any{
		"enabled_tool_count":          len(c.EnabledTools),
		"enabled_tools":               names,
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
	Reason string `json:"reason"`
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
	if input.ToolName != ToolNameCurrentTime && !IsMCPToolName(input.ToolName) {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool is not supported.")
	}
	if input.ApprovalStatus != ToolCallApprovalRequired || input.ExecutionStatus != ToolCallExecutionBlocked {
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
	if input.Metadata == nil {
		input.Metadata = map[string]any{}
	}
	input.ErrorCode = strings.TrimSpace(input.ErrorCode)
	input.ErrorMessage = RedactEventText(strings.TrimSpace(input.ErrorMessage))
	return input, nil
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
	for _, marker := range []string{"postgres://", "postgresql://", "password=", "api_key", "bearer ", "secret", "token", "credential", "authorization", "sk-", ".ssh", "id_ed25519", "id_rsa", ".env"} {
		if strings.Contains(lower, marker) {
			return "[redacted]"
		}
	}
	if strings.Contains(value, "/Users/") || strings.Contains(value, "\\Users\\") {
		return "[redacted]"
	}
	return value
}
