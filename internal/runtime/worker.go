package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type WorkerRunner interface {
	Run(context.Context, productdata.Run, productdata.BackgroundJob) error
}

type Worker struct {
	Service      productdata.Service
	Broadcaster  *Broadcaster
	Runner       WorkerRunner
	WorkerID     string
	LeaseSeconds int
	PollInterval time.Duration
}

func NewWorker(service productdata.Service, broadcaster *Broadcaster, runner WorkerRunner) *Worker {
	return &Worker{Service: service, Broadcaster: broadcaster, Runner: runner, WorkerID: fmt.Sprintf("worker_%d", time.Now().UnixNano()), LeaseSeconds: 30, PollInterval: 250 * time.Millisecond}
}

func (w *Worker) Start(ctx context.Context) {
	if w == nil || w.Service == nil || w.Runner == nil {
		return
	}
	go w.loop(ctx)
}

func (w *Worker) ProcessOne(ctx context.Context) (bool, error) {
	if w == nil || w.Service == nil || w.Runner == nil {
		return false, nil
	}
	coordinator := JobCoordinator{Service: w.Service, WorkerID: w.WorkerID, LeaseSeconds: w.LeaseSeconds}
	recovery, err := coordinator.RecoverExpired(ctx, 10)
	if err != nil {
		return false, err
	}
	w.publishRecoveryEvents(recovery)
	claim, err := coordinator.Claim(ctx)
	if err != nil || !claim.OK {
		return claim.OK, err
	}
	w.publishRunEvents(ctx, claim.Run.ID, 0)
	afterRenew := 0
	if w.Broadcaster != nil {
		afterRenew = w.highestRunSequence(ctx, claim.Run.ID)
	}
	if _, err := coordinator.RenewLease(ctx, claim.Job); err != nil {
		return true, err
	}
	w.publishRunEvents(ctx, claim.Run.ID, afterRenew)
	if err := w.Runner.Run(ctx, claim.Run, claim.Job); err != nil {
		afterFail := 0
		if w.Broadcaster != nil {
			afterFail = w.highestRunSequence(ctx, claim.Run.ID)
		}
		_, _ = coordinator.Fail(ctx, claim.Job, "worker_run_failed", "Worker run failed.")
		w.publishRunEvents(ctx, claim.Run.ID, afterFail)
		return true, err
	}
	_, err = coordinator.Complete(ctx, claim.Job)
	return true, err
}

func (w *Worker) publishRecoveryEvents(recovery JobRecoveryResult) {
	if w.Broadcaster == nil {
		return
	}
	for _, recovered := range recovery.Recoveries {
		for _, event := range recovered.Events {
			w.Broadcaster.Publish(event)
		}
	}
}

func (w *Worker) publishRunEvents(ctx context.Context, runID string, afterSequence int) {
	if w.Broadcaster == nil {
		return
	}
	events, err := w.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, afterSequence)
	if err != nil {
		return
	}
	for _, event := range events {
		w.Broadcaster.Publish(event)
	}
}

func (w *Worker) highestRunSequence(ctx context.Context, runID string) int {
	events, err := w.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, 0)
	if err != nil || len(events) == 0 {
		return 0
	}
	return events[len(events)-1].Sequence
}

func (w *Worker) loop(ctx context.Context) {
	interval := w.PollInterval
	if interval <= 0 {
		interval = 250 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		_, _ = w.ProcessOne(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
