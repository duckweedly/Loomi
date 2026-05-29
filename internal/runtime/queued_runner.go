package runtime

import (
	"context"
	"strings"
	"sync"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type QueuedRunRouter struct {
	Local            *LocalRunner
	Gateway          *Gateway
	MCPExecutor      MCPToolExecutor
	SandboxStore     *SandboxProcessStore
	WebExecutor      WebToolExecutor
	BrowserExecutor  BrowserToolExecutor
	ArtifactService  productdata.ArtifactService
	AgentTaskService productdata.AgentTaskService
	MemoryService    productdata.Service
	toolExecutor     ToolExecutor
}

type MCPToolExecutor interface {
	ExecuteMCPTool(context.Context, productdata.ToolCall) (map[string]any, error)
}

func (r QueuedRunRouter) Run(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
	svc := r.service()
	var prepared *productdata.RunContext
	if svc != nil && job.ID != "" {
		if terminal, err := runIsTerminal(ctx, svc, job.RunID); err != nil || terminal {
			return err
		}
		state := &PipelineState{RunContext: productdata.RunContext{Run: run, Job: job}}
		pipeline := Pipeline{Recorder: PipelineRecorder{Service: svc, Broadcaster: r.broadcaster()}}
		if err := pipeline.Execute(ctx, state, []PipelineStage{
			{Name: productdata.PipelineStepPrepareContext, Run: func(ctx context.Context, state *PipelineState) error {
				context, err := svc.PrepareRunContext(ctx, identity.LocalDevIdentity(), job)
				if err != nil {
					return err
				}
				state.RunContext = context
				return nil
			}},
			{Name: productdata.PipelineStepResolveTools},
			{Name: productdata.PipelineStepInvokeRuntime},
			{Name: productdata.PipelineStepFinalize},
		}); err != nil {
			return err
		}
		run = state.RunContext.Run
		job = state.RunContext.Job
		prepared = &state.RunContext
	}
	return r.dispatch(ctx, run, job, prepared)
}

func (r QueuedRunRouter) dispatch(ctx context.Context, run productdata.Run, job productdata.BackgroundJob, prepared *productdata.RunContext) error {
	if toolCallID := metadataString(job.Metadata, "tool_call_id"); toolCallID != "" {
		return r.runApprovedTool(ctx, run, job, toolCallID, prepared, true)
	}
	if run.Source == productdata.RunSourceModelGateway {
		if r.Gateway != nil {
			r.Gateway.runWithContext(ctx, run, gatewayInputFromJob(run, job), prepared)
			lastExecutedToolCallID, err := r.executeReadyAutoApprovedTools(ctx, run, prepared)
			if err != nil {
				return err
			}
			if lastExecutedToolCallID != "" {
				return r.continueAfterToolBatch(ctx, run, lastExecutedToolCallID, prepared)
			}
			return r.gatewayResult(ctx, run.ID)
		}
		return nil
	}
	if r.Local != nil {
		return r.Local.Run(ctx, run, job)
	}
	return nil
}

func (r QueuedRunRouter) dispatchWithExecutorForTest(ctx context.Context, run productdata.Run, job productdata.BackgroundJob, prepared *productdata.RunContext, executor ToolExecutor) error {
	r.toolExecutor = executor
	return r.dispatch(ctx, run, job, prepared)
}

func (r QueuedRunRouter) nextAutoApprovedToolCall(ctx context.Context, runID string) string {
	toolCallIDs := r.readyAutoApprovedToolCalls(ctx, runID)
	if len(toolCallIDs) == 0 {
		return ""
	}
	return toolCallIDs[0]
}

func (r QueuedRunRouter) readyAutoApprovedToolCalls(ctx context.Context, runID string) []string {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return nil
	}
	state, err := r.Gateway.Service.GetRunStepState(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return nil
	}
	toolCallIDs := make([]string, 0, len(state.PendingToolCalls))
	for _, step := range state.PendingToolCalls {
		if step.Status == productdata.RunStepStatusApproved {
			toolCallIDs = append(toolCallIDs, step.ToolCallID)
		}
	}
	return toolCallIDs
}

