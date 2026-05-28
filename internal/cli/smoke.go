package cli

import (
	"context"
	"fmt"
	"strings"
)

type SmokeAgentOptions struct {
	ThreadID    string
	Prompt      string
	Mode        string
	Provider    string
	Model       string
	Persona     string
	Workspace   string
	AutoApprove bool
}

type SmokeAgentResult struct {
	OK                   bool               `json:"ok"`
	Stage                string             `json:"stage"`
	ThreadID             string             `json:"thread_id,omitempty"`
	RunID                string             `json:"run_id,omitempty"`
	FinalStage           string             `json:"final_stage,omitempty"`
	Status               string             `json:"status,omitempty"`
	BlockedReason        string             `json:"blocked_reason,omitempty"`
	Remedy               string             `json:"remedy,omitempty"`
	Provider             ProviderCapability `json:"provider,omitempty"`
	Workspace            string             `json:"workspace,omitempty"`
	EventCount           int                `json:"event_count,omitempty"`
	ToolEventCount       int                `json:"tool_event_count,omitempty"`
	ApprovalCount        int                `json:"approval_count,omitempty"`
	PendingApprovalCount int                `json:"pending_approval_count,omitempty"`
	ToolChain            []string           `json:"tool_chain,omitempty"`
	ToolOrder            []string           `json:"tool_order,omitempty"`
	FinalPersisted       bool               `json:"final_persisted,omitempty"`
	ReplayOK             bool               `json:"replay_ok,omitempty"`
	ReplayEventCount     int                `json:"replay_event_count,omitempty"`
	ReplayTerminalStage  string             `json:"replay_terminal_stage,omitempty"`
	FailureLogPath       string             `json:"failure_log_path,omitempty"`
	FinalMessage         string             `json:"final_message,omitempty"`
	LastEvents           []string           `json:"last_events,omitempty"`
}

