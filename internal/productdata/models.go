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

type Code string

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
	PipelineStepInvokeRuntime  PipelineStepName = "invoke_runtime"
	PipelineStepFinalize       PipelineStepName = "finalize"
	PipelineStepRecover        PipelineStepName = "recover"
	PipelineStepFail           PipelineStepName = "fail"

	EventRunQueued                = "run_queued"
	EventJobClaimed               = "job_claimed"
	EventLeaseRenewed             = "lease_renewed"
	EventPipelineStepStarted      = "pipeline_step_started"
	EventPipelineStepCompleted    = "pipeline_step_completed"
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

	CodeInvalidRequest        Code = "invalid_request"
	CodeThreadNotFound        Code = "thread_not_found"
	CodeRunNotFound           Code = "run_not_found"
	CodeActiveRunExists       Code = "active_run_exists"
	CodeProviderUnavailable   Code = "provider_unavailable"
	CodeProviderMisconfigured Code = "provider_misconfigured"
	CodeMethodNotAllowed      Code = "method_not_allowed"
	CodeInternalError         Code = "internal_error"
	CodeConflict              Code = "conflict"
)

const (
	MaxThreadTitleLength         = 120
	MaxClientMessageIDLength     = 120
	ToolNameCurrentTime          = "runtime.get_current_time"
	ToolNameTodoWrite            = "runtime.todo_write"
	ToolNameMCPCallTool          = "mcp.call_tool"
	ToolNameWorkspaceGlob        = "workspace.glob"
	ToolNameWorkspaceGrep        = "workspace.grep"
	ToolNameWorkspaceReadFile    = "workspace.read_file"
	ToolNameWorkspaceWriteFile   = "workspace.write_file"
	ToolNameWorkspaceEdit        = "workspace.edit"
	ToolNameWorkspaceExecCommand = "workspace.exec_command"
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
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	StopRequestedAt *time.Time `json:"stop_requested_at,omitempty"`
	ErrorCode       *string    `json:"error_code,omitempty"`
	ErrorMessage    *string    `json:"error_message,omitempty"`
}

type ToolCall struct {
	ID               string                  `json:"id"`
	ThreadID         string                  `json:"thread_id"`
	RunID            string                  `json:"run_id"`
	ToolCallID       string                  `json:"tool_call_id"`
	ToolName         string                  `json:"tool_name"`
	ArgumentsSummary map[string]any          `json:"arguments_summary"`
	ApprovalStatus   ToolCallApprovalStatus  `json:"approval_status"`
	ExecutionStatus  ToolCallExecutionStatus `json:"execution_status"`
	ResultSummary    map[string]any          `json:"result_summary,omitempty"`
	ErrorCode        *string                 `json:"error_code,omitempty"`
	ErrorMessage     *string                 `json:"error_message,omitempty"`
	RequestedAt      time.Time               `json:"requested_at"`
	UpdatedAt        time.Time               `json:"updated_at"`
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
	Title string
	Mode  ThreadMode
}

