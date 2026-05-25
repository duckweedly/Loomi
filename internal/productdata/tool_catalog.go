package productdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

func ToolCatalogFromEvents(events []RunEvent) []ToolCatalogEntry {
	return toolCatalogFromEvents(events, ToolExecutionStateExecutable)
}

func SafeToolCatalogFromEvents(events []RunEvent) []ToolCatalogEntry {
	return toolCatalogFromEvents(events, ToolExecutionStateNonExecutable)
}

func toolCatalogFromEvents(events []RunEvent, mcpExecutionState ToolExecutionState) []ToolCatalogEntry {
	entries := []ToolCatalogEntry{builtinCurrentTimeCatalogEntry()}
	byName := map[string]ToolCatalogEntry{ToolNameCurrentTime: entries[0]}
	orderedEvents := append([]RunEvent(nil), events...)
	sort.SliceStable(orderedEvents, func(i, j int) bool {
		if orderedEvents[i].CreatedAt.Equal(orderedEvents[j].CreatedAt) {
			if orderedEvents[i].Sequence == orderedEvents[j].Sequence {
				return orderedEvents[i].ID < orderedEvents[j].ID
			}
			return orderedEvents[i].Sequence < orderedEvents[j].Sequence
		}
		return orderedEvents[i].CreatedAt.Before(orderedEvents[j].CreatedAt)
	})
	for _, event := range orderedEvents {
		if event.Type != "mcp_discovery_succeeded" || metadataStringValue(event.Metadata, "status") != "succeeded" {
			continue
		}
		slug := metadataStringValue(event.Metadata, "server_slug")
		for _, name := range metadataStringSlice(event.Metadata, "candidate_names") {
			if !IsMCPToolName(name) || mcpServerSlugFromToolName(name) != slug {
				continue
			}
			byName[name] = ToolCatalogEntry{
				Name:            name,
				DisplayName:     safeMCPDisplayName(name),
				Description:     "Discovered local MCP tool.",
				Source:          ToolCatalogSourceMCP,
				Group:           ToolCatalogGroupMCP,
				InputSchemaHash: mcpCandidateSchemaHash(event.Metadata, name),
				RiskLevel:       ToolRiskMedium,
				ApprovalPolicy:  ToolApprovalAlwaysRequired,
				Enabled:         true,
				ExecutionState:  mcpExecutionState,
				SafeMetadata: RedactEventMetadata(map[string]any{
					"server_slug":      slug,
					"discovery_status": "succeeded",
				}),
			}
		}
	}
	entries = entries[:0]
	for _, entry := range byName {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name < entries[j].Name })
	return entries
}

func builtinCurrentTimeCatalogEntry() ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:            ToolNameCurrentTime,
		DisplayName:     "Current time",
		Description:     "Returns the current UTC time.",
		Source:          ToolCatalogSourceBuiltin,
		Group:           ToolCatalogGroupRuntime,
		InputSchemaHash: inputSchemaHash(map[string]any{"type": "object", "properties": map[string]any{"timezone": map[string]any{"type": "string", "enum": []string{"UTC"}}}}),
		RiskLevel:       ToolRiskLow,
		ApprovalPolicy:  ToolApprovalAlwaysRequired,
		Enabled:         true,
		ExecutionState:  ToolExecutionStateExecutable,
		SafeMetadata:    map[string]any{"arguments": []string{"timezone"}},
	}
}

func inputSchemaHash(schema map[string]any) string {
	raw, err := json.Marshal(schema)
	if err != nil {
		raw = []byte("{}")
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func safeMCPDisplayName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) != 3 {
		return name
	}
	return parts[1] + " " + parts[2]
}

func mcpCandidateSchemaHash(metadata map[string]any, toolName string) string {
	for _, key := range []string{"candidate_schema_hashes", "schema_hashes"} {
		hashes, ok := metadata[key].(map[string]any)
		if !ok {
			continue
		}
		if hash, ok := hashes[toolName].(string); ok && strings.TrimSpace(hash) != "" {
			return strings.TrimSpace(hash)
		}
	}
	for _, key := range []string{"candidate_schema_hash", "schema_hash"} {
		if hash := metadataStringValue(metadata, key); hash != "" {
			return hash
		}
	}
	return ""
}
