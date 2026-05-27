package httpapi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

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

type memoryCreateEntryRequest struct {
	ScopeType      productdata.MemoryScopeType `json:"scope_type"`
	ScopeID        string                      `json:"scope_id"`
	Title          string                      `json:"title"`
	Content        string                      `json:"content"`
	SourceThreadID string                      `json:"source_thread_id"`
	SourceRunID    string                      `json:"source_run_id"`
	SourceEventID  string                      `json:"source_event_id"`
}

type memoryProposalResponse struct {
	Proposal  productdata.MemoryWriteProposal `json:"proposal"`
	RequestID string                          `json:"request_id"`
}

type memoryProposalListResponse struct {
	Items     []productdata.MemoryWriteProposal `json:"items"`
	RequestID string                            `json:"request_id"`
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

type memoryProviderResponse struct {
	Status    productdata.MemoryProviderStatus `json:"status"`
	RequestID string                           `json:"request_id"`
}

type memorySnapshotResponse struct {
	Snapshot  productdata.MemoryOverviewSnapshot `json:"snapshot"`
	RequestID string                             `json:"request_id"`
}

type memoryImpressionResponse struct {
	Impression productdata.MemoryImpressionSnapshot `json:"impression"`
	RequestID  string                               `json:"request_id"`
}

type memoryContentResponse struct {
	Content   string `json:"content"`
	Layer     string `json:"layer"`
	URI       string `json:"uri"`
	RequestID string `json:"request_id"`
}

type memoryErrorEvent struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Provider  string `json:"provider"`
	State     string `json:"state"`
	CheckedAt string `json:"checked_at,omitempty"`
	RunID     string `json:"run_id,omitempty"`
	EventType string `json:"event_type,omitempty"`
}

type memoryErrorsResponse struct {
	Errors    []memoryErrorEvent `json:"errors"`
	RequestID string             `json:"request_id"`
}

