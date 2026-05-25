package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type startRunRequest struct {
	ScriptName string                `json:"script_name"`
	MessageID  string                `json:"message_id"`
	Source     productdata.RunSource `json:"source"`
	ProviderID string                `json:"provider_id"`
	Model      string                `json:"model"`
}

type modelProviderListResponse struct {
	Providers []productruntime.ProviderCapability `json:"providers"`
	RequestID string                              `json:"request_id"`
}

type checkModelProviderRequest struct {
	ProviderID string `json:"provider_id"`
}

type saveModelProviderRequest struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type modelProviderSaveResponse struct {
	Provider  productruntime.ProviderCapability `json:"provider"`
	RequestID string                            `json:"request_id"`
}

type modelProviderCheckResponse struct {
	Provider  productruntime.ProviderCapability `json:"provider"`
	RequestID string                            `json:"request_id"`
}

type runResponse struct {
	Run       productdata.Run `json:"run"`
	RequestID string          `json:"request_id"`
}

type runEventListResponse struct {
	Events    []productdata.RunEvent `json:"events"`
	RequestID string                 `json:"request_id"`
}

type toolCallResponse struct {
	ToolCall  productdata.ToolCall `json:"tool_call"`
	RequestID string               `json:"request_id"`
}

type stopRunResponse struct {
	Run       productdata.Run           `json:"run"`
	Result    productdata.StopRunResult `json:"result"`
	RequestID string                    `json:"request_id"`
}

func (s *Server) handleModelProviders(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, modelProviderListResponse{Providers: productruntime.ProviderCapabilities(s.providers), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		s.handleModelProviderSave(w, r)
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleModelProviderSave(w http.ResponseWriter, r *http.Request) {
	var req saveModelProviderRequest
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
		return
	}
	provider := productruntime.ProviderConfig{ID: "custom", Family: productruntime.ProviderFamilyOpenAICompatible, BaseURL: strings.TrimSpace(req.BaseURL), APIKey: strings.TrimSpace(req.APIKey), Model: strings.TrimSpace(req.Model), Enabled: true}
	capability := provider.Capability()
	if capability.Status == productruntime.ProviderStatusMisconfigured {
		writeAPIError(w, productdata.NewError(productdata.CodeProviderMisconfigured, capability.Message))
		return
	}
	provider = s.saveProviderConfig(provider)
	writeJSON(w, http.StatusOK, modelProviderSaveResponse{Provider: provider.Capability(), RequestID: diagnostics.NewRequestID()})
}

func (s *Server) saveProviderConfig(provider productruntime.ProviderConfig) productruntime.ProviderConfig {
	if s.gatewayRunner != nil {
		provider = s.gatewayRunner.SaveProviderConfig(provider)
	}
	for index, candidate := range s.providers {
		if candidate.ID == provider.ID {
			s.providers[index] = provider
			return provider
		}
	}
	s.providers = append(s.providers, provider)
	return provider
}

func (s *Server) handleModelProviderCheck(w http.ResponseWriter, r *http.Request) {
	var req checkModelProviderRequest
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
		return
	}
	for _, provider := range s.providers {
		if provider.ID != req.ProviderID {
			continue
		}
		capability := provider.Capability()
		if capability.Status == productruntime.ProviderStatusUnavailable {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, capability.Message))
			return
		}
		if capability.Status == productruntime.ProviderStatusMisconfigured {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderMisconfigured, capability.Message))
			return
		}
		writeJSON(w, http.StatusOK, modelProviderCheckResponse{Provider: capability, RequestID: diagnostics.NewRequestID()})
		return
	}
	writeAPIError(w, productdata.NewError(productdata.CodeProviderMisconfigured, "Provider is not configured."))
}

func (s *Server) handleThreadRuns(w http.ResponseWriter, r *http.Request, threadID string) {
	if !s.productAvailable(w) {
		return
	}
	if strings.HasSuffix(r.URL.Path, "/runs/current") {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, "GET")
			return
		}
		run, err := s.product.GetCurrentRun(r.Context(), identity.LocalDevIdentity(), threadID)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, runResponse{Run: run, RequestID: diagnostics.NewRequestID()})
		return
	}
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, "POST")
		return
	}
	var req startRunRequest
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
	}
	run, err := s.product.StartRun(r.Context(), identity.LocalDevIdentity(), threadID, productdata.StartRunInput{ScriptName: req.ScriptName, Source: req.Source, MessageID: req.MessageID, ProviderID: req.ProviderID, Model: req.Model})
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, runResponse{Run: run, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleRunByID(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	runID, suffix := splitRunPath(r.URL.Path)
	if runID == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
		return
	}
	switch suffix {
	case "":
		s.handleRun(w, r, runID)
	case "events":
		s.handleRunEvents(w, r, runID)
	case "events/stream":
		s.handleRunEventStream(w, r, runID)
	case "stop":
		s.handleStopRun(w, r, runID)
	default:
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
	}
}

