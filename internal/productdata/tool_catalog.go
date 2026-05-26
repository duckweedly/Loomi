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
	entries = append(entries, builtinWorkspaceCatalogEntries()...)
	entries = append(entries, builtinSandboxCatalogEntries()...)
	entries = append(entries, builtinLSPCatalogEntries()...)
	entries = append(entries, builtinWebCatalogEntries()...)
	entries = append(entries, builtinBrowserCatalogEntries()...)
	entries = append(entries, builtinArtifactCatalogEntries()...)
	entries = append(entries, builtinAgentCatalogEntries()...)
	byName := map[string]ToolCatalogEntry{}
	for _, entry := range entries {
		byName[entry.Name] = entry
	}
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

func builtinWorkspaceCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinWorkspaceReadCatalogEntry(ToolNameWorkspaceGlob, "Workspace glob", "Find files under the configured workspace root.", []string{"pattern", "path", "limit"}),
		builtinWorkspaceReadCatalogEntry(ToolNameWorkspaceGrep, "Workspace grep", "Search text files under the configured workspace root.", []string{"query", "path", "include", "case_sensitive", "limit"}),
		builtinWorkspaceReadCatalogEntry(ToolNameWorkspaceRead, "Workspace read", "Read a bounded UTF-8 text slice from one workspace file.", []string{"path", "offset", "limit", "max_bytes"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspaceWriteFile, "Workspace write file", "Create a bounded UTF-8 text file under the configured workspace root.", []string{"path", "content", "max_bytes"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspaceEdit, "Workspace edit", "Apply one bounded exact text replacement inside a workspace file.", []string{"path", "old_text", "new_text", "max_bytes"}),
	}
}

func builtinWorkspaceReadCatalogEntry(name string, displayName string, description string, arguments []string) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupWorkspace,
		RiskLevel:      ToolRiskLow,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments": append([]string(nil), arguments...),
			"read_only": true,
			"scope":     "workspace",
		},
	}
}

func builtinWorkspaceMutationCatalogEntry(name string, displayName string, description string, arguments []string) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupWorkspace,
		RiskLevel:      ToolRiskHigh,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":     append([]string(nil), arguments...),
			"read_only":     false,
			"scope":         "workspace",
			"write_capable": true,
		},
	}
}

func builtinSandboxCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		{
			Name:           ToolNameSandboxExecCommand,
			DisplayName:    "Bounded read-only command",
			Description:    "Run one approved read-only argv-form command under the configured workspace root. This is not an isolated sandbox.",
			Source:         ToolCatalogSourceBuiltin,
			Group:          ToolCatalogGroupSandbox,
			RiskLevel:      ToolRiskHigh,
			ApprovalPolicy: ToolApprovalAlwaysRequired,
			Enabled:        true,
			ExecutionState: ToolExecutionStateExecutable,
			SafeMetadata: map[string]any{
				"allowed_commands": []string{"pwd", "ls", "git status"},
				"arguments":        []string{"argv", "cwd", "timeout_ms", "max_output_bytes"},
				"argv_only":        true,
				"exec_capable":     true,
				"read_only":        true,
				"isolated_sandbox": false,
				"scope":            "bounded_read_only_command",
			},
		},
	}
}

func builtinLSPCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinLSPCatalogEntry(ToolNameLSPDiagnostics, "LSP diagnostics", "Read bounded diagnostics for a workspace source file.", []string{"path", "language", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPSymbols, "LSP symbols", "Read bounded symbol summaries for a workspace source file.", []string{"path", "query", "language", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPReferences, "LSP references", "Read bounded workspace references for a source position.", []string{"path", "line", "column", "include_declaration", "limit"}),
	}
}

func builtinLSPCatalogEntry(name string, displayName string, description string, arguments []string) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupLSP,
		RiskLevel:      ToolRiskLow,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments": append([]string(nil), arguments...),
			"read_only": true,
			"scope":     "lsp",
		},
	}
}

func builtinWebCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{{
		Name:           ToolNameWebFetch,
		DisplayName:    "Web fetch",
		Description:    "Fetch one bounded public HTTP(S) URL and return a safe text summary.",
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupWeb,
		RiskLevel:      ToolRiskMedium,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":      []string{"url", "max_bytes", "timeout_ms"},
			"network_access": "public_http_only",
			"read_only":      true,
			"scope":          "web",
		},
	}}
}

func builtinBrowserCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinBrowserCatalogEntry(ToolNameBrowserOpen, "Browser open", "Open one bounded public HTTP(S) page in a run-scoped browser session.", []string{"url", "max_bytes", "timeout_ms"}),
		builtinBrowserCatalogEntry(ToolNameBrowserSnapshot, "Browser snapshot", "Return the current safe snapshot for a run-scoped browser session.", []string{"session_id"}),
		builtinBrowserCatalogEntry(ToolNameBrowserClickLink, "Browser click link", "Navigate one safe link from a run-scoped browser session.", []string{"session_id", "link_index", "max_bytes", "timeout_ms"}),
	}
}

func builtinBrowserCatalogEntry(name string, displayName string, description string, arguments []string) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupBrowser,
		RiskLevel:      ToolRiskMedium,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":      append([]string(nil), arguments...),
			"network_access": "public_http_only",
			"scope":          "browser",
			"stateful":       true,
		},
	}
}

func builtinArtifactCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinArtifactCatalogEntry(ToolNameArtifactCreateText, "Artifact create text", "Create one bounded non-executable text artifact.", []string{"title", "content", "max_bytes"}, false),
		builtinArtifactCatalogEntry(ToolNameArtifactRead, "Artifact read", "Read one bounded text artifact excerpt.", []string{"artifact_id", "max_bytes"}, true),
		builtinArtifactCatalogEntry(ToolNameArtifactList, "Artifact list", "List bounded safe artifact summaries.", []string{"limit"}, true),
	}
}

func builtinArtifactCatalogEntry(name string, displayName string, description string, arguments []string, readOnly bool) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupArtifact,
		RiskLevel:      ToolRiskMedium,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":      append([]string(nil), arguments...),
			"non_executable": true,
			"read_only":      readOnly,
			"scope":          "artifact",
		},
	}
}

func builtinAgentCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinAgentCatalogEntry(ToolNameAgentSpawn, "Agent spawn", "Create one bounded child coordination task.", []string{"role", "goal"}, false),
		builtinAgentCatalogEntry(ToolNameAgentList, "Agent list", "List bounded child coordination task summaries.", []string{"limit"}, true),
		builtinAgentCatalogEntry(ToolNameAgentComplete, "Agent complete", "Complete one child coordination task with a bounded result summary.", []string{"task_id", "result_summary"}, false),
	}
}

func builtinAgentCatalogEntry(name string, displayName string, description string, arguments []string, readOnly bool) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupAgent,
		RiskLevel:      ToolRiskMedium,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":            append([]string(nil), arguments...),
			"autonomous_execution": false,
			"coordination_only":    true,
			"read_only":            readOnly,
			"scope":                "agent",
		},
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
