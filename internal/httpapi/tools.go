package httpapi

import (
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type toolCatalogResponse struct {
	Tools     []productdata.ToolCatalogEntry `json:"tools"`
	RequestID string                         `json:"request_id"`
}

func (s *Server) handleToolsCatalog(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	tools, err := s.product.ListToolCatalog(r.Context(), identity.LocalDevIdentity())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	if tools == nil {
		tools = []productdata.ToolCatalogEntry{}
	}
	writeJSON(w, http.StatusOK, toolCatalogResponse{Tools: tools, RequestID: diagnostics.NewRequestID()})
}
