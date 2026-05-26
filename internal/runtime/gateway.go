package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"sync"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type Gateway struct {
	Service     productdata.Service
	Broadcaster *Broadcaster
	mu          sync.RWMutex
	Providers   []Provider
}

type GatewayRunInput struct {
	ThreadID   string
	MessageID  string
	ProviderID string
	Model      string
}

type GatewayContinuationInput struct {
	ThreadID   string
	MessageID  string
	ProviderID string
	Model      string
	ToolCallID string
}

func NewGateway(service productdata.Service, broadcaster *Broadcaster, providers []Provider) *Gateway {
	return &Gateway{Service: service, Broadcaster: broadcaster, Providers: providers}
}

func (g *Gateway) SaveProviderConfig(provider ProviderConfig) ProviderConfig {
	if g == nil {
		return provider
	}
	provider.ID = "custom"
	provider.Family = ProviderFamilyOpenAICompatible
	provider.Enabled = true
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Providers = replaceProvider(g.Providers, NewHTTPProvider(provider, http.DefaultClient))
	return provider
}

func (g *Gateway) SaveProvider(provider Provider) {
	if g == nil || provider == nil {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Providers = replaceProvider(g.Providers, provider)
}

func (g *Gateway) RemoveProvider(providerID string) {
	if g == nil || strings.TrimSpace(providerID) == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Providers = removeProvider(g.Providers, providerID)
}

func (g *Gateway) RunAsync(_ context.Context, run productdata.Run, input GatewayRunInput) {
	if g == nil || g.Service == nil || run.Source != productdata.RunSourceModelGateway {
		return
	}
	go g.run(context.Background(), run, input)
}

func (g *Gateway) run(ctx context.Context, run productdata.Run, input GatewayRunInput) {
	provider, err := g.selectProvider(input.ProviderID)
	if err != nil {
		g.fail(ctx, run.ID, "provider_misconfigured", "Provider configuration is incomplete.")
		return
	}
	capability := provider.Config().Capability()
	if capability.Status != ProviderStatusAvailable {
		g.fail(ctx, run.ID, string(capability.Status), capability.Message)
		return
	}
	messages, err := g.loadRequestMessages(ctx, input.ThreadID, input.MessageID)
	if err != nil {
		g.fail(ctx, run.ID, "invalid_request", "Model request context could not be loaded.")
		return
	}
	g.streamProviderResponse(ctx, run, provider, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, Messages: messages, Model: selectedModel(input.Model, provider.Config().Model), Tools: g.providerToolsForRun(ctx, run.ID)}, "initial", true)
}

func (g *Gateway) ContinueAfterToolResult(ctx context.Context, run productdata.Run, input GatewayContinuationInput) {
	provider, err := g.selectProvider(input.ProviderID)
	if err != nil {
		g.fail(ctx, run.ID, "provider_misconfigured", "Provider configuration is incomplete.")
		return
	}
	capability := provider.Config().Capability()
	if capability.Status != ProviderStatusAvailable {
		g.fail(ctx, run.ID, string(capability.Status), capability.Message)
		return
	}
	messages, err := g.loadContinuationMessages(ctx, input.ThreadID, input.MessageID, run.ID, input.ToolCallID)
	if err != nil {
		g.fail(ctx, run.ID, "tool_result_context_unavailable", "Tool result context could not be loaded.")
		return
	}
	g.streamProviderResponse(ctx, run, provider, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, Messages: messages, Model: selectedModel(input.Model, provider.Config().Model), Tools: g.providerToolsForContinuation(ctx, run.ID)}, "continuation", false)
}

