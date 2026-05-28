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

const maxProviderToolResultContentBytes = 4096

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
	g.runWithContext(ctx, run, input, nil)
}

func (g *Gateway) runWithContext(ctx context.Context, run productdata.Run, input GatewayRunInput, prepared *productdata.RunContext) {
	provider, err := g.selectProvider(input.ProviderID)
	if err != nil {
		g.fail(ctx, run.ID, "provider_misconfigured", "Provider configuration is incomplete.")
		return
	}
	capability := provider.Config().Capability()
	if !providerStatusCanRun(capability.Status) {
		g.fail(ctx, run.ID, string(capability.Status), capability.Message)
		return
	}
	messages, err := g.loadRequestMessages(ctx, input.ThreadID, input.MessageID)
	if err != nil {
		g.fail(ctx, run.ID, "invalid_request", "Model request context could not be loaded.")
		return
	}
	prepared = g.withExternalMemorySnapshot(ctx, prepared, messages)
	g.streamProviderResponse(ctx, run, provider, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, SystemPrompt: runSystemPrompt(prepared), Messages: messages, Model: selectedModel(input.Model, provider.Config().Model), Tools: g.providerToolsForRun(ctx, run.ID)}, "initial", true)
}

func (g *Gateway) ContinueAfterToolResult(ctx context.Context, run productdata.Run, input GatewayContinuationInput) {
	provider, err := g.selectProvider(input.ProviderID)
	if err != nil {
		g.fail(ctx, run.ID, "provider_misconfigured", "Provider configuration is incomplete.")
		return
	}
	capability := provider.Config().Capability()
	if !providerStatusCanRun(capability.Status) {
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
			content = naturalLanguageFinalContent(content)
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
			if g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed", Metadata: providerMetadata(provider.Config())}) {
				_ = proposePostRunMemory(ctx, g.Service, identity.LocalDevIdentity(), run.ID)
			}
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
			if guard := g.toolPlannerGuardrail(ctx, run, event); guard != nil {
				g.failWithMetadata(ctx, run.ID, guard.Code, guard.Message, guard.Metadata)
				return
			}
			_, err := g.recordToolCallRequest(ctx, run, event)
			if err != nil {
				code, message := toolRequestFailure(err)
				g.failWithMetadata(ctx, run.ID, code, message, map[string]any{"tool_name": event.ToolName, "arguments_summary": toolArgumentsSummary(event.Metadata), "error_code": string(productdata.ErrorCode(err)), "error_message": productdata.RedactEventText(err.Error())})
				return
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

func toolRequestFailure(err error) (string, string) {
	if err == nil {
		return "tool_validation_failed", "Tool request failed validation."
	}
	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "not allowed") || strings.Contains(lower, "chat mode") || strings.Contains(lower, "work mode"):
		return "tool_call_rejected", "Permission was not granted for this tool in the current run. Switch to the right mode or enable the tool, then retry."
	case strings.Contains(lower, "workspace root") || strings.Contains(lower, "workspace folder"):
		return "tool_call_rejected", "Workspace access is not available for this run. Select a workspace folder, then retry."
	case message != "":
		return "tool_call_rejected", "Tool request failed validation: " + message
	default:
		return "tool_call_rejected", "Tool request failed validation."
	}
}

type toolPlannerGuardrail struct {
	Code     string
	Message  string
	Metadata map[string]any
}