func (r QueuedRunRouter) executeReadyAutoApprovedTools(ctx context.Context, run productdata.Run, prepared *productdata.RunContext) (string, error) {
	lastExecutedToolCallID := ""
	for {
		if terminal, err := runIsTerminal(ctx, r.Gateway.Service, run.ID); err != nil || terminal {
			return lastExecutedToolCallID, err
		}
		toolCallIDs := r.readyAutoApprovedToolCalls(ctx, run.ID)
		if len(toolCallIDs) == 0 {
			return lastExecutedToolCallID, nil
		}
		if len(toolCallIDs) == 1 {
			if err := r.runApprovedTool(ctx, run, preparedJob(prepared), toolCallIDs[0], prepared, false); err != nil {
				return "", err
			}
			lastExecutedToolCallID = toolCallIDs[0]
			continue
		}
		errs := make(chan error, len(toolCallIDs))
		var wg sync.WaitGroup
		for _, toolCallID := range toolCallIDs {
			wg.Add(1)
			go func(toolCallID string) {
				defer wg.Done()
				errs <- r.runApprovedTool(ctx, run, preparedJob(prepared), toolCallID, prepared, false)
			}(toolCallID)
		}
		wg.Wait()
		close(errs)
		for err := range errs {
			if err != nil {
				return "", err
			}
		}
		lastExecutedToolCallID = toolCallIDs[len(toolCallIDs)-1]
	}
}

func (r QueuedRunRouter) continueAfterToolBatch(ctx context.Context, run productdata.Run, seedToolCallID string, prepared *productdata.RunContext) error {
	lastToolCallID := seedToolCallID
	for {
		if drainedToolCallID, err := r.executeReadyAutoApprovedTools(ctx, run, prepared); err != nil {
			return err
		} else if drainedToolCallID != "" {
			lastToolCallID = drainedToolCallID
		}
		if lastToolCallID == "" {
			return r.gatewayResult(ctx, run.ID)
		}
		if r.hasUnresolvedToolCalls(ctx, run.ID) {
			return r.gatewayResult(ctx, run.ID)
		}
		input := r.gatewayContinuationInput(ctx, run, lastToolCallID, prepared)
		input.SystemPrompt = runSystemPrompt(prepared)
		if input.MessageID == "" || input.ProviderID == "" {
			return productdata.NewError(productdata.CodeInvalidRequest, "Continuation context is missing.")
		}
		jobID := continuationClaimJobID(run, lastToolCallID, prepared)
		claimed, ok, err := r.Gateway.Service.ClaimToolContinuation(ctx, identity.LocalDevIdentity(), productdata.ClaimToolContinuationInput{ThreadID: run.ThreadID, RunID: run.ID, ToolCallID: lastToolCallID, JobID: jobID, ProviderID: input.ProviderID, Model: input.Model})
		if err != nil {
			return err
		}
		if !ok {
			return r.gatewayResult(ctx, run.ID)
		}
		r.publishRunEvents([]productdata.RunEvent{claimed})
		input.ContinuationPreclaimed = true
		r.Gateway.ContinueAfterToolResult(ctx, run, input)
		lastToolCallID = ""
		if r.nextAutoApprovedToolCall(ctx, run.ID) == "" {
			return r.gatewayResult(ctx, run.ID)
		}
	}
}

func continuationClaimJobID(run productdata.Run, toolCallID string, prepared *productdata.RunContext) string {
	if prepared != nil && strings.TrimSpace(prepared.Job.ID) != "" {
		return prepared.Job.ID
	}
	return "direct:" + run.ID + ":" + strings.TrimSpace(toolCallID)
}

func (r QueuedRunRouter) hasUnresolvedToolCalls(ctx context.Context, runID string) bool {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return false
	}
	state, err := r.Gateway.Service.GetRunStepState(ctx, identity.LocalDevIdentity(), runID)
	return err == nil && len(state.PendingToolCalls) > 0
}

func (r QueuedRunRouter) service() productdata.Service {
	if r.Gateway != nil && r.Gateway.Service != nil {
		return r.Gateway.Service
	}
	if r.Local != nil && r.Local.Service != nil {
		return r.Local.Service
	}
	return nil
}

