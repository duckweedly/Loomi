package runtime

import (
	"context"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type QueuedRunRouter struct {
	Local            *LocalRunner
	Gateway          *Gateway
	MCPExecutor      MCPToolExecutor
	WebExecutor      WebToolExecutor
	BrowserExecutor  BrowserToolExecutor
	ArtifactService  productdata.ArtifactService
	AgentTaskService productdata.AgentTaskService
	MemoryService    productdata.Service
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
		return r.runApprovedTool(ctx, run, toolCallID, prepared)
	}
	if run.Source == productdata.RunSourceModelGateway {
		if r.Gateway != nil {
			r.Gateway.runWithContext(ctx, run, gatewayInputFromJob(run, job), prepared)
			for {
				toolCallID := r.nextAutoApprovedToolCall(ctx, run.ID)
				if toolCallID == "" {
					break
				}
				if err := r.runApprovedTool(ctx, run, toolCallID, prepared); err != nil {
					return err
				}
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

func (r QueuedRunRouter) nextAutoApprovedToolCall(ctx context.Context, runID string) string {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return ""
	}
	events, err := r.Gateway.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return ""
	}
	completed := map[string]bool{}
	approved := []string{}
	for _, event := range events {
		toolCallID := metadataString(event.Metadata, "tool_call_id")
		if toolCallID == "" {
			continue
		}
		switch event.Type {
		case productdata.EventToolCallSucceeded, productdata.EventToolCallFailed:
			completed[toolCallID] = true
		case productdata.EventToolCallApproved:
			approved = append(approved, toolCallID)
		}
	}
	for _, toolCallID := range approved {
		if !completed[toolCallID] {
			return toolCallID
		}
	}
	return ""
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

func (r QueuedRunRouter) runApprovedTool(ctx context.Context, run productdata.Run, toolCallID string, prepared *productdata.RunContext) error {
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
		if r.shouldResumeContinuationAfterSucceededTool(ctx, run.ID, toolCallID) {
			input := r.gatewayContinuationInput(ctx, run, toolCallID)
			if input.MessageID == "" || input.ProviderID == "" {
				return productdata.NewError(productdata.CodeInvalidRequest, "Continuation context is missing.")
			}
			r.Gateway.ContinueAfterToolResult(ctx, run, input)
			return r.gatewayResult(ctx, run.ID)
		}
		return nil
	}
	if existing.ExecutionStatus != productdata.ToolCallExecutionNotStarted {
		return nil
	}
	call, _, err := r.Gateway.Service.StartToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID)
	if err != nil {
		if terminal, terminalErr := runIsTerminal(ctx, r.Gateway.Service, run.ID); terminalErr != nil || terminal {
			return terminalErr
		}
		return err
	}
	enabledTools := []productdata.ToolResolution(nil)
	if prepared != nil {
		enabledTools = prepared.EnabledTools
	}
	catalog := toolCatalogForExecution(enabledTools)
	broker := ToolBroker{Executor: DefaultToolExecutor{MCPExecutor: r.MCPExecutor, WorkspaceExecutor: WorkspaceToolExecutor{}, SandboxExecutor: SandboxToolExecutor{}, LSPExecutor: LSPToolExecutor{}, WebExecutor: r.WebExecutor, BrowserExecutor: r.BrowserExecutor, ArtifactExecutor: ArtifactToolExecutor{Artifacts: r.artifactService()}, AgentExecutor: AgentToolExecutor{Tasks: r.agentTaskService()}, MemoryExecutor: MemoryToolExecutor{Service: r.memoryService(), Ident: identity.LocalDevIdentity()}}}
	result, err := broker.Execute(ctx, ToolInvocationFromCall(call, catalog, enabledTools))
	if err != nil {
		_, _, _ = r.Gateway.Service.FailToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, "tool_execution_failed", err.Error())
		return nil
	}
	if _, _, err = r.Gateway.Service.CompleteToolCallSuccess(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, result.ResultSummary); err != nil {
		return err
	}
	if result.ToolName == productdata.ToolNameTodoWrite {
		if event, ok := appendProviderWorkTodoSnapshot(ctx, r.Gateway.Service, run, result.ResultSummary); ok && r.broadcaster() != nil {
			r.broadcaster().Publish(event)
		}
	} else {
		_, _ = appendWorkTodoSnapshot(ctx, r.Gateway.Service, run, "runtime")
	}
	input := r.gatewayContinuationInput(ctx, run, toolCallID)
	if input.MessageID == "" || input.ProviderID == "" {
		return productdata.NewError(productdata.CodeInvalidRequest, "Continuation context is missing.")
	}
	r.Gateway.ContinueAfterToolResult(ctx, run, input)
	for {
		nextToolCallID := r.nextAutoApprovedToolCall(ctx, run.ID)
		if nextToolCallID == "" {
			break
		}
		if err := r.runApprovedTool(ctx, run, nextToolCallID, prepared); err != nil {
			return err
		}
	}
	return r.gatewayResult(ctx, run.ID)
}

func (r QueuedRunRouter) shouldResumeContinuationAfterSucceededTool(ctx context.Context, runID string, toolCallID string) bool {
	if r.Gateway == nil || r.Gateway.Service == nil || strings.TrimSpace(toolCallID) == "" {
		return false
	}
	events, err := r.Gateway.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil {
		return false
	}
	succeededSequence := 0
	for _, event := range events {
		if event.Type == productdata.EventToolCallSucceeded && metadataString(event.Metadata, "tool_call_id") == toolCallID {
			succeededSequence = event.Sequence
		}
	}
	if succeededSequence == 0 {
		return false
	}
	for _, event := range events {
		if event.Sequence <= succeededSequence {
			continue
		}
		if event.Category == productdata.RunEventCategoryFinal || event.Type == productdata.EventToolCallRequested {
			return false
		}
		if event.Type == "model_request_started" && metadataString(event.Metadata, "model_phase") == "continuation" {
			return false
		}
	}
	return true
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

func (r QueuedRunRouter) gatewayContinuationInput(ctx context.Context, run productdata.Run, toolCallID string) GatewayContinuationInput {
	input := GatewayContinuationInput{ThreadID: run.ThreadID, ToolCallID: toolCallID}
	if r.Gateway == nil || r.Gateway.Service == nil {
		return input
	}
	events, err := r.Gateway.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		return input
	}
	for _, event := range events {
		if event.Type != "run_created" {
			continue
		}
		input.MessageID = metadataString(event.Metadata, "message_id")
		input.ProviderID = metadataString(event.Metadata, "provider_id")
		input.Model = metadataString(event.Metadata, "model")
		return input
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
