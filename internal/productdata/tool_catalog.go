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
	entries = append(entries, builtinDiscoveryCatalogEntries()...)
	entries = append(entries, builtinWorkspaceCatalogEntries()...)
	entries = append(entries, builtinSandboxCatalogEntries()...)
	entries = append(entries, builtinLSPCatalogEntries()...)
	entries = append(entries, builtinWebCatalogEntries()...)
	entries = append(entries, builtinBrowserCatalogEntries()...)
	entries = append(entries, builtinArtifactCatalogEntries()...)
	entries = append(entries, builtinAgentCatalogEntries()...)
	entries = append(entries, builtinMemoryCatalogEntries()...)
	entries = append(entries, builtinTodoCatalogEntries()...)
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

func builtinDiscoveryCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		{
			Name:           ToolNameLoadTools,
			DisplayName:    "Load tools",
			Description:    "Return safe descriptions for enabled runtime tools by exact name or keyword.",
			Source:         ToolCatalogSourceBuiltin,
			Group:          ToolCatalogGroupDiscovery,
			RiskLevel:      ToolRiskLow,
			ApprovalPolicy: ToolApprovalReadOnly,
			Enabled:        true,
			ExecutionState: ToolExecutionStateExecutable,
			SafeMetadata: map[string]any{
				"arguments":             []string{"queries", "names", "limit"},
				"read_only":             true,
				"scope":                 "runtime_catalog",
				"dynamic_schema_loader": false,
			},
		},
		{
			Name:           ToolNameLoadSkill,
			DisplayName:    "Load skill",
			Description:    "Return a safe installed skill summary by name.",
			Source:         ToolCatalogSourceBuiltin,
			Group:          ToolCatalogGroupDiscovery,
			RiskLevel:      ToolRiskLow,
			ApprovalPolicy: ToolApprovalReadOnly,
			Enabled:        true,
			ExecutionState: ToolExecutionStateExecutable,
			SafeMetadata: map[string]any{
				"arguments":           []string{"name", "limit"},
				"read_only":           true,
				"scope":               "skill_manifest",
				"returns_skill_body":  false,
				"executes_skill_code": false,
			},
		},
	}
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
		builtinWorkspaceReadCatalogEntry(ToolNameWorkspaceListDirectory, "Workspace list directory", "Read a bounded directory listing with safe relative paths.", []string{"path", "max_entries", "depth", "include_hidden", "sort"}),
		builtinWorkspaceReadCatalogEntry(ToolNameWorkspaceTreeSummary, "Workspace tree summary", "Summarize and classify a bounded directory tree.", []string{"path", "max_entries", "depth", "include_hidden", "sort"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspaceWriteFile, "Workspace write file", "Create a bounded UTF-8 text file under the configured workspace root.", []string{"path", "content", "max_bytes"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspaceEdit, "Workspace edit", "Apply one bounded exact text replacement inside a workspace file.", []string{"path", "old_text", "new_text", "max_bytes"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspacePatchPreview, "Workspace patch preview", "Preview one bounded exact text replacement before applying it.", []string{"path", "old_text", "new_text", "max_bytes"}),
		builtinWorkspaceMutationCatalogEntry(ToolNameWorkspacePatchApply, "Workspace patch apply", "Apply one previously previewed bounded text replacement.", []string{"path", "old_text", "new_text", "max_bytes"}),
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
		ApprovalPolicy: ToolApprovalReadOnly,
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
	metadata := map[string]any{
		"arguments":     append([]string(nil), arguments...),
		"read_only":     false,
		"scope":         "workspace",
		"write_capable": true,
	}
	if name == ToolNameWorkspaceEdit {
		metadata["requires_read_before_edit"] = true
		metadata["returns_diff"] = true
		metadata["normalizes_line_endings"] = true
		metadata["preserves_indentation"] = true
		metadata["strips_trailing_whitespace_except_markdown"] = true
	}
	if name == ToolNameWorkspacePatchPreview {
		metadata["read_only"] = true
		metadata["write_capable"] = false
		metadata["requires_read_before_preview"] = true
		metadata["returns_diff"] = true
		metadata["preview_only"] = true
		metadata["normalizes_line_endings"] = true
		metadata["preserves_indentation"] = true
		metadata["strips_trailing_whitespace_except_markdown"] = true
	}
	if name == ToolNameWorkspacePatchApply {
		metadata["requires_patch_preview"] = true
		metadata["returns_diff"] = true
		metadata["normalizes_line_endings"] = true
		metadata["preserves_indentation"] = true
		metadata["strips_trailing_whitespace_except_markdown"] = true
	}
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
		SafeMetadata:   metadata,
	}
}

func builtinSandboxCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinSandboxCatalogEntry(ToolNameSandboxExecCommand, "Bounded read-only command", "Run one approved argv-form read or validation command under the configured workspace root.", []string{"argv", "cwd", "timeout_ms", "max_output_bytes"}, "bounded_command"),
		builtinSandboxCatalogEntry(ToolNameSandboxStartProcess, "Sandbox start process", "Start one approved argv-form read or validation process under the configured workspace root.", []string{"argv", "cwd", "timeout_ms", "max_output_bytes", "stdin"}, "bounded_process"),
		builtinSandboxCatalogEntry(ToolNameSandboxContinueProcess, "Sandbox continue process", "Read current output/status for one run-scoped sandbox process and optionally write bounded stdin.", []string{"process_id", "cursor", "stdin_text", "input_seq", "close_stdin"}, "bounded_process"),
		builtinSandboxCatalogEntry(ToolNameSandboxTerminateProcess, "Sandbox terminate process", "Terminate one run-scoped sandbox process.", []string{"process_id"}, "bounded_process"),
	}
}