func RunAgentSmoke(ctx context.Context, client *Client, cfg Config, opts SmokeAgentOptions) (SmokeAgentResult, error) {
	if client == nil {
		client = NewClientFromConfig(cfg)
	}
	result := SmokeAgentResult{Stage: "api_ready"}
	if err := client.CheckReady(ctx); err != nil {
		result.Stage = "api_ready"
		result.BlockedReason = "api_unavailable"
		result.Remedy = "Start loomi-api or set LOOMI_HOST / loomi config set host."
		return result, nil
	}
	if workspace := strings.TrimSpace(opts.Workspace); workspace != "" {
		result.Stage = "workspace_root"
		config, err := client.SaveWorkspaceRoot(ctx, workspace)
		if err != nil {
			result.BlockedReason = "workspace_root_unavailable"
			result.Remedy = err.Error()
			return result, nil
		}
		result.Workspace = config.DisplayName
	}

	providerID := strings.TrimSpace(firstNonEmpty(opts.Provider, cfg.Provider))
	if providerID == "" {
		providerID = "local_codex"
	}
	if providerID != "" {
		result.Stage = "provider_check"
		provider, err := client.CheckModelProvider(ctx, providerID)
		if err != nil {
			result.BlockedReason = "provider_check_error"
			result.Remedy = "Run loomi doctor and check /v1/model-providers/check."
			return result, nil
		}
		result.Provider = provider
		if !providerReadyForAgentSmoke(provider) {
			result.BlockedReason = providerBlockedReason(provider)
			result.Remedy = providerBlockedRemedy(provider)
			return result, nil
		}
	}

	events := []RunEvent{}
	result.Stage = "run_started"
	runResult, err := (Runner{Client: client}).Execute(ctx, RunOptions{
		ThreadID: opts.ThreadID,
		Prompt:   firstNonEmpty(opts.Prompt, "Loomi harness smoke: read AGENTS.md, then reply with a one sentence completion marker."),
		Mode:     firstNonEmpty(opts.Mode, cfg.Mode, "work"),
		Provider: providerID,
		Model:    firstNonEmpty(opts.Model, cfg.Model),
		Persona:  firstNonEmpty(opts.Persona, cfg.Persona),
		OnEvent: func(event RunEvent) {
			events = append(events, event)
		},
		OnApproval: func(PendingApproval, RunEvent) (string, error) {
			if opts.AutoApprove {
				return "approve", nil
			}
			return "skip", nil
		},
	})
	if err != nil {
		result.Stage = "run_stream"
		result.BlockedReason = "run_error"
		result.Remedy = err.Error()
		return result, nil
	}
	result.ThreadID = runResult.ThreadID
	result.RunID = runResult.RunID
	result.Status = runResult.Status
	result.FinalStage = finalStage(events, runResult.Status)
	result.EventCount = len(events)
	result.ToolEventCount = countToolEvents(events)
	result.ApprovalCount = countApprovalRequired(events)
	result.PendingApprovalCount = len(runResult.PendingApprovals)
	result.ToolChain = toolChain(events)
	result.ToolOrder = result.ToolChain
	result.LastEvents = lastEventSummaries(events, 3)
	finalMessage, err := finalAssistantMessage(ctx, client, result.ThreadID, result.RunID)
	if err != nil {
		result.Stage = "final_message"
		result.BlockedReason = "final_message_read_error"
		result.Remedy = err.Error()
		return result, nil
	}
	result.FinalMessage = finalMessage
	result.FinalPersisted = strings.TrimSpace(result.FinalMessage) != ""
	result.OK = runResult.Status == "completed"
	if !result.OK {
		result.Stage = finalStage(events, runResult.Status)
		if len(runResult.PendingApprovals) > 0 {
			result.BlockedReason = "tool_approval_required"
			result.Remedy = "Re-run with --auto-approve for harness-only approval, or approve listed tool calls manually."
		} else {
			result.BlockedReason = "run_" + strings.TrimSpace(runResult.Status)
			result.Remedy = "Inspect loomi runs attach " + runResult.RunID + " --compact."
		}
		return result, nil
	}
	if result.PendingApprovalCount > 0 {
		result.Stage = "pending_approval"
		result.OK = false
		result.BlockedReason = "pending_approval"
		result.Remedy = "Run reached a terminal state with unresolved tool approval events; inspect approval and tool terminal events."
		return result, nil
	}
	if strings.TrimSpace(result.FinalMessage) == "" {
		result.Stage = "final_message"
		result.OK = false
		result.BlockedReason = "final_message_missing"
		result.Remedy = "Run completed but no persisted assistant message was returned by /v1/threads/{thread_id}/messages."
		return result, nil
	}
	if strings.TrimSpace(result.FinalMessage) == "[redacted]" {
		result.Stage = "final_message"
		result.OK = false
		result.BlockedReason = "final_message_redacted"
		result.Remedy = "The final assistant message was only [redacted]; inspect tool-result continuation and finalization."
		return result, nil
	}
	if isInvalidSmokeFinalMessage(result.FinalMessage) {
		result.Stage = "final_message"
		result.OK = false
		result.BlockedReason = "final_message_placeholder"
		result.Remedy = "The final assistant message is a generated failure placeholder; inspect run finalization and persisted assistant messages."
		return result, nil
	}
	replayEvents, err := client.ListEvents(ctx, result.RunID, 0)
	if err != nil {
		result.Stage = "replay"
		result.OK = false
		result.BlockedReason = "replay_read_error"
		result.Remedy = err.Error()
		return result, nil
	}
	result.ReplayEventCount = len(replayEvents)
	result.ReplayTerminalStage = finalStage(replayEvents, runResult.Status)
	result.ReplayOK = result.ReplayEventCount > 0 && result.ReplayTerminalStage == result.FinalStage
	result.EventCount = len(replayEvents)
	result.ToolEventCount = countToolEvents(replayEvents)
	result.ApprovalCount = countApprovalRequired(replayEvents)
	result.PendingApprovalCount = len(PendingApprovals(replayEvents))
	result.ToolChain = toolChain(replayEvents)
	result.ToolOrder = result.ToolChain
	result.LastEvents = lastEventSummaries(replayEvents, 3)
	if !result.ReplayOK {
		result.Stage = "replay"
		result.OK = false
		result.BlockedReason = "replay_incomplete"
		result.Remedy = "Run completed but persisted replay events did not match the live terminal stage; inspect loomi runs attach " + result.RunID + " --compact."
		return result, nil
	}
	if result.PendingApprovalCount > 0 {
		result.Stage = "pending_approval"
		result.OK = false
		result.BlockedReason = "pending_approval"
		result.Remedy = "Persisted replay still has unresolved tool approval events; inspect approval and tool terminal events."
		return result, nil
	}
	result.Stage = "run_completed"
	return result, nil
}

