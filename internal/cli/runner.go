package cli

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

type StartRunInput struct {
	MessageID  string
	Source     string
	ProviderID string
	Model      string
	PersonaID  string
	ScriptName string
}

type RunOptions struct {
	ThreadID   string
	Prompt     string
	Mode       string
	Provider   string
	Model      string
	Persona    string
	Script     string
	OnEvent    func(RunEvent)
	OnApproval func(PendingApproval, RunEvent) (string, error)
}

type RunResult struct {
	ThreadID         string
	RunID            string
	Status           string
	Output           string
	PendingApprovals []PendingApproval
}

type PendingApproval struct {
	ThreadID   string
	RunID      string
	ToolCallID string
	ToolName   string
}

type Runner struct {
	Client *Client
}

func (r Runner) Execute(ctx context.Context, opts RunOptions) (RunResult, error) {
	client := r.Client
	if client == nil {
		client = NewClient("")
	}
	threadID := strings.TrimSpace(opts.ThreadID)
	if threadID == "" {
		mode := strings.TrimSpace(opts.Mode)
		if mode == "" {
			mode = "work"
		}
		thread, err := client.CreateThread(ctx, mode)
		if err != nil {
			return RunResult{}, err
		}
		threadID = thread.ID
	}
	message, err := client.AddMessage(ctx, threadID, opts.Prompt)
	if err != nil {
		return RunResult{}, err
	}
	input := StartRunInput{
		MessageID:  message.ID,
		Source:     "model_gateway",
		ProviderID: strings.TrimSpace(opts.Provider),
		Model:      strings.TrimSpace(opts.Model),
		PersonaID:  strings.TrimSpace(opts.Persona),
		ScriptName: strings.TrimSpace(opts.Script),
	}
	if input.ProviderID == "" && input.ScriptName == "" {
		input.ProviderID = "local_codex"
	}
	if input.ScriptName != "" {
		input.Source = "local_simulated"
		input.ProviderID = ""
		input.Model = ""
	}
	run, err := client.StartRun(ctx, threadID, input)
	if err != nil {
		return RunResult{}, err
	}
	var output strings.Builder
	events := []RunEvent{}
	seen := map[string]struct{}{}
	decided := map[string]struct{}{}
	status := run.Status
	afterSequence := 0
	for attempt := 0; attempt < 4; attempt++ {
		var callbackErr error
		err = client.StreamEvents(ctx, run.ID, afterSequence, func(event RunEvent) {
			if callbackErr != nil {
				return
			}
			key := event.ID
			if key == "" {
				key = event.RunID + ":" + strconv.Itoa(event.Sequence)
			}
			if _, ok := seen[key]; ok {
				return
			}
			seen[key] = struct{}{}
			if event.Sequence > afterSequence {
				afterSequence = event.Sequence
			}
			events = append(events, event)
			if isToolCallTerminalEvent(event.Type) {
				delete(decided, eventToolCallID(event))
			}
			if event.Content != nil && event.Type == "model_output_delta" {
				output.WriteString(*event.Content)
			}
			switch event.Type {
			case "tool_call_approval_required":
				status = "blocked_on_tool_approval"
			case "run_completed":
				status = "completed"
			case "run_failed":
				status = "failed"
			case "run_stopped":
				status = "stopped"
			}
			if opts.OnEvent != nil {
				opts.OnEvent(event)
			}
			if event.Type == "tool_call_approval_required" && opts.OnApproval != nil {
				approval := PendingApproval{
					ThreadID:   strings.TrimSpace(event.ThreadID),
					RunID:      strings.TrimSpace(event.RunID),
					ToolCallID: eventToolCallID(event),
					ToolName:   metadataString(event.Metadata, "tool_name"),
				}
				if approval.ThreadID == "" {
					approval.ThreadID = threadID
				}
				if approval.RunID == "" {
					approval.RunID = run.ID
				}
				if approval.ToolCallID == "" {
					return
				}
				if _, ok := decided[approval.ToolCallID]; ok {
					return
				}
				action, err := opts.OnApproval(approval, event)
				if err != nil {
					callbackErr = err
					return
				}
				action = strings.TrimSpace(strings.ToLower(action))
				if action == "" || action == "skip" {
					return
				}
				if action != "approve" && action != "deny" {
					callbackErr = errors.New("approval action must be approve, deny, or skip")
					return
				}
				decided[approval.ToolCallID] = struct{}{}
				if _, err := client.DecideToolCall(ctx, approval.ThreadID, approval.RunID, approval.ToolCallID, action); err != nil {
					callbackErr = err
					return
				}
			}
		})
		if callbackErr != nil {
			return RunResult{}, callbackErr
		}
		if err != nil {
			return RunResult{}, err
		}
		if isTerminalRunStatus(status) || hasUndecidedPendingApprovals(events, decided) {
			break
		}
	}
	return RunResult{ThreadID: threadID, RunID: run.ID, Status: status, Output: output.String(), PendingApprovals: PendingApprovals(events)}, nil
}

func hasUndecidedPendingApprovals(events []RunEvent, decided map[string]struct{}) bool {
	for _, approval := range PendingApprovals(events) {
		if _, ok := decided[approval.ToolCallID]; !ok {
			return true
		}
	}
	return false
}

func isToolCallTerminalEvent(eventType string) bool {
	switch eventType {
	case "tool_call_approved", "tool_call_denied", "tool_call_executing", "tool_call_succeeded", "tool_call_failed", "tool_call_cancelled":
		return true
	default:
		return false
	}
}

func isTerminalRunStatus(status string) bool {
	switch status {
	case "completed", "failed", "stopped":
		return true
	default:
		return false
	}
}

func PendingApprovalEvents(events []RunEvent) []RunEvent {
	byToolCallID := map[string]RunEvent{}
	order := []string{}
	for _, event := range events {
		toolCallID := eventToolCallID(event)
		if toolCallID == "" {
			continue
		}
		switch event.Type {
		case "tool_call_approval_required":
			if _, ok := byToolCallID[toolCallID]; !ok {
				order = append(order, toolCallID)
			}
			byToolCallID[toolCallID] = event
		case "tool_call_approved", "tool_call_denied", "tool_call_executing", "tool_call_succeeded", "tool_call_failed", "tool_call_cancelled":
			delete(byToolCallID, toolCallID)
		}
	}
	result := []RunEvent{}
	for _, toolCallID := range order {
		if event, ok := byToolCallID[toolCallID]; ok {
			result = append(result, event)
		}
	}
	return result
}

func PendingApprovals(events []RunEvent) []PendingApproval {
	approvals := []PendingApproval{}
	for _, event := range PendingApprovalEvents(events) {
		approvals = append(approvals, PendingApproval{
			ThreadID:   strings.TrimSpace(event.ThreadID),
			RunID:      strings.TrimSpace(event.RunID),
			ToolCallID: eventToolCallID(event),
			ToolName:   metadataString(event.Metadata, "tool_name"),
		})
	}
	return approvals
}

func eventToolCallID(event RunEvent) string {
	return metadataString(event.Metadata, "tool_call_id")
}

func metadataString(metadata map[string]any, key string) string {
	if metadata == nil {
		return ""
	}
	value, _ := metadata[key].(string)
	return strings.TrimSpace(value)
}