type memoryProviderDetectResponse struct {
	Detected  bool   `json:"detected"`
	BaseURL   string `json:"base_url,omitempty"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

type memoryProviderRequest struct {
	Enabled          bool                         `json:"enabled"`
	Provider         productdata.MemoryProviderID `json:"provider"`
	CommitAfterRun   bool                         `json:"commit_after_run"`
	SemanticEndpoint string                       `json:"semantic_endpoint"`
	OpenViking       memoryOpenVikingRequest      `json:"openviking"`
	Nowledge         memoryNowledgeRequest        `json:"nowledge"`
}

type memoryOpenVikingRequest struct {
	BaseURL            string `json:"base_url"`
	RootAPIKey         string `json:"root_api_key"`
	EmbeddingSelector  string `json:"embedding_selector"`
	EmbeddingProvider  string `json:"embedding_provider"`
	EmbeddingModel     string `json:"embedding_model"`
	EmbeddingAPIKey    string `json:"embedding_api_key"`
	EmbeddingAPIBase   string `json:"embedding_api_base"`
	EmbeddingDimension int    `json:"embedding_dimension"`
	VLMSelector        string `json:"vlm_selector"`
	VLMProvider        string `json:"vlm_provider"`
	VLMModel           string `json:"vlm_model"`
	VLMAPIKey          string `json:"vlm_api_key"`
	VLMAPIBase         string `json:"vlm_api_base"`
	RerankSelector     string `json:"rerank_selector"`
	RerankProvider     string `json:"rerank_provider"`
	RerankModel        string `json:"rerank_model"`
	RerankAPIKey       string `json:"rerank_api_key"`
	RerankAPIBase      string `json:"rerank_api_base"`
}

type memoryNowledgeRequest struct {
	BaseURL          string `json:"base_url"`
	APIKey           string `json:"api_key"`
	RequestTimeoutMS int    `json:"request_timeout_ms"`
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

func (s *Server) handleMemoryEntries(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		s.handleMemory(w, r)
		return
	}
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "GET, POST")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	var req memoryCreateEntryRequest
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, err)
		return
	}
	entry, err := s.product.CreateMemoryEntry(r.Context(), identity.LocalDevIdentity(), productdata.CreateMemoryEntryInput{
		ScopeType:      req.ScopeType,
		ScopeID:        req.ScopeID,
		Title:          req.Title,
		Content:        req.Content,
		SourceThreadID: req.SourceThreadID,
		SourceRunID:    req.SourceRunID,
		SourceEventID:  req.SourceEventID,
	})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, memoryEntryResponse{Entry: productdata.MemorySearchResult{
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
		RedactionApplied: entry.SafetyState != productdata.MemorySafetySafe,
	}, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemoryByID(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/v1/memory/")
	if rest == "provider" {
		s.handleMemoryProvider(w, r)
		return
	}
	if rest == "provider/nowledge/detect" {
		s.handleNowledgeProviderDetect(w, r)
		return
	}
	if rest == "provider/openviking/detect" {
		s.handleOpenVikingProviderDetect(w, r)
		return
	}
	if rest == "snapshot" || rest == "snapshot/rebuild" {
		s.handleMemorySnapshot(w, r, rest == "snapshot/rebuild")
		return
	}
	if rest == "impression" || rest == "impression/rebuild" {
		s.handleMemoryImpression(w, r, rest == "impression/rebuild")
		return
	}
	if rest == "content" {
		s.handleMemoryContent(w, r)
		return
	}
	if rest == "errors" {
		s.handleMemoryErrors(w, r)
		return
	}
	if rest == "search" && r.Method == http.MethodPost {
		s.handleMemory(w, r)
		return
	}
	if rest == "audit" && r.Method == http.MethodGet {
		s.handleMemoryAudit(w, r)
		return
	}
	if rest == "entries" {
		s.handleMemoryEntries(w, r)
		return
	}
	if strings.HasPrefix(rest, "entries/") {
		rest = strings.TrimPrefix(rest, "entries/")
	}
	if rest == "write-proposals" && (r.Method == http.MethodGet || r.Method == http.MethodPost) {
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

func (s *Server) handleMemoryContent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	uri := strings.TrimSpace(r.URL.Query().Get("uri"))
	layer := strings.TrimSpace(r.URL.Query().Get("layer"))
	if layer == "" {
		layer = "overview"
	}
	if uri == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "uri is required."))
		return
	}
	if layer != "overview" && layer != "read" {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "layer must be overview or read."))
		return
	}
	entryID, ok := strings.CutPrefix(uri, "memory://")
	if !ok || strings.TrimSpace(entryID) == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "unsupported memory uri."))
		return
	}
	entry, err := s.product.GetMemoryEntry(r.Context(), identity.LocalDevIdentity(), entryID, memoryAccessInputFromQuery(r))
	if err != nil {
		writeAPIError(w, err)
		return
	}
	content := strings.TrimSpace(entry.Summary)
	if layer == "read" && strings.TrimSpace(entry.Title) != "" {
		if content != "" {
			content = strings.TrimSpace(entry.Title) + "\n\n" + content
		} else {
			content = strings.TrimSpace(entry.Title)
		}
	}
	writeJSON(w, http.StatusOK, memoryContentResponse{Content: content, Layer: layer, URI: uri, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemoryErrors(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	store, ok := s.product.(productdata.ModelProviderConfigStore)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Memory provider configuration is unavailable."))
		return
	}
	errors, err := store.ListMemoryProviderErrors(r.Context(), identity.LocalDevIdentity(), 10)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	responseErrors := make([]memoryErrorEvent, 0, len(errors))
	for _, item := range errors {
		checkedAt := ""
		if !item.CheckedAt.IsZero() {
			checkedAt = item.CheckedAt.Format("2006-01-02T15:04:05Z07:00")
		}
		responseErrors = append(responseErrors, memoryErrorEvent{Code: item.Code, Message: item.Message, Provider: string(item.Provider), State: string(item.State), CheckedAt: checkedAt, RunID: item.RunID, EventType: item.EventType})
	}
	writeJSON(w, http.StatusOK, memoryErrorsResponse{Errors: responseErrors, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemorySnapshot(w http.ResponseWriter, r *http.Request, rebuild bool) {
	store, ok := s.product.(productdata.MemorySnapshotService)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Memory snapshot is unavailable."))
		return
	}
	if (!rebuild && r.Method != http.MethodGet) || (rebuild && r.Method != http.MethodPost) {
		w.Header().Set("Allow", map[bool]string{false: "GET", true: "POST"}[rebuild])
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	var snapshot productdata.MemoryOverviewSnapshot
	var err error
	if rebuild {
		snapshot, err = store.RebuildMemoryOverviewSnapshot(r.Context(), identity.LocalDevIdentity())
	} else {
		snapshot, err = store.GetMemoryOverviewSnapshot(r.Context(), identity.LocalDevIdentity())
	}
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, memorySnapshotResponse{Snapshot: snapshot, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemoryImpression(w http.ResponseWriter, r *http.Request, rebuild bool) {
	store, ok := s.product.(productdata.MemorySnapshotService)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Memory impression is unavailable."))
		return
	}
	if (!rebuild && r.Method != http.MethodGet) || (rebuild && r.Method != http.MethodPost) {
		w.Header().Set("Allow", map[bool]string{false: "GET", true: "POST"}[rebuild])
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	var impression productdata.MemoryImpressionSnapshot
	var err error
	if rebuild {
		impression, err = store.RebuildMemoryImpressionSnapshot(r.Context(), identity.LocalDevIdentity())
	} else {
		impression, err = store.GetMemoryImpressionSnapshot(r.Context(), identity.LocalDevIdentity())
	}
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, memoryImpressionResponse{Impression: impression, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleMemoryProvider(w http.ResponseWriter, r *http.Request) {
	store, ok := s.product.(productdata.ModelProviderConfigStore)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Memory provider configuration is unavailable."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		status, err := store.GetMemoryProviderStatus(r.Context(), identity.LocalDevIdentity())
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryProviderResponse{Status: status, RequestID: diagnostics.NewRequestID()})
	case http.MethodPut:
		var req memoryProviderRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
		if _, err := store.SaveMemoryProviderConfig(r.Context(), identity.LocalDevIdentity(), productdata.MemoryProviderConfig{
			Enabled:          req.Enabled,
			Provider:         req.Provider,
			CommitAfterRun:   req.CommitAfterRun,
			SemanticEndpoint: req.SemanticEndpoint,
			OpenViking: productdata.OpenVikingMemoryConfig{
				BaseURL:            req.OpenViking.BaseURL,
				RootAPIKey:         req.OpenViking.RootAPIKey,
				EmbeddingSelector:  req.OpenViking.EmbeddingSelector,
				EmbeddingProvider:  req.OpenViking.EmbeddingProvider,
				EmbeddingModel:     req.OpenViking.EmbeddingModel,
				EmbeddingAPIKey:    req.OpenViking.EmbeddingAPIKey,
				EmbeddingAPIBase:   req.OpenViking.EmbeddingAPIBase,
				EmbeddingDimension: req.OpenViking.EmbeddingDimension,
				VLMSelector:        req.OpenViking.VLMSelector,
				VLMProvider:        req.OpenViking.VLMProvider,
				VLMModel:           req.OpenViking.VLMModel,
				VLMAPIKey:          req.OpenViking.VLMAPIKey,
				VLMAPIBase:         req.OpenViking.VLMAPIBase,
				RerankSelector:     req.OpenViking.RerankSelector,
				RerankProvider:     req.OpenViking.RerankProvider,
				RerankModel:        req.OpenViking.RerankModel,
				RerankAPIKey:       req.OpenViking.RerankAPIKey,
				RerankAPIBase:      req.OpenViking.RerankAPIBase,
			},
			Nowledge: productdata.NowledgeMemoryConfig{
				BaseURL:          req.Nowledge.BaseURL,
				APIKey:           req.Nowledge.APIKey,
				RequestTimeoutMS: req.Nowledge.RequestTimeoutMS,
			},
		}); err != nil {
			writeAPIError(w, err)
			return
		}
		status, err := store.GetMemoryProviderStatus(r.Context(), identity.LocalDevIdentity())
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryProviderResponse{Status: status, RequestID: diagnostics.NewRequestID()})
	default:
		w.Header().Set("Allow", "GET, PUT")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
	}
}

func (s *Server) handleNowledgeProviderDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	baseURL := "http://127.0.0.1:14242"
	ctx, cancel := context.WithTimeout(r.Context(), 1200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Nowledge detect request failed."))
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: false, Message: "Nowledge local instance was not detected.", RequestID: diagnostics.NewRequestID()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: true, BaseURL: baseURL, Message: "Nowledge local instance detected.", RequestID: diagnostics.NewRequestID()})
		return
	}
	writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: false, Message: "Nowledge local instance was not detected.", RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleOpenVikingProviderDetect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", "GET")
		writeAPIError(w, productdata.NewError(productdata.CodeMethodNotAllowed, "Method not allowed."))
		return
	}
	baseURL := "http://127.0.0.1:8282"
	ctx, cancel := context.WithTimeout(r.Context(), 1200*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/fs/ls?uri=viking%3A%2F%2Fmemory", nil)
	if err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "OpenViking detect request failed."))
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: false, Message: "OpenViking local instance was not detected.", RequestID: diagnostics.NewRequestID()})
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: true, BaseURL: baseURL, Message: "OpenViking local instance detected.", RequestID: diagnostics.NewRequestID()})
		return
	}
	writeJSON(w, http.StatusOK, memoryProviderDetectResponse{Detected: false, Message: "OpenViking local instance was not detected.", RequestID: diagnostics.NewRequestID()})
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
	if r.Method == http.MethodGet {
		values := r.URL.Query()
		limit, _ := strconv.Atoi(values.Get("limit"))
		status := productdata.MemoryWriteStatus(values.Get("status"))
		if status == "" {
			status = productdata.MemoryWritePending
		}
		output, err := s.product.ListMemoryWriteProposals(r.Context(), identity.LocalDevIdentity(), productdata.MemoryWriteProposalListInput{
			Status:      status,
			ScopeType:   productdata.MemoryScopeType(values.Get("scope_type")),
			ScopeID:     values.Get("scope_id"),
			SourceRunID: values.Get("source_run_id"),
			Limit:       limit,
		})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		if output.Items == nil {
			output.Items = []productdata.MemoryWriteProposal{}
		}
		writeJSON(w, http.StatusOK, memoryProposalListResponse{Items: output.Items, RequestID: diagnostics.NewRequestID()})
		return
	}
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
	if len(parts) == 1 && r.Method == http.MethodPatch {
		var req productdata.MemoryWriteProposalUpdateInput
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
		proposal, err := s.product.UpdateMemoryWriteProposal(r.Context(), identity.LocalDevIdentity(), parts[0], req)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, memoryProposalResponse{Proposal: proposal, RequestID: diagnostics.NewRequestID()})
		return
	}
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
