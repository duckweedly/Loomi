package httpapi

import (
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type workerQueueDiagnosticsResponse struct {
	Diagnostics productdata.WorkerQueueDiagnostics `json:"diagnostics"`
	RequestID   string                             `json:"request_id"`
}

func (s *Server) handleWorkerQueueDiagnostics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	if !s.productAvailable(w) {
		return
	}
	status, err := s.product.WorkerQueueDiagnostics(r.Context(), identity.LocalDevIdentity())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if s.cfg.WorkerQueuePaused {
		status.QueueStatus = productdata.WorkerQueueStatusPaused
		status.WorkerStatus = productdata.WorkerStatusPaused
	}
	if !s.cfg.WorkerQueueEnabled {
		status.WorkerStatus = productdata.WorkerStatusStopped
	}
	writeJSON(w, http.StatusOK, workerQueueDiagnosticsResponse{Diagnostics: status, RequestID: diagnostics.NewRequestID()})
}