func (g *Gateway) streamProviderResponse(ctx context.Context, run productdata.Run, provider Provider, request ProviderRequest, modelPhase string, allowToolCalls bool) {
	metadata := providerMetadata(provider.Config())
	metadata["model_phase"] = modelPhase
	if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "model_request_started", Summary: "Model request started", Metadata: metadata}) {
		return
	}
	events, err := provider.Stream(ctx, request)
	if err != nil {
		g.fail(ctx, run.ID, "provider_error", "Provider request failed.")
		return
	}
	var final strings.Builder
	for event := range events {
		if productdata.IsRunTerminal(g.currentStatus(ctx, run.ID)) {
			return
		}
		switch event.Type {
		case ProviderEventTextDelta:
			final.WriteString(event.Text)
			if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryMessage, Type: "model_output_delta", Summary: "Model output delta", Content: &event.Text, Metadata: mergePhaseMetadata(provider.Config(), event.Metadata, modelPhase)}) {
				return
			}
		case ProviderEventCompleted:
			content := final.String()
			if event.Text != "" {
				content = event.Text
			}
			if strings.TrimSpace(content) == "" {
				g.fail(ctx, run.ID, "empty_response", "Model returned an empty response.")
				return
			}
			assistantMetadata := mergePhaseMetadata(provider.Config(), event.Metadata, modelPhase)
			assistantMetadata["run_id"] = run.ID
			if _, err := g.Service.AppendAssistantMessage(ctx, identity.LocalDevIdentity(), run.ThreadID, productdata.AppendAssistantMessageInput{Content: content, Metadata: assistantMetadata}); err != nil {
				g.fail(ctx, run.ID, "assistant_message_persist_failed", "Assistant message could not be persisted.")
				return
			}
			if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryMessage, Type: "model_output_completed", Summary: "Model output completed", Content: &content, Metadata: mergePhaseMetadata(provider.Config(), event.Metadata, modelPhase)}) {
				return
			}
			g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed", Metadata: providerMetadata(provider.Config())})
			return
		case ProviderEventToolCall:
			if !allowToolCalls {
				if g.continuationToolCallIDExists(ctx, run.ID, event) {
					g.fail(ctx, run.ID, "duplicate_tool_call_id", "Tool call id was already used in this run.")
					return
				}
				if g.continuationToolLimitReached(ctx, run, event) {
					g.failWithMetadata(ctx, run.ID, "tool_loop_limit_reached", "Tool loop limit reached.", map[string]any{"loop_count": productdata.DefaultMaxBoundedToolCallsPerRun, "max_tool_calls": productdata.DefaultMaxBoundedToolCallsPerRun})
					return
				}
				if !g.canRequestContinuationTool(ctx, run, event) {
					g.fail(ctx, run.ID, "unsupported_tool_loop", "Additional tool calls are not supported in this run.")
					return
				}
			}
			if !g.recordToolCallRequest(ctx, run, event) {
				g.fail(ctx, run.ID, "tool_call_rejected", "Tool request could not be accepted.")
			}
			return
		case ProviderEventRefusal:
			g.fail(ctx, run.ID, "model_refusal", fallbackMessage(event.Message, "Model response was refused."))
			return
		case ProviderEventTimeout:
			g.fail(ctx, run.ID, "provider_timeout", "Provider request timed out.")
			return
		case ProviderEventRateLimited:
			g.failWithMetadata(ctx, run.ID, "provider_rate_limited", "Provider rate limit reached.", event.Metadata)
			return
		case ProviderEventEmptyResponse:
			g.fail(ctx, run.ID, "empty_response", "Model returned an empty response.")
			return
		case ProviderEventMisconfigured:
			g.failWithMetadata(ctx, run.ID, "provider_misconfigured", fallbackMessage(event.Message, "Provider configuration is incomplete."), event.Metadata)
			return
		case ProviderEventError:
			g.failWithMetadata(ctx, run.ID, fallbackMessage(event.ErrorCode, "provider_error"), fallbackMessage(event.Message, "Provider request failed."), event.Metadata)
			return
		}
	}
	g.fail(ctx, run.ID, "empty_response", "Model returned an empty response.")
}

func (g *Gateway) selectProvider(providerID string) (Provider, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, provider := range g.Providers {
		if provider.Config().ID == providerID {
			return provider, nil
		}
	}
	return nil, errors.New("provider not found")
}

