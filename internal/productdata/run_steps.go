package productdata

import (
	"fmt"
	"strings"
)

type RunStepKind string
type RunStepStatus string
type RunStepNextAction string

const (
	RunStepKindModelRequest  RunStepKind = "model_request"
	RunStepKindToolRequested RunStepKind = "tool_requested"
	RunStepKindApproval      RunStepKind = "approval"
	RunStepKindToolExecution RunStepKind = "tool_execution"
	RunStepKindContinuation  RunStepKind = "continuation"
	RunStepKindTerminal      RunStepKind = "terminal"

	RunStepStatusPending   RunStepStatus = "pending"
	RunStepStatusRequired  RunStepStatus = "required"
	RunStepStatusApproved  RunStepStatus = "approved"
	RunStepStatusDenied    RunStepStatus = "denied"
	RunStepStatusRunning   RunStepStatus = "running"
	RunStepStatusSucceeded RunStepStatus = "succeeded"
	RunStepStatusFailed    RunStepStatus = "failed"
	RunStepStatusCompleted RunStepStatus = "completed"
	RunStepStatusStopped   RunStepStatus = "stopped"

	RunStepNextActionStartModel          RunStepNextAction = "start_model"
	RunStepNextActionWaitForToolApproval RunStepNextAction = "wait_for_tool_approval"
	RunStepNextActionExecuteTool         RunStepNextAction = "execute_tool"
	RunStepNextActionContinueModel       RunStepNextAction = "continue_model"
	RunStepNextActionTerminal            RunStepNextAction = "terminal"
	RunStepNextActionNone                RunStepNextAction = "none"
)

type RunStep struct {
	ID            string         `json:"id"`
	RunID         string         `json:"run_id"`
	Sequence      int            `json:"sequence"`
	Kind          RunStepKind    `json:"kind"`
	Status        RunStepStatus  `json:"status"`
	ToolCallID    string         `json:"tool_call_id,omitempty"`
	ToolName      string         `json:"tool_name,omitempty"`
	ModelPhase    string         `json:"model_phase,omitempty"`
	Summary       string         `json:"summary"`
	SafeMetadata  map[string]any `json:"safe_metadata,omitempty"`
	SourceEventID string         `json:"source_event_id,omitempty"`
}

type RunStepState struct {
	Steps                    []RunStep         `json:"steps"`
	PendingToolCalls         []RunStep         `json:"pending_tool_calls"`
	CompletedToolResults     []RunStep         `json:"completed_tool_results"`
	Terminal                 *RunStep          `json:"terminal,omitempty"`
	NextAction               RunStepNextAction `json:"next_action"`
	lastCompletedSequence    int
	lastContinuationSequence int
}

func BuildRunStepLedger(events []RunEvent) []RunStep {
	steps := make([]RunStep, 0, len(events))
	for _, event := range events {
		for _, step := range runStepsFromEvent(event) {
			steps = append(steps, step)
		}
	}
	return steps
}

func AnnotateRunStepMetadata(eventType string, summary string, metadata map[string]any) map[string]any {
	kind, status, ok := runStepKindStatus(eventType, metadata)
	if !ok {
		return metadata
	}
	next := RedactEventMetadata(metadata)
	if next == nil {
		next = map[string]any{}
	}
	next["run_step_kind"] = string(kind)
	next["run_step_status"] = string(status)
	next["run_step_summary"] = RedactEventText(strings.TrimSpace(summary))
	return next
}

