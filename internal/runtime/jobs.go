package runtime

import (
	"context"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type JobCoordinator struct {
	Service      productdata.Service
	WorkerID     string
	LeaseSeconds int
}

type JobClaimResult struct {
	Job productdata.BackgroundJob
	Run productdata.Run
	OK  bool
}

type JobRecoveryResult struct {
	Recoveries []productdata.BackgroundJobRecovery
}

func (c JobCoordinator) Claim(ctx context.Context) (JobClaimResult, error) {
	if c.Service == nil {
		return JobClaimResult{}, nil
	}
	job, run, ok, err := c.Service.ClaimBackgroundJob(ctx, identity.LocalDevIdentity(), productdata.ClaimBackgroundJobInput{WorkerID: c.workerID(), LeaseSeconds: c.LeaseSeconds})
	if err != nil {
		return JobClaimResult{}, err
	}
	return JobClaimResult{Job: job, Run: run, OK: ok}, nil
}

func (c JobCoordinator) RecoverExpired(ctx context.Context, limit int) (JobRecoveryResult, error) {
	if c.Service == nil {
		return JobRecoveryResult{}, nil
	}
	recoveries, err := c.Service.RecoverBackgroundJobs(ctx, identity.LocalDevIdentity(), productdata.RecoverBackgroundJobsInput{Limit: limit})
	if err != nil {
		return JobRecoveryResult{}, err
	}
	return JobRecoveryResult{Recoveries: recoveries}, nil
}

func (c JobCoordinator) RenewLease(ctx context.Context, job productdata.BackgroundJob) (bool, error) {
	if c.Service == nil {
		return false, nil
	}
	_, changed, err := c.Service.RenewBackgroundJobLease(ctx, identity.LocalDevIdentity(), productdata.RenewBackgroundJobLeaseInput{JobID: job.ID, WorkerID: c.workerID(), OwnershipVersion: job.OwnershipVersion, LeaseSeconds: c.LeaseSeconds})
	return changed, err
}

func (c JobCoordinator) Complete(ctx context.Context, job productdata.BackgroundJob) (bool, error) {
	if c.Service == nil {
		return false, nil
	}
	_, changed, err := c.Service.CompleteBackgroundJob(ctx, identity.LocalDevIdentity(), productdata.CompleteBackgroundJobInput{JobID: job.ID, WorkerID: c.workerID(), OwnershipVersion: job.OwnershipVersion})
	return changed, err
}

func (c JobCoordinator) Fail(ctx context.Context, job productdata.BackgroundJob, errorCode string, errorMessage string) (bool, error) {
	if c.Service == nil {
		return false, nil
	}
	_, changed, err := c.Service.FailBackgroundJob(ctx, identity.LocalDevIdentity(), productdata.FailBackgroundJobInput{JobID: job.ID, WorkerID: c.workerID(), OwnershipVersion: job.OwnershipVersion, ErrorCode: errorCode, ErrorMessage: errorMessage})
	return changed, err
}

func (c JobCoordinator) workerID() string {
	if c.WorkerID == "" {
		return "worker_local"
	}
	return c.WorkerID
}