func replaceProvider(providers []Provider, provider Provider) []Provider {
	config := provider.Config()
	for index, candidate := range providers {
		if candidate.Config().ID == config.ID {
			next := append([]Provider{}, providers...)
			next[index] = provider
			return next
		}
	}
	return append(providers, provider)
}

func removeProvider(providers []Provider, providerID string) []Provider {
	next := make([]Provider, 0, len(providers))
	for _, provider := range providers {
		if provider.Config().ID != providerID {
			next = append(next, provider)
		}
	}
	return next
}

func (g *Gateway) loadRequestMessages(ctx context.Context, threadID string, triggerMessageID string) ([]ProviderMessage, error) {
	messages, err := g.Service.ListMessages(ctx, identity.LocalDevIdentity(), threadID)
	if err != nil {
		return nil, err
	}
	result := make([]ProviderMessage, 0, len(messages))
	seenTrigger := triggerMessageID == ""
	for _, message := range messages {
		role := "user"
		if message.Role == productdata.MessageRoleAssistant {
			role = "assistant"
		}
		result = append(result, ProviderMessage{Role: role, Content: message.Content})
		if message.ID == triggerMessageID {
			seenTrigger = true
			break
		}
	}
	if !seenTrigger || len(result) == 0 {
		return nil, productdata.NewError(productdata.CodeInvalidRequest, "Message not found.")
	}
	return result, nil
}

func (g *Gateway) loadContinuationMessages(ctx context.Context, threadID string, triggerMessageID string, runID string, toolCallID string) ([]ProviderMessage, error) {
	messages, err := g.loadRequestMessages(ctx, threadID, triggerMessageID)
	if err != nil {
		return nil, err
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return nil, err
	}
	toolCall, result, err := continuationToolResult(events, toolCallID)
	if err != nil {
		return nil, err
	}
	resultContent, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	messages = append(messages, ProviderMessage{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: toolCall.ToolCallID, ToolName: toolCall.ToolName, ArgumentsSummary: toolCall.ArgumentsSummary})
	messages = append(messages, ProviderMessage{Role: ProviderMessageRoleToolResult, ToolCallID: toolCall.ToolCallID, ToolName: toolCall.ToolName, Content: string(resultContent)})
	return messages, nil
}

func (g *Gateway) append(ctx context.Context, runID string, input productdata.AppendRunEventInput) bool {
	event, err := g.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), runID, input)
	if err != nil {
		return false
	}
	if g.Broadcaster != nil {
		g.Broadcaster.Publish(event)
	}
	return true
}

func (g *Gateway) recordToolCallRequest(ctx context.Context, run productdata.Run, event ProviderEvent) bool {
	toolCallID := metadataString(event.Metadata, "tool_call_id")
	if toolCallID == "" {
		toolCallID = "tc_1"
	}
	arguments := toolArgumentsSummary(event.Metadata)
	candidateSchemaHash := ""
	if IsMCPToolName(event.ToolName) {
		var allowed bool
		allowed, candidateSchemaHash = g.mcpToolAllowedForRun(ctx, run.ID, event.ToolName)
		if !allowed {
			return false
		}
	} else if (productdata.IsDiscoveryToolName(event.ToolName) || productdata.IsWorkspaceToolName(event.ToolName) || productdata.IsSandboxToolName(event.ToolName) || productdata.IsLSPToolName(event.ToolName) || productdata.IsWebToolName(event.ToolName) || productdata.IsBrowserToolName(event.ToolName) || productdata.IsArtifactToolName(event.ToolName) || productdata.IsAgentToolName(event.ToolName) || productdata.IsTodoToolName(event.ToolName)) && !g.scopedToolAllowedForRun(ctx, run.ID, event.ToolName) {
		return false
	}
	approvalStatus := productdata.ToolCallApprovalRequired
	executionStatus := productdata.ToolCallExecutionBlocked
	if autoApproveToolCall(event.ToolName) {
		approvalStatus = productdata.ToolCallApprovalApproved
		executionStatus = productdata.ToolCallExecutionNotStarted
	}
	_, events, err := g.Service.RecordToolCallRequest(ctx, identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: event.ToolName, CandidateSchemaHash: candidateSchemaHash, ArgumentsSummary: arguments, ArgumentsHash: argumentsHash(arguments), ApprovalStatus: approvalStatus, ExecutionStatus: executionStatus})
	if err != nil {
		return false
	}
	if todo, ok := appendWorkTodoSnapshot(ctx, g.Service, run, "runtime"); ok {
		events = append(events, todo)
	}
	if g.Broadcaster != nil {
		for _, recorded := range events {
			g.Broadcaster.Publish(recorded)
		}
	}
	return true
}

