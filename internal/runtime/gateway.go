package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type Gateway struct {
	Service     productdata.Service
	Broadcaster *Broadcaster
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
	g.Providers = replaceProvider(g.Providers, NewHTTPProvider(provider, http.DefaultClient))
	return provider
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
	g.streamProviderResponse(ctx, run, provider, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, Messages: messages, Model: selectedModel(input.Model, provider.Config().Model)}, "initial", true)
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
	g.streamProviderResponse(ctx, run, provider, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, Messages: messages, Model: selectedModel(input.Model, provider.Config().Model)}, "continuation", false)
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
				g.fail(ctx, run.ID, "unsupported_tool_loop", "Additional tool calls are not supported in this run.")
				return
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
			g.fail(ctx, run.ID, "provider_rate_limited", "Provider rate limit reached.")
			return
		case ProviderEventEmptyResponse:
			g.fail(ctx, run.ID, "empty_response", "Model returned an empty response.")
			return
		case ProviderEventMisconfigured:
			g.fail(ctx, run.ID, "provider_misconfigured", fallbackMessage(event.Message, "Provider configuration is incomplete."))
			return
		case ProviderEventError:
			g.fail(ctx, run.ID, fallbackMessage(event.ErrorCode, "provider_error"), fallbackMessage(event.Message, "Provider request failed."))
			return
		}
	}
	g.fail(ctx, run.ID, "empty_response", "Model returned an empty response.")
}

func (g *Gateway) selectProvider(providerID string) (Provider, error) {
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
	}
	_, events, err := g.Service.RecordToolCallRequest(ctx, identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: event.ToolName, CandidateSchemaHash: candidateSchemaHash, ArgumentsSummary: arguments, ArgumentsHash: argumentsHash(arguments), ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked})
	if err != nil {
		return false
	}
	if g.Broadcaster != nil {
		for _, recorded := range events {
			g.Broadcaster.Publish(recorded)
		}
	}
	return true
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
	_ = g.append(ctx, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryError, Type: code, Summary: message, ErrorCode: code, ErrorMessage: message})
	_ = g.append(ctx, runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_failed", Summary: message, ErrorCode: code, ErrorMessage: message})
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