func RebuildRunStepState(events []RunEvent) RunStepState {
	steps := BuildRunStepLedger(events)
	state := RunStepState{Steps: steps, NextAction: RunStepNextActionStartModel}
	pending := map[string]RunStep{}
	order := []string{}
	for _, step := range steps {
		switch step.Kind {
		case RunStepKindContinuation:
			state.lastContinuationSequence = step.Sequence
		case RunStepKindToolRequested:
			if step.ToolCallID != "" {
				pending[step.ToolCallID] = step
				order = appendUniqueRunStepID(order, step.ToolCallID)
			}
		case RunStepKindApproval:
			if step.ToolCallID == "" {
				continue
			}
			if step.Status == RunStepStatusApproved {
				pending[step.ToolCallID] = step
				order = appendUniqueRunStepID(order, step.ToolCallID)
			}
			if step.Status == RunStepStatusDenied {
				delete(pending, step.ToolCallID)
			}
		case RunStepKindToolExecution:
			if step.ToolCallID == "" {
				continue
			}
			switch step.Status {
			case RunStepStatusRunning:
				pending[step.ToolCallID] = step
				order = appendUniqueRunStepID(order, step.ToolCallID)
			case RunStepStatusSucceeded:
				state.CompletedToolResults = append(state.CompletedToolResults, step)
				state.lastCompletedSequence = step.Sequence
				delete(pending, step.ToolCallID)
			case RunStepStatusFailed:
				delete(pending, step.ToolCallID)
			}
		case RunStepKindTerminal:
			stepCopy := step
			state.Terminal = &stepCopy
		}
	}
	for _, id := range order {
		if step, ok := pending[id]; ok {
			state.PendingToolCalls = append(state.PendingToolCalls, step)
		}
	}
	state.NextAction = nextRunStepAction(state)
	return state
}

func runStepsFromEvent(event RunEvent) []RunStep {
	kind, status, ok := runStepKindStatus(event.Type, event.Metadata)
	if !ok {
		return nil
	}
	return []RunStep{runStepFromEvent(event, kind, status)}
}

func runStepKindStatus(eventType string, metadata map[string]any) (RunStepKind, RunStepStatus, bool) {
	if eventType == "model_request_started" {
		if metadataStringValue(metadata, "model_phase") == "continuation" {
			return RunStepKindContinuation, RunStepStatusRunning, true
		}
		return RunStepKindModelRequest, RunStepStatusRunning, true
	}
	switch eventType {
	case EventToolCallRequested:
		return RunStepKindToolRequested, RunStepStatusPending, true
	case EventToolCallApprovalRequired:
		return RunStepKindApproval, RunStepStatusRequired, true
	case EventToolCallApproved:
		return RunStepKindApproval, RunStepStatusApproved, true
	case EventToolCallDenied:
		return RunStepKindApproval, RunStepStatusDenied, true
	case EventToolCallExecuting:
		return RunStepKindToolExecution, RunStepStatusRunning, true
	case EventToolCallSucceeded:
		return RunStepKindToolExecution, RunStepStatusSucceeded, true
	case EventToolCallFailed:
		return RunStepKindToolExecution, RunStepStatusFailed, true
	case EventRunCompleted:
		return RunStepKindTerminal, RunStepStatusCompleted, true
	case EventRunFailed:
		return RunStepKindTerminal, RunStepStatusFailed, true
	case EventRunStopped:
		return RunStepKindTerminal, RunStepStatusStopped, true
	default:
		return "", "", false
	}
}

func runStepFromEvent(event RunEvent, kind RunStepKind, status RunStepStatus) RunStep {
	return RunStep{
		ID:            fmt.Sprintf("step_%d", event.Sequence),
		RunID:         event.RunID,
		Sequence:      event.Sequence,
		Kind:          kind,
		Status:        status,
		ToolCallID:    metadataStringValue(event.Metadata, "tool_call_id"),
		ToolName:      metadataStringValue(event.Metadata, "tool_name"),
		ModelPhase:    metadataStringValue(event.Metadata, "model_phase"),
		Summary:       RedactEventText(strings.TrimSpace(event.Summary)),
		SafeMetadata:  RedactEventMetadata(event.Metadata),
		SourceEventID: event.ID,
	}
}

func nextRunStepAction(state RunStepState) RunStepNextAction {
	if state.Terminal != nil {
		return RunStepNextActionTerminal
	}
	if len(state.PendingToolCalls) > 0 {
		latest := state.PendingToolCalls[len(state.PendingToolCalls)-1]
		switch latest.Status {
		case RunStepStatusApproved:
			return RunStepNextActionExecuteTool
		case RunStepStatusRunning:
			return RunStepNextActionNone
		default:
			return RunStepNextActionWaitForToolApproval
		}
	}
	if len(state.CompletedToolResults) > 0 {
		if state.lastContinuationSequence > state.lastCompletedSequence {
			return RunStepNextActionNone
		}
		return RunStepNextActionContinueModel
	}
	return RunStepNextActionStartModel
}

func appendUniqueRunStepID(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}
