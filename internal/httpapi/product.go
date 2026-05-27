package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type identityResponse struct {
	User      productdata.User `json:"user"`
	RequestID string           `json:"request_id"`
}

type threadResponse struct {
	Thread    productdata.Thread `json:"thread"`
	RequestID string             `json:"request_id"`
}

type threadListResponse struct {
	Threads   []productdata.Thread `json:"threads"`
	RequestID string               `json:"request_id"`
}

func nonNilThreads(threads []productdata.Thread) []productdata.Thread {
	if threads == nil {
		return []productdata.Thread{}
	}
	return threads
}

type personaListResponse struct {
	Personas  []productdata.Persona `json:"personas"`
	RequestID string                `json:"request_id"`
}

type skillListResponse struct {
	Skills    []productruntime.InstalledSkill `json:"skills"`
	RequestID string                          `json:"request_id"`
}

type messageResponse struct {
	Message   apiMessage `json:"message"`
	RequestID string     `json:"request_id"`
}

type messageListResponse struct {
	Messages  []apiMessage `json:"messages"`
	RequestID string       `json:"request_id"`
}

type apiMessage struct {
	ID                 string                  `json:"id"`
	ThreadID           string                  `json:"thread_id"`
	Role               productdata.MessageRole `json:"role"`
	Content            string                  `json:"content"`
	Metadata           map[string]any          `json:"metadata"`
	ClientMessageID    *string                 `json:"client_message_id,omitempty"`
	RunID              string                  `json:"run_id,omitempty"`
	AttemptOfMessageID string                  `json:"attempt_of_message_id,omitempty"`
	CreatedAt          time.Time               `json:"created_at"`
}

func newAPIMessage(message productdata.Message) apiMessage {
	return apiMessage{
		ID:                 message.ID,
		ThreadID:           message.ThreadID,
		Role:               message.Role,
		Content:            message.Content,
		Metadata:           message.Metadata,
		ClientMessageID:    message.ClientMessageID,
		RunID:              metadataString(message.Metadata, "run_id"),
		AttemptOfMessageID: metadataString(message.Metadata, "attempt_of_message_id"),
		CreatedAt:          message.CreatedAt,
	}
}

func newAPIMessages(messages []productdata.Message) []apiMessage {
	items := make([]apiMessage, 0, len(messages))
	for _, message := range messages {
		items = append(items, newAPIMessage(message))
	}
	return items
}

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

type createThreadRequest struct {
	Title     string                 `json:"title"`
	Mode      productdata.ThreadMode `json:"mode"`
	PersonaID string                 `json:"persona_id"`
}

type updateThreadRequest struct {
	Title     *string                 `json:"title"`
	Mode      *productdata.ThreadMode `json:"mode"`
	PersonaID *string                 `json:"persona_id"`
}

type createMessageRequest struct {
	Content         string `json:"content"`
	ClientMessageID string `json:"client_message_id"`
}

func (s *Server) handleCurrentIdentity(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	user, err := s.product.CurrentIdentity(r.Context(), identity.LocalDevIdentity())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, identityResponse{User: user, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleThreads(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		includeArchived := r.URL.Query().Get("include_archived") == "true"
		threads, err := s.product.ListThreads(r.Context(), identity.LocalDevIdentity(), includeArchived)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, threadListResponse{Threads: nonNilThreads(threads), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req createThreadRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		thread, err := s.product.CreateThread(r.Context(), identity.LocalDevIdentity(), productdata.CreateThreadInput{Title: req.Title, Mode: req.Mode, PersonaID: req.PersonaID})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, threadResponse{Thread: thread, RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleThreadByID(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	threadID, suffix := splitThreadPath(r.URL.Path)
	if threadID == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeThreadNotFound, "Thread not found."))
		return
	}
	if suffix == "messages" {
		s.handleThreadMessages(w, r, threadID)
		return
	}
	if suffix == "artifacts" || strings.HasPrefix(suffix, "artifacts/") {
		s.handleThreadArtifacts(w, r, threadID, suffix)
		return
	}
	if suffix == "agent-tasks" {
		s.handleThreadAgentTasks(w, r, threadID, suffix)
		return
	}
	if suffix == "runs" || suffix == "runs/current" {
		s.handleThreadRuns(w, r, threadID)
		return
	}
	if strings.HasPrefix(suffix, "runs/") {
		s.handleThreadRunResource(w, r, threadID, strings.TrimPrefix(suffix, "runs/"))
		return
	}
	if suffix == "archive" {
		s.handleArchiveThread(w, r, threadID)
		return
	}
	if suffix != "" {
		writeAPIError(w, productdata.NewError(productdata.CodeThreadNotFound, "Thread not found."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		thread, err := s.product.GetThread(r.Context(), identity.LocalDevIdentity(), threadID)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, threadResponse{Thread: thread, RequestID: diagnostics.NewRequestID()})
	case http.MethodPatch:
		var req updateThreadRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		if req.Title == nil && req.Mode == nil && req.PersonaID == nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Thread update requires title or mode."))
			return
		}
		thread, err := s.product.UpdateThread(r.Context(), identity.LocalDevIdentity(), threadID, productdata.UpdateThreadInput{Title: req.Title, Mode: req.Mode, PersonaID: req.PersonaID})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, threadResponse{Thread: thread, RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, PATCH")
	}
}

func (s *Server) handlePersonas(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	personas, err := s.product.ListPersonas(r.Context(), identity.LocalDevIdentity())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if personas == nil {
		personas = []productdata.Persona{}
	}
	writeJSON(w, http.StatusOK, personaListResponse{Personas: personas, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleSkills(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	skills, err := productruntime.DiscoverInstalledSkills(s.skillDiscoveryInput)
	if err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Skill discovery failed."))
		return
	}
	if skills == nil {
		skills = []productruntime.InstalledSkill{}
	}
	writeJSON(w, http.StatusOK, skillListResponse{Skills: skills, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleArchiveThread(w http.ResponseWriter, r *http.Request, threadID string) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, "POST")
		return
	}
	thread, err := s.product.ArchiveThread(r.Context(), identity.LocalDevIdentity(), threadID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, threadResponse{Thread: thread, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleThreadMessages(w http.ResponseWriter, r *http.Request, threadID string) {
	switch r.Method {
	case http.MethodGet:
		messages, err := s.product.ListMessages(r.Context(), identity.LocalDevIdentity(), threadID)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		if messages == nil {
			messages = []productdata.Message{}
		}
		writeJSON(w, http.StatusOK, messageListResponse{Messages: newAPIMessages(messages), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req createMessageRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		message, created, err := s.product.CreateMessage(r.Context(), identity.LocalDevIdentity(), threadID, productdata.CreateMessageInput{Content: req.Content, ClientMessageID: req.ClientMessageID})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		status := http.StatusCreated
		if !created {
			status = http.StatusOK
		}
		writeJSON(w, status, messageResponse{Message: newAPIMessage(message), RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func decodeJSONRequest(r *http.Request, v any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(v)
}

func writeMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Unsupported method."))
}

func (s *Server) productAvailable(w http.ResponseWriter) bool {
	if s.product != nil {
		return true
	}
	writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Product data is unavailable."))
	return false
}

func splitThreadPath(path string) (string, string) {
	return splitResourcePath(path, "/v1/threads/")
}