func autoApproveToolCall(toolName string) bool {
	return toolName == productdata.ToolNameWebSearch || productdata.IsDiscoveryToolName(toolName)
}

func (g *Gateway) canRequestContinuationTool(ctx context.Context, run productdata.Run, event ProviderEvent) bool {
	if event.ToolName == "" || !g.continuationToolNameSupported(event.ToolName) {
		return false
	}
	if !g.scopedToolAllowedForRun(ctx, run.ID, event.ToolName) {
		return false
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		return false
	}
	return acceptedToolCallCount(events) < productdata.DefaultMaxBoundedToolCallsPerRun
}

func acceptedToolCallCount(events []productdata.RunEvent) int {
	seen := map[string]bool{}
	for _, event := range events {
		if event.Type != productdata.EventToolCallRequested {
			continue
		}
		toolCallID := metadataString(event.Metadata, "tool_call_id")
		if toolCallID == "" || seen[toolCallID] {
			continue
		}
		seen[toolCallID] = true
	}
	return len(seen)
}

func (g *Gateway) continuationToolLimitReached(ctx context.Context, run productdata.Run, event ProviderEvent) bool {
	if event.ToolName == "" || !g.continuationToolNameSupported(event.ToolName) {
		return false
	}
	if !g.scopedToolAllowedForRun(ctx, run.ID, event.ToolName) {
		return false
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		return false
	}
	return acceptedToolCallCount(events) >= productdata.DefaultMaxBoundedToolCallsPerRun
}

func (g *Gateway) continuationToolNameSupported(toolName string) bool {
	return productdata.IsDiscoveryToolName(toolName) ||
		productdata.IsWorkspaceToolName(toolName) ||
		productdata.IsSandboxToolName(toolName) ||
		productdata.IsLSPToolName(toolName) ||
		productdata.IsWebToolName(toolName) ||
		productdata.IsBrowserToolName(toolName) ||
		productdata.IsArtifactToolName(toolName) ||
		productdata.IsAgentToolName(toolName) ||
		productdata.IsTodoToolName(toolName)
}

func (g *Gateway) providerToolsForRun(ctx context.Context, runID string) []ProviderToolDefinition {
	if g == nil || g.Service == nil {
		return nil
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return nil
	}
	return providerToolsFromEvents(events)
}