func (r QueuedRunRouter) broadcaster() *Broadcaster {
	if r.Gateway != nil && r.Gateway.Broadcaster != nil {
		return r.Gateway.Broadcaster
	}
	if r.Local != nil && r.Local.Broadcaster != nil {
		return r.Local.Broadcaster
	}
	return nil
}

func (r QueuedRunRouter) runApprovedTool(ctx context.Context, run productdata.Run, job productdata.BackgroundJob, toolCallID string, prepared *productdata.RunContext, continueAfter bool) error {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return nil
	}
	if terminal, err := runIsTerminal(ctx, r.Gateway.Service, run.ID); err != nil || terminal {
		return err
	}
	existing, err := r.Gateway.Service.GetToolCall(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID)
	if err != nil {
		return err
	}
	if existing.ExecutionStatus == productdata.ToolCallExecutionSucceeded {
		if continueAfter && r.shouldResumeContinuationAfterSucceededTool(ctx, run.ID, toolCallID) {
			return r.continueAfterToolBatch(ctx, run, toolCallID, prepared)
		}
		return nil
	}
	if existing.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		return nil
	}
	call, startedEvents, err := r.Gateway.Service.StartToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID)
	if err != nil {
		if terminal, terminalErr := runIsTerminal(ctx, r.Gateway.Service, run.ID); terminalErr != nil || terminal {
			return terminalErr
		}
		return err
	}
	r.publishRunEvents(startedEvents)
	if call.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		return nil
	}
	enabledTools := []productdata.ToolResolution(nil)
	if prepared != nil {
		enabledTools = prepared.EnabledTools
	}
	catalog := toolCatalogForExecution(enabledTools)
	executor := r.toolExecutor
	if executor == nil {
		executor = DefaultToolExecutor{MCPExecutor: r.MCPExecutor, WorkspaceExecutor: WorkspaceToolExecutor{}, SandboxExecutor: SandboxToolExecutor{Store: r.SandboxStore}, LSPExecutor: LSPToolExecutor{}, WebExecutor: r.WebExecutor, BrowserExecutor: r.BrowserExecutor, ArtifactExecutor: ArtifactToolExecutor{Artifacts: r.artifactService()}, AgentExecutor: AgentToolExecutor{Tasks: r.agentTaskService()}, MemoryExecutor: MemoryToolExecutor{Service: r.memoryService(), Ident: identity.LocalDevIdentity()}}
	}
	broker := ToolBroker{Executor: executor}
	invocation := ToolInvocationFromCall(call, catalog, enabledTools)
	invocation.RunStatus = run.Status
	if prepared != nil {
		invocation.WorkspaceRoot = prepared.WorkspaceRoot.Path
	}
	result, err := broker.Execute(ctx, invocation)
	if err != nil {
		if !r.toolExecutionOwnerStillCurrent(ctx, job) {
			return nil
		}
		code, message := toolExecutionFailure(err)
		_, failedEvents, _ := r.Gateway.Service.FailToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, code, message)
		r.publishRunEvents(failedEvents)
		return nil
	}
	if result.ToolName == productdata.ToolNameAgentDelegate && result.ResultSummary["child_run_id"] != nil {
		if err := r.appendAgentDelegateHandoffEvent(ctx, run, toolCallID, result.ResultSummary); err != nil {
			return err
		}
		return nil
	}
	var completedEvents []productdata.RunEvent
	if !r.toolExecutionOwnerStillCurrent(ctx, job) {
		return nil
	}
	if _, completedEvents, err = r.Gateway.Service.CompleteToolCallSuccess(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, result.ResultSummary); err != nil {
		return err
	}
	r.publishRunEvents(completedEvents)
	if result.ToolName == productdata.ToolNameTodoWrite {
		if event, ok := appendProviderWorkTodoSnapshot(ctx, r.Gateway.Service, run, result.ResultSummary); ok && r.broadcaster() != nil {
			r.broadcaster().Publish(event)
		}
	} else {
		_, _ = appendWorkTodoSnapshot(ctx, r.Gateway.Service, run, "runtime")
	}
	if !continueAfter {
		return nil
	}
	return r.continueAfterToolBatch(ctx, run, toolCallID, prepared)
}