func (s *Server) handleRun(w http.ResponseWriter, r *http.Request, runID string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	run, err := s.product.GetRun(r.Context(), identity.LocalDevIdentity(), runID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, runResponse{Run: run, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleThreadRunResource(w http.ResponseWriter, r *http.Request, threadID string, suffix string) {
	runID, rest := splitResourcePath(suffix, "")
	if runID == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
		return
	}
	if strings.HasPrefix(rest, "tool-calls/") {
		s.handleToolCall(w, r, threadID, runID, strings.TrimPrefix(rest, "tool-calls/"))
		return
	}
	writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
}

func (s *Server) handleToolCall(w http.ResponseWriter, r *http.Request, threadID string, runID string, toolCallID string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	if toolCallID == "" || strings.Contains(toolCallID, "/") {
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
		return
	}
	call, err := s.product.GetToolCall(r.Context(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toolCallResponse{ToolCall: call, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleRunEvents(w http.ResponseWriter, r *http.Request, runID string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	afterSequence, ok := parseAfterSequence(w, r)
	if !ok {
		return
	}
	events, err := s.product.ListRunEvents(r.Context(), identity.LocalDevIdentity(), runID, afterSequence)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, runEventListResponse{Events: events, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleRunEventStream(w http.ResponseWriter, r *http.Request, runID string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	afterSequence, ok := parseAfterSequence(w, r)
	if !ok {
		return
	}
	run, err := s.product.GetRun(r.Context(), identity.LocalDevIdentity(), runID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	var live <-chan productdata.RunEvent
	if s.broadcaster != nil && !productdata.IsRunTerminal(run.Status) {
		live = s.broadcaster.Subscribe(r.Context(), runID)
	}
	events, err := s.product.ListRunEvents(r.Context(), identity.LocalDevIdentity(), runID, afterSequence)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	sent := map[string]struct{}{}
	highestSequence := afterSequence
	terminalDelivered := productdata.IsRunTerminal(run.Status)
	for _, event := range events {
		writeSSEEvent(w, event)
		flushSSE(w)
		sent[event.ID] = struct{}{}
		if event.Sequence > highestSequence {
			highestSequence = event.Sequence
		}
		if productdata.IsRunTerminal(statusFromStreamEvent(event)) {
			terminalDelivered = true
		}
	}
	if terminalDelivered || live == nil {
		writeSSEClose(w, run.ID)
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case event, ok := <-live:
			if !ok {
				return
			}
			if event.Sequence <= highestSequence {
				continue
			}
			if _, ok := sent[event.ID]; ok {
				continue
			}
			writeSSEEvent(w, event)
			flushSSE(w)
			sent[event.ID] = struct{}{}
			highestSequence = event.Sequence
			if productdata.IsRunTerminal(statusFromStreamEvent(event)) {
				writeSSEClose(w, run.ID)
				return
			}
		}
	}
}

func (s *Server) handleStopRun(w http.ResponseWriter, r *http.Request, runID string) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, "POST")
		return
	}
	output, err := s.product.StopRun(r.Context(), identity.LocalDevIdentity(), runID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if s.broadcaster != nil {
		for _, event := range output.Events {
			s.broadcaster.Publish(event)
		}
	}
	writeJSON(w, http.StatusOK, stopRunResponse{Run: output.Run, Result: output.Result, RequestID: diagnostics.NewRequestID()})
}

func parseAfterSequence(w http.ResponseWriter, r *http.Request) (int, bool) {
	raw := r.URL.Query().Get("after_sequence")
	if raw == "" {
		return 0, true
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "after_sequence must be a non-negative integer."))
		return 0, false
	}
	return value, true
}

func statusFromStreamEvent(event productdata.RunEvent) productdata.RunStatus {
	if event.Category != productdata.RunEventCategoryFinal {
		return productdata.RunStatusRunning
	}
	switch event.Type {
	case "run_failed":
		return productdata.RunStatusFailed
	case "run_stopped":
		return productdata.RunStatusStopped
	default:
		return productdata.RunStatusCompleted
	}
}

func writeSSEClose(w http.ResponseWriter, runID string) {
	fmt.Fprintf(w, "event: stream_closed\ndata: {\"run_id\":%q,\"reason\":\"terminal\"}\n\n", runID)
	flushSSE(w)
}

func flushSSE(w http.ResponseWriter) {
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}
}

func writeSSEEvent(w http.ResponseWriter, event productdata.RunEvent) {
	raw, _ := json.Marshal(struct {
		Event productdata.RunEvent `json:"event"`
	}{Event: event})
	fmt.Fprintf(w, "id: %s\nevent: run_event\ndata: %s\n\n", event.ID, raw)
}

func splitRunPath(path string) (string, string) {
	return splitResourcePath(path, "/v1/runs/")
}