func (g *Gateway) providerToolsForContinuation(ctx context.Context, runID string) []ProviderToolDefinition {
	if g == nil || g.Service == nil {
		return nil
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil || acceptedToolCallCount(events) >= productdata.DefaultMaxBoundedToolCallsPerRun {
		return nil
	}
	return providerToolsFromEvents(events)
}

func providerToolsFromEvents(events []productdata.RunEvent) []ProviderToolDefinition {
	enabled := map[string]bool{}
	ordered := []string{}
	for _, event := range events {
		for _, name := range metadataStringList(event.Metadata, "enabled_tools") {
			if !enabled[name] {
				ordered = append(ordered, name)
			}
			enabled[name] = true
		}
	}
	tools := make([]ProviderToolDefinition, 0, len(ordered))
	for _, name := range ordered {
		tool, ok := builtinProviderToolDefinition(name)
		if !ok {
			continue
		}
		tools = append(tools, tool)
	}
	return tools
}

func builtinProviderToolDefinition(name string) (ProviderToolDefinition, bool) {
	switch name {
	case productdata.ToolNameLoadTools:
		return providerTool(name, "Return safe descriptions for currently enabled Loomi tools by name or keyword.", map[string]any{"queries": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 5}, "names": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 20}, "limit": integerSchema(1, 30)}, []string{}), true
	case productdata.ToolNameLoadSkill:
		return providerTool(name, "Return a safe installed skill summary by name without loading the instruction body.", map[string]any{"name": stringSchema("Installed skill name or keyword."), "limit": integerSchema(1, 20)}, []string{"name"}), true
	case productdata.ToolNameWorkspaceGlob:
		return providerTool(name, "Find files under the workspace root.", map[string]any{"pattern": stringSchema("Glob pattern."), "path": stringSchema("Optional relative directory."), "limit": integerSchema(1, 500)}, []string{"pattern"}), true
	case productdata.ToolNameWorkspaceGrep:
		return providerTool(name, "Search text files under the workspace root.", map[string]any{"query": stringSchema("Search query."), "path": stringSchema("Optional relative directory."), "include": stringSchema("Optional file glob."), "case_sensitive": map[string]any{"type": "boolean"}, "limit": integerSchema(1, 500)}, []string{"query"}), true
	case productdata.ToolNameWorkspaceRead:
		return providerTool(name, "Read a bounded UTF-8 slice from one workspace file.", map[string]any{"path": stringSchema("Relative file path."), "offset": integerSchema(0, 1000000), "limit": integerSchema(1, 1000000), "max_bytes": integerSchema(1, 131072)}, []string{"path"}), true
	case productdata.ToolNameWorkspaceWriteFile:
		return providerTool(name, "Create a new bounded UTF-8 text file under the workspace root.", map[string]any{"path": stringSchema("Relative file path."), "content": stringSchema("File content."), "max_bytes": integerSchema(1, 131072)}, []string{"path", "content"}), true
	case productdata.ToolNameWorkspaceEdit:
		return providerTool(name, "Apply one bounded replacement in a workspace file after reading it.", map[string]any{"path": stringSchema("Relative file path."), "old_text": stringSchema("Existing text to replace exactly once."), "new_text": stringSchema("Replacement text."), "max_bytes": integerSchema(1, 131072)}, []string{"path", "old_text", "new_text"}), true
	case productdata.ToolNameWorkspacePatchPreview:
		return providerTool(name, "Preview one bounded replacement in a workspace file after reading it.", map[string]any{"path": stringSchema("Relative file path."), "old_text": stringSchema("Existing text to replace exactly once."), "new_text": stringSchema("Replacement text."), "max_bytes": integerSchema(1, 131072)}, []string{"path", "old_text", "new_text"}), true
	case productdata.ToolNameWorkspacePatchApply:
		return providerTool(name, "Apply one previously previewed bounded replacement in a workspace file.", map[string]any{"path": stringSchema("Relative file path."), "old_text": stringSchema("Existing text to replace exactly once."), "new_text": stringSchema("Replacement text."), "max_bytes": integerSchema(1, 131072)}, []string{"path", "old_text", "new_text"}), true
	case productdata.ToolNameSandboxExecCommand:
		return providerTool(name, "Run one approved argv-form read or validation command under the workspace root.", map[string]any{"argv": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "minItems": 1}, "cwd": stringSchema("Optional relative working directory."), "timeout_ms": integerSchema(1000, 30000), "max_output_bytes": integerSchema(1, 32768)}, []string{"argv"}), true
	case productdata.ToolNameSandboxStartProcess:
		return providerTool(name, "Start one approved argv-form read or validation process under the workspace root.", map[string]any{"argv": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "minItems": 1}, "cwd": stringSchema("Optional relative working directory."), "timeout_ms": integerSchema(1000, 120000), "max_output_bytes": integerSchema(1, 65536), "stdin": map[string]any{"type": "boolean"}}, []string{"argv"}), true
	case productdata.ToolNameSandboxContinueProcess:
		return providerTool(name, "Read current output/status for one run-scoped sandbox process, optionally writing bounded stdin.", map[string]any{"process_id": stringSchema("Sandbox process id returned by sandbox.start_process."), "cursor": integerSchema(0, 65536), "stdin_text": stringSchema("Optional bounded stdin text for stdin-enabled processes."), "input_seq": integerSchema(1, 1000000), "close_stdin": map[string]any{"type": "boolean"}}, []string{"process_id"}), true
	case productdata.ToolNameSandboxTerminateProcess:
		return providerTool(name, "Terminate one run-scoped sandbox process.", map[string]any{"process_id": stringSchema("Sandbox process id returned by sandbox.start_process.")}, []string{"process_id"}), true
	case productdata.ToolNameLSPDiagnostics:
		return providerTool(name, "Read bounded diagnostics for a workspace source file.", map[string]any{"path": stringSchema("Relative source file path."), "language": stringSchema("Optional language id."), "limit": integerSchema(1, 100)}, []string{"path"}), true
	case productdata.ToolNameLSPSymbols:
		return providerTool(name, "Read bounded symbol summaries for a workspace source file.", map[string]any{"path": stringSchema("Relative source file path."), "query": stringSchema("Optional symbol query."), "language": stringSchema("Optional language id."), "limit": integerSchema(1, 100)}, []string{"path"}), true
	case productdata.ToolNameLSPReferences:
		return providerTool(name, "Read bounded references for a source position.", map[string]any{"path": stringSchema("Relative source file path."), "line": integerSchema(1, 1000000), "column": integerSchema(1, 1000000), "include_declaration": map[string]any{"type": "boolean"}, "limit": integerSchema(1, 100)}, []string{"path", "line", "column"}), true
	case productdata.ToolNameLSPDefinition:
		return providerTool(name, "Find a bounded best-effort definition for a source position.", map[string]any{"path": stringSchema("Relative source file path."), "line": integerSchema(1, 1000000), "column": integerSchema(1, 1000000), "language": stringSchema("Optional language id."), "limit": integerSchema(1, 100)}, []string{"path", "line", "column"}), true
	case productdata.ToolNameLSPHover:
		return providerTool(name, "Read a bounded best-effort hover summary for a source position.", map[string]any{"path": stringSchema("Relative source file path."), "line": integerSchema(1, 1000000), "column": integerSchema(1, 1000000), "language": stringSchema("Optional language id.")}, []string{"path", "line", "column"}), true
	case productdata.ToolNameWebSearch:
		return WebSearchProviderToolDefinition(), true
	case productdata.ToolNameWebFetch:
		return providerTool(name, "Fetch one bounded public HTTP(S) URL and return a safe text summary.", map[string]any{"url": stringSchema("Public HTTP(S) URL."), "max_bytes": integerSchema(1, 131072), "timeout_ms": integerSchema(1000, 30000)}, []string{"url"}), true
	case productdata.ToolNameBrowserOpen:
		return providerTool(name, "Open one bounded public HTTP(S) page in a run-scoped browser session.", map[string]any{"url": stringSchema("Public HTTP(S) URL."), "max_bytes": integerSchema(1, 131072), "timeout_ms": integerSchema(1000, 30000)}, []string{"url"}), true
	case productdata.ToolNameBrowserSnapshot:
		return providerTool(name, "Return the current safe snapshot for a run-scoped browser session.", map[string]any{"session_id": stringSchema("Browser session id.")}, []string{"session_id"}), true
	case productdata.ToolNameBrowserClickLink:
		return providerTool(name, "Navigate one safe link from a run-scoped browser session.", map[string]any{"session_id": stringSchema("Browser session id."), "link_index": integerSchema(0, 100), "max_bytes": integerSchema(1, 131072), "timeout_ms": integerSchema(1000, 30000)}, []string{"session_id", "link_index"}), true
	case productdata.ToolNameBrowserScreenshot:
		return providerTool(name, "Return a bounded text screenshot summary for a run-scoped browser session.", map[string]any{"session_id": stringSchema("Browser session id.")}, []string{"session_id"}), true
	case productdata.ToolNameBrowserType:
		return providerTool(name, "Record bounded text into a discovered input target in a run-scoped browser session.", map[string]any{"session_id": stringSchema("Browser session id."), "target": stringSchema("Input target from browser snapshot."), "text": stringSchema("Text to type.")}, []string{"session_id", "target", "text"}), true
	case productdata.ToolNameBrowserPress:
		return providerTool(name, "Record one bounded key press in a run-scoped browser session.", map[string]any{"session_id": stringSchema("Browser session id."), "key": map[string]any{"type": "string", "enum": []string{"Enter", "Escape", "Tab", "ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight"}}}, []string{"session_id", "key"}), true
	case productdata.ToolNameTodoWrite:
		itemSchema := map[string]any{"type": "object", "additionalProperties": false, "properties": map[string]any{"id": stringSchema("Stable todo id."), "title": stringSchema("Short todo title."), "status": map[string]any{"type": "string", "enum": []string{"pending", "running", "completed", "blocked", "failed"}}, "summary": stringSchema("Optional safe progress summary.")}, "required": []string{"id", "title", "status"}}
		return providerTool(name, "Replace the current Work plan todo snapshot with bounded safe todo items.", map[string]any{"items": map[string]any{"type": "array", "items": itemSchema, "minItems": 1, "maxItems": productdata.MaxWorkTodoItems}}, []string{"items"}), true
	default:
		return ProviderToolDefinition{}, false
	}
}

