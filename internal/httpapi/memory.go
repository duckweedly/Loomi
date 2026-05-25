package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type memoryListResponse struct {
	Items     []productdata.MemorySearchResult `json:"items"`
	Filters   memoryAppliedFilters             `json:"filters"`
	RequestID string                           `json:"request_id"`
}

type memorySearchResponse struct {
	Items         []productdata.MemorySearchResult `json:"items"`
	ExcludedCount int                              `json:"excluded_count"`
	RequestID     string                           `json:"request_id"`
}

type memoryEntryResponse struct {
	Entry     productdata.MemorySearchResult `json:"entry"`
	RequestID string                         `json:"request_id"`
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
	Query             string                      `json:"query"`
	ScopeType         productdata.MemoryScopeType `json:"scope_type"`
	ScopeID           string                      `json:"scope_id"`
	SourceThreadID    string                      `json:"source_thread_id"`
	SourceRunID       string                      `json:"source_run_id"`
	SourceType        string                      `json:"source_type"`
	IncludeTombstoned bool                        `json:"include_tombstoned"`
	Limit             int                         `json:"limit"`
}

type memoryAppliedFilters struct {
	Query             string                      `json:"query,omitempty"`
	ScopeType         productdata.MemoryScopeType `json:"scope_type,omitempty"`
	ScopeID           string                      `json:"scope_id,omitempty"`
	SourceThreadID    string                      `json:"source_thread_id,omitempty"`
	SourceRunID       string                      `json:"source_run_id,omitempty"`
	SourceType        string                      `json:"source_type,omitempty"`
	IncludeTombstoned bool                        `json:"include_tombstoned,omitempty"`
}

type memoryAuditResponse struct {
	Items     []productdata.MemoryAuditItem `json:"items"`
	RequestID string                        `json:"request_id"`
}

