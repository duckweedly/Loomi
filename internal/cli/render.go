package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

type Renderer struct {
	Out io.Writer
}

func (r Renderer) PrintStatus(client *Client) error {
	if client == nil {
		client = NewClient("")
	}
	_, err := fmt.Fprintf(r.out(), "Loomi API: %s\n", client.BaseURL())
	return err
}

func (r Renderer) PrintTools(tools []ToolCatalogEntry) error {
	grouped := map[string][]ToolCatalogEntry{}
	for _, tool := range tools {
		group := strings.TrimSpace(tool.Group)
		if group == "" {
			group = "other"
		}
		grouped[group] = append(grouped[group], tool)
	}
	groups := make([]string, 0, len(grouped))
	for group := range grouped {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	for groupIndex, group := range groups {
		if groupIndex > 0 {
			if _, err := fmt.Fprintln(r.out()); err != nil {
				return err
			}
		}
		if _, err := fmt.Fprintf(r.out(), "[%s]\n", group); err != nil {
			return err
		}
		items := grouped[group]
		sort.Slice(items, func(i, j int) bool { return items[i].Name < items[j].Name })
		for _, tool := range items {
			if _, err := fmt.Fprintf(r.out(), "  %s\t%s\t%s\t%s\t%v\n", tool.Name, tool.ExecutionState, tool.ApprovalPolicy, tool.RiskLevel, tool.Enabled); err != nil {
				return err
			}
		}
	}
	return nil
}

func (r Renderer) PrintToolsFlat(tools []ToolCatalogEntry) error {
	for _, tool := range tools {
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%v\n", tool.Name, tool.Group, tool.ExecutionState, tool.Enabled); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintMCPServers(servers []MCPServerStatus) error {
	for _, server := range servers {
		candidates := "-"
		if len(server.CandidateNames) > 0 {
			candidates = strings.Join(server.CandidateNames, ",")
		}
		errorCode := server.RedactedErrorCode
		if errorCode == "" {
			errorCode = "-"
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\tenabled=%v\tdiscovery=%s\tcandidates=%d\texecution=%s\terror=%s\t%s\n", server.ServerSlug, server.DisplayName, server.Transport, server.Enabled, server.DiscoveryStatus, server.CandidateCount, server.ExecutionMode, errorCode, candidates); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintArtifacts(artifacts []Artifact) error {
	for _, artifact := range artifacts {
		excerpt := strings.ReplaceAll(strings.TrimSpace(artifact.TextExcerpt), "\n", "\\n")
		if len(excerpt) > 120 {
			excerpt = excerpt[:117] + "..."
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\tbytes=%d\ttruncated=%v\t%s\n", artifact.ID, artifact.ArtifactType, artifact.Title, artifact.ContentBytes, artifact.Truncated, excerpt); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintArtifact(artifact Artifact) error {
	if _, err := fmt.Fprintf(r.out(), "artifact %s %s\nthread %s\nrun %s\ntitle %s\nbytes %d\ntruncated %v\n\n", artifact.ID, artifact.ArtifactType, artifact.ThreadID, artifact.RunID, artifact.Title, artifact.ContentBytes, artifact.Truncated); err != nil {
		return err
	}
	_, err := fmt.Fprintln(r.out(), artifact.TextExcerpt)
	return err
}

func (r Renderer) PrintMemoryItems(items []MemorySearchResult) error {
	for _, item := range items {
		summary := strings.ReplaceAll(strings.TrimSpace(item.Summary), "\n", "\\n")
		if len(summary) > 120 {
			summary = summary[:117] + "..."
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s/%s\tredacted=%v\t%s\n", item.ID, item.Status, item.SafetyState, item.ScopeType, item.ScopeID, item.RedactionApplied, summary); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintMemoryItem(item MemorySearchResult) error {
	if _, err := fmt.Fprintf(r.out(), "memory %s\nstatus %s\nsafety %s\nscope %s/%s\nsource %s %s\nredacted %v\n\ntitle %s\nsummary %s\n", item.ID, item.Status, item.SafetyState, item.ScopeType, item.ScopeID, item.SourceType, item.SourceRunID, item.RedactionApplied, item.Title, item.Summary); err != nil {
		return err
	}
	return nil
}

func (r Renderer) PrintMemoryAudit(items []MemoryAuditItem) error {
	for _, item := range items {
		target := firstNonEmpty(item.MemoryEntryID, item.MemoryProposalID, "-")
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\tthread=%s\trun=%s\tstatus=%s\tredacted=%v\n", item.ID, item.EventType, target, item.ThreadID, item.RunID, item.Status, item.RedactionApplied); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintAgentTasks(tasks []AgentTask) error {
	for _, task := range tasks {
		goal := strings.ReplaceAll(strings.TrimSpace(task.Goal), "\n", "\\n")
		if len(goal) > 120 {
			goal = goal[:117] + "..."
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s\t%s\n", task.ID, task.Status, task.Role, task.RunID, goal); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintThreads(threads []Thread) error {
	for _, thread := range threads {
		title := strings.TrimSpace(thread.Title)
		if title == "" {
			title = "(untitled)"
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s\t%s\n", thread.ID, thread.Mode, thread.LifecycleStatus, thread.UpdatedAt, title); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintRun(run Run) error {
	_, err := fmt.Fprintf(r.out(), "run %s %s\nthread %s\n", run.ID, run.Status, run.ThreadID)
	return err
}

func (r Renderer) PrintPersonas(personas []Persona) error {
	for _, persona := range personas {
		defaultFlag := ""
		if persona.IsDefault {
			defaultFlag = "default"
		}
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s\tv%d\t%s\n", persona.ID, persona.Slug, persona.Name, defaultFlag, persona.ActiveVersion, persona.Source); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintModelProviders(providers []ProviderCapability) error {
	for _, provider := range providers {
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s\t%s\n", provider.ID, provider.Family, provider.Model, provider.Status, provider.ExecutionState); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintStopRun(result StopRunResult) error {
	_, err := fmt.Fprintf(r.out(), "run %s %s %s\n", result.Run.ID, result.Run.Status, result.Result)
	return err
}

func (r Renderer) PrintEvent(event RunEvent) error {
	summary := strings.TrimSpace(event.Summary)
	if summary == "" {
		summary = event.Type
	}
	if event.Content != nil && *event.Content != "" {
		summary = *event.Content
	}
	if toolSummary := formatToolEventSummary(event); toolSummary != "" {
		summary = toolSummary
	}
	_, err := fmt.Fprintf(r.out(), "%04d %s %s\n", event.Sequence, event.Type, summary)
	return err
}

func (r Renderer) PrintEventCompact(event RunEvent) error {
	if IsToolEvent(event) {
		toolName := metadataString(event.Metadata, "tool_name")
		if toolName == "" {
			toolName = "tool"
		}
		toolCallID := eventToolCallID(event)
		state := strings.TrimPrefix(event.Type, "tool_call_")
		detail := ""
		if text := formatToolEventDetail(event); text != "" {
			detail = " " + text
		}
		_, err := fmt.Fprintf(r.out(), "%04d %s %s %s%s\n", event.Sequence, state, toolName, toolCallID, detail)
		return err
	}
	summary := strings.TrimSpace(event.Summary)
	if summary == "" {
		summary = event.Type
	}
	if event.Content != nil && *event.Content != "" {
		summary = *event.Content
	}
	_, err := fmt.Fprintf(r.out(), "%04d %s\n", event.Sequence, summary)
	return err
}

func IsToolEvent(event RunEvent) bool {
	return strings.HasPrefix(event.Type, "tool_call_") || eventToolCallID(event) != ""
}

func (r Renderer) PrintRunResult(result RunResult) error {
	if _, err := fmt.Fprintf(r.out(), "\nrun %s %s\nthread %s\n", result.RunID, result.Status, result.ThreadID); err != nil {
		return err
	}
	if len(result.PendingApprovals) == 0 {
		return nil
	}
	if _, err := fmt.Fprintln(r.out(), "pending approvals:"); err != nil {
		return err
	}
	for _, approval := range result.PendingApprovals {
		toolName := approval.ToolName
		if toolName == "" {
			toolName = "tool"
		}
		if _, err := fmt.Fprintf(r.out(), "- %s %s\n  approve: loomi approvals approve %s %s %s\n  deny:    loomi approvals deny %s %s %s\n", toolName, approval.ToolCallID, approval.ThreadID, approval.RunID, approval.ToolCallID, approval.ThreadID, approval.RunID, approval.ToolCallID); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintApprovals(events []RunEvent) error {
	for _, event := range events {
		toolCallID := eventToolCallID(event)
		toolName := metadataString(event.Metadata, "tool_name")
		if _, err := fmt.Fprintf(r.out(), "%s\t%s\t%s\t%s\n", event.ThreadID, event.RunID, toolCallID, toolName); err != nil {
			return err
		}
	}
	return nil
}

func (r Renderer) PrintApprovalNotice(event RunEvent) error {
	toolCallID := eventToolCallID(event)
	if toolCallID == "" {
		return nil
	}
	toolName := metadataString(event.Metadata, "tool_name")
	if toolName == "" {
		toolName = "tool"
	}
	switch event.Type {
	case "tool_call_approval_required":
		_, err := fmt.Fprintf(r.out(), "approval required: %s %s\n  approve: loomi approvals approve %s %s %s\n  deny:    loomi approvals deny %s %s %s\n", toolName, toolCallID, event.ThreadID, event.RunID, toolCallID, event.ThreadID, event.RunID, toolCallID)
		return err
	case "tool_call_approved", "tool_call_denied", "tool_call_succeeded", "tool_call_failed", "tool_call_cancelled":
		_, err := fmt.Fprintf(r.out(), "%s: %s %s\n", event.Type, toolName, toolCallID)
		return err
	default:
		return nil
	}
}

func formatToolEventSummary(event RunEvent) string {
	toolCallID := eventToolCallID(event)
	toolName := metadataString(event.Metadata, "tool_name")
	if toolCallID == "" && toolName == "" {
		return ""
	}
	if toolName == "" {
		toolName = "tool"
	}
	prefix := strings.TrimSpace(toolName + " " + toolCallID)
	if detail := formatToolEventDetail(event); detail != "" {
		return prefix + " " + detail
	}
	if event.Type == "tool_call_approved" || event.Type == "tool_call_denied" || event.Type == "tool_call_executing" || event.Type == "tool_call_cancelled" {
		return prefix
	}
	return ""
}

func formatToolEventDetail(event RunEvent) string {
	toolName := metadataString(event.Metadata, "tool_name")
	switch event.Type {
	case "tool_call_requested", "tool_call_approval_required":
		if text := formatToolArguments(toolName, metadataMapValue(event.Metadata["arguments_summary"])); text != "" {
			return "args=" + text
		}
		if text := compactMetadataValue(event.Metadata["arguments_summary"]); text != "" {
			return "args=" + text
		}
	case "tool_call_succeeded":
		if text := formatToolResult(toolName, metadataMapValue(event.Metadata["result_summary"])); text != "" {
			return "result=" + text
		}
		if text := compactMetadataValue(event.Metadata["result_summary"]); text != "" {
			return "result=" + text
		}
	case "tool_call_failed":
		if code := metadataString(event.Metadata, "error_code"); code != "" {
			return "failed=" + code
		}
	}
	return ""
}

func formatToolArguments(toolName string, args map[string]any) string {
	if len(args) == 0 {
		return ""
	}
	switch toolName {
	case "workspace.glob", "workspace.grep", "workspace.read", "workspace.write_file", "workspace.edit", "workspace.patch_preview", "workspace.patch_apply", "lsp.diagnostics", "lsp.symbols", "lsp.references", "lsp.definition", "lsp.hover":
		return joinToolFields(
			toolField("path", metadataStringValue(args, "path")),
			toolField("pattern", metadataStringValue(args, "pattern")),
			toolField("query", metadataStringValue(args, "query")),
			toolField("limit", metadataNumberValue(args, "limit")),
		)
	case "sandbox.exec_command", "sandbox.start_process":
		return joinToolFields(
			toolField("argv", quoteToolText(joinStringListValue(args["argv"], " "))),
			toolField("cwd", metadataStringValue(args, "cwd")),
			toolField("timeout_ms", metadataNumberValue(args, "timeout_ms")),
			toolField("stdin", metadataBoolValue(args, "stdin")),
		)
	case "sandbox.continue_process", "sandbox.terminate_process":
		return joinToolFields(
			toolField("process", metadataStringValue(args, "process_id")),
			toolField("cursor", metadataNumberValue(args, "cursor")),
			toolField("close_stdin", metadataBoolValue(args, "close_stdin")),
		)
	case "browser.open":
		return joinToolFields(toolField("url", metadataStringValue(args, "url")))
	case "browser.snapshot", "browser.click_link", "browser.screenshot", "browser.type", "browser.press":
		return joinToolFields(
			toolField("session", metadataStringValue(args, "session_id")),
			toolField("index", metadataNumberValue(args, "index")),
			toolField("text", quoteToolText(metadataStringValue(args, "text"))),
			toolField("key", metadataStringValue(args, "key")),
		)
	case "artifact.create_text", "artifact.read", "artifact.list":
		return joinToolFields(
			toolField("artifact", metadataStringValue(args, "artifact_id")),
			toolField("title", quoteToolText(metadataStringValue(args, "title"))),
			toolField("limit", metadataNumberValue(args, "limit")),
		)
	case "web.fetch":
		return joinToolFields(toolField("url", metadataStringValue(args, "url")))
	case "web.search":
		return joinToolFields(
			toolField("query", quoteToolText(metadataStringValue(args, "query"))),
			toolField("limit", metadataNumberValue(args, "limit")),
		)
	case "agent.spawn", "agent.complete":
		return joinToolFields(
			toolField("task", metadataStringValue(args, "task_id")),
			toolField("role", metadataStringValue(args, "role")),
			toolField("goal", quoteToolText(metadataStringValue(args, "goal"))),
		)
	}
	return ""
}

func formatToolResult(toolName string, result map[string]any) string {
	if len(result) == 0 {
		return ""
	}
	switch toolName {
	case "workspace.glob":
		return joinToolFields(
			toolField("matches", metadataLenValue(result, "matches")),
			toolField("truncated", metadataBoolValue(result, "truncated")),
		)
	case "workspace.grep":
		return joinToolFields(
			toolField("matches", firstNonEmpty(metadataLenValue(result, "matches"), metadataNumberValue(result, "match_count"))),
			toolField("truncated", metadataBoolValue(result, "truncated")),
		)
	case "workspace.read":
		return joinToolFields(
			toolField("path", metadataStringValue(result, "path")),
			toolField("bytes", firstNonEmpty(metadataNumberValue(result, "bytes_read"), metadataNumberValue(result, "content_bytes"), metadataNumberValue(result, "bytes"))),
			toolField("truncated", metadataBoolValue(result, "truncated")),
		)
	case "workspace.write_file", "workspace.edit", "workspace.patch_preview", "workspace.patch_apply":
		return joinToolFields(
			toolField("path", metadataStringValue(result, "path")),
			toolField("changed", metadataBoolValue(result, "changed")),
			toolField("preview", metadataStringValue(result, "preview_id")),
			toolField("bytes", firstNonEmpty(metadataNumberValue(result, "bytes_written"), metadataNumberValue(result, "bytes_after"))),
			toolField("lines", metadataNumberValue(result, "line_count_after")),
		)
	case "sandbox.exec_command":
		return joinToolFields(
			toolField("exit", metadataNumberValue(result, "exit_code")),
			toolField("timeout", metadataBoolValue(result, "timed_out")),
			toolField("stdout", quoteToolText(firstLine(metadataStringValue(result, "stdout")))),
		)
	case "sandbox.start_process", "sandbox.continue_process", "sandbox.terminate_process":
		return joinToolFields(
			toolField("process", metadataStringValue(result, "process_id")),
			toolField("status", metadataStringValue(result, "status")),
			toolField("exit", metadataNumberValue(result, "exit_code")),
			toolField("cursor", firstNonEmpty(metadataNumberValue(result, "next_cursor"), metadataNumberValue(result, "cursor"))),
			toolField("stdout", quoteToolText(firstLine(metadataStringValue(result, "stdout")))),
		)
	case "browser.open", "browser.snapshot", "browser.click_link", "browser.screenshot", "browser.type", "browser.press":
		return joinToolFields(
			toolField("session", metadataStringValue(result, "session_id")),
			toolField("title", quoteToolText(metadataStringValue(result, "title"))),
			toolField("url", metadataStringValue(result, "url")),
			toolField("links", metadataLenValue(result, "links")),
			toolField("inputs", metadataLenValue(result, "inputs")),
		)
	case "artifact.create_text", "artifact.read":
		return joinToolFields(
			toolField("artifact", metadataStringValue(result, "artifact_id")),
			toolField("title", quoteToolText(metadataStringValue(result, "title"))),
			toolField("bytes", firstNonEmpty(metadataNumberValue(result, "size_bytes"), metadataNumberValue(result, "content_bytes"))),
			toolField("truncated", metadataBoolValue(result, "truncated")),
		)
	case "artifact.list":
		return joinToolFields(toolField("items", metadataLenValue(result, "artifacts")))
	case "web.fetch":
		return joinToolFields(
			toolField("status", metadataNumberValue(result, "status_code")),
			toolField("url", metadataStringValue(result, "url")),
			toolField("bytes", firstNonEmpty(metadataNumberValue(result, "body_bytes"), metadataNumberValue(result, "content_bytes"))),
		)
	case "web.search":
		return joinToolFields(
			toolField("results", metadataLenValue(result, "results")),
			toolField("provider", metadataStringValue(result, "provider")),
		)
	case "todo.write":
		return joinToolFields(toolField("items", metadataLenValue(result, "items")))
	case "tool.load_tools":
		return joinToolFields(toolField("tools", metadataLenValue(result, "tools")))
	case "agent.spawn", "agent.complete":
		return joinToolFields(
			toolField("task", metadataStringValue(result, "task_id")),
			toolField("status", metadataStringValue(result, "status")),
		)
	case "agent.list":
		return joinToolFields(toolField("tasks", metadataLenValue(result, "tasks")))
	case "lsp.diagnostics", "lsp.symbols", "lsp.references", "lsp.definition", "lsp.hover":
		return joinToolFields(
			toolField("path", metadataStringValue(result, "path")),
			toolField("items", firstNonEmpty(metadataLenValue(result, "items"), metadataLenValue(result, "diagnostics"), metadataLenValue(result, "symbols"), metadataLenValue(result, "references"))),
			toolField("truncated", metadataBoolValue(result, "truncated")),
		)
	}
	return ""
}

func compactMetadataValue(value any) string {
	if value == nil {
		return ""
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	text := strings.Join(strings.Fields(string(raw)), " ")
	const max = 240
	if len(text) > max {
		return text[:max-3] + "..."
	}
	return text
}

func metadataMapValue(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func metadataStringValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return strings.TrimSpace(text)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func metadataBoolValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	if typed, ok := value.(bool); ok {
		return fmt.Sprint(typed)
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func metadataNumberValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case int:
		return fmt.Sprint(typed)
	case int64:
		return fmt.Sprint(typed)
	case float64:
		if typed == float64(int64(typed)) {
			return fmt.Sprint(int64(typed))
		}
		return fmt.Sprint(typed)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func metadataLenValue(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case []any:
		return fmt.Sprint(len(typed))
	case []string:
		return fmt.Sprint(len(typed))
	case []map[string]any:
		return fmt.Sprint(len(typed))
	default:
		return ""
	}
}

func joinStringListValue(value any, sep string) string {
	if value == nil {
		return ""
	}
	switch typed := value.(type) {
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, fmt.Sprint(item))
		}
		return strings.Join(items, sep)
	case []string:
		return strings.Join(typed, sep)
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func joinToolFields(fields ...string) string {
	kept := make([]string, 0, len(fields))
	for _, field := range fields {
		if field != "" {
			kept = append(kept, field)
		}
	}
	return strings.Join(kept, " ")
}

func toolField(name, value string) string {
	if value == "" {
		return ""
	}
	return name + "=" + value
}

func quoteToolText(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	const max = 80
	if len(text) > max {
		text = text[:max-3] + "..."
	}
	return fmt.Sprintf("%q", text)
}

func firstLine(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	if index := strings.IndexByte(text, '\n'); index >= 0 {
		return strings.TrimSpace(text[:index])
	}
	return text
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func (r Renderer) PrintJSON(value any) error {
	encoder := json.NewEncoder(r.out())
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func (r Renderer) PrintJSONLine(value any) error {
	return json.NewEncoder(r.out()).Encode(value)
}

func (r Renderer) out() io.Writer {
	if r.Out == nil {
		return io.Discard
	}
	return r.Out
}
