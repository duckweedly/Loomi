package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "model_request_started", Summary: "Model request started", Metadata: providerMetadata(provider.Config())}) {
		return
	}
	messages, err := g.loadRequestMessages(ctx, input.ThreadID, input.MessageID)
	if err != nil {
		g.fail(ctx, run.ID, "invalid_request", "Model request context could not be loaded.")
		return
	}
	events, err := provider.Stream(ctx, ProviderRequest{ThreadID: input.ThreadID, MessageID: input.MessageID, Messages: messages, Model: selectedModel(input.Model, provider.Config().Model)})
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
			if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryMessage, Type: "model_output_delta", Summary: "Model output delta", Content: &event.Text, Metadata: mergeMetadata(provider.Config(), event.Metadata)}) {
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
			assistantMetadata := mergeMetadata(provider.Config(), event.Metadata)
			assistantMetadata["run_id"] = run.ID
			if _, err := g.Service.AppendAssistantMessage(ctx, identity.LocalDevIdentity(), run.ThreadID, productdata.AppendAssistantMessageInput{Content: content, Metadata: assistantMetadata}); err != nil {
				g.fail(ctx, run.ID, "assistant_message_persist_failed", "Assistant message could not be persisted.")
				return
			}
			if !g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryMessage, Type: "model_output_completed", Summary: "Model output completed", Content: &content, Metadata: mergeMetadata(provider.Config(), event.Metadata)}) {
				return
			}
			g.append(ctx, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed", Metadata: providerMetadata(provider.Config())})
			return
		case ProviderEventToolCall:
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
	_, events, err := g.Service.RecordToolCallRequest(ctx, identity.LocalDevIdentity(), run.ID, productdata.RecordToolCallRequestInput{ToolCallID: toolCallID, ToolName: event.ToolName, ArgumentsSummary: arguments, ArgumentsHash: argumentsHash(arguments), ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked})
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

func argumentsHash(arguments map[string]any) string {
	value := "timezone=UTC"
	if timezone, ok := arguments["timezone"].(string); ok {
		value = "timezone=" + timezone
	}
	sum := sha256.Sum256([]byte(value))
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