func providerTool(name string, description string, properties map[string]any, required []string) ProviderToolDefinition {
	return ProviderToolDefinition{
		Name:         name,
		ProviderName: providerToolName(name),
		Description:  description,
		Parameters: map[string]any{
			"type":                 "object",
			"additionalProperties": false,
			"properties":           properties,
			"required":             required,
		},
	}
}

func stringSchema(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}

func integerSchema(min int, max int) map[string]any {
	return map[string]any{"type": "integer", "minimum": min, "maximum": max}
}

func (g *Gateway) continuationToolCallIDExists(ctx context.Context, runID string, event ProviderEvent) bool {
	toolCallID := metadataString(event.Metadata, "tool_call_id")
	if toolCallID == "" {
		return false
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return false
	}
	for _, runEvent := range events {
		if runEvent.Type != productdata.EventToolCallRequested {
			continue
		}
		if metadataString(runEvent.Metadata, "tool_call_id") == toolCallID {
			return true
		}
	}
	return false
}

func (g *Gateway) workspaceToolAllowedForRun(ctx context.Context, runID string, toolName string) bool {
	return g.scopedToolAllowedForRun(ctx, runID, toolName)
}

func (g *Gateway) scopedToolAllowedForRun(ctx context.Context, runID string, toolName string) bool {
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return false
	}
	for _, event := range events {
		for _, name := range metadataStringList(event.Metadata, "enabled_tools") {
			if name == toolName {
				return true
			}
		}
	}
	return false
}

