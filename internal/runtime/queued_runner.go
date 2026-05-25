package runtime

import (
	"context"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type QueuedRunRouter struct {
	Local   *LocalRunner
	Gateway *Gateway
}

func (r QueuedRunRouter) Run(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
	if toolCallID := metadataString(job.Metadata, "tool_call_id"); toolCallID != "" {
		return r.runApprovedTool(ctx, run, toolCallID)
	}
	if run.Source == productdata.RunSourceModelGateway {
		if r.Gateway != nil {
			r.Gateway.run(ctx, run, gatewayInputFromJob(run, job))
			return r.gatewayResult(ctx, run.ID)
		}
		return nil
	}
	if r.Local != nil {
		return r.Local.Run(ctx, run, job)
	}
	return nil
}

func (r QueuedRunRouter) runApprovedTool(ctx context.Context, run productdata.Run, toolCallID string) error {
	if r.Gateway == nil || r.Gateway.Service == nil {
		return nil
	}
	call, _, err := r.Gateway.Service.StartToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID)
	if err != nil {
		return err
	}
	tool := CurrentTimeToolDefinition()
	if call.ToolName != tool.Name {
		_, _, _ = r.Gateway.Service.FailToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, "unsupported_tool", "Tool is not supported.")
		return nil
	}
	args, err := tool.NormalizeArguments(call.ArgumentsSummary)
	if err != nil {
		_, _, _ = r.Gateway.Service.FailToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, "invalid_tool_arguments", err.Error())
		return nil
	}
	result, err := tool.Execute(args)
	if err != nil {
		_, _, _ = r.Gateway.Service.FailToolCallExecution(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, "tool_execution_failed", err.Error())
		return nil
	}
	if _, _, err = r.Gateway.Service.CompleteToolCallSuccess(ctx, identity.LocalDevIdentity(), run.ThreadID, run.ID, toolCallID, result); err != nil {
		return err
	}
	input := r.gatewayContinuationInput(ctx, run, toolCallID)
	if input.MessageID == "" || input.ProviderID == "" {
		return productdata.NewError(productdata.CodeInvalidRequest, "Continuation context is missing.")
	}
	r.Gateway.ContinueAfterToolResult(ctx, run, input)
	return r.gatewayResult(ctx, run.ID)
}

func (r QueuedRunRouter) gatewayResult(ctx context.Context, runID string) error {
	run, err := r.Gateway.Service.GetRun(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return err
	}
	if run.Status == productdata.RunStatusFailed || run.Status == productdata.RunStatusStopped {
		return productdata.NewError(productdata.CodeInternalError, "Queued gateway run did not complete.")
	}
	return nil
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