func builtinSandboxCatalogEntry(name string, displayName string, description string, arguments []string, scope string) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description + " This is not an isolated sandbox.",
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupSandbox,
		RiskLevel:      ToolRiskHigh,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"allowed_commands":   []string{"pwd", "ls", "cat", "head", "tail", "sed -n", "wc", "rg", "git status", "git diff", "git log", "git show", "go test", "bun test", "bun run build", "npm test", "npm run build", "pnpm test", "pnpm run build"},
			"arguments":          append([]string(nil), arguments...),
			"argv_only":          true,
			"exec_capable":       true,
			"read_only":          false,
			"cursor_capable":     name == ToolNameSandboxContinueProcess,
			"stdin_capable":      name == ToolNameSandboxStartProcess || name == ToolNameSandboxContinueProcess,
			"validation_capable": true,
			"isolated_sandbox":   false,
			"scope":              scope,
		},
	}
}

func builtinLSPCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinLSPCatalogEntry(ToolNameLSPDiagnostics, "LSP diagnostics", "Read bounded diagnostics for a workspace source file.", []string{"path", "language", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPSymbols, "LSP symbols", "Read bounded symbol summaries for a workspace source file.", []string{"path", "query", "language", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPReferences, "LSP references", "Read bounded workspace references for a source position.", []string{"path", "line", "column", "include_declaration", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPDefinition, "LSP definition", "Find a bounded best-effort definition for a source position.", []string{"path", "line", "column", "language", "limit"}),
		builtinLSPCatalogEntry(ToolNameLSPHover, "LSP hover", "Read a bounded best-effort hover summary for a source position.", []string{"path", "line", "column", "language"}),
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
	return []ToolCatalogEntry{
		{
			Name:           ToolNameWebFetch,
			DisplayName:    "Web fetch",
			Description:    "Fetch one bounded public HTTP(S) URL and return a safe text summary.",
			Source:         ToolCatalogSourceBuiltin,
			Group:          ToolCatalogGroupWeb,
			RiskLevel:      ToolRiskMedium,
			ApprovalPolicy: ToolApprovalReadOnly,
			Enabled:        true,
			ExecutionState: ToolExecutionStateExecutable,
			SafeMetadata: map[string]any{
				"arguments":      []string{"url", "max_bytes", "timeout_ms"},
				"network_access": "public_http_only",
				"read_only":      true,
				"scope":          "web",
			},
		},
		{
			Name:           ToolNameWebSearch,
			DisplayName:    "Web search",
			Description:    "Search the public web through a configured Brave or Tavily provider and return bounded safe results.",
			Source:         ToolCatalogSourceBuiltin,
			Group:          ToolCatalogGroupWeb,
			RiskLevel:      ToolRiskMedium,
			ApprovalPolicy: ToolApprovalReadOnly,
			Enabled:        true,
			ExecutionState: ToolExecutionStateExecutable,
			SafeMetadata: map[string]any{
				"arguments":      []string{"query", "provider", "limit", "timeout_ms"},
				"network_access": "search_provider_api",
				"providers":      []string{"tavily", "brave"},
				"read_only":      true,
				"scope":          "web",
			},
		},
	}
}

func builtinBrowserCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinBrowserCatalogEntry(ToolNameBrowserOpen, "Browser open", "Open one bounded public HTTP(S) page in a run-scoped browser session.", []string{"url", "max_bytes", "timeout_ms"}),
		builtinBrowserCatalogEntry(ToolNameBrowserSnapshot, "Browser snapshot", "Return the current safe snapshot for a run-scoped browser session.", []string{"session_id"}),
		builtinBrowserCatalogEntry(ToolNameBrowserClickLink, "Browser click link", "Navigate one safe link from a run-scoped browser session.", []string{"session_id", "link_index", "max_bytes", "timeout_ms"}),
		builtinBrowserCatalogEntry(ToolNameBrowserScreenshot, "Browser screenshot", "Return a bounded text screenshot summary for a run-scoped browser session.", []string{"session_id"}),
		builtinBrowserCatalogEntry(ToolNameBrowserType, "Browser type", "Record bounded text into a discovered input target in a run-scoped browser session.", []string{"session_id", "target", "text"}),
		builtinBrowserCatalogEntry(ToolNameBrowserPress, "Browser press", "Record one bounded key press in a run-scoped browser session.", []string{"session_id", "key"}),
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
			"javascript":     false,
			"network_access": "public_http_only",
			"read_only":      name != ToolNameBrowserType && name != ToolNameBrowserPress,
			"scope":          "browser",
			"stateful":       true,
		},
	}
}

func builtinArtifactCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinArtifactCatalogEntry(ToolNameArtifactCreateText, "Artifact create text", "Create one bounded non-executable text artifact.", []string{"title", "filename", "mime_type", "display", "content", "max_bytes"}, false),
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

func builtinMemoryCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{
		builtinMemoryCatalogEntry(ToolNameMemorySearch, "Memory search", "Search approved memory summaries in the current safe scope.", []string{"query", "limit", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_type"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryList, "Memory list", "List approved memory summaries in the current safe scope.", []string{"limit", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_type"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryRead, "Memory read", "Read one approved memory summary without raw content.", []string{"entry_id", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryWrite, "Memory write", "Create one approval-gated memory write proposal.", []string{"title", "content", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_event_id", "idempotency_key"}, false),
		builtinMemoryCatalogEntry(ToolNameMemoryEdit, "Memory edit", "Edit a pending memory proposal or create an approval-gated replacement proposal.", []string{"proposal_id", "entry_id", "title", "content", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_event_id", "idempotency_key"}, false),
		builtinMemoryCatalogEntry(ToolNameMemoryForget, "Memory forget", "Tombstone one approved memory entry through the audited memory boundary.", []string{"entry_id", "reason", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, false),
		builtinMemoryCatalogEntry(ToolNameMemoryContext, "Memory context", "Return provider status plus bounded relevant memory summaries.", []string{"query", "limit", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_type"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryTimeline, "Memory timeline", "List safe memory audit timeline items.", []string{"limit", "scope_type", "scope_id", "source_thread_id", "source_run_id", "source_type"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryConnections, "Memory connections", "Return bounded related memory summaries for one entry or query.", []string{"entry_id", "query", "limit", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryThreadSearch, "Memory thread search", "Search local thread and message history with safe excerpts.", []string{"query", "limit"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryThreadFetch, "Memory thread fetch", "Fetch safe local thread message excerpts.", []string{"thread_id", "limit"}, true),
		builtinMemoryCatalogEntry(ToolNameMemoryStatus, "Memory status", "Return memory provider readiness and configuration state.", []string{}, true),
		builtinMemoryCatalogEntry(ToolNameNotebookRead, "Notebook read", "Read one approved structured notebook entry.", []string{"entry_id", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, true),
		builtinMemoryCatalogEntry(ToolNameNotebookWrite, "Notebook write", "Write one approved structured notebook entry through the audited memory boundary.", []string{"title", "content", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, false),
		builtinMemoryCatalogEntry(ToolNameNotebookEdit, "Notebook edit", "Replace one structured notebook entry by tombstoning the old entry and writing a new approved entry.", []string{"entry_id", "title", "content", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, false),
		builtinMemoryCatalogEntry(ToolNameNotebookForget, "Notebook forget", "Tombstone one structured notebook entry.", []string{"entry_id", "reason", "scope_type", "scope_id", "source_thread_id", "source_run_id"}, false),
	}
}

func ApplyMemoryToolAvailability(entries []ToolCatalogEntry, status MemoryProviderStatus) []ToolCatalogEntry {
	allowed := memoryToolProviderAllowlist(status)
	next := make([]ToolCatalogEntry, 0, len(entries))
	for _, entry := range entries {
		if !IsMemoryToolName(entry.Name) {
			next = append(next, entry)
			continue
		}
		entry.SafeMetadata = cloneSafeMetadata(entry.SafeMetadata)
		entry.SafeMetadata["active_provider"] = string(status.Provider)
		entry.SafeMetadata["available_providers"] = memoryToolAvailableProviders(entry.Name)
		if !allowed[entry.Name] {
			entry.Enabled = false
			entry.ExecutionState = ToolExecutionStateDisabled
			entry.ApprovalPolicy = ToolApprovalDisabled
			entry.SafeMetadata["disabled_reason"] = memoryToolDisabledReason(status)
		}
		next = append(next, entry)
	}
	return next
}

func FilterMemoryToolResolutionsForProvider(tools []ToolResolution, status MemoryProviderStatus) []ToolResolution {
	allowed := memoryToolProviderAllowlist(status)
	next := make([]ToolResolution, 0, len(tools))
	for _, tool := range tools {
		if IsMemoryToolName(tool.Name) && !allowed[tool.Name] {
			continue
		}
		next = append(next, tool)
	}
	return next
}

func memoryToolProviderAllowlist(status MemoryProviderStatus) map[string]bool {
	tools := map[string]bool{}
	if !status.Enabled || status.State == MemoryProviderStateDisabled || status.State == MemoryProviderStateUnconfigured {
		return tools
	}
	for _, name := range []string{ToolNameMemorySearch, ToolNameMemoryList, ToolNameMemoryRead, ToolNameMemoryWrite, ToolNameMemoryEdit, ToolNameMemoryForget, ToolNameMemoryContext, ToolNameMemoryTimeline, ToolNameMemoryConnections, ToolNameMemoryThreadSearch, ToolNameMemoryThreadFetch, ToolNameMemoryStatus, ToolNameNotebookRead, ToolNameNotebookWrite, ToolNameNotebookEdit, ToolNameNotebookForget} {
		tools[name] = true
	}
	if status.Provider == MemoryProviderNowledge {
		delete(tools, ToolNameMemoryEdit)
	}
	return tools
}

func memoryToolAvailableProviders(name string) []string {
	providers := []string{string(MemoryProviderLocal), string(MemoryProviderSemantic), string(MemoryProviderOpenViking)}
	if name != ToolNameMemoryEdit {
		providers = append(providers, string(MemoryProviderNowledge))
	}
	return providers
}

func memoryToolDisabledReason(status MemoryProviderStatus) string {
	if !status.Enabled || status.State == MemoryProviderStateDisabled {
		return "memory_disabled"
	}
	if status.State == MemoryProviderStateUnconfigured {
		return "memory_provider_unconfigured"
	}
	if status.Provider == MemoryProviderNowledge {
		return "not_supported_by_nowledge"
	}
	return "not_supported_by_provider"
}

func cloneSafeMetadata(metadata map[string]any) map[string]any {
	next := make(map[string]any, len(metadata)+3)
	for key, value := range metadata {
		next[key] = value
	}
	return next
}

func builtinMemoryCatalogEntry(name string, displayName string, description string, arguments []string, readOnly bool) ToolCatalogEntry {
	return ToolCatalogEntry{
		Name:           name,
		DisplayName:    displayName,
		Description:    description,
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupMemory,
		RiskLevel:      ToolRiskMedium,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":           append([]string(nil), arguments...),
			"approval_gated":      true,
			"read_only":           readOnly,
			"returns_raw_content": false,
			"scope":               "memory",
		},
	}
}

func builtinTodoCatalogEntries() []ToolCatalogEntry {
	return []ToolCatalogEntry{{
		Name:           ToolNameTodoWrite,
		DisplayName:    "Todo write",
		Description:    "Replace the current Work-mode todo snapshot with bounded safe todo items.",
		Source:         ToolCatalogSourceBuiltin,
		Group:          ToolCatalogGroupTodo,
		RiskLevel:      ToolRiskLow,
		ApprovalPolicy: ToolApprovalAlwaysRequired,
		Enabled:        true,
		ExecutionState: ToolExecutionStateExecutable,
		SafeMetadata: map[string]any{
			"arguments":       []string{"items"},
			"read_only":       false,
			"scope":           "work_todo",
			"max_items":       MaxWorkTodoItems,
			"updates_plan_ui": true,
		},
	}}
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
