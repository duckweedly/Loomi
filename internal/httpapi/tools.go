package httpapi

import (
	"net/http"
	"time"

	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func (s *Server) handleToolCatalog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tools":      productruntime.ToolCatalog().Tools,
		"updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	})
}