func preparedJob(prepared *productdata.RunContext) productdata.BackgroundJob {
	if prepared == nil {
		return productdata.BackgroundJob{}
	}
	return prepared.Job
}

func (r QueuedRunRouter) toolExecutionOwnerStillCurrent(ctx context.Context, job productdata.BackgroundJob) bool {
	if r.Gateway == nil || r.Gateway.Service == nil || job.ID == "" || job.LeasedBy == nil || *job.LeasedBy == "" || job.OwnershipVersion == 0 {
		return true
	}
	_, ok, err := r.Gateway.Service.RenewBackgroundJobLease(ctx, identity.LocalDevIdentity(), productdata.RenewBackgroundJobLeaseInput{JobID: job.ID, WorkerID: *job.LeasedBy, OwnershipVersion: job.OwnershipVersion, LeaseSeconds: 30})
	return err == nil && ok
}

func (r QueuedRunRouter) appendAgentDelegateHandoffEvent(ctx context.Context, run productdata.Run, toolCallID string, result map[string]any) error {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return nil
	}
	childRunID := metadataString(result, "child_run_id")
	childThreadID := metadataString(result, "child_thread_id")
	if childRunID == "" || childThreadID == "" {
		return nil
	}
	parentToolCallID := metadataString(result, "parent_tool_call_id")
	if parentToolCallID == "" {
		parentToolCallID = toolCallID
	}
	metadata := map[string]any{
		"scope":                "agent",
		"operation":            "delegate",
		"tool_call_id":         toolCallID,
		"parent_tool_call_id":  parentToolCallID,
		"task_id":              metadataString(result, "task_id"),
		"child_thread_id":      childThreadID,
		"child_run_id":         childRunID,
		"status":               metadataString(result, "status"),
		"autonomous_execution": true,
		"redaction_applied":    true,
	}
	event, err := r.Gateway.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: productdata.EventAgentChildRunStarted, Summary: "Delegated agent child run started", Metadata: metadata})
	if err != nil {
		return err
	}
	r.publishRunEvents([]productdata.RunEvent{event})
	return nil
}

func (r QueuedRunRouter) publishRunEvents(events []productdata.RunEvent) {
	broadcaster := r.broadcaster()
	if broadcaster == nil {
		return
	}
	for _, event := range events {
		broadcaster.Publish(event)
	}
}

func toolExecutionFailure(err error) (string, string) {
	if err == nil {
		return "tool_execution_failed", "Tool execution failed."
	}
	message := strings.TrimSpace(err.Error())
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "workspace root is not authorized"):
		return "permission_not_authorized", "Permission was not granted for this workspace tool. Re-select or approve the workspace, then retry."
	case strings.Contains(lower, "workspace root is unavailable"):
		return "workspace_unbound", "No usable workspace folder is bound for this run. Select a workspace folder, then retry."
	case strings.Contains(lower, "not allowed"):
		return "permission_not_authorized", "Permission was not granted for this tool or command. Approve an allowed tool action, then retry."
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "deadline exceeded"):
		return "bounded_limit_reached", "The tool hit its timeout or bounded output limit. Try a narrower command or smaller input."
	case message != "":
		return "tool_execution_failed", "Tool execution failed: " + message
	default:
		return "tool_execution_failed", "Tool execution failed."
	}
}

func (r QueuedRunRouter) shouldResumeContinuationAfterSucceededTool(ctx context.Context, runID string, toolCallID string) bool {
	if r.Gateway == nil || r.Gateway.Service == nil || strings.TrimSpace(toolCallID) == "" {
		return false
	}
	state, err := r.Gateway.Service.GetRunStepState(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return false
	}
	if state.NextAction != productdata.RunStepNextActionContinueModel {
		return false
	}
	if len(state.CompletedToolResults) == 0 {
		return false
	}
	return state.CompletedToolResults[len(state.CompletedToolResults)-1].ToolCallID == toolCallID
}