func (g *Gateway) toolPlannerGuardrail(ctx context.Context, run productdata.Run, event ProviderEvent) *toolPlannerGuardrail {
	if g == nil || g.Service == nil || !productdata.IsWorkspaceToolName(event.ToolName) {
		return nil
	}
	thread, err := g.Service.GetThread(ctx, identity.LocalDevIdentity(), run.ThreadID)
	if err != nil || thread.Mode != productdata.ThreadModeWork {
		return nil
	}
	events, err := g.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		return nil
	}
	arguments := toolArgumentsSummary(event.Metadata)
	if hasNonTerminalToolCall(events) {
		return nil
	}
	if repeatedWorkspaceToolRequest(events, event.ToolName, arguments) {
		return &toolPlannerGuardrail{
			Code:    "tool_planner_guardrail",
			Message: "Repeated workspace tool request was blocked. Use the existing tool result or continue with a narrower next step.",
			Metadata: map[string]any{
				"tool_name":         event.ToolName,
				"arguments_summary": arguments,
				"guardrail":         "repeated_workspace_tool_arguments",
			},
		}
	}
	if !isFirstWorkspaceToolRequest(events) || !directoryInventoryIntent(g.latestUserMessage(ctx, run.ThreadID)) {
		return nil
	}
	if event.ToolName == productdata.ToolNameWorkspaceTreeSummary || event.ToolName == productdata.ToolNameWorkspaceListDirectory {
		return nil
	}
	return &toolPlannerGuardrail{
		Code:    "tool_planner_guardrail",
		Message: "Directory inventory should start with workspace_tree_summary or workspace_list_directory, not grep/glob/read.",
		Metadata: map[string]any{
			"tool_name":         event.ToolName,
			"arguments_summary": arguments,
			"guardrail":         "directory_inventory_first_tool",
			"recommended_tool":  productdata.ToolNameWorkspaceTreeSummary,
		},
	}
}

func (g *Gateway) latestUserMessage(ctx context.Context, threadID string) string {
	if g == nil || g.Service == nil {
		return ""
	}
	messages, err := g.Service.ListMessages(ctx, identity.LocalDevIdentity(), threadID)
	if err != nil {
		return ""
	}
	for index := len(messages) - 1; index >= 0; index-- {
		if messages[index].Role == productdata.MessageRoleUser {
			return messages[index].Content
		}
	}
	return ""
}

func isFirstWorkspaceToolRequest(events []productdata.RunEvent) bool {
	for _, event := range events {
		if event.Type == productdata.EventToolCallRequested && productdata.IsWorkspaceToolName(metadataString(event.Metadata, "tool_name")) {
			return false
		}
	}
	return true
}

func hasNonTerminalToolCall(events []productdata.RunEvent) bool {
	pending := map[string]bool{}
	for _, event := range events {
		toolCallID := metadataString(event.Metadata, "tool_call_id")
		if toolCallID == "" {
			continue
		}
		switch event.Type {
		case productdata.EventToolCallRequested:
			pending[toolCallID] = true
		case productdata.EventToolCallSucceeded, productdata.EventToolCallFailed, productdata.EventToolCallDenied, productdata.EventToolCallCancelled:
			delete(pending, toolCallID)
		}
	}
	return len(pending) > 0
}

func repeatedWorkspaceToolRequest(events []productdata.RunEvent, toolName string, arguments map[string]any) bool {
	switch toolName {
	case productdata.ToolNameWorkspaceRead, productdata.ToolNameWorkspaceListDirectory, productdata.ToolNameWorkspaceGrep:
	default:
		return false
	}
	hash := argumentsHash(arguments)
	stateChanged := false
	for index := len(events) - 1; index >= 0; index-- {
		event := events[index]
		if workspaceRepeatResetEvent(event) {
			stateChanged = true
		}
		if event.Type != productdata.EventToolCallRequested {
			continue
		}
		if metadataString(event.Metadata, "tool_name") != toolName {
			continue
		}
		previous, ok := event.Metadata["arguments_summary"].(map[string]any)
		if !ok {
			continue
		}
		if argumentsHash(previous) == hash {
			return !stateChanged
		}
	}
	return false
}

func workspaceRepeatResetEvent(event productdata.RunEvent) bool {
	if event.Type != productdata.EventToolCallSucceeded {
		return false
	}
	switch metadataString(event.Metadata, "tool_name") {
	case productdata.ToolNameWorkspaceWriteFile, productdata.ToolNameWorkspaceEdit, productdata.ToolNameWorkspacePatchApply:
		return true
	case productdata.ToolNameSandboxExecCommand, productdata.ToolNameSandboxStartProcess, productdata.ToolNameSandboxContinueProcess, productdata.ToolNameSandboxTerminateProcess:
		return true
	default:
		return false
	}
}