type UpdateThreadInput struct {
	Title *string
	Mode  *ThreadMode
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
	ToolCallID       string
	ToolName         string
	ArgumentsSummary map[string]any
	ArgumentsHash    string
	ApprovalStatus   ToolCallApprovalStatus
	ExecutionStatus  ToolCallExecutionStatus
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

func NewThreadID() string        { return prefixedID("thr") }
func NewMessageID() string       { return prefixedID("msg") }
func NewRunID() string           { return prefixedID("run") }
func NewRunEventID() string      { return prefixedID("evt") }
func NewBackgroundJobID() string { return prefixedID("job") }
func NewToolCallID() string      { return prefixedID("tool") }

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
	input.ArgumentsHash = strings.TrimSpace(input.ArgumentsHash)
	if input.ToolCallID == "" || input.ToolName == "" {
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool call id and name are required.")
	}
	if !isSupportedToolName(input.ToolName) {
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
	return normalizeToolArguments(input)
}

func isSupportedToolName(name string) bool {
	switch name {
	case ToolNameCurrentTime, ToolNameTodoWrite, ToolNameMCPCallTool, ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceReadFile, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspaceExecCommand:
		return true
	default:
		return false
	}
}

func normalizeToolArguments(input RecordToolCallRequestInput) (RecordToolCallRequestInput, error) {
	switch input.ToolName {
	case ToolNameCurrentTime:
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
	case ToolNameTodoWrite:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "items"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		items, err := normalizeTodoWriteSummary(input.ArgumentsSummary["items"])
		if err != nil {
			return RecordToolCallRequestInput{}, err
		}
		input.ArgumentsSummary["items"] = items
		return input, nil
	case ToolNameMCPCallTool:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "server", "tool", "arguments"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		arguments, err := normalizeMCPCallToolSummary(input.ArgumentsSummary)
		if err != nil {
			return RecordToolCallRequestInput{}, err
		}
		input.ArgumentsSummary["server"] = "local"
		input.ArgumentsSummary["tool"] = "echo"
		input.ArgumentsSummary["arguments"] = arguments
		return input, nil
	case ToolNameWorkspaceGlob:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "pattern", "limit"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		pattern, ok := input.ArgumentsSummary["pattern"].(string)
		if !ok || strings.TrimSpace(pattern) == "" || sensitiveWorkspacePath(pattern) || workspacePathEscapes(pattern) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace glob pattern is invalid.")
		}
		input.ArgumentsSummary["pattern"] = strings.TrimSpace(pattern)
		return input, nil
	case ToolNameWorkspaceGrep:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "query", "path", "limit"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		query, ok := input.ArgumentsSummary["query"].(string)
		if !ok || strings.TrimSpace(query) == "" {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace grep query is invalid.")
		}
		input.ArgumentsSummary["query"] = strings.TrimSpace(query)
		if value, ok := input.ArgumentsSummary["path"]; ok && value != nil {
			pathValue, ok := value.(string)
			if !ok || sensitiveWorkspacePath(pathValue) || workspacePathEscapes(pathValue) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace grep path is invalid.")
			}
			input.ArgumentsSummary["path"] = strings.TrimSpace(pathValue)
		}
		return input, nil
	case ToolNameWorkspaceReadFile:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "path", "max_bytes"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		pathValue, ok := input.ArgumentsSummary["path"].(string)
		if !ok || strings.TrimSpace(pathValue) == "" || sensitiveWorkspacePath(pathValue) || workspacePathEscapes(pathValue) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace read path is invalid.")
		}
		input.ArgumentsSummary["path"] = strings.TrimSpace(pathValue)
		return input, nil
	case ToolNameWorkspaceWriteFile:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "path", "content"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		pathValue, ok := input.ArgumentsSummary["path"].(string)
		if !ok || strings.TrimSpace(pathValue) == "" || sensitiveWorkspacePath(pathValue) || workspacePathEscapes(pathValue) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace write path is invalid.")
		}
		content, ok := input.ArgumentsSummary["content"].(string)
		if !ok || len(content) > 65536 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace write content is invalid.")
		}
		input.ArgumentsSummary["path"] = strings.TrimSpace(pathValue)
		return input, nil
	case ToolNameWorkspaceEdit:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "path", "old_text", "new_text"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		pathValue, ok := input.ArgumentsSummary["path"].(string)
		if !ok || strings.TrimSpace(pathValue) == "" || sensitiveWorkspacePath(pathValue) || workspacePathEscapes(pathValue) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace edit path is invalid.")
		}
		oldText, ok := input.ArgumentsSummary["old_text"].(string)
		if !ok || oldText == "" || len(oldText) > 65536 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace edit old_text is invalid.")
		}
		newText, ok := input.ArgumentsSummary["new_text"].(string)
		if !ok || len(newText) > 65536 {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace edit new_text is invalid.")
		}
		input.ArgumentsSummary["path"] = strings.TrimSpace(pathValue)
		return input, nil
	case ToolNameWorkspaceExecCommand:
		if err := validateAllowedArgumentKeys(input.ArgumentsSummary, "command", "cwd", "timeout_seconds"); err != nil {
			return RecordToolCallRequestInput{}, err
		}
		command, err := normalizeCommandSummary(input.ArgumentsSummary["command"])
		if err != nil || dangerousCommandSummary(command) {
			return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace command is invalid.")
		}
		input.ArgumentsSummary["command"] = command
		if value, ok := input.ArgumentsSummary["cwd"]; ok && value != nil {
			cwd, ok := value.(string)
			if !ok || strings.TrimSpace(cwd) == "" || sensitiveWorkspacePath(cwd) || workspacePathEscapes(cwd) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace command cwd is invalid.")
			}
			input.ArgumentsSummary["cwd"] = strings.TrimSpace(cwd)
		}
		if value, ok := input.ArgumentsSummary["timeout_seconds"]; ok && value != nil {
			if !validBoundedNumber(value, 1, 120) {
				return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Workspace command timeout is invalid.")
			}
		}
		return input, nil
	default:
		return RecordToolCallRequestInput{}, NewError(CodeInvalidRequest, "Tool is not supported.")
	}
}

func validateAllowedArgumentKeys(arguments map[string]any, allowed ...string) error {
	allowedSet := map[string]struct{}{}
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range arguments {
		if _, ok := allowedSet[key]; !ok {
			return NewError(CodeInvalidRequest, "Tool call argument is not supported.")
		}
	}
	return nil
}

func normalizeCommandSummary(value any) ([]any, error) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 || len(items) > 32 {
		return nil, NewError(CodeInvalidRequest, "Workspace command argv is invalid.")
	}
	for _, item := range items {
		part, ok := item.(string)
		if !ok || strings.TrimSpace(part) == "" || len(part) > 65536 {
			return nil, NewError(CodeInvalidRequest, "Workspace command argv is invalid.")
		}
	}
	return items, nil
}

