package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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
	PersonaID  string                `json:"persona_id"`
}

type modelProviderListResponse struct {
	Providers []productruntime.ProviderCapability `json:"providers"`
	RequestID string                              `json:"request_id"`
}

type localProviderDetectionResponse struct {
	Providers []productruntime.LocalProviderCapability `json:"providers"`
	RequestID string                                   `json:"request_id"`
}

type checkModelProviderRequest struct {
	ProviderID string `json:"provider_id"`
}

type saveModelProviderRequest struct {
	BaseURL string `json:"base_url"`
	Model   string `json:"model"`
	APIKey  string `json:"api_key"`
}

type webSearchConfigRequest struct {
	TavilyAPIKey string `json:"tavily_api_key"`
	BraveAPIKey  string `json:"brave_api_key"`
}

type webSearchConfig struct {
	HasTavilyKey bool `json:"has_tavily_key"`
	HasBraveKey  bool `json:"has_brave_key"`
	Enabled      bool `json:"enabled"`
}

type webSearchConfigResponse struct {
	Config    webSearchConfig `json:"config"`
	RequestID string          `json:"request_id"`
}

type workspaceRootConfig struct {
	Configured  bool   `json:"configured"`
	DisplayName string `json:"display_name"`
}

type workspaceRootResponse struct {
	Config    workspaceRootConfig `json:"config"`
	RequestID string              `json:"request_id"`
}

