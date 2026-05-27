package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type mcpServersResponse struct {
	Servers   []productruntime.MCPServerStatus `json:"servers"`
	RequestID string                           `json:"request_id"`
}

type mcpServerResponse struct {
	Server    productruntime.MCPServerStatus `json:"server"`
	RequestID string                         `json:"request_id"`
}

type saveMCPServerRequest struct {
	Slug        string            `json:"slug"`
	DisplayName string            `json:"display_name"`
	Enabled     bool              `json:"enabled"`
	Transport   string            `json:"transport"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	TimeoutMS   int               `json:"timeout_ms"`
}

func (s *Server) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		servers, err := s.mcpServerStatuses(r.Context())
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, mcpServersResponse{Servers: servers, RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		s.handleSaveMCPServer(w, r)
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}

func (s *Server) handleMCPServerBySlug(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/v1/mcp/servers/")
	if path == "" {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "MCP server slug is required."))
		return
	}
	if slug, ok := strings.CutSuffix(path, "/discover"); ok {
		if r.Method != http.MethodPost {
			writeMethodNotAllowed(w, "POST")
			return
		}
		s.handleDiscoverMCPServer(w, strings.Trim(slug, "/"))
		return
	}
	if strings.Contains(path, "/") {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "MCP server action is invalid."))
		return
	}
	switch r.Method {
	case http.MethodDelete:
		s.handleDeleteMCPServer(w, r, path)
	default:
		writeMethodNotAllowed(w, "DELETE")
	}
}

func (s *Server) handleSaveMCPServer(w http.ResponseWriter, r *http.Request) {
	var req saveMCPServerRequest
	if err := decodeJSONRequest(r, &req); err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "Invalid JSON request."))
		return
	}
	config := productruntime.MCPServerConfig{Slug: req.Slug, DisplayName: req.DisplayName, Enabled: req.Enabled, Transport: productruntime.MCPTransport(strings.TrimSpace(req.Transport)), Command: req.Command, Args: req.Args, Env: req.Env, TimeoutMS: req.TimeoutMS}
	validated, err := productruntime.ValidateMCPServerConfig(config)
	if err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, productruntime.RedactMCPText(err.Error())))
		return
	}
	if store, ok := s.product.(productdata.MCPServerConfigStore); ok {
		if _, err := store.SaveMCPServerConfig(r.Context(), identity.LocalDevIdentity(), productdata.MCPServerConfigRecord{Slug: validated.Slug, DisplayName: validated.DisplayName, Enabled: validated.Enabled, Transport: string(validated.Transport), Command: validated.Command, Args: validated.Args, Env: validated.Env, TimeoutMS: validated.TimeoutMS}); err != nil {
			writeAPIError(w, err)
			return
		}
	}
	status, err := s.mcpStatusForConfig(r.Context(), validated)
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mcpServerResponse{Server: status, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleDeleteMCPServer(w http.ResponseWriter, r *http.Request, slug string) {
	if store, ok := s.product.(productdata.MCPServerConfigStore); ok {
		if err := store.DeleteMCPServerConfig(r.Context(), identity.LocalDevIdentity(), slug); err != nil {
			writeAPIError(w, err)
			return
		}
	}
	servers, err := s.mcpServerStatuses(r.Context())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mcpServersResponse{Servers: servers, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) handleDiscoverMCPServer(w http.ResponseWriter, slug string) {
	configs, err := s.mcpServerConfigs()
	if err != nil {
		writeAPIError(w, err)
		return
	}
	config, ok := configs[slug]
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "MCP server config was not found."))
		return
	}
	result, _ := productruntime.DiscoverMCPTools(nil, config)
	event := productdata.RunEvent{ID: productdata.NewRunEventID(), UserID: identity.LocalDevIdentity().UserID, Type: "mcp_discovery_" + string(result.Status), Category: productdata.RunEventCategoryProgress, Summary: "MCP discovery " + string(result.Status), Metadata: productruntime.MCPDiscoveryEventMetadata(result), CreatedAt: time.Now().UTC()}
	s.providerMu.Lock()
	s.mcpDiscoveryEvents = append(s.mcpDiscoveryEvents, event)
	s.providerMu.Unlock()
	status := productruntime.MCPServerStatuses(map[string]productruntime.MCPServerConfig{config.Slug: config}, []productdata.RunEvent{event})[0]
	writeJSON(w, http.StatusOK, mcpServerResponse{Server: status, RequestID: diagnostics.NewRequestID()})
}

func (s *Server) mcpServerStatuses(ctx context.Context) ([]productruntime.MCPServerStatus, error) {
	if s.product == nil {
		return []productruntime.MCPServerStatus{}, nil
	}
	configs, err := s.mcpServerConfigs()
	if err != nil {
		return nil, err
	}
	events, err := s.product.ListMCPDiscoveryEvents(ctx, identity.LocalDevIdentity())
	if err != nil {
		return nil, err
	}
	s.providerMu.RLock()
	events = append(events, s.mcpDiscoveryEvents...)
	s.providerMu.RUnlock()
	return productruntime.MCPServerStatuses(configs, events), nil
}

func (s *Server) mcpStatusForConfig(ctx context.Context, config productruntime.MCPServerConfig) (productruntime.MCPServerStatus, error) {
	events, err := s.product.ListMCPDiscoveryEvents(ctx, identity.LocalDevIdentity())
	if err != nil {
		return productruntime.MCPServerStatus{}, err
	}
	s.providerMu.RLock()
	events = append(events, s.mcpDiscoveryEvents...)
	s.providerMu.RUnlock()
	statuses := productruntime.MCPServerStatuses(map[string]productruntime.MCPServerConfig{config.Slug: config}, events)
	if len(statuses) == 0 {
		return productruntime.MCPServerStatus{}, productdata.NewError(productdata.CodeInvalidRequest, "MCP server status unavailable.")
	}
	return statuses[0], nil
}

func (s *Server) mcpServerConfigs() (map[string]productruntime.MCPServerConfig, error) {
	configs, err := productruntime.MCPServerConfigsFromEnv()
	if err != nil {
		return nil, productdata.NewError(productdata.CodeInvalidRequest, "MCP server config is invalid.")
	}
	if store, ok := s.product.(productdata.MCPServerConfigStore); ok {
		saved, err := store.ListMCPServerConfigs(context.Background(), identity.LocalDevIdentity())
		if err != nil {
			return nil, err
		}
		for _, record := range saved {
			config := productruntime.MCPServerConfig{Slug: record.Slug, DisplayName: record.DisplayName, Enabled: record.Enabled, Transport: productruntime.MCPTransport(record.Transport), Command: record.Command, Args: record.Args, Env: record.Env, TimeoutMS: record.TimeoutMS}
			validated, err := productruntime.ValidateMCPServerConfig(config)
			if err != nil {
				return nil, productdata.NewError(productdata.CodeInvalidRequest, "Saved MCP server config is invalid.")
			}
			configs[validated.Slug] = validated
		}
	}
	return configs, nil
}