func normalizeTodoWriteSummary(value any) ([]any, error) {
	items, ok := value.([]any)
	if !ok || len(items) == 0 || len(items) > 20 {
		return nil, NewError(CodeInvalidRequest, "Todo items are invalid.")
	}
	normalized := make([]any, 0, len(items))
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			return nil, NewError(CodeInvalidRequest, "Todo item is invalid.")
		}
		title, ok := item["title"].(string)
		title = strings.TrimSpace(title)
		if !ok || title == "" || len(title) > 160 {
			return nil, NewError(CodeInvalidRequest, "Todo item title is invalid.")
		}
		status := "pending"
		if rawStatus, ok := item["status"]; ok && rawStatus != nil {
			statusValue, ok := rawStatus.(string)
			if !ok {
				return nil, NewError(CodeInvalidRequest, "Todo item status is invalid.")
			}
			status = strings.TrimSpace(statusValue)
		}
		if status != "pending" && status != "in_progress" && status != "completed" {
			return nil, NewError(CodeInvalidRequest, "Todo item status is invalid.")
		}
		normalized = append(normalized, map[string]any{"title": title, "status": status})
	}
	return normalized, nil
}

func normalizeMCPCallToolSummary(arguments map[string]any) (map[string]any, error) {
	server, ok := arguments["server"].(string)
	if !ok || strings.TrimSpace(server) != "local" {
		return nil, NewError(CodeInvalidRequest, "MCP server is invalid.")
	}
	tool, ok := arguments["tool"].(string)
	if !ok || strings.TrimSpace(tool) != "echo" {
		return nil, NewError(CodeInvalidRequest, "MCP tool is invalid.")
	}
	rawArguments, ok := arguments["arguments"].(map[string]any)
	if !ok {
		return nil, NewError(CodeInvalidRequest, "MCP arguments are invalid.")
	}
	if err := validateAllowedArgumentKeys(rawArguments, "message"); err != nil {
		return nil, err
	}
	message, ok := rawArguments["message"].(string)
	message = strings.TrimSpace(message)
	if !ok || message == "" || len(message) > 500 || secretLookingText(message) {
		return nil, NewError(CodeInvalidRequest, "MCP message is invalid.")
	}
	return map[string]any{"message": message}, nil
}

func secretLookingText(value string) bool {
	lower := strings.ToLower(value)
	return RedactEventText(value) == "[redacted]" || strings.Contains(lower, "sk-")
}

func dangerousCommandSummary(command []any) bool {
	if len(command) == 0 {
		return true
	}
	first, _ := command[0].(string)
	base := strings.ToLower(strings.TrimSpace(first))
	if slash := strings.LastIndexAny(base, "/\\"); slash >= 0 {
		base = base[slash+1:]
	}
	switch base {
	case "sh", "bash", "zsh", "fish", "rm", "dd", "mkfs", "chmod", "chown", "kill", "killall", "shutdown", "reboot", "sudo", "su":
		return true
	case "git":
		if len(command) > 1 {
			sub, _ := command[1].(string)
			sub = strings.ToLower(strings.TrimSpace(sub))
			return sub == "push" || sub == "reset" || sub == "clean" || sub == "checkout"
		}
	}
	return false
}

func validBoundedNumber(value any, min int, max int) bool {
	var number int
	switch typed := value.(type) {
	case int:
		number = typed
	case int64:
		number = int(typed)
	case float64:
		number = int(typed)
	default:
		return false
	}
	return number >= min && number <= max
}

func workspacePathEscapes(value string) bool {
	value = strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	return value == "" || strings.HasPrefix(value, "/") || strings.Contains(value, "../") || strings.HasPrefix(value, "..")
}

func sensitiveWorkspacePath(value string) bool {
	value = strings.ToLower(strings.ReplaceAll(value, "\\", "/"))
	for _, segment := range strings.Split(value, "/") {
		if segment == "" {
			continue
		}
		if segment == ".ssh" || segment == ".aws" || segment == "secrets" || segment == "credentials" || strings.HasPrefix(segment, ".env") || strings.HasPrefix(segment, "id_rsa") || strings.HasPrefix(segment, "id_ed25519") || strings.HasSuffix(segment, ".pem") {
			return true
		}
	}
	return false
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

func IsToolCallTerminal(status ToolCallExecutionStatus) bool {
	switch status {
	case ToolCallExecutionSucceeded, ToolCallExecutionFailed, ToolCallExecutionCancelled:
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
	for _, marker := range []string{"postgres://", "postgresql://", "password=", "api_key", "bearer ", "secret", "token"} {
		if strings.Contains(lower, marker) {
			return "[redacted]"
		}
	}
	return value
}
