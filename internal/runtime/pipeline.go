package runtime

import (
	"context"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type PipelineRecorder struct {
	Service     productdata.Service
	Broadcaster *Broadcaster
}

type PipelineState struct {
	RunContext productdata.RunContext
}

type PipelineStage struct {
	Name productdata.PipelineStepName
	Run  func(context.Context, *PipelineState) error
}

type Pipeline struct {
	Recorder PipelineRecorder
}

func (p Pipeline) Execute(ctx context.Context, state *PipelineState, stages []PipelineStage) error {
	for _, stage := range stages {
		if !p.Recorder.StepStarted(ctx, state.RunContext.Run.ID, stage.Name, p.stageMetadata(state, stage.Name)) {
			return productdata.NewError(productdata.CodeInternalError, "Pipeline step could not be recorded.")
		}
		if stage.Run != nil {
			if err := stage.Run(ctx, state); err != nil {
				p.Recorder.StepFailed(ctx, state.RunContext.Run.ID, stage.Name, map[string]any{"error_code": string(productdata.ErrorCode(err))})
				return err
			}
		}
		if !p.Recorder.StepCompleted(ctx, state.RunContext.Run.ID, stage.Name, p.stageMetadata(state, stage.Name)) {
			return productdata.NewError(productdata.CodeInternalError, "Pipeline step could not be completed.")
		}
	}
	return nil
}

func (p Pipeline) stageMetadata(state *PipelineState, stage productdata.PipelineStepName) map[string]any {
	if state == nil {
		return map[string]any{}
	}
	metadata := map[string]any{
		"job_id":  state.RunContext.Job.ID,
		"attempt": state.RunContext.Job.AttemptCount,
	}
	switch stage {
	case productdata.PipelineStepPrepareContext:
		for key, value := range state.RunContext.SafeSummary() {
			metadata[key] = value
		}
	case productdata.PipelineStepResolveTools:
		for key, value := range state.RunContext.ToolResolutionSummary() {
			metadata[key] = value
		}
	case productdata.PipelineStepInvokeRuntime:
		metadata["source"] = string(state.RunContext.Run.Source)
		if state.RunContext.ProviderRoute.ProviderID != "" {
			metadata["provider_id"] = state.RunContext.ProviderRoute.ProviderID
		}
		if state.RunContext.ProviderRoute.Model != "" {
			metadata["model"] = state.RunContext.ProviderRoute.Model
		}
	case productdata.PipelineStepFinalize:
		metadata["terminal_outcome_pending"] = true
	}
	return metadata
}

func (r PipelineRecorder) StepStarted(ctx context.Context, runID string, step productdata.PipelineStepName, metadata map[string]any) bool {
	return r.append(ctx, runID, productdata.EventPipelineStepStarted, "Pipeline step started", step, metadata)
}

func (r PipelineRecorder) StepCompleted(ctx context.Context, runID string, step productdata.PipelineStepName, metadata map[string]any) bool {
	return r.append(ctx, runID, productdata.EventPipelineStepCompleted, "Pipeline step completed", step, metadata)
}

func (r PipelineRecorder) StepFailed(ctx context.Context, runID string, step productdata.PipelineStepName, metadata map[string]any) bool {
	return r.append(ctx, runID, productdata.EventPipelineStepFailed, "Pipeline step failed", step, metadata)
}

func (r PipelineRecorder) append(ctx context.Context, runID string, eventType string, summary string, step productdata.PipelineStepName, metadata map[string]any) bool {
	if r.Service == nil {
		return false
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	metadata["step"] = string(step)
	event, err := r.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), runID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: eventType, Summary: summary, Metadata: metadata})
	if err != nil {
		return false
	}
	if r.Broadcaster != nil {
		r.Broadcaster.Publish(event)
	}
	return true
}
