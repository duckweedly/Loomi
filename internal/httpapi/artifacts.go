package httpapi

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type artifactResponse struct {
	Artifact  artifactDTO `json:"artifact"`
	RequestID string      `json:"request_id"`
}

type artifactListResponse struct {
	Artifacts []artifactDTO `json:"artifacts"`
	RequestID string        `json:"request_id"`
}

type artifactDTO struct {
	ID           string `json:"id"`
	ThreadID     string `json:"thread_id"`
	RunID        string `json:"run_id"`
	Title        string `json:"title"`
	ArtifactType string `json:"artifact_type"`
	ContentBytes int    `json:"content_bytes"`
	TextExcerpt  string `json:"text_excerpt"`
	Truncated    bool   `json:"truncated"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

func (s *Server) handleThreadArtifacts(w http.ResponseWriter, r *http.Request, threadID string, suffix string) {
	artifacts, ok := s.product.(productdata.ArtifactService)
	if !ok {
		writeAPIError(w, productdata.NewError(productdata.CodeInternalError, "Artifact service is unavailable."))
		return
	}
	switch {
	case suffix == "artifacts":
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, "GET")
			return
		}
		limit := intQuery(r, "limit", 20)
		items, err := artifacts.ListArtifacts(r.Context(), identity.LocalDevIdentity(), productdata.ListArtifactsInput{ThreadID: threadID, Limit: limit})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, artifactListResponse{Artifacts: artifactDTOs(items), RequestID: diagnostics.NewRequestID()})
	case strings.HasPrefix(suffix, "artifacts/"):
		if r.Method != http.MethodGet {
			writeMethodNotAllowed(w, "GET")
			return
		}
		artifactID := strings.TrimSpace(strings.TrimPrefix(suffix, "artifacts/"))
		if artifactID == "" || strings.Contains(artifactID, "/") {
			writeAPIError(w, productdata.NewError(productdata.CodeArtifactNotFound, "Artifact not found."))
			return
		}
		maxBytes := intQuery(r, "max_bytes", 4096)
		artifact, err := artifacts.ReadArtifact(r.Context(), identity.LocalDevIdentity(), productdata.ReadArtifactInput{ThreadID: threadID, ArtifactID: artifactID, MaxBytes: maxBytes})
		if err != nil {
			writeAPIError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, artifactResponse{Artifact: artifactDTOFromArtifact(artifact), RequestID: diagnostics.NewRequestID()})
	default:
		writeAPIError(w, productdata.NewError(productdata.CodeThreadNotFound, "Thread not found."))
	}
}

func artifactDTOs(artifacts []productdata.Artifact) []artifactDTO {
	items := make([]artifactDTO, 0, len(artifacts))
	for _, artifact := range artifacts {
		items = append(items, artifactDTOFromArtifact(artifact))
	}
	return items
}

func artifactDTOFromArtifact(artifact productdata.Artifact) artifactDTO {
	return artifactDTO{
		ID:           artifact.ID,
		ThreadID:     artifact.ThreadID,
		RunID:        artifact.RunID,
		Title:        artifact.Title,
		ArtifactType: artifact.ArtifactType,
		ContentBytes: artifact.ContentBytes,
		TextExcerpt:  artifact.TextExcerpt,
		Truncated:    artifact.Truncated,
		CreatedAt:    artifact.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		UpdatedAt:    artifact.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}

func intQuery(r *http.Request, key string, fallback int) int {
	value := strings.TrimSpace(r.URL.Query().Get(key))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