func directoryInventoryIntent(message string) bool {
	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	hasSubject := false
	for _, marker := range []string{"目录", "文件夹", "folder", "directory", "tree"} {
		if strings.Contains(lower, marker) {
			hasSubject = true
			break
		}
	}
	if !hasSubject {
		return false
	}
	for _, marker := range []string{"盘点", "分类", "有哪些", "有什么", "列出", "结构", "inventory", "classify", "what files", "what is in", "contains", "list"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func naturalLanguageFinalContent(input string) string {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || !looksLikeStructuredFinalPayload(trimmed) {
		return trimmed
	}
	var payload any
	if err := json.Unmarshal([]byte(trimmed), &payload); err == nil {
		if candidate, ok := extractStructuredFinalText(payload); ok {
			return candidate
		}
	}
	return "The tool run produced results, but the provider returned a structured payload instead of a final natural-language answer."
}

func extractStructuredFinalText(value any) (string, bool) {
	switch typed := value.(type) {
	case string:
		candidate := strings.TrimSpace(typed)
		if candidate != "" && !looksLikeStructuredFinalPayload(candidate) {
			return candidate, true
		}
	case []any:
		parts := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := extractStructuredFinalText(item); ok {
				parts = append(parts, text)
			}
		}
		if len(parts) > 0 {
			return strings.Join(parts, "\n\n"), true
		}
	case map[string]any:
		for _, key := range []string{"answer", "final", "message", "summary", "output_text", "content", "text", "result", "output"} {
			if nested, ok := typed[key]; ok {
				if text, ok := extractStructuredFinalText(nested); ok {
					return text, true
				}
			}
		}
	}
	return "", false
}

func looksLikeStructuredFinalPayload(input string) bool {
	if input == "" {
		return false
	}
	if strings.HasPrefix(input, "{") || strings.HasPrefix(input, "[") {
		return json.Valid([]byte(input))
	}
	lower := strings.ToLower(input)
	return strings.Contains(lower, `"tool_calls"`) || strings.Contains(lower, `"tool_call_id"`) || strings.Contains(lower, "<tool_call")
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
	results, err := continuationToolResults(events, toolCallID)
	if err != nil {
		return nil, err
	}
	for _, item := range results {
		resultContent, err := json.Marshal(compactToolResultPayload(item.Result, maxProviderToolResultContentBytes))
		if err != nil {
			return nil, err
		}
		messages = append(messages, ProviderMessage{Role: ProviderMessageRoleAssistantToolCall, ToolCallID: item.ToolCall.ToolCallID, ToolName: item.ToolCall.ToolName, ArgumentsSummary: item.ToolCall.ArgumentsSummary})
		messages = append(messages, ProviderMessage{Role: ProviderMessageRoleToolResult, ToolCallID: item.ToolCall.ToolCallID, ToolName: item.ToolCall.ToolName, Content: string(resultContent)})
	}
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

func (g *Gateway) recordToolCallRequest(ctx context.Context, run productdata.Run, event ProviderEvent) (bool, error) {
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
			return false, productdata.NewError(productdata.CodeInvalidRequest, "MCP tool is not allowed for this run.")
		}
	} else if (productdata.IsDiscoveryToolName(event.ToolName) || productdata.IsWorkspaceToolName(event.ToolName) || productdata.IsSandboxToolName(event.ToolName) || productdata.IsLSPToolName(event.ToolName) || productdata.IsWebToolName(event.ToolName) || productdata.IsBrowserToolName(event.ToolName) || productdata.IsArtifactToolName(event.ToolName) || productdata.IsAgentToolName(event.ToolName) || productdata.IsTodoToolName(event.ToolName)) && !g.scopedToolAllowedForRun(ctx, run.ID, event.ToolName) {
		return false, productdata.NewError(productdata.CodeInvalidRequest, "Tool is not allowed for this run.")
	}
	approvalStatus := productdata.ToolCallApprovalRequired
	executionStatus := productdata.ToolCallExecutionBlocked
	if autoApproveToolCall(event.ToolName) {
		approvalStatus = productdata.ToolCallApprovalApproved
		executionStatus = productdata.ToolCallExecutionNotStarted
	}
	_, events, err := g.Service.RecordToolCallRequest(ctx, identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: event.ToolName, CandidateSchemaHash: candidateSchemaHash, ArgumentsSummary: arguments, ArgumentsHash: argumentsHash(arguments), ApprovalStatus: approvalStatus, ExecutionStatus: executionStatus})
	if err != nil {
		return false, err
	}
	if todo, ok := appendWorkTodoSnapshot(ctx, g.Service, run, "runtime"); ok {
		events = append(events, todo)
	}
	if g.Broadcaster != nil {
		for _, recorded := range events {
			g.Broadcaster.Publish(recorded)
		}
	}
	return autoApproveToolCall(event.ToolName), nil
}

