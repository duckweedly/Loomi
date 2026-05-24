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
}

func NewLocalRunner(service productdata.Service, broadcaster *Broadcaster) *LocalRunner {
	return &LocalRunner{Service: service, Broadcaster: broadcaster, StepDelay: 200 * time.Millisecond}
}

func (r *LocalRunner) RunAsync(run productdata.Run, scriptName string) {
	if r == nil || r.Service == nil {
		return
	}
	go r.run(context.Background(), run, scriptName)
}

func (r *LocalRunner) run(ctx context.Context, run productdata.Run, scriptName string) {
	sim := NewSimulator(scriptName)
	for _, step := range sim.Steps() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(r.StepDelay):
		}
		current, err := r.Service.GetRun(ctx, identity.LocalDevIdentity(), run.ID)
		if err != nil || productdata.IsRunTerminal(current.Status) {
			return
		}
		event, err := r.Service.AppendRunEvent(ctx, identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{Category: step.Category, Type: step.Type, Summary: step.Summary, Content: step.Content, Metadata: step.Metadata})
		if err != nil {
			return
		}
		if r.Broadcaster != nil {
			r.Broadcaster.Publish(event)
		}
	}
}