func (s *Server) handleMemory(w http.ResponseWriter, r *http.Request) {
	if s.product == nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Product service is unavailable."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		input := memorySearchInputFromQuery(r)
		if err := validateMemorySearchInput(input); err != nil {
			writeAPIError(w, err)
			return
		}
		output, err := s.product.ListMemoryEntries(r.Context(), identity.LocalDevIdentity(), input)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		output.Items = memoryItems(output.Items)
		writeJSON(w, http.StatusOK, memoryListResponse{Items: output.Items, Filters: memoryFilters(input), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req memorySearchRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
		input := productdata.MemorySearchInput{Query: req.Query, ScopeType: req.ScopeType, ScopeID: req.ScopeID, SourceThreadID: req.SourceThreadID, SourceRunID: req.SourceRunID, SourceType: req.SourceType, IncludeTombstoned: req.IncludeTombstoned, Limit: req.Limit}
		if err := validateMemorySearchInput(input); err != nil {
			writeAPIError(w, err)
			return
		}
		output, err := s.product.SearchMemory(r.Context(), identity.LocalDevIdentity(), input)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		output.Items = memoryItems(output.Items)
		writeJSON(w, http.StatusOK, memorySearchResponse{Items: output.Items, ExcludedCount: output.ExcludedCount, RequestID: diagnostics.NewRequestID()})
	default:
		w.Header().Set("Allow", "GET, POST")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
	}
}

func validateMemorySearchInput(input productdata.MemorySearchInput) error {
	if input.ScopeType == productdata.MemoryScopeThread && strings.TrimSpace(input.ScopeID) == "" {
		return productdata.NewError(productdata.CodeInvalidRequest, "Thread memory scope id is required.")
	}
	return nil
}

func memorySearchInputFromQuery(r *http.Request) productdata.MemorySearchInput {
	values := r.URL.Query()
	limit, _ := strconv.Atoi(values.Get("limit"))
	includeTombstoned := values.Get("include_tombstoned") == "true"
	return productdata.MemorySearchInput{
		Query:             values.Get("q"),
		ScopeType:         productdata.MemoryScopeType(values.Get("scope_type")),
		ScopeID:           values.Get("scope_id"),
		SourceThreadID:    values.Get("source_thread_id"),
		SourceRunID:       values.Get("source_run_id"),
		SourceType:        values.Get("source_type"),
		IncludeTombstoned: includeTombstoned,
		Limit:             limit,
	}
}

func memoryFilters(input productdata.MemorySearchInput) memoryAppliedFilters {
	return memoryAppliedFilters{Query: input.Query, ScopeType: input.ScopeType, ScopeID: input.ScopeID, SourceThreadID: input.SourceThreadID, SourceRunID: input.SourceRunID, SourceType: input.SourceType, IncludeTombstoned: input.IncludeTombstoned}
}

func memoryItems(items []productdata.MemorySearchResult) []productdata.MemorySearchResult {
	if items == nil {
		return []productdata.MemorySearchResult{}
	}
	return items
}

func (s *Server) handleMemoryByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/memory/")
	if rest == "search" && r.Method == http.MethodPost {
		s.handleMemory(w, r)
		return
	}
	if rest == "audit" && r.Method == http.MethodGet {
		s.handleMemoryAudit(w, r)
		return
	}
	if rest == "entries" {
		s.handleMemory(w, r)
		return
	}
	if strings.HasPrefix(rest, "entries/") {
		rest = strings.TrimPrefix(rest, "entries/")
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
		entry, err := s.product.GetMemoryEntry(r.Context(), identity.LocalDevIdentity(), entryID, memoryAccessInputFromQuery(r))
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryEntryResponse{Entry: productdata.MemorySearchResult{
			ID:               entry.ID,
			Title:            entry.Title,
			Summary:          entry.Summary,
			ScopeType:        entry.ScopeType,
			ScopeID:          entry.ScopeID,
			Status:           string(entry.Status),
			SafetyState:      string(entry.SafetyState),
			SourceThreadID:   entry.SourceThreadID,
			SourceRunID:      entry.SourceRunID,
			SourceEventID:    entry.SourceEventID,
			SourceType:       memorySourceType(entry.SourceThreadID, entry.SourceRunID),
			CreatedAt:        entry.CreatedAt,
			UpdatedAt:        entry.UpdatedAt,
			DeletedAt:        entry.DeletedAt,
			RedactionApplied: entry.SafetyState != productdata.MemorySafetySafe,
		}, RequestID: diagnostics.NewRequestID()})
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

func memoryAccessInputFromQuery(r *http.Request) productdata.MemoryEntryAccessInput {
	values := r.URL.Query()
	return productdata.MemoryEntryAccessInput{
		ScopeType:      productdata.MemoryScopeType(values.Get("scope_type")),
		ScopeID:        values.Get("scope_id"),
		SourceThreadID: values.Get("source_thread_id"),
		SourceRunID:    values.Get("source_run_id"),
	}
}

func memorySourceType(sourceThreadID string, sourceRunID string) string {
	if strings.TrimSpace(sourceRunID) != "" {
		return "run"
	}
	if strings.TrimSpace(sourceThreadID) != "" {
		return "thread"
	}
	return "manual"
}

func (s *Server) handleMemoryAudit(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	limit, _ := strconv.Atoi(values.Get("limit"))
	threadID := strings.TrimSpace(values.Get("thread_id"))
	if threadID == "" && productdata.MemoryScopeType(values.Get("scope_type")) == productdata.MemoryScopeThread {
		threadID = strings.TrimSpace(values.Get("scope_id"))
	}
	if threadID == "" {
		threadID = strings.TrimSpace(values.Get("source_thread_id"))
	}
	output, err := s.product.ListMemoryAudit(r.Context(), identity.LocalDevIdentity(), productdata.MemoryAuditInput{
		ThreadID:    threadID,
		SourceRunID: values.Get("source_run_id"),
		EventType:   values.Get("event_type"),
		Limit:       limit,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if output.Items == nil {
		output.Items = []productdata.MemoryAuditItem{}
	}
	writeJSON(w, http.StatusOK, memoryAuditResponse{Items: output.Items, RequestID: diagnostics.NewRequestID()})
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
