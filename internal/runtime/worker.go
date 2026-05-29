package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type WorkerRunner interface {
	Run(context.Context, productdata.Run, productdata.BackgroundJob) error
}

type runStepStateProjectionEnsurer interface {
	EnsureRunStepStateProjection(context.Context, identity.LocalIdentity, string) error
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
	if tasks, ok := w.Service.(productdata.AgentTaskService); ok {
		reconciled, err := tasks.ReconcileAgentTaskChildRuns(ctx, identity.LocalDevIdentity(), 10)
		if err != nil {
			return false, err
		}
		w.publishAgentTaskReconciliationEvents(reconciled)
	}
	recovery, err := coordinator.RecoverExpired(ctx, 10)
	if err != nil {
		return false, err
	}
	w.publishRecoveryEvents(recovery)
	claim, err := coordinator.Claim(ctx)
	if err != nil || !claim.OK {
		return claim.OK, err
	}
	if ensurer, ok := w.Service.(runStepStateProjectionEnsurer); ok {
		if err := ensurer.EnsureRunStepStateProjection(ctx, identity.LocalDevIdentity(), claim.Run.ID); err != nil {
			return true, err
		}
	}
	claimPublishCursor, err := w.claimPublishCursor(ctx, claim.Run.ID)
	if err != nil {
		return true, err
	}
	afterRenew := w.publishRunEvents(ctx, claim.Run.ID, claimPublishCursor)
	if _, err := coordinator.RenewLease(ctx, claim.Job); err != nil {
		return true, err
	}
	w.publishRunEvents(ctx, claim.Run.ID, afterRenew)
	runCtx, cancelRun := context.WithCancel(ctx)
	stopHeartbeat := w.startLeaseHeartbeat(runCtx, coordinator, claim.Job, cancelRun)
	stopWatcher := w.startRunStopWatcher(runCtx, claim.Run.ID, cancelRun)
	err = w.Runner.Run(runCtx, claim.Run, claim.Job)
	stopWatcher()
	stopHeartbeat()
	if err != nil {
		afterFail := w.publishRunEvents(ctx, claim.Run.ID, afterRenew)
		if _, failErr := coordinator.Fail(ctx, claim.Job, "worker_run_failed", "Worker run failed."); failErr != nil {
			w.publishRunEvents(ctx, claim.Run.ID, afterFail)
			return true, errors.Join(err, failErr)
		}
		w.publishRunEvents(ctx, claim.Run.ID, afterFail)
		return true, err
	}
	_, err = coordinator.Complete(ctx, claim.Job)
	return true, err
}

func (w *Worker) claimPublishCursor(ctx context.Context, runID string) (int, error) {
	if w.Broadcaster == nil {
		return 0, nil
	}
	state, err := w.Service.GetRunStepState(ctx, identity.LocalDevIdentity(), runID)
	if err != nil {
		return 0, err
	}
	if state.LastEventSequence <= 0 {
		return 0, nil
	}
	return state.LastEventSequence - 1, nil
}

func (w *Worker) startLeaseHeartbeat(ctx context.Context, coordinator JobCoordinator, job productdata.BackgroundJob, cancelRun context.CancelFunc) context.CancelFunc {
	leaseSeconds := w.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	interval := time.Duration(leaseSeconds) * time.Second / 2
	if interval < 100*time.Millisecond {
		interval = 100 * time.Millisecond
	}
	heartbeatCtx, stop := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-ticker.C:
				changed, err := coordinator.RenewLease(heartbeatCtx, job)
				if err != nil || !changed {
					cancelRun()
					return
				}
			}
		}
	}()
	return stop
}

func (w *Worker) startRunStopWatcher(ctx context.Context, runID string, cancelRun context.CancelFunc) context.CancelFunc {
	if w.Service == nil || strings.TrimSpace(runID) == "" {
		return func() {}
	}
	watcherCtx, stop := context.WithCancel(ctx)
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-watcherCtx.Done():
				return
			case <-ticker.C:
				run, err := w.Service.GetRun(watcherCtx, identity.LocalDevIdentity(), runID)
				if err != nil {
					continue
				}
				if run.StopRequestedAt != nil || productdata.IsRunTerminal(run.Status) {
					cancelRun()
					return
				}
			}
		}
	}()
	return stop
}

func (w *Worker) publishAgentTaskReconciliationEvents(reconciled []productdata.AgentTaskChildRunReconciliation) {
	if w.Broadcaster == nil {
		return
	}
	for _, item := range reconciled {
		for _, event := range item.Events {
			w.Broadcaster.Publish(event)
		}
	}
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

func (w *Worker) publishRunEvents(ctx context.Context, runID string, afterSequence int) int {
	if w.Broadcaster == nil {
		return afterSequence
	}
	events, err := w.Service.ListRunEvents(ctx, identity.LocalDevIdentity(), runID, afterSequence)
	if err != nil {
		return afterSequence
	}
	highest := afterSequence
	for _, event := range events {
		w.Broadcaster.Publish(event)
		if event.Sequence > highest {
			highest = event.Sequence
		}
	}
	return highest
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
