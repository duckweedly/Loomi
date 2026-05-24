package runtime

import (
	"context"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	EventModelRequestStarted  = "model_request_started"
	EventModelOutputDelta     = "model_output_delta"
	EventModelOutputCompleted = "model_output_completed"
	EventModelRefusal         = "model_refusal"
	EventToolCallBlocked      = "tool_call_blocked"
	EventProviderError        = "provider_error"
	EventProviderTimeout      = "provider_timeout"
	EventProviderRateLimited  = "provider_rate_limited"
)

type LocalRunner struct {
	Service     productdata.Service
	Broadcaster *Broadcaster
	StepDelay   time.Duration
	Pipeline    PipelineRecorder
}

func NewLocalRunner(service productdata.Service, broadcaster *Broadcaster) *LocalRunner {
	return &LocalRunner{Service: service, Broadcaster: broadcaster, StepDelay: 200 * time.Millisecond, Pipeline: PipelineRecorder{Service: service, Broadcaster: broadcaster}}
}

func (r *LocalRunner) RunAsync(run productdata.Run, scriptName string) {
	if r == nil || r.Service == nil {
		return
	}
	go r.run(context.Background(), run, productdata.BackgroundJob{}, scriptName)
}

func (r *LocalRunner) Run(ctx context.Context, run productdata.Run, job productdata.BackgroundJob) error {
	return r.run(ctx, run, job, "")
}

func ownedWorkerID(job productdata.BackgroundJob) string {
	if job.LeasedBy == nil {
		return ""
	}
	return *job.LeasedBy
}

func (r *LocalRunner) run(ctx context.Context, run productdata.Run, job productdata.BackgroundJob, scriptName string) error {
	sim := NewSimulator(scriptName)
	current, err := r.Service.GetRun(ctx, identity.LocalDevIdentity(), run.ID)
	if err != nil {
		return err
	}
	if productdata.IsRunTerminal(current.Status) {
		return nil
	}
	if job.ID != "" {
		if err := r.renewOwnedJob(ctx, job); err != nil {
			return err
		}
	}
	if !r.Pipeline.StepStarted(ctx, run.ID, productdata.PipelineStepInvokeRuntime, map[string]any{}) {
		return productdata.NewError(productdata.CodeInternalError, "Pipeline step could not be recorded.")
	}
	steps := sim.Steps()
	for index, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(r.StepDelay):
		}
		current, err := r.Service.GetRun(ctx, identity.LocalDevIdentity(), run.ID)
		if err != nil {
			return err
		}
		if productdata.IsRunTerminal(current.Status) {
			return nil
		}
		if job.ID != "" {
			if err := r.renewOwnedJob(ctx, job); err != nil {
				return err
			}
		}
		if step.Category == productdata.RunEventCategoryMessage {
			if err := r.appendAssistantMessage(ctx, run, sim.ScriptName, contentString(step.Content)); err != nil {
				return err
			}
		}
		if step.Category == productdata.RunEventCategoryFinal && index == len(steps)-1 {
			if !r.Pipeline.StepCompleted(ctx, run.ID, productdata.PipelineStepInvokeRuntime, map[string]any{}) {
				return productdata.NewError(productdata.CodeInternalError, "Pipeline step could not be completed.")
			}
		}
		event, err := r.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{Category: step.Category, Type: step.Type, Summary: step.Summary, Content: step.Content, Metadata: step.Metadata})
		if err != nil {
			return err
		}
		if r.Broadcaster != nil {
			r.Broadcaster.Publish(event)
		}
	}
	return nil
}

func (r *LocalRunner) appendAssistantMessage(ctx context.Context, run productdata.Run, scriptName string, content string) error {
	metadata := map[string]any{"run_id": run.ID, "script_name": scriptName}
	_, err := r.Service.AppendAssistantMessage(ctx, identity.LocalDevIdentity(), run.ThreadID, productdata.AppendAssistantMessageInput{Content: content, Metadata: metadata})
	if err == nil {
		return nil
	}
	if productdata.ErrorCode(err) == productdata.CodeInvalidRequest {
		return nil
	}
	return err
}

func (r *LocalRunner) renewOwnedJob(ctx context.Context, job productdata.BackgroundJob) error {
	_, changed, err := r.Service.RenewBackgroundJobLease(ctx, identity.LocalDevIdentity(), productdata.RenewBackgroundJobLeaseInput{JobID: job.ID, WorkerID: ownedWorkerID(job), OwnershipVersion: job.OwnershipVersion, LeaseSeconds: 30})
	if err != nil {
		return err
	}
	if !changed {
		return productdata.NewError(productdata.CodeInvalidRequest, "Worker no longer owns the job.")
	}
	return nil
}

func contentString(content *string) string {
	if content == nil {
		return ""
	}
	return *content
}