func autoApproveToolCall(toolName string) bool {
	return toolName == productdata.ToolNameWebSearch || toolName == productdata.ToolNameWebFetch || productdata.IsDiscoveryToolName(toolName) || productdata.IsWorkspaceReadOnlyToolName(toolName)
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
	tools := providerToolsFromEvents(events)
	if workspaceGlobSucceeded(events) {
		tools = omitProviderTool(tools, productdata.ToolNameWorkspaceGlob)
	}
	return tools
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

func workspaceGlobSucceeded(events []productdata.RunEvent) bool {
	for _, event := range events {
		if event.Type == productdata.EventToolCallSucceeded && metadataString(event.Metadata, "tool_name") == productdata.ToolNameWorkspaceGlob {
			return true
		}
	}
	return false
}

func omitProviderTool(tools []ProviderToolDefinition, name string) []ProviderToolDefinition {
	if name == "" || len(tools) == 0 {
		return tools
	}
	filtered := tools[:0]
	for _, tool := range tools {
		if tool.Name != name {
			filtered = append(filtered, tool)
		}
	}
	return filtered
}

func builtinProviderToolDefinition(name string) (ProviderToolDefinition, bool) {
	switch name {
	case productdata.ToolNameLoadTools:
		return providerTool(name, "query-only catalog lookup. Return safe descriptions for currently enabled Loomi tools by catalog keyword; omit query to list a bounded safe catalog.", map[string]any{"query": stringSchema("Optional catalog search phrase."), "queries": map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 5}, "limit": integerSchema(1, 30)}, []string{}), true
	case productdata.ToolNameLoadSkill:
		return providerTool(name, "Return a safe installed skill summary by name without loading the instruction body.", map[string]any{"name": stringSchema("Installed skill name or keyword."), "limit": integerSchema(1, 20)}, []string{"name"}), true
	case productdata.ToolNameWorkspaceGlob:
		return providerTool(name, "Find files under the selected workspace root. Use path \".\" for the selected folder; do not repeat the root folder name.", map[string]any{"pattern": stringSchema("Glob pattern."), "path": stringSchema("Optional relative directory from the selected workspace root. Use . for the root."), "limit": integerSchema(1, 500)}, []string{"pattern"}), true
	case productdata.ToolNameWorkspaceGrep:
		return providerTool(name, "Search text files under the selected workspace root. Use path \".\" for the selected folder; do not repeat the root folder name.", map[string]any{"query": stringSchema("Search query."), "path": stringSchema("Optional relative directory from the selected workspace root. Use . for the root."), "include": stringSchema("Optional file glob."), "case_sensitive": map[string]any{"type": "boolean"}, "limit": integerSchema(1, 500)}, []string{"query"}), true
	case productdata.ToolNameWorkspaceRead:
		return providerTool(name, "Read a bounded UTF-8 slice from one file under the selected workspace root. Paths are relative to that root.", map[string]any{"path": stringSchema("Relative file path from the selected workspace root; do not repeat the root folder name."), "offset": integerSchema(0, 1000000), "limit": integerSchema(1, 1000000), "max_bytes": integerSchema(1, 131072)}, []string{"path"}), true
	case productdata.ToolNameWorkspaceListDirectory:
		return providerTool(name, "Read a bounded directory listing under the selected workspace root. Use this before grep for folder listing, inventory, or classification questions. Paths are relative; use path \".\" for the selected folder.", map[string]any{"path": stringSchema("Relative directory from the selected workspace root. Use . for the selected folder."), "max_entries": integerSchema(1, 500), "depth": integerSchema(1, 3), "include_hidden": map[string]any{"type": "boolean"}, "sort": map[string]any{"type": "string", "enum": []string{"name", "modified", "size"}}}, []string{}), true
	case productdata.ToolNameWorkspaceTreeSummary:
		return providerTool(name, "Return a bounded classified summary of a directory tree. Prefer this over grep when the user asks what a folder contains or how files are grouped by kind.", map[string]any{"path": stringSchema("Relative directory from the selected workspace root. Use . for the selected folder."), "max_entries": integerSchema(1, 500), "depth": integerSchema(1, 3), "include_hidden": map[string]any{"type": "boolean"}, "sort": map[string]any{"type": "string", "enum": []string{"name", "modified", "size"}}}, []string{}), true
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
	case productdata.ToolNameArtifactCreateText:
		return providerTool(name, "Create one bounded non-executable text artifact for reports, articles, Markdown, or saveable documents. The result returns an artifacts array; cite its returned key only.", map[string]any{"title": stringSchema("Optional artifact title."), "filename": stringSchema("Optional display filename."), "mime_type": stringSchema("Optional MIME type."), "display": map[string]any{"type": "string", "enum": []string{"inline", "panel"}}, "content": stringSchema("Artifact text content."), "max_bytes": integerSchema(1, defaultArtifactMaxBytes)}, []string{"content"}), true
	case productdata.ToolNameArtifactRead:
		return providerTool(name, "Read one bounded text artifact excerpt by id without returning raw full content.", map[string]any{"artifact_id": stringSchema("Artifact id or key."), "max_bytes": integerSchema(1, defaultArtifactMaxBytes)}, []string{"artifact_id"}), true
	case productdata.ToolNameArtifactList:
		return providerTool(name, "List bounded safe artifact reference metadata for the current thread without raw content.", map[string]any{"limit": integerSchema(1, 50)}, []string{}), true
	case productdata.ToolNameMemorySearch:
		return providerTool(name, "Search approved Loomi memory summaries in the current safe scope.", map[string]any{"query": stringSchema("Memory search query."), "limit": integerSchema(1, 20), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_type": stringSchema("Optional source type filter.")}, []string{"query"}), true
	case productdata.ToolNameMemoryList:
		return providerTool(name, "List approved Loomi memory summaries in the current safe scope.", map[string]any{"limit": integerSchema(1, 20), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_type": stringSchema("Optional source type filter.")}, []string{}), true
	case productdata.ToolNameMemoryRead:
		return providerTool(name, "Read one approved Loomi memory summary without raw content.", map[string]any{"entry_id": stringSchema("Memory entry id."), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"entry_id"}), true
	case productdata.ToolNameMemoryWrite:
		return providerTool(name, "Create one approval-gated Loomi memory write proposal.", map[string]any{"title": stringSchema("Short memory title."), "content": stringSchema("Memory content to propose."), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_event_id": stringSchema("Optional source event id."), "idempotency_key": stringSchema("Optional idempotency key.")}, []string{"title", "content"}), true
	case productdata.ToolNameMemoryEdit:
		return providerTool(name, "Edit a pending Loomi memory proposal or create an approval-gated replacement proposal.", map[string]any{"proposal_id": stringSchema("Optional pending proposal id."), "entry_id": stringSchema("Optional existing memory entry id."), "title": stringSchema("Short memory title."), "content": stringSchema("Replacement safe memory content."), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_event_id": stringSchema("Optional source event id."), "idempotency_key": stringSchema("Optional idempotency key.")}, []string{"title", "content"}), true
	case productdata.ToolNameMemoryForget:
		return providerTool(name, "Tombstone one approved Loomi memory entry through the audited memory boundary.", map[string]any{"entry_id": stringSchema("Memory entry id."), "reason": stringSchema("Optional safe deletion reason."), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"entry_id"}), true
	case productdata.ToolNameMemoryContext:
		return providerTool(name, "Return Loomi memory provider status plus bounded relevant memory summaries.", map[string]any{"query": stringSchema("Optional memory query."), "limit": integerSchema(1, 20), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_type": stringSchema("Optional source type filter.")}, []string{}), true
	case productdata.ToolNameMemoryTimeline:
		return providerTool(name, "List safe Loomi memory audit timeline items.", map[string]any{"limit": integerSchema(1, 50), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id."), "source_type": stringSchema("Optional event type filter.")}, []string{}), true
	case productdata.ToolNameMemoryConnections:
		return providerTool(name, "Return bounded related Loomi memory summaries for one entry or query.", map[string]any{"entry_id": stringSchema("Optional memory entry id."), "query": stringSchema("Optional related memory query."), "limit": integerSchema(1, 20), "scope_type": stringSchema("Optional memory scope type."), "scope_id": stringSchema("Optional memory scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{}), true
	case productdata.ToolNameMemoryThreadSearch:
		return providerTool(name, "Search local thread and message history with safe excerpts.", map[string]any{"query": stringSchema("Thread search query."), "limit": integerSchema(1, 20)}, []string{"query"}), true
	case productdata.ToolNameMemoryThreadFetch:
		return providerTool(name, "Fetch safe local thread message excerpts.", map[string]any{"thread_id": stringSchema("Thread id."), "limit": integerSchema(1, 50)}, []string{"thread_id"}), true
	case productdata.ToolNameMemoryStatus:
		return providerTool(name, "Return Loomi memory provider readiness and configuration state.", map[string]any{}, []string{}), true
	case productdata.ToolNameNotebookRead:
		return providerTool(name, "Read one approved structured Loomi notebook entry without raw unsafe content.", map[string]any{"entry_id": stringSchema("Notebook entry id."), "scope_type": stringSchema("Optional notebook scope type."), "scope_id": stringSchema("Optional notebook scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"entry_id"}), true
	case productdata.ToolNameNotebookWrite:
		return providerTool(name, "Write one approval-gated structured Loomi notebook entry.", map[string]any{"title": stringSchema("Short notebook title."), "content": stringSchema("Notebook content to store."), "scope_type": stringSchema("Optional notebook scope type."), "scope_id": stringSchema("Optional notebook scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"title", "content"}), true
	case productdata.ToolNameNotebookEdit:
		return providerTool(name, "Replace one structured Loomi notebook entry by tombstoning the old entry and writing the new entry.", map[string]any{"entry_id": stringSchema("Notebook entry id."), "title": stringSchema("Short notebook title."), "content": stringSchema("Replacement notebook content."), "scope_type": stringSchema("Optional notebook scope type."), "scope_id": stringSchema("Optional notebook scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"entry_id", "title", "content"}), true
	case productdata.ToolNameNotebookForget:
		return providerTool(name, "Tombstone one structured Loomi notebook entry through the audited memory boundary.", map[string]any{"entry_id": stringSchema("Notebook entry id."), "reason": stringSchema("Optional safe deletion reason."), "scope_type": stringSchema("Optional notebook scope type."), "scope_id": stringSchema("Optional notebook scope id."), "source_thread_id": stringSchema("Optional source thread id."), "source_run_id": stringSchema("Optional source run id.")}, []string{"entry_id"}), true
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

type continuationToolResultItem struct {
	ToolCall continuationToolCall
	Result   map[string]any
}

func continuationToolResults(events []productdata.RunEvent, toolCallID string) ([]continuationToolResultItem, error) {
	requested := map[string]continuationToolCall{}
	completed := map[string]bool{}
	results := []continuationToolResultItem{}
	for _, event := range events {
		if event.Type != productdata.EventToolCallRequested && event.Type != productdata.EventToolCallSucceeded {
			continue
		}
		eventToolCallID := metadataString(event.Metadata, "tool_call_id")
		if eventToolCallID == "" {
			continue
		}
		if event.Type == productdata.EventToolCallRequested {
			requested[eventToolCallID] = continuationToolCall{ToolCallID: eventToolCallID, ToolName: metadataString(event.Metadata, "tool_name"), ArgumentsSummary: toolArgumentsSummary(event.Metadata)}
			continue
		}
		if completed[eventToolCallID] {
			continue
		}
		toolCall, ok := requested[eventToolCallID]
		if !ok {
			return nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool call request was not found.")
		}
		result := metadataMap(event.Metadata, "result_for_model_redacted")
		if len(result) == 0 {
			result = metadataMap(event.Metadata, "result_summary")
		}
		if len(result) == 0 {
			return nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool result was not found.")
		}
		completed[eventToolCallID] = true
		results = append(results, continuationToolResultItem{ToolCall: toolCall, Result: productdata.RedactEventMetadata(result)})
		if toolCallID != "" && eventToolCallID == toolCallID {
			return results, nil
		}
	}
	return nil, productdata.NewError(productdata.CodeInvalidRequest, "Tool result was not found.")
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

func runSystemPrompt(prepared *productdata.RunContext) string {
	base := "You are Loomi, a careful local assistant. Answer naturally and concisely."
	if prepared != nil && strings.TrimSpace(prepared.Persona.SystemPrompt) != "" {
		base = strings.TrimSpace(prepared.Persona.SystemPrompt)
	}
	policy := "\n\nOutput style:\n- Answer first, then give only the context needed to act.\n- No preface such as \"Sure\", \"Certainly\", or \"Here is\".\n- Do not repeat the user's request back to them.\n- For code changes, report what changed and what was verified; do not narrate every step.\n\nTool policy:\n- Use tools only when they are needed. Do not use tools for greetings, small talk, or stable general knowledge.\n- Use web_search for current events, latest information, news, prices, or external facts that may have changed.\n- Use web_fetch when the user gives a public URL or when search results need one source opened for analysis.\n- Do not call workspace, sandbox, LSP, browser, artifact, todo, or agent tools in Chat mode. Use those only when the current tool list truly exposes them for a Work run.\n- If a useful tool is not available, answer with the limitation and the next best safe option. Do not fabricate tool calls or tool results.\n- Final user-facing output must be natural language, not JSON or a tool protocol transcript."
	if prepared != nil && prepared.Thread.Mode == productdata.ThreadModeWork {
		workspaceLabel := strings.TrimSpace(prepared.WorkspaceRoot.DisplayName)
		if workspaceLabel == "" {
			workspaceLabel = productdata.WorkspaceDisplayNameFromPath(prepared.WorkspaceRoot.Path)
		}
		policy += "\n\nWork mode policy:\n- This is a Work run. File and folder tasks are tool-first: inspect/list/search/classify/summarize files with workspace_list_directory, workspace_tree_summary, workspace_grep, and workspace_read before answering.\n- For folder listing, inventory, or classification, start with workspace_tree_summary or workspace_list_directory using path \".\" and a bounded max_entries/depth. Do not start with grep for directory inventory.\n- Use workspace_grep only for content search. Read specific files with workspace_read only after you have a relative path.\n- Do not tell the user to run shell commands, paste directory listings, export files, or grant oral permission when workspace tools are available. Request the tool and continue from the result.\n- Do not use sandbox commands for file reads, globbing, grep, cat/head/tail/ls-style listing, or simple classification. Use sandbox commands only for build/test/lint or commands that truly need a process.\n- Workspace tool paths are always relative to the selected workspace root. The selected folder is already the root; use \".\" for it and do not repeat the root folder name such as Downloads, Desktop, or Documents in tool paths.\n- If a requested path is outside the selected root, ask the user to choose that folder in the UI.\n- Final output should summarize the work clearly and omit raw tool protocol details."
		if workspaceLabel != "" {
			policy += "\n\nWorkspace reference policy:\n- Selected workspace: " + productdata.RedactEventText(workspaceLabel) + "\n- When the user says current directory, this directory, selected directory, just selected directory, 当前目录, 这个目录, 刚选目录, use the selected workspace root and path \".\".\n- When the user says download directory or 下载目录, only treat it as Downloads when the selected workspace label is Downloads; otherwise ask the user to choose Downloads in the UI before using workspace tools.\n- Never infer Loomi, Arkloop, a previous thread folder, or the process working directory as the workspace for this run."
		} else {
			policy += "\n\nWorkspace reference policy:\n- No workspace label is available for this run. Ask the user to choose a workspace folder before using workspace tools for current directory, this directory, selected directory, just selected directory, 当前目录, 这个目录, 刚选目录, download directory, or 下载目录."
		}
		policy += "\n\nTool selection strategy:\n- Directory questions: use workspace_tree_summary or workspace_list_directory first with path \".\" and bounded max_entries/depth before summarizing categories.\n- Use workspace_glob only for file-name pattern matching or a narrow follow-up after directory inventory.\n- Content questions: use workspace_grep or workspace_read after you have a relative path; use grep to find candidates, then read the specific file.\n- Modification questions: use workspace_read first, then workspace_patch_preview, then workspace_patch_apply only after approval.\n- Use sandbox commands only when the user explicitly asks for a shell/process action or when verifying a change with build/test/lint."
		policy += "\n\nArtifact/document contract:\n- Reports, articles, Markdown, and saveable documents should use artifact.create_text instead of placing the full document only in the final reply.\n- Reference saved artifacts as [title](artifact:<key>) using a key returned by artifact.create_text or artifact.list.\n- Do not invent artifact keys."
	}
	return base + memoryPromptContext(prepared) + policy
}

func memoryPromptContext(prepared *productdata.RunContext) string {
	if prepared == nil {
		return ""
	}
	var builder strings.Builder
	appendSnapshotBlock := func(tag string, snapshot productdata.MemorySnapshot, include func(productdata.MemorySearchResult) bool) {
		wrote := false
		for _, entry := range snapshot.Entries {
			if !include(entry) {
				continue
			}
			if !wrote {
				builder.WriteString("\n\n<")
				builder.WriteString(tag)
				builder.WriteString(">\n")
				wrote = true
			}
			builder.WriteString("- ")
			builder.WriteString(productdata.RedactEventText(strings.TrimSpace(entry.Title)))
			if strings.TrimSpace(entry.Summary) != "" {
				builder.WriteString(": ")
				builder.WriteString(productdata.RedactEventText(strings.TrimSpace(entry.Summary)))
			}
			builder.WriteString("\n")
		}
		if wrote {
			builder.WriteString("</")
			builder.WriteString(tag)
			builder.WriteString(">")
		}
	}
	appendSnapshotBlock("memory", prepared.MemorySnapshot, func(entry productdata.MemorySearchResult) bool {
		return entry.SourceType != "notebook"
	})
	appendSnapshotBlock("notebook", prepared.NotebookSnapshot, func(entry productdata.MemorySearchResult) bool {
		return entry.SourceType == "notebook"
	})
	return builder.String()
}

func fallbackMessage(value string, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
