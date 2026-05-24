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

func (r PipelineRecorder) StepStarted(ctx context.Context, runID string, step productdata.PipelineStepName, metadata map[string]any) bool {
	return r.append(ctx, runID, productdata.EventPipelineStepStarted, "Pipeline step started", step, metadata)
}

func (r PipelineRecorder) StepCompleted(ctx context.Context, runID string, step productdata.PipelineStepName, metadata map[string]any) bool {
	return r.append(ctx, runID, productdata.EventPipelineStepCompleted, "Pipeline step completed", step, metadata)
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
