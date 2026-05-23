package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
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

type messageResponse struct {
	Message   productdata.Message `json:"message"`
	RequestID string              `json:"request_id"`
}

type messageListResponse struct {
	Messages  []productdata.Message `json:"messages"`
	RequestID string                `json:"request_id"`
}

type createThreadRequest struct {
	Title string                 `json:"title"`
	Mode  productdata.ThreadMode `json:"mode"`
}

type updateThreadRequest struct {
	Title *string                 `json:"title"`
	Mode  *productdata.ThreadMode `json:"mode"`
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
		writeJSON(w, http.StatusOK, threadListResponse{Threads: threads, RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req createThreadRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		thread, err := s.product.CreateThread(r.Context(), identity.LocalDevIdentity(), productdata.CreateThreadInput{Title: req.Title, Mode: req.Mode})
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
	if suffix == "runs" || suffix == "runs/current" {
		s.handleThreadRuns(w, r, threadID)
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
		if req.Title == nil && req.Mode == nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Thread update requires title or mode."))
			return
		}
		thread, err := s.product.UpdateThread(r.Context(), identity.LocalDevIdentity(), threadID, productdata.UpdateThreadInput{Title: req.Title, Mode: req.Mode})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, threadResponse{Thread: thread, RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, PATCH")
	}
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
		writeJSON(w, http.StatusOK, messageListResponse{Messages: messages, RequestID: diagnostics.NewRequestID()})
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
		writeJSON(w, status, messageResponse{Message: message, RequestID: diagnostics.NewRequestID()})
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
