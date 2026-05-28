package httpapi

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type RuntimeRunner interface {
	RunAsync(productdata.Run, string)
}

type GatewayRunner interface {
	RunAsync(context.Context, productdata.Run, productruntime.GatewayRunInput)
	SaveProviderConfig(productruntime.ProviderConfig) productruntime.ProviderConfig
	SaveProvider(productruntime.Provider)
	RemoveProvider(string)
}

type Server struct {
	cfg                         config.Config
	checker                     db.Checker
	product                     productdata.Service
	broadcaster                 *productruntime.Broadcaster
	runner                      RuntimeRunner
	gatewayRunner               GatewayRunner
	providerMu                  sync.RWMutex
	providers                   []productruntime.ProviderConfig
	localProviderDetectionInput productruntime.LocalProviderDetectionInput
	skillDiscoveryInput         productruntime.SkillDiscoveryInput
	localProviderEnablements    map[string]productruntime.LocalProviderCapability
	mcpDiscoveryEvents          []productdata.RunEvent
	mux                         *http.ServeMux
}

func NewServer(cfg config.Config, checker db.Checker) *Server {
	return NewServerWithProduct(cfg, checker, productdata.NewMemoryService())
}

func NewServerWithProduct(cfg config.Config, checker db.Checker, product productdata.Service) *Server {
	return NewServerWithRuntime(cfg, checker, product, productruntime.NewBroadcaster(), nil)
}

func NewServerWithRuntime(cfg config.Config, checker db.Checker, product productdata.Service, broadcaster *productruntime.Broadcaster, runner RuntimeRunner) *Server {
	s := NewServerWithRuntimes(cfg, checker, product, broadcaster, runner, nil)
	return s
}

func NewServerWithRuntimes(cfg config.Config, checker db.Checker, product productdata.Service, broadcaster *productruntime.Broadcaster, runner RuntimeRunner, gatewayRunner GatewayRunner) *Server {
	providers := append(productruntime.ProviderConfigsFromConfig(cfg), savedProviderConfigs(product)...)
	s := &Server{cfg: cfg, checker: checker, product: product, broadcaster: broadcaster, runner: runner, gatewayRunner: gatewayRunner, providers: providers, skillDiscoveryInput: productruntime.DefaultSkillDiscoveryInput(), localProviderEnablements: map[string]productruntime.LocalProviderCapability{}, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.HandleFunc("GET /v1/me", s.handleCurrentIdentity)
	s.mux.HandleFunc("GET /v1/diagnostics/worker-queue", s.handleWorkerQueueDiagnostics)
	s.mux.HandleFunc("/v1/personas", s.handlePersonas)
	s.mux.HandleFunc("/v1/model-providers", s.handleModelProviders)
	s.mux.HandleFunc("POST /v1/model-providers/check", s.handleModelProviderCheck)
	s.mux.HandleFunc("GET /v1/local-provider-detections", s.handleLocalProviderDetections)
	s.mux.HandleFunc("/v1/local-provider-detections/", s.handleLocalProviderDetectionByID)
	s.mux.HandleFunc("/v1/skills", s.handleSkills)
	s.mux.HandleFunc("/v1/tools/catalog", s.handleToolsCatalog)
	s.mux.HandleFunc("/v1/web-search/config", s.handleWebSearchConfig)
	s.mux.HandleFunc("/v1/workspace/root", s.handleWorkspaceRoot)
	s.mux.HandleFunc("/v1/mcp/servers", s.handleMCPServers)
	s.mux.HandleFunc("/v1/mcp/servers/", s.handleMCPServerBySlug)
	s.mux.HandleFunc("/v1/memory", s.handleMemory)
	s.mux.HandleFunc("/v1/memory/", s.handleMemoryByID)
	s.mux.HandleFunc("/v1/threads", s.handleThreads)
	s.mux.HandleFunc("/v1/threads/", s.handleThreadByID)
	s.mux.HandleFunc("/v1/runs/", s.handleRunByID)
	return s
}

func savedProviderConfigs(product productdata.Service) []productruntime.ProviderConfig {
	store, ok := product.(productdata.ModelProviderConfigStore)
	if !ok {
		return nil
	}
	saved, err := store.ListModelProviderConfigs(context.Background(), identity.LocalDevIdentity())
	if err != nil {
		return nil
	}
	providers := make([]productruntime.ProviderConfig, 0, len(saved))
	for _, provider := range saved {
		providers = append(providers, productruntime.ProviderConfig{ID: provider.ID, Family: productruntime.ProviderFamily(provider.Family), BaseURL: provider.BaseURL, APIKey: provider.APIKey, Model: provider.Model, Enabled: provider.Enabled})
	}
	return providers
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.isLocalAPIDiagnosticPath(r.URL.Path) || strings.HasPrefix(r.URL.Path, "/v1/") || r.URL.Path == "/v1" {
		s.setCORSHeaders(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) isLocalAPIDiagnosticPath(path string) bool {
	return path == "/healthz" || path == "/readyz"
}

func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" || !s.isLocalWebDevOrigin(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func (s *Server) isLocalWebDevOrigin(origin string) bool {
	if s.cfg.AppEnv != "local" && s.cfg.AppEnv != "development" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" || parsed.Port() == "" {
		return false
	}
	return parsed.Hostname() == "127.0.0.1" || parsed.Hostname() == "localhost"
}
