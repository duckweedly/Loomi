package httpapi

import (
	"net/http"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type contextSourceListResponse struct {
	Sources   []productdata.ContextSource `json:"sources"`
	RequestID string                      `json:"request_id"`
}

type contextSourceResponse struct {
	Source    productdata.ContextSource `json:"source"`
	RequestID string                    `json:"request_id"`
}

type createContextSourceRequest struct {
	Kind     productdata.ContextSourceKind `json:"kind"`
	Title    string                        `json:"title"`
	Locator  string                        `json:"locator"`
	Summary  string                        `json:"summary"`
	Metadata map[string]any                `json:"metadata"`
}

func (s *Server) handleThreadContextSources(w http.ResponseWriter, r *http.Request, threadID string, suffix string) {
	if suffix != "sources" {
		writeAPIError(w, productdata.NewError(productdata.CodeThreadNotFound, "Thread not found."))
		return
	}
	sources, ok := s.product.(productdata.ContextSourceService)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Context source service is unavailable."))
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := sources.ListContextSources(r.Context(), identity.LocalDevIdentity(), productdata.ListContextSourcesInput{ThreadID: threadID, Limit: intQuery(r, "limit", 20)})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, contextSourceListResponse{Sources: items, RequestID: diagnostics.NewRequestID()})
	case http.MethodPost:
		var req createContextSourceRequest
		if err := decodeJSONRequest(r, &req); err != nil {
			writeAPIError(w, err)
			return
		}
		source, err := sources.CreateContextSource(r.Context(), identity.LocalDevIdentity(), productdata.CreateContextSourceInput{
			ThreadID: threadID,
			Kind:     req.Kind,
			Title:    req.Title,
			Locator:  req.Locator,
			Summary:  req.Summary,
			Metadata: req.Metadata,
		})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, contextSourceResponse{Source: source, RequestID: diagnostics.NewRequestID()})
	default:
		writeMethodNotAllowed(w, "GET, POST")
	}
}
