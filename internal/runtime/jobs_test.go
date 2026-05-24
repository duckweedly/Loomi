package runtime

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestJobCoordinatorClaimMapsServiceResult(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Jobs", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	coordinator := JobCoordinator{Service: svc, WorkerID: "worker_test", LeaseSeconds: 5}

	result, err := coordinator.Claim(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.OK || result.Job.RunID != run.ID || result.Run.Status != productdata.RunStatusRunning {
		t.Fatalf("result = %+v", result)
	}
	if result.Job.LeasedBy == nil || *result.Job.LeasedBy != "worker_test" || result.Job.LeaseExpiresAt == nil {
		t.Fatalf("job lease = %+v", result.Job)
	}
}

func TestJobCoordinatorRecoversExpiredLeasesAndGuardsOwnership(t *testing.T) {
	svc := &jobCoordinatorFakeService{recoveries: []productdata.BackgroundJobRecovery{{Job: productdata.BackgroundJob{ID: "job_1", Status: productdata.BackgroundJobStatusQueued}}}}
	coordinator := JobCoordinator{Service: svc, WorkerID: "worker_test", LeaseSeconds: 7}

	recovery, err := coordinator.RecoverExpired(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(recovery.Recoveries) != 1 || svc.recoverLimit != 3 {
		t.Fatalf("recovery=%+v limit=%d", recovery, svc.recoverLimit)
	}
	job := productdata.BackgroundJob{ID: "job_1", OwnershipVersion: 4}
	if changed, err := coordinator.RenewLease(context.Background(), job); err != nil || !changed {
		t.Fatalf("renew changed=%v err=%v", changed, err)
	}
	if svc.renew.WorkerID != "worker_test" || svc.renew.OwnershipVersion != 4 || svc.renew.LeaseSeconds != 7 {
		t.Fatalf("renew input = %+v", svc.renew)
	}
	if changed, err := coordinator.Complete(context.Background(), job); err != nil || !changed {
		t.Fatalf("complete changed=%v err=%v", changed, err)
	}
	if svc.complete.WorkerID != "worker_test" || svc.complete.OwnershipVersion != 4 {
		t.Fatalf("complete input = %+v", svc.complete)
	}
}

func TestJobCoordinatorClaimReturnsEmptyWhenQueueEmpty(t *testing.T) {
	coordinator := JobCoordinator{Service: productdata.NewMemoryService(), WorkerID: "worker_test"}

	result, err := coordinator.Claim(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.OK || result.Job.ID != "" || result.Run.ID != "" {
		t.Fatalf("result = %+v", result)
	}
}

type jobCoordinatorFakeService struct {
	productdata.Service
	recoveries   []productdata.BackgroundJobRecovery
	recoverLimit int
	renew        productdata.RenewBackgroundJobLeaseInput
	complete     productdata.CompleteBackgroundJobInput
}

func (s *jobCoordinatorFakeService) RecoverBackgroundJobs(_ context.Context, _ identity.LocalIdentity, input productdata.RecoverBackgroundJobsInput) ([]productdata.BackgroundJobRecovery, error) {
	s.recoverLimit = input.Limit
	return s.recoveries, nil
}

func (s *jobCoordinatorFakeService) RenewBackgroundJobLease(_ context.Context, _ identity.LocalIdentity, input productdata.RenewBackgroundJobLeaseInput) (productdata.BackgroundJob, bool, error) {
	s.renew = input
	return productdata.BackgroundJob{ID: input.JobID}, true, nil
}

func (s *jobCoordinatorFakeService) CompleteBackgroundJob(_ context.Context, _ identity.LocalIdentity, input productdata.CompleteBackgroundJobInput) (productdata.BackgroundJob, bool, error) {
	s.complete = input
	return productdata.BackgroundJob{ID: input.JobID}, true, nil
}
