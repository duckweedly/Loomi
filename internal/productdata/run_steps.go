package productdata

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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
	Steps                             []RunStep                  `json:"steps"`
	PendingToolCalls                  []RunStep                  `json:"pending_tool_calls"`
	CompletedToolResults              []RunStep                  `json:"completed_tool_results"`
	Terminal                          *RunStep                   `json:"terminal,omitempty"`
	NextAction                        RunStepNextAction          `json:"next_action"`
	AcceptedToolCallCount             int                        `json:"accepted_tool_call_count,omitempty"`
	SeenToolCallIDs                   []string                   `json:"seen_tool_call_ids,omitempty"`
	EnabledToolNames                  []string                   `json:"enabled_tool_names,omitempty"`
	TriggerMessageID                  string                     `json:"trigger_message_id,omitempty"`
	ProviderID                        string                     `json:"provider_id,omitempty"`
	Model                             string                     `json:"model,omitempty"`
	MCPToolSchemaHashes               map[string]string          `json:"mcp_tool_schema_hashes,omitempty"`
	MCPAvailability                   MCPToolAvailabilitySummary `json:"mcp_availability,omitempty"`
	WorkspaceGlobSucceeded            bool                       `json:"workspace_glob_succeeded,omitempty"`
	WorkspaceToolRequestCount         int                        `json:"workspace_tool_request_count,omitempty"`
	WorkspaceArgumentHashesSinceReset []string                   `json:"workspace_argument_hashes_since_reset,omitempty"`
	LastCompletedSequence             int                        `json:"last_completed_sequence,omitempty"`
	LastContinuationSequence          int                        `json:"last_continuation_sequence,omitempty"`
	LastContinuationOutputSequence    int                        `json:"last_continuation_output_sequence,omitempty"`
	LastEventSequence                 int                        `json:"last_event_sequence,omitempty"`
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
	state := RebuildRunStepStateFromSteps(steps)
	for _, event := range events {
		state = applyRunStepStateEventFacts(state, event)
	}
	if len(events) > 0 {
		state.LastEventSequence = events[len(events)-1].Sequence
	}
	state.NextAction = nextRunStepAction(state)
	return state
}

func RebuildRunStepStateFromSteps(steps []RunStep) RunStepState {
	state := RunStepState{Steps: steps, NextAction: RunStepNextActionStartModel}
	pending := map[string]RunStep{}
	order := []string{}
	for _, step := range steps {
		switch step.Kind {
		case RunStepKindContinuation:
			state.LastContinuationSequence = step.Sequence
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
				state.LastCompletedSequence = step.Sequence
				if step.ToolName == ToolNameWorkspaceGlob {
					state.WorkspaceGlobSucceeded = true
				}
				delete(pending, step.ToolCallID)
			case RunStepStatusFailed:
				delete(pending, step.ToolCallID)
			}
		case RunStepKindTerminal:
			stepCopy := step
			state.Terminal = &stepCopy
		}
	}
	for _, step := range steps {
		if step.Kind == RunStepKindToolRequested && step.ToolCallID != "" {
			state.SeenToolCallIDs = appendUniqueRunStepID(state.SeenToolCallIDs, step.ToolCallID)
		}
	}
	state.AcceptedToolCallCount = len(state.SeenToolCallIDs)
	for _, id := range order {
		if step, ok := pending[id]; ok {
			state.PendingToolCalls = append(state.PendingToolCalls, step)
		}
	}
	state.NextAction = nextRunStepAction(state)
	return state
}

func AdvanceRunStepState(state RunStepState, event RunEvent) RunStepState {
	state = applyRunStepStateEventFacts(state, event)
	steps := runStepsFromEvent(event)
	if len(steps) == 0 {
		if event.Sequence > state.LastEventSequence {
			state.LastEventSequence = event.Sequence
		}
		state.NextAction = nextRunStepAction(state)
		return state
	}
	for _, step := range steps {
		state.Steps = append(state.Steps, step)
		switch step.Kind {
		case RunStepKindContinuation:
			state.LastContinuationSequence = step.Sequence
		case RunStepKindToolRequested:
			if step.ToolCallID != "" && !containsString(state.SeenToolCallIDs, step.ToolCallID) {
				state.SeenToolCallIDs = append(state.SeenToolCallIDs, step.ToolCallID)
				state.AcceptedToolCallCount = len(state.SeenToolCallIDs)
			}
			state.PendingToolCalls = upsertRunStepByToolCallID(state.PendingToolCalls, step)
		case RunStepKindApproval:
			switch step.Status {
			case RunStepStatusApproved:
				state.PendingToolCalls = upsertRunStepByToolCallID(state.PendingToolCalls, step)
			case RunStepStatusDenied:
				state.PendingToolCalls = removeRunStepByToolCallID(state.PendingToolCalls, step.ToolCallID)
			}
		case RunStepKindToolExecution:
			switch step.Status {
			case RunStepStatusRunning:
				state.PendingToolCalls = upsertRunStepByToolCallID(state.PendingToolCalls, step)
			case RunStepStatusSucceeded:
				state.CompletedToolResults = append(state.CompletedToolResults, step)
				state.LastCompletedSequence = step.Sequence
				if step.ToolName == ToolNameWorkspaceGlob {
					state.WorkspaceGlobSucceeded = true
				}
				state.PendingToolCalls = removeRunStepByToolCallID(state.PendingToolCalls, step.ToolCallID)
			case RunStepStatusFailed:
				state.PendingToolCalls = removeRunStepByToolCallID(state.PendingToolCalls, step.ToolCallID)
			}
		case RunStepKindTerminal:
			stepCopy := step
			state.Terminal = &stepCopy
		}
	}
	if event.Sequence > state.LastEventSequence {
		state.LastEventSequence = event.Sequence
	}
	state.NextAction = nextRunStepAction(state)
	return state
}

