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
	AutoApprove bool
}

type SmokeAgentResult struct {
	OK             bool               `json:"ok"`
	Stage          string             `json:"stage"`
	ThreadID       string             `json:"thread_id,omitempty"`
	RunID          string             `json:"run_id,omitempty"`
	FinalStage     string             `json:"final_stage,omitempty"`
	Status         string             `json:"status,omitempty"`
	BlockedReason  string             `json:"blocked_reason,omitempty"`
	Remedy         string             `json:"remedy,omitempty"`
	Provider       ProviderCapability `json:"provider,omitempty"`
	EventCount     int                `json:"event_count,omitempty"`
	ToolEventCount int                `json:"tool_event_count,omitempty"`
	ApprovalCount  int                `json:"approval_count,omitempty"`
	LastEvents     []string           `json:"last_events,omitempty"`
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
	result.LastEvents = lastEventSummaries(events, 3)
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
	result.Stage = "run_completed"
	return result, nil
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
