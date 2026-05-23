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

type Code string

const (
	ThreadModeChat ThreadMode = "chat"
	ThreadModeWork ThreadMode = "work"

	ThreadLifecycleActive   ThreadLifecycleStatus = "active"
	ThreadLifecycleArchived ThreadLifecycleStatus = "archived"

	MessageRoleUser MessageRole = "user"

	RunStatusPending   RunStatus = "pending"
	RunStatusRunning   RunStatus = "running"
	RunStatusCompleted RunStatus = "completed"
	RunStatusFailed    RunStatus = "failed"
	RunStatusStopped   RunStatus = "stopped"

	RunSourceLocalSimulated RunSource = "local_simulated"

	RunEventCategoryLifecycle RunEventCategory = "lifecycle"
	RunEventCategoryProgress  RunEventCategory = "progress"
	RunEventCategoryMessage   RunEventCategory = "message"
	RunEventCategoryError     RunEventCategory = "error"
	RunEventCategoryFinal     RunEventCategory = "final"

	CodeInvalidRequest   Code = "invalid_request"
	CodeThreadNotFound   Code = "thread_not_found"
	CodeRunNotFound      Code = "run_not_found"
	CodeActiveRunExists  Code = "active_run_exists"
	CodeMethodNotAllowed Code = "method_not_allowed"
	CodeInternalError    Code = "internal_error"
)

const (
	MaxThreadTitleLength     = 120
	MaxClientMessageIDLength = 120
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
	ID           string     `json:"id"`
	ThreadID     string     `json:"thread_id"`
	UserID       string     `json:"-"`
	Status       RunStatus  `json:"status"`
	Source       RunSource  `json:"source"`
	Title        string     `json:"title"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorCode    *string    `json:"error_code,omitempty"`
	ErrorMessage *string    `json:"error_message,omitempty"`
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
}

type AppendRunEventInput struct {
	Category RunEventCategory
	Type     string
	Summary  string
	Content  *string
	Metadata map[string]any
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

func NewThreadID() string   { return prefixedID("thr") }
func NewMessageID() string  { return prefixedID("msg") }
func NewRunID() string      { return prefixedID("run") }
func NewRunEventID() string { return prefixedID("evt") }

func prefixedID(prefix string) string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(buf))
}

func ValidateRunStatus(status RunStatus) error {
	switch status {
	case RunStatusPending, RunStatusRunning, RunStatusCompleted, RunStatusFailed, RunStatusStopped:
		return nil
	default:
		return NewError(CodeInvalidRequest, "Run status is invalid.")
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
	case RunStatusPending, RunStatusRunning:
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
	return input, nil
}

func RedactEventMetadata(metadata map[string]any) map[string]any {
	if metadata == nil {
		return map[string]any{}
	}
	redacted := make(map[string]any, len(metadata))
	for key, value := range metadata {
		redacted[key] = redactMetadataValue(value)
	}
	return redacted
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
