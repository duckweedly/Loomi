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

type Code string

const (
	ThreadModeChat ThreadMode = "chat"
	ThreadModeWork ThreadMode = "work"

	ThreadLifecycleActive   ThreadLifecycleStatus = "active"
	ThreadLifecycleArchived ThreadLifecycleStatus = "archived"

	MessageRoleUser MessageRole = "user"

	CodeInvalidRequest   Code = "invalid_request"
	CodeThreadNotFound   Code = "thread_not_found"
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

func NewThreadID() string  { return prefixedID("thr") }
func NewMessageID() string { return prefixedID("msg") }

func prefixedID(prefix string) string {
	buf := make([]byte, 6)
	if _, err := rand.Read(buf); err != nil {
		return fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().UnixNano(), hex.EncodeToString(buf))
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
