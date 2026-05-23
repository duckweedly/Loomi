package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/diagnostics"
)

type healthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	RequestID   string `json:"request_id"`
}

type readinessResponse struct {
	Status      string              `json:"status"`
	Service     string              `json:"service"`
	Environment string              `json:"environment"`
	RequestID   string              `json:"request_id"`
	Checks      []db.ReadinessCheck `json:"checks"`
}

func (s *Server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:      "alive",
		Service:     "loomi-api",
		Environment: s.cfg.AppEnv,
		RequestID:   diagnostics.NewRequestID(),
	})
}

func (s *Server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	// Readiness must fail visibly instead of hanging on slow dependencies.
	timeout := time.Duration(s.cfg.ReadinessTimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(r.Context(), timeout)
	defer cancel()
	checks := db.CheckReadiness(ctx, s.checker)
	status := "ready"
	code := http.StatusOK
	for _, check := range checks {
		if check.Status == db.StatusFailed {
			status = "not_ready"
			code = http.StatusServiceUnavailable
			break
		}
	}
	writeJSON(w, code, readinessResponse{
		Status:      status,
		Service:     "loomi-api",
		Environment: s.cfg.AppEnv,
		RequestID:   diagnostics.NewRequestID(),
		Checks:      checks,
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
