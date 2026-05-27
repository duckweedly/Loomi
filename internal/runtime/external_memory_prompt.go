package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func (g *Gateway) withExternalMemorySnapshot(ctx context.Context, prepared *productdata.RunContext, messages []ProviderMessage) *productdata.RunContext {
	if g == nil || g.Service == nil || prepared == nil {
		return prepared
	}
	query := lastUserPromptText(messages)
	if query == "" {
		return prepared
	}
	executor := MemoryToolExecutor{Service: g.Service}
	hits, handled, err := executor.externalMemorySearch(ctx, query, 5)
	if !handled {
		return prepared
	}
	if err != nil {
		g.appendExternalMemorySnapshotFailure(ctx, prepared, err)
		return prepared
	}
	if len(hits) == 0 {
		return prepared
	}
	next := *prepared
	entries := make([]productdata.MemorySearchResult, 0, len(hits))
	now := time.Now()
	for _, hit := range hits {
		entries = append(entries, productdata.MemorySearchResult{
			ID:               hit.URI,
			Title:            productdata.RedactEventText(hit.Title),
			Summary:          productdata.RedactEventText(hit.Summary),
			ScopeType:        productdata.MemoryScopeUser,
			ScopeID:          prepared.Thread.UserID,
			Status:           string(productdata.MemoryEntryApproved),
			SafetyState:      string(productdata.MemorySafetySafe),
			SourceThreadID:   prepared.Thread.ID,
			SourceRunID:      prepared.Run.ID,
			SourceType:       string(hit.Provider),
			CreatedAt:        now,
			UpdatedAt:        now,
			RankReason:       productdata.RedactEventText(hit.RankReason),
			RedactionApplied: true,
		})
	}
	next.MemorySnapshot = productdata.MemorySnapshot{
		RunID:            prepared.Run.ID,
		ThreadID:         prepared.Thread.ID,
		Entries:          entries,
		Limit:            5,
		TotalCandidates:  len(entries),
		LoadStatus:       "loaded_external",
		RedactionApplied: true,
	}
	_, _ = g.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), prepared.Run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     productdata.EventMemoryExternalSnapshotLoaded,
		Summary:  "External memory snapshot loaded",
		Metadata: productdata.RedactEventMetadata(map[string]any{
			"provider":          string(hits[0].Provider),
			"status":            next.MemorySnapshot.LoadStatus,
			"entry_count":       len(entries),
			"limit":             next.MemorySnapshot.Limit,
			"redaction_applied": true,
		}),
	})
	return &next
}

func (g *Gateway) appendExternalMemorySnapshotFailure(ctx context.Context, prepared *productdata.RunContext, err error) {
	if g == nil || g.Service == nil || prepared == nil || err == nil {
		return
	}
	status, statusErr := g.Service.GetMemoryProviderStatus(ctx, identity.LocalDevIdentity())
	if statusErr != nil {
		return
	}
	_, _ = g.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), prepared.Run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryError,
		Type:     productdata.EventMemoryExternalSnapshotFailed,
		Summary:  "External memory snapshot failed",
		Metadata: productdata.RedactEventMetadata(map[string]any{
			"provider":          string(status.Provider),
			"status":            "failed",
			"error_code":        "memory_external_snapshot_failed",
			"redaction_applied": true,
		}),
		ErrorCode:    "memory_external_snapshot_failed",
		ErrorMessage: "External memory snapshot failed",
	})
}

func lastUserPromptText(messages []ProviderMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == "user" && strings.TrimSpace(messages[i].Content) != "" {
			return productdata.RedactEventText(strings.TrimSpace(messages[i].Content))
		}
	}
	return ""
}