func (r QueuedRunRouter) artifactService() productdata.ArtifactService {
	if r.ArtifactService != nil {
		return r.ArtifactService
	}
	service := r.service()
	if service == nil {
		return nil
	}
	if svc, ok := service.(productdata.ArtifactService); ok {
		return svc
	}
	return nil
}

func (r QueuedRunRouter) agentTaskService() productdata.AgentTaskService {
	if r.AgentTaskService != nil {
		return r.AgentTaskService
	}
	service := r.service()
	if service == nil {
		return nil
	}
	if svc, ok := service.(productdata.AgentTaskService); ok {
		return svc
	}
	return nil
}

func (r QueuedRunRouter) memoryService() productdata.Service {
	if r.MemoryService != nil {
		return r.MemoryService
	}
	return r.service()
}

func toolCatalogForExecution(enabledTools []productdata.ToolResolution) []productdata.ToolCatalogEntry {
	catalog := []productdata.ToolCatalogEntry{}
	for _, tool := range enabledTools {
		entry := productdata.ToolCatalogEntry{
			Name:            tool.Name,
			DisplayName:     tool.Name,
			Source:          productdata.ToolCatalogSource(tool.Source),
			Group:           productdata.ToolCatalogGroup(tool.Group),
			InputSchemaHash: tool.InputSchemaHash,
			RiskLevel:       productdata.ToolRiskLevel(tool.RiskLevel),
			ApprovalPolicy:  productdata.ToolApprovalPolicy(tool.ApprovalPolicy),
			Enabled:         true,
			ExecutionState:  productdata.ToolExecutionState(tool.ExecutionState),
		}
		if entry.Source == "" {
			entry.Source = productdata.ToolCatalogSourceBuiltin
		}
		if entry.Group == "" {
			entry.Group = productdata.ToolCatalogGroupRuntime
		}
		if entry.RiskLevel == "" {
			entry.RiskLevel = productdata.ToolRiskLow
		}
		if entry.ApprovalPolicy == "" {
			entry.ApprovalPolicy = productdata.ToolApprovalAlwaysRequired
		}
		catalog = append(catalog, entry)
	}
	return catalog
}

func (r QueuedRunRouter) gatewayResult(ctx context.Context, runID string) error {
	run, err := r.Gateway.Service.GetRun(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return err
	}
	if run.Status == productdata.RunStatusStopped {
		return nil
	}
	if run.Status == productdata.RunStatusFailed {
		return productdata.NewError(productdata.CodeInternalError, "Queued gateway run did not complete.")
	}
	return nil
}

func runIsTerminal(ctx context.Context, svc productdata.Service, runID string) (bool, error) {
	run, err := svc.GetRun(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return false, err
	}
	return productdata.IsRunTerminal(run.Status), nil
}

func gatewayInputFromJob(run productdata.Run, job productdata.BackgroundJob) GatewayRunInput {
	return GatewayRunInput{
		ThreadID:   run.ThreadID,
		MessageID:  metadataString(job.Metadata, "message_id"),
		ProviderID: metadataString(job.Metadata, "provider_id"),
		Model:      metadataString(job.Metadata, "model"),
	}
}

func (r QueuedRunRouter) gatewayContinuationInput(ctx context.Context, run productdata.Run, toolCallID string, prepared *productdata.RunContext) GatewayContinuationInput {
	input := GatewayContinuationInput{ThreadID: run.ThreadID, ToolCallID: toolCallID}
	if prepared != nil {
		input.MessageID = metadataString(prepared.Job.Metadata, "message_id")
		input.ProviderID = prepared.ProviderRoute.ProviderID
		input.Model = prepared.ProviderRoute.Model
		if input.MessageID != "" && input.ProviderID != "" {
			return input
		}
	}
	if r.Gateway == nil || r.Gateway.Service == nil {
		return input
	}
	state, err := r.Gateway.Service.GetRunStepState(ctx, identity.LocalDevIdentity(), run.ID)
	if err == nil {
		input.MessageID = state.TriggerMessageID
		input.ProviderID = state.ProviderID
		input.Model = state.Model
		if input.MessageID != "" && input.ProviderID != "" {
			return input
		}
	}
	return input
}

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}
