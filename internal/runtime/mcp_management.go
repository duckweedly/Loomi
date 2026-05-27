package runtime

import (
	"sort"
	"strings"

	"github.com/sheridiany/loomi/internal/productdata"
)

type MCPServerStatus struct {
	ServerSafeID      string   `json:"server_safe_id"`
	ServerSlug        string   `json:"server_slug"`
	DisplayName       string   `json:"display_name"`
	Transport         string   `json:"transport"`
	Enabled           bool     `json:"enabled"`
	ConfigSource      string   `json:"config_source"`
	DiscoveryStatus   string   `json:"discovery_status"`
	CandidateCount    int      `json:"candidate_count"`
	CandidateNames    []string `json:"candidate_names"`
	ExecutionMode     string   `json:"execution_mode"`
	RedactedErrorCode string   `json:"redacted_error_code,omitempty"`
	LastDiscoveredAt  string   `json:"last_discovered_at,omitempty"`
}

func MCPServerStatuses(configs map[string]MCPServerConfig, events []productdata.RunEvent) []MCPServerStatus {
	slugs := make([]string, 0, len(configs))
	for slug := range configs {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	latest := latestMCPDiscoveryByServer(events)
	statuses := make([]MCPServerStatus, 0, len(slugs))
	for _, slug := range slugs {
		config := configs[slug]
		status := MCPServerStatus{
			ServerSafeID:    "mcp:" + config.Slug,
			ServerSlug:      config.Slug,
			DisplayName:     RedactMCPText(config.DisplayName),
			Transport:       string(config.Transport),
			Enabled:         config.Enabled,
			ConfigSource:    "local",
			DiscoveryStatus: "not_discovered",
			CandidateNames:  []string{},
			ExecutionMode:   "disabled",
		}
		if !config.Enabled {
			status.DiscoveryStatus = string(MCPDiscoveryDisabled)
		}
		if event, ok := latest[slug]; ok {
			status.DiscoveryStatus = discoveryStatusFromEvent(event)
			status.CandidateNames = safeMCPCandidateNames(slug, event.Metadata)
			status.CandidateCount = len(status.CandidateNames)
			status.RedactedErrorCode = RedactMCPText(metadataString(event.Metadata, "error_code"))
			if !event.CreatedAt.IsZero() {
				status.LastDiscoveredAt = event.CreatedAt.UTC().Format("2006-01-02T15:04:05Z")
			}
		}
		if config.Enabled && status.DiscoveryStatus == string(MCPDiscoverySucceeded) && status.CandidateCount > 0 {
			status.ExecutionMode = "approval_gated"
		}
		statuses = append(statuses, status)
	}
	return statuses
}

func latestMCPDiscoveryByServer(events []productdata.RunEvent) map[string]productdata.RunEvent {
	latest := map[string]productdata.RunEvent{}
	for _, event := range events {
		slug := mcpStatusMetadataString(event.Metadata, "server_slug")
		if slug == "" {
			continue
		}
		existing, ok := latest[slug]
		if !ok || existing.CreatedAt.Before(event.CreatedAt) || (existing.CreatedAt.Equal(event.CreatedAt) && existing.Sequence <= event.Sequence) {
			latest[slug] = event
		}
	}
	return latest
}

func discoveryStatusFromEvent(event productdata.RunEvent) string {
	if status := mcpStatusMetadataString(event.Metadata, "status"); status != "" {
		return RedactMCPText(status)
	}
	switch event.Type {
	case "mcp_discovery_succeeded":
		return string(MCPDiscoverySucceeded)
	case "mcp_discovery_failed":
		return string(MCPDiscoveryFailed)
	case "mcp_discovery_rejected":
		return string(MCPDiscoveryRejected)
	default:
		return "unknown"
	}
}

func safeMCPCandidateNames(serverSlug string, metadata map[string]any) []string {
	values := mcpStatusMetadataStringSlice(metadata, "candidate_names")
	names := make([]string, 0, len(values))
	prefix := "mcp." + serverSlug + "."
	for _, value := range values {
		if productdata.IsMCPToolName(value) && strings.HasPrefix(value, prefix) {
			names = append(names, value)
		}
	}
	sort.Strings(names)
	return names
}

func mcpStatusMetadataString(metadata map[string]any, key string) string {
	if value, ok := metadata[key].(string); ok {
		return value
	}
	return ""
}

func mcpStatusMetadataStringSlice(metadata map[string]any, key string) []string {
	switch values := metadata[key].(type) {
	case []string:
		return append([]string(nil), values...)
	case []any:
		out := make([]string, 0, len(values))
		for _, value := range values {
			if text, ok := value.(string); ok {
				out = append(out, text)
			}
		}
		return out
	default:
		return nil
	}
}