func applyRunStepStateEventFacts(state RunStepState, event RunEvent) RunStepState {
	state.EnabledToolNames = appendUniqueStrings(state.EnabledToolNames, runStepMetadataStringList(event.Metadata, "enabled_tools")...)
	if event.Type == "run_created" {
		state.TriggerMessageID = firstNonEmptyRunStepString(state.TriggerMessageID, metadataStringValue(event.Metadata, "message_id"))
		state.ProviderID = firstNonEmptyRunStepString(state.ProviderID, metadataStringValue(event.Metadata, "provider_id"))
		state.Model = firstNonEmptyRunStepString(state.Model, metadataStringValue(event.Metadata, "model"))
	}
	if metadataStringValue(event.Metadata, "model_phase") == "continuation" && (event.Type == "model_output_delta" || event.Type == "model_output_completed") {
		state.LastContinuationOutputSequence = event.Sequence
	}
	if event.Type == "mcp_discovery_succeeded" || event.Type == "mcp_discovery_failed" || event.Type == "mcp_discovery_rejected" {
		state.MCPAvailability = advanceMCPAvailabilitySummary(state.MCPAvailability, event)
	}
	if event.Type == "mcp_discovery_succeeded" && metadataStringValue(event.Metadata, "status") == "succeeded" {
		for _, name := range metadataStringSlice(event.Metadata, "candidate_names") {
			if !IsMCPToolName(name) {
				continue
			}
			if hash := mcpCandidateSchemaHash(event.Metadata, name); hash != "" {
				if state.MCPToolSchemaHashes == nil {
					state.MCPToolSchemaHashes = map[string]string{}
				}
				state.MCPToolSchemaHashes[name] = hash
			}
		}
	}
	toolName := metadataStringValue(event.Metadata, "tool_name")
	if event.Type == EventToolCallSucceeded && workspaceArgumentResetTool(toolName) {
		state.WorkspaceArgumentHashesSinceReset = nil
	}
	if event.Type == EventToolCallRequested && IsWorkspaceToolName(toolName) {
		state.WorkspaceToolRequestCount++
		if hash := workspaceToolArgumentHash(toolName, event.Metadata); hash != "" {
			state.WorkspaceArgumentHashesSinceReset = appendUniqueRunStepID(state.WorkspaceArgumentHashesSinceReset, hash)
		}
	}
	return state
}

func firstNonEmptyRunStepString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func workspaceArgumentResetTool(toolName string) bool {
	switch toolName {
	case ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchApply:
		return true
	case ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess:
		return true
	default:
		return false
	}
}

func workspaceToolArgumentHash(toolName string, metadata map[string]any) string {
	switch toolName {
	case ToolNameWorkspaceRead, ToolNameWorkspaceListDirectory, ToolNameWorkspaceGrep, ToolNameWorkspaceGlob, ToolNameWorkspaceTreeSummary:
	default:
		return ""
	}
	arguments, ok := metadata["arguments_summary"].(map[string]any)
	if !ok {
		return ""
	}
	raw, err := json.Marshal(RedactEventMetadata(arguments))
	if err != nil {
		raw = []byte("{}")
	}
	sum := sha256.Sum256(raw)
	return toolName + ":" + hex.EncodeToString(sum[:])
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
	case EventToolCallCancelled:
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
		if state.LastContinuationOutputSequence > state.LastCompletedSequence {
			return RunStepNextActionNone
		}
		return RunStepNextActionContinueModel
	}
	return RunStepNextActionStartModel
}

func upsertRunStepByToolCallID(steps []RunStep, step RunStep) []RunStep {
	if step.ToolCallID == "" {
		return steps
	}
	for index := range steps {
		if steps[index].ToolCallID == step.ToolCallID {
			steps[index] = step
			return steps
		}
	}
	return append(steps, step)
}

func removeRunStepByToolCallID(steps []RunStep, toolCallID string) []RunStep {
	if toolCallID == "" {
		return steps
	}
	next := steps[:0]
	for _, step := range steps {
		if step.ToolCallID != toolCallID {
			next = append(next, step)
		}
	}
	return next
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

func appendUniqueStrings(values []string, candidates ...string) []string {
	for _, candidate := range candidates {
		values = appendUniqueRunStepID(values, candidate)
	}
	return values
}

func containsString(values []string, value string) bool {
	value = strings.TrimSpace(value)
	for _, existing := range values {
		if existing == value {
			return true
		}
	}
	return false
}

func runStepMetadataStringList(metadata map[string]any, key string) []string {
	value, ok := metadata[key]
	if !ok {
		return nil
	}
	switch typed := value.(type) {
	case []string:
		return typed
	case []any:
		result := make([]string, 0, len(typed))
		for _, item := range typed {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				result = append(result, strings.TrimSpace(text))
			}
		}
		return result
	default:
		return nil
	}
}
