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

func metadataString(metadata map[string]any, key string) string {
	value, ok := metadata[key].(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}
