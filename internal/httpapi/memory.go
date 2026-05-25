package httpapi

import (
	"net/http"
	"strings"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type memoryListResponse struct {
	Items     []productdata.MemorySearchResult `json:"items"`
	RequestID string                           `json:"request_id"`
}

type memorySearchResponse struct {
	Items         []productdata.MemorySearchResult `json:"items"`
	ExcludedCount int                              `json:"excluded_count"`
	RequestID     string                           `json:"request_id"`
}

type memoryEntryResponse struct {
	Entry     productdata.MemoryEntry `json:"entry"`
	RequestID string                  `json:"request_id"`
}

type memoryProposalResponse struct {
	Proposal  productdata.MemoryWriteProposal `json:"proposal"`
	RequestID string                          `json:"request_id"`
}

type memoryDecisionResponse struct {
	Proposal  productdata.MemoryWriteProposal `json:"proposal"`
	Entry     productdata.MemoryEntry         `json:"entry,omitempty"`
	RequestID string                          `json:"request_id"`
}

type memoryDeleteResponse struct {
	EntryID   string `json:"entry_id"`
	Status    string `json:"status"`
	DeletedAt string `json:"deleted_at"`
	RequestID string `json:"request_id"`
}

type memorySearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	if s.product == nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Product service is unavailable."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		output, err := s.product.ListMemoryEntries(r.Context(), identity.LocalDevIdentity(), productdata.MemorySearchInput{Limit: 50})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryListResponse{Items: output.Items, RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req memorySearchRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
		output, err := s.product.SearchMemory(r.Context(), identity.LocalDevIdentity(), productdata.MemorySearchInput{Query: req.Query, Limit: req.Limit})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memorySearchResponse{Items: output.Items, ExcludedCount: output.ExcludedCount, RequestID: diagnostics.NewRequestID()})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
	}
}

func (s *Server) handleMemoryByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/memory/")
	if rest == "search" && r.Method == http.MethodPost {
		s.handleMemory(w, r)
		return
	}
	if rest == "write-proposals" && r.Method == http.MethodPost {
		s.handleMemoryProposal(w, r)
		return
	}
	if strings.HasPrefix(rest, "write-proposals/") {
		s.handleMemoryProposalDecision(w, r, strings.TrimPrefix(rest, "write-proposals/"))
		return
	}
	entryID := strings.Trim(rest, "/")
	if entryID == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeMemoryNotFound, "Memory not found."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		entry, err := s.product.GetMemoryEntry(r.Context(), identity.LocalDevIdentity(), entryID)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryEntryResponse{Entry: entry, RequestID: diagnostics.NewRequestID()})
	case http.MethodDelete:
		var req productdata.DeleteMemoryEntryInput
		if r.Body != nil && r.ContentLength != 0 {
			if err := decodeJSONRequest(r, &req); err != nil {
				writeAPIError(w, err)
				return
			}
		}
		tombstone, err := s.product.DeleteMemoryEntry(r.Context(), identity.LocalDevIdentity(), entryID, req)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryDeleteResponse{EntryID: tombstone.EntryID, Status: tombstone.Status, DeletedAt: tombstone.DeletedAt.Format("2006-01-02T15:04:05Z07:00"), RequestID: diagnostics.NewRequestID()})
	default:
		w.Header().Set("Allow", "GET, DELETE")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
	}
}

func (s *Server) handleMemoryProposal(w http.ResponseWriter, r *http.Request) {
	var req productdata.ProposeMemoryWriteInput
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, err)
		return
	}
	proposal, err := s.product.ProposeMemoryWrite(r.Context(), identity.LocalDevIdentity(), req)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, memoryProposalResponse{Proposal: proposal, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemoryProposalDecision(w http.ResponseWriter, r *http.Request, rest string) {
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) != 2 {
		writeAPIError(w, productdata.NewError(productdata.CodeMemoryNotFound, "Memory proposal not found."))
		return
	}
	var req productdata.MemoryWriteDecisionInput
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
	}
	var decision productdata.MemoryWriteDecision
	var err error
	switch {
	case r.Method == http.MethodPost && parts[1] == "approve":
		decision, err = s.product.ApproveMemoryWrite(r.Context(), identity.LocalDevIdentity(), parts[0], req)
	case r.Method == http.MethodPost && parts[1] == "deny":
		decision, err = s.product.DenyMemoryWrite(r.Context(), identity.LocalDevIdentity(), parts[0], req)
	default:
		w.Header().Set("Allow", "POST")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, memoryDecisionResponse{Proposal: decision.Proposal, Entry: decision.Entry, RequestID: diagnostics.NewRequestID()})
}
