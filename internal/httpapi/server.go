package httpapi

import (
	"context"
	"net/http"
	"strings"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type RuntimeRunner interface {
	RunAsync(productdata.Run, string)
}

type GatewayRunner interface {
	RunAsync(context.Context, productdata.Run, productruntime.GatewayRunInput)
	SaveProviderConfig(productruntime.ProviderConfig) productruntime.ProviderConfig
}

type Server struct {
	cfg           config.Config
	checker       db.Checker
	product       productdata.Service
	broadcaster   *productruntime.Broadcaster
	runner        RuntimeRunner
	gatewayRunner GatewayRunner
	providers     []productruntime.ProviderConfig
	mux           *http.ServeMux
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
	s := &Server{cfg: cfg, checker: checker, product: product, broadcaster: broadcaster, runner: runner, gatewayRunner: gatewayRunner, providers: productruntime.ProviderConfigsFromConfig(cfg), mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /readyz", s.handleReadyz)
	s.mux.HandleFunc("GET /v1/me", s.handleCurrentIdentity)
	s.mux.HandleFunc("GET /v1/diagnostics/worker-queue", s.handleWorkerQueueDiagnostics)
	s.mux.HandleFunc("/v1/tools/catalog", s.handleToolCatalog)
	s.mux.HandleFunc("/v1/model-providers", s.handleModelProviders)
	s.mux.HandleFunc("POST /v1/model-providers/check", s.handleModelProviderCheck)
	s.mux.HandleFunc("/v1/threads", s.handleThreads)
	s.mux.HandleFunc("/v1/threads/", s.handleThreadByID)
	s.mux.HandleFunc("/v1/runs/", s.handleRunByID)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/v1/") || r.URL.Path == "/v1" {
		s.setCORSHeaders(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin == "" || !s.isLocalWebDevOrigin(origin) {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Vary", "Origin")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func (s *Server) isLocalWebDevOrigin(origin string) bool {
	if s.cfg.AppEnv != "local" && s.cfg.AppEnv != "development" {
		return false
	}
	return origin == "http://127.0.0.1:5173" || origin == "http://localhost:5173" || origin == "http://127.0.0.1:5180" || origin == "http://localhost:5180"
}