type workspaceRootRequest struct {
	Path string `json:"path"`
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
		providers := s.modelProviderCapabilities()
		writeJSON(w, http.StatusOK, modelProviderListResponse{Providers: providers, RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		s.handleModelProviderSave(w, r)
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleLocalProviderDetections(w http.ResponseWriter, r *http.Request) {
	input := s.localProviderDetectionInput
	if input.HomeDir == "" && input.CodexHome == "" && input.ClaudeConfigDir == "" && input.Env == nil && !input.Disabled {
		input = productruntime.LocalProviderDetectionInputFromProcess()
	}
	writeJSON(w, http.StatusOK, localProviderDetectionResponse{Providers: productruntime.DetectLocalProviders(input), RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleWebSearchConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, webSearchConfigResponse{Config: s.webSearchConfig(), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req webSearchConfigRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		tavily := strings.TrimSpace(req.TavilyAPIKey)
		brave := strings.TrimSpace(req.BraveAPIKey)
		if tavily == "" && brave == "" && !s.webSearchConfig().Enabled {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "At least one web search API key is required."))
			return
		}
		if store, ok := s.product.(productdata.ModelProviderConfigStore); ok {
			saved, err := store.SaveWebSearchConfig(r.Context(), identity.LocalDevIdentity(), productdata.WebSearchConfig{TavilyAPIKey: tavily, BraveAPIKey: brave})
			if err != nil {
				writeAPIError(w, err)
				return
			}
			tavily = saved.TavilyAPIKey
			brave = saved.BraveAPIKey
		}
		s.providerMu.Lock()
		if tavily != "" {
			s.cfg.TavilyAPIKey = tavily
			_ = os.Setenv("LOOMI_TAVILY_API_KEY", tavily)
		}
		if brave != "" {
			s.cfg.BraveSearchAPIKey = brave
			_ = os.Setenv("LOOMI_BRAVE_SEARCH_API_KEY", brave)
		}
		s.providerMu.Unlock()
		writeJSON(w, http.StatusOK, webSearchConfigResponse{Config: s.webSearchConfig(), RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleWorkspaceRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, workspaceRootResponse{Config: s.currentWorkspaceRootConfig(r.Context()), RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req workspaceRootRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
			return
		}
		real, err := resolveWorkspaceRootPath(req.Path)
		if err != nil {
			writeAPIError(w, err)
			return
		}
		if store, ok := s.product.(productdata.WorkspaceRootConfigStore); ok {
			if _, err := store.SaveWorkspaceRootConfig(r.Context(), identity.LocalDevIdentity(), productdata.WorkspaceRootConfig{Path: real}); err != nil {
				writeAPIError(w, err)
				return
			}
		}
		if err := os.Setenv("LOOMI_WORKSPACE_ROOT", real); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Workspace folder could not be saved."))
			return
		}
		writeJSON(w, http.StatusOK, workspaceRootResponse{Config: workspaceRootConfigFromPath(real, true), RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) currentWorkspaceRootConfig(ctx context.Context) workspaceRootConfig {
	root := ""
	if store, ok := s.product.(productdata.WorkspaceRootConfigStore); ok {
		if saved, err := store.GetWorkspaceRootConfig(ctx, identity.LocalDevIdentity()); err == nil {
			root = strings.TrimSpace(saved.Path)
		}
	}
	if root == "" {
		root = strings.TrimSpace(os.Getenv("LOOMI_WORKSPACE_ROOT"))
	}
	if root == "" {
		return workspaceRootConfig{Configured: false, DisplayName: "Home"}
	}
	if real, err := resolveWorkspaceRootPath(root); err == nil {
		_ = os.Setenv("LOOMI_WORKSPACE_ROOT", real)
		return workspaceRootConfigFromPath(real, true)
	}
	return workspaceRootConfigFromPath(root, false)
}

func (s *Server) applySavedWorkspaceRoot(ctx context.Context) {
	store, ok := s.product.(productdata.WorkspaceRootConfigStore)
	if !ok {
		return
	}
	saved, err := store.GetWorkspaceRootConfig(ctx, identity.LocalDevIdentity())
	if err != nil || strings.TrimSpace(saved.Path) == "" {
		return
	}
	real, err := resolveWorkspaceRootPath(saved.Path)
	if err != nil {
		return
	}
	_ = os.Setenv("LOOMI_WORKSPACE_ROOT", real)
}

func resolveWorkspaceRootPath(path string) (string, error) {
	root := strings.TrimSpace(path)
	if root == "" || !filepath.IsAbs(root) {
		return "", productdata.NewError(productdata.CodeInvalidRequest, "Workspace folder must be an absolute path.")
	}
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", productdata.NewError(productdata.CodeInvalidRequest, "Workspace folder is unavailable.")
	}
	info, err := os.Stat(real)
	if err != nil || !info.IsDir() {
		return "", productdata.NewError(productdata.CodeInvalidRequest, "Workspace folder is unavailable.")
	}
	return real, nil
}

func workspaceRootConfigFromPath(root string, configured bool) workspaceRootConfig {
	name := filepath.Base(root)
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = "Selected folder"
	}
	return workspaceRootConfig{Configured: configured, DisplayName: name}
}

func (s *Server) webSearchConfig() webSearchConfig {
	s.providerMu.RLock()
	tavily := strings.TrimSpace(s.cfg.TavilyAPIKey) != ""
	brave := strings.TrimSpace(s.cfg.BraveSearchAPIKey) != ""
	s.providerMu.RUnlock()
	if store, ok := s.product.(productdata.ModelProviderConfigStore); ok {
		if saved, err := store.GetWebSearchConfig(context.Background(), identity.LocalDevIdentity()); err == nil {
			tavily = tavily || strings.TrimSpace(saved.TavilyAPIKey) != ""
			brave = brave || strings.TrimSpace(saved.BraveAPIKey) != ""
		}
	}
	if !tavily {
		tavily = strings.TrimSpace(os.Getenv("LOOMI_TAVILY_API_KEY")) != ""
	}
	if !brave {
		brave = strings.TrimSpace(os.Getenv("LOOMI_BRAVE_SEARCH_API_KEY")) != ""
	}
	return webSearchConfig{HasTavilyKey: tavily, HasBraveKey: brave, Enabled: tavily || brave}
}

func (s *Server) handleLocalProviderDetectionByID(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/v1/local-provider-detections/" {
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, "GET")
			return
		}
		s.handleLocalProviderDetections(w, r)
		return
	}
	providerID, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/v1/local-provider-detections/"), "/enable")
	if !ok || strings.TrimSpace(providerID) == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Local provider action is invalid."))
		return
	}
	switch r.Method {
	case http.MethodPost:
		s.handleLocalProviderEnable(w, providerID)
	case http.MethodDelete:
		s.handleLocalProviderDisable(w, providerID)
	default:
		writeMethodNotAllowed(w, "POST, DELETE")
	}
}

func (s *Server) handleLocalProviderEnable(w http.ResponseWriter, providerID string) {
	if providerID != "local_codex" {
		writeAPIError(w, productdata.NewError(productdata.CodeProviderMisconfigured, "Local provider execution is unsupported."))
		return
	}
	input := s.localProviderDetectionInput
	if input.HomeDir == "" && input.CodexHome == "" && input.ClaudeConfigDir == "" && input.Env == nil && !input.Disabled {
		input = productruntime.LocalProviderDetectionInputFromProcess()
	}
	for _, provider := range productruntime.DetectLocalProviders(input) {
		if provider.ProviderID != providerID {
			continue
		}
		if provider.Status != productruntime.LocalProviderStatusAvailable {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local provider is not available."))
			return
		}
		if s.gatewayRunner == nil {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local Codex execution bridge is unavailable."))
			return
		}
		if _, err := productruntime.LoadLocalCodexCredentialSnapshot(input); err != nil {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local Codex login is unavailable."))
			return
		}
		s.gatewayRunner.SaveProvider(productruntime.NewLocalCodexProvider(input))
		s.providerMu.Lock()
		s.localProviderEnablements[providerID] = provider
		s.providerMu.Unlock()
		writeJSON(w, http.StatusOK, modelProviderSaveResponse{Provider: productruntime.LocalProviderRouteCapability(provider), RequestID: diagnostics.NewRequestID()})
		return
	}
	writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local provider is not available."))
}

func (s *Server) handleLocalProviderDisable(w http.ResponseWriter, providerID string) {
	s.providerMu.Lock()
	provider, ok := s.localProviderEnablements[providerID]
	if ok {
		delete(s.localProviderEnablements, providerID)
		s.providerMu.Unlock()
		if s.gatewayRunner != nil {
			s.gatewayRunner.RemoveProvider(providerID)
		}
		writeJSON(w, http.StatusOK, modelProviderSaveResponse{Provider: productruntime.LocalProviderRouteCapability(provider), RequestID: diagnostics.NewRequestID()})
		return
	}
	s.providerMu.Unlock()
	writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local provider is not enabled."))
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
	if store, ok := s.product.(productdata.ModelProviderConfigStore); ok {
		saved, err := store.SaveModelProviderConfig(r.Context(), identity.LocalDevIdentity(), productdata.ModelProviderConfig{ID: provider.ID, Family: string(provider.Family), BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model, Enabled: provider.Enabled})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		provider = providerConfigFromProduct(saved)
	}
	provider = s.saveProviderConfig(provider)
	writeJSON(w, http.StatusOK, modelProviderSaveResponse{Provider: provider.Capability(), RequestID: diagnostics.NewRequestID()})
}

func (s *Server) saveProviderConfig(provider productruntime.ProviderConfig) productruntime.ProviderConfig {
	if s.gatewayRunner != nil {
		provider = s.gatewayRunner.SaveProviderConfig(provider)
	}
	s.providerMu.Lock()
	defer s.providerMu.Unlock()
	for index, candidate := range s.providers {
		if candidate.ID == provider.ID {
			s.providers[index] = provider
			return provider
		}
	}
	s.providers = append(s.providers, provider)
	return provider
}

func providerConfigFromProduct(provider productdata.ModelProviderConfig) productruntime.ProviderConfig {
	return productruntime.ProviderConfig{ID: provider.ID, Family: productruntime.ProviderFamily(provider.Family), BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model, Enabled: provider.Enabled}
}

func (s *Server) handleModelProviderCheck(w http.ResponseWriter, r *http.Request) {
	var req checkModelProviderRequest
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
		return
	}
	provider, ok := s.findProviderConfig(req.ProviderID)
	if ok {
		capability := provider.Capability()
		if capability.Status == productruntime.ProviderStatusUnavailable {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, capability.Message))
			return
		}
		if capability.Status == productruntime.ProviderStatusMisconfigured {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderMisconfigured, capability.Message))
			return
		}
		writeJSON(w, http.StatusOK, modelProviderCheckResponse{Provider: productruntime.CheckProviderCompletion(r.Context(), provider, nil), RequestID: diagnostics.NewRequestID()})
		return
	}
	capability, ok := s.localProviderRouteCapability(req.ProviderID)
	if ok {
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

func (s *Server) modelProviderCapabilities() []productruntime.ProviderCapability {
	s.providerMu.RLock()
	defer s.providerMu.RUnlock()
	providers := productruntime.ProviderCapabilities(s.providers)
	for _, provider := range s.localProviderEnablements {
		providers = append(providers, productruntime.LocalProviderRouteCapability(provider))
	}
	return providers
}

func (s *Server) findProviderConfig(providerID string) (productruntime.ProviderConfig, bool) {
	s.providerMu.RLock()
	defer s.providerMu.RUnlock()
	for _, provider := range s.providers {
		if provider.ID == providerID {
			return provider, true
		}
	}
	return productruntime.ProviderConfig{}, false
}

func (s *Server) localProviderRouteCapability(providerID string) (productruntime.ProviderCapability, bool) {
	s.providerMu.RLock()
	defer s.providerMu.RUnlock()
	provider, ok := s.localProviderEnablements[providerID]
	if !ok {
		return productruntime.ProviderCapability{}, false
	}
	return productruntime.LocalProviderRouteCapability(provider), true
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
	if req.ProviderID != "" {
		capability, ok := s.localProviderRouteCapability(req.ProviderID)
		if req.ProviderID == "local_codex" && !ok {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, "Local Codex is not enabled for this session."))
			return
		}
		if ok && (capability.Status != productruntime.ProviderStatusAvailable || capability.ExecutionState != "supported") {
			writeAPIError(w, productdata.NewError(productdata.CodeProviderUnavailable, capability.Message))
			return
		}
	}
	run, err := s.product.StartRun(r.Context(), identity.LocalDevIdentity(), threadID, productdata.StartRunInput{ScriptName: req.ScriptName, Source: req.Source, MessageID: req.MessageID, ProviderID: req.ProviderID, Model: req.Model, PersonaID: req.PersonaID})
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
	toolCallID, action := splitResourcePath(toolCallID, "")
	if toolCallID == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
		return
	}
	if action != "" {
		s.handleToolCallDecision(w, r, threadID, runID, toolCallID, action)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	call, err := s.product.GetToolCall(r.Context(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toolCallResponse{ToolCall: call, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleToolCallDecision(w http.ResponseWriter, r *http.Request, threadID string, runID string, toolCallID string, action string) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, "POST")
		return
	}
	var (
		call productdata.ToolCall
		err  error
	)
	switch action {
	case "approve":
		call, _, err = s.product.ApproveToolCall(r.Context(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	case "deny":
		call, _, err = s.product.DenyToolCall(r.Context(), identity.LocalDevIdentity(), threadID, runID, toolCallID)
	default:
		writeAPIError(w, productdata.NewError(productdata.CodeRunNotFound, "Run not found."))
		return
	}
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