func isInvalidSmokeFinalMessage(message string) bool {
	switch strings.TrimSpace(message) {
	case "未生成成功回复":
		return true
	default:
		return false
	}
}

func providerReadyForAgentSmoke(provider ProviderCapability) bool {
	if provider.ExecutionState == "unsupported" {
		return false
	}
	switch provider.Status {
	case "available", "completion-ok":
		return true
	default:
		return false
	}
}

func providerBlockedReason(provider ProviderCapability) string {
	switch provider.HTTPStatus {
	case 401, 403:
		return "provider_auth"
	case 429:
		return "provider_rate_limited"
	case 503:
		return "provider_unavailable"
	}
	if provider.CheckCode != "" {
		return provider.CheckCode
	}
	if provider.Status != "" {
		return "provider_" + string(provider.Status)
	}
	return "provider_unavailable"
}

func providerBlockedRemedy(provider ProviderCapability) string {
	switch provider.HTTPStatus {
	case 401, 403:
		return "Refresh the provider API token, then run loomi doctor again."
	case 429:
		return "Wait for quota reset or switch LOOMI_PROVIDER / loomi config provider."
	case 503:
		return "Retry later or switch to another configured provider."
	}
	if strings.TrimSpace(provider.Message) != "" {
		return provider.Message
	}
	return "Fix provider configuration or upstream completion before a live loomi run."
}

func finalStage(events []RunEvent, status string) string {
	for i := len(events) - 1; i >= 0; i-- {
		if strings.TrimSpace(events[i].Type) != "" {
			return events[i].Type
		}
	}
	if strings.TrimSpace(status) != "" {
		return status
	}
	return "unknown"
}

func countToolEvents(events []RunEvent) int {
	count := 0
	for _, event := range events {
		if IsToolEvent(event) {
			count++
		}
	}
	return count
}

func countApprovalRequired(events []RunEvent) int {
	count := 0
	for _, event := range events {
		if event.Type == "tool_call_approval_required" {
			count++
		}
	}
	return count
}

func toolChain(events []RunEvent) []string {
	chain := []string{}
	seen := map[string]struct{}{}
	for _, event := range events {
		if !IsToolEvent(event) {
			continue
		}
		toolName := metadataString(event.Metadata, "tool_name")
		if toolName == "" {
			continue
		}
		toolCallID := eventToolCallID(event)
		key := toolName + ":" + toolCallID
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		chain = append(chain, toolName)
	}
	return chain
}

func finalAssistantMessage(ctx context.Context, client *Client, threadID string, runID string) (string, error) {
	messages, err := client.ListMessages(ctx, threadID)
	if err != nil {
		return "", err
	}
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		if message.Role != "assistant" {
			continue
		}
		if runID != "" && messageRunID(message) != runID {
			continue
		}
		return strings.TrimSpace(message.Content), nil
	}
	return "", nil
}

func messageRunID(message Message) string {
	if strings.TrimSpace(message.RunID) != "" {
		return strings.TrimSpace(message.RunID)
	}
	return metadataString(message.Metadata, "run_id")
}

func lastEventSummaries(events []RunEvent, limit int) []string {
	if limit <= 0 || len(events) == 0 {
		return nil
	}
	start := len(events) - limit
	if start < 0 {
		start = 0
	}
	summaries := make([]string, 0, len(events)-start)
	for _, event := range events[start:] {
		summary := fmt.Sprintf("%04d %s", event.Sequence, event.Type)
		if IsToolEvent(event) {
			if toolName := metadataString(event.Metadata, "tool_name"); toolName != "" {
				summary += " " + toolName
			}
			if toolCallID := eventToolCallID(event); toolCallID != "" {
				summary += " " + toolCallID
			}
		}
		summaries = append(summaries, summary)
	}
	return summaries
}