func (g *Gateway) mcpToolAllowedForRun(ctx context.Context, runID string, toolName string) (bool, string) {
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return false, ""
	}
	discovered := false
	allowed := false
	candidateSchemaHash := ""
	for _, event := range events {
		if event.Type == "mcp_discovery_succeeded" && metadataString(event.Metadata, "status") == "succeeded" {
			for _, name := range metadataStringList(event.Metadata, "candidate_names", "mcp_candidate_names") {
				if name == toolName {
					discovered = true
					candidateSchemaHash = mcpCandidateSchemaHash(event.Metadata, toolName)
				}
			}
		}
		for _, name := range metadataStringList(event.Metadata, "enabled_tools") {
			if name == toolName {
				allowed = true
			}
		}
	}
	return discovered && allowed && candidateSchemaHash != "", candidateSchemaHash
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
		if hash := metadataString(metadata, key); hash != "" {
			return hash
		}
	}
	return ""
}

func argumentsHash(arguments map[string]any) string {
	raw, err := json.Marshal(productdata.RedactEventMetadata(arguments))
	if err != nil {
		raw = []byte("{}")
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func toolArgumentsSummary(metadata map[string]any) map[string]any {
	arguments, ok := metadata["arguments_summary"].(map[string]any)
	if ok {
		return arguments
	}
	return map[string]any{}
}

func (g *Gateway) fail(ctx context.Context, runID string, code string, message string) {
	g.failWithMetadata(ctx, runID, code, message, nil)
}

func (g *Gateway) failWithMetadata(ctx context.Context, runID string, code string, message string, metadata map[string]any) {
	_ = g.append(ctx, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryError, Type: code, Summary: message, ErrorCode: code, ErrorMessage: message, Metadata: metadata})
	_ = g.append(ctx, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_failed", Summary: message, ErrorCode: code, ErrorMessage: message, Metadata: metadata})
}

func (g *Gateway) currentStatus(ctx context.Context, runID string) productdata.RunStatus {
	run, err := g.Service.GetRun(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return productdata.RunStatusFailed
	}
	return run.Status
}

func providerMetadata(provider ProviderConfig) map[string]any {
	return map[string]any{"provider_id": provider.ID, "provider_family": string(provider.Family), "model": provider.Model}
}

func mergeMetadata(provider ProviderConfig, metadata map[string]any) map[string]any {
	merged := providerMetadata(provider)
	for key, value := range metadata {
		merged[key] = value
	}
	return merged
}

func mergePhaseMetadata(provider ProviderConfig, metadata map[string]any, modelPhase string) map[string]any {
	merged := mergeMetadata(provider, metadata)
	if modelPhase != "" {
		merged["model_phase"] = modelPhase
	}
	return merged
}

type continuationToolCall struct {
	ToolCallID       string
	ToolName         string
	ArgumentsSummary map[string]any
}

func continuationToolResult(events []productdata.RunEvent, toolCallID string) (continuationToolCall, map[string]any, error) {
	var requested *continuationToolCall
	for _, event := range events {
		if event.Type != productdata.EventToolCallRequested && event.Type != productdata.EventToolCallSucceeded {
			continue
		}
		eventToolCallID := metadataString(event.Metadata, "tool_call_id")
		if toolCallID != "" && eventToolCallID != toolCallID {
			continue
		}
		if event.Type == productdata.EventToolCallRequested {
			requested = &continuationToolCall{ToolCallID: eventToolCallID, ToolName: metadataString(event.Metadata, "tool_name"), ArgumentsSummary: toolArgumentsSummary(event.Metadata)}
			continue
		}
		if requested == nil {
			return continuationToolCall{}, nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool call request was not found.")
		}
		result := metadataMap(event.Metadata, "result_for_model_redacted")
		if len(result) == 0 {
			result = metadataMap(event.Metadata, "result_summary")
		}
		if len(result) == 0 {
			return continuationToolCall{}, nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool result was not found.")
		}
		return *requested, productdata.RedactEventMetadata(result), nil
	}
	return continuationToolCall{}, nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool result was not found.")
}

func metadataMap(metadata map[string]any, key string) map[string]any {
	value, ok := metadata[key].(map[string]any)
	if ok {
		return value
	}
	return map[string]any{}
}

func metadataStringList(metadata map[string]any, keys ...string) []string {
	var values []string
	for _, key := range keys {
		switch typed := metadata[key].(type) {
		case []string:
			values = append(values, typed...)
		case []any:
			for _, item := range typed {
				if text, ok := item.(string); ok {
					values = append(values, strings.TrimSpace(text))
				}
			}
		}
	}
	return values
}

func selectedModel(override string, configured string) string {
	if override != "" {
		return override
	}
	return configured
}

func fallbackMessage(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
