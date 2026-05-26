package httpapi

import (
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

type mcpServersResponse struct {
	Servers   []productruntime.MCPServerStatus `json:"servers"`
	RequestID string                           `json:"request_id"`
}

func (s *Server) handleMCPServers(w http.ResponseWriter, r *http.Request) {
	if !s.productAvailable(w) {
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, "GET")
		return
	}
	configs, err := productruntime.MCPServerConfigsFromEnv()
	if err != nil {
		writeAPIError(w, productdata.NewError(productdata.CodeInvalidRequest, "MCP server config is invalid."))
		return
	}
	events, err := s.product.ListMCPDiscoveryEvents(r.Context(), identity.LocalDevIdentity())
	if err != nil {
		writeAPIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, mcpServersResponse{Servers: productruntime.MCPServerStatuses(configs, events), RequestID: diagnostics.NewRequestID()})
}
