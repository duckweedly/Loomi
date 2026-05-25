package runtime

import (
	"context"
	"testing"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestPipelineRecorderPersistsAndPublishesStepEvents(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Pipeline", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	broadcaster := NewBroadcaster()
	published := broadcaster.Subscribe(context.Background(), run.ID)
	recorder := PipelineRecorder{Service: svc, Broadcaster: broadcaster}

	if !recorder.StepStarted(context.Background(), run.ID, productdata.PipelineStepInvokeRuntime, map[string]any{"token": "secret"}) {
		t.Fatal("StepStarted() = false")
	}
	if !recorder.StepCompleted(context.Background(), run.ID, productdata.PipelineStepInvokeRuntime, nil) {
		t.Fatal("StepCompleted() = false")
	}

	first := <-published
	second := <-published
	if first.Type != productdata.EventPipelineStepStarted || second.Type != productdata.EventPipelineStepCompleted {
		t.Fatalf("published = %+v %+v", first, second)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	started := events[len(events)-2]
	if started.Metadata["step"] != string(productdata.PipelineStepInvokeRuntime) || started.Metadata["token"] != "[redacted]" {
		t.Fatalf("started metadata = %+v", started.Metadata)
	}
}

func TestPipelineExecutesInsertedStagesInOrder(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Pipeline", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_pipeline", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	state := &PipelineState{RunContext: ctxData}
	var order []productdata.PipelineStepName
	pipeline := Pipeline{Recorder: PipelineRecorder{Service: svc}}

	err = pipeline.Execute(context.Background(), state, []PipelineStage{
		{Name: productdata.PipelineStepPrepareContext, Run: func(_ context.Context, _ *PipelineState) error {
			order = append(order, productdata.PipelineStepPrepareContext)
			return nil
		}},
		{Name: productdata.PipelineStepName("test_stage"), Run: func(_ context.Context, _ *PipelineState) error {
			order = append(order, productdata.PipelineStepName("test_stage"))
			return nil
		}},
		{Name: productdata.PipelineStepFinalize, Run: func(_ context.Context, _ *PipelineState) error {
			order = append(order, productdata.PipelineStepFinalize)
			return nil
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := []productdata.PipelineStepName{order[0], order[1], order[2]}; got[0] != productdata.PipelineStepPrepareContext || got[1] != "test_stage" || got[2] != productdata.PipelineStepFinalize {
		t.Fatalf("order = %+v", order)
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	var found bool
	for _, event := range events {
		if event.Type == productdata.EventPipelineStepCompleted && event.Metadata["step"] == "test_stage" {
			found = true
		}
	}
	if !found {
		t.Fatalf("inserted stage event not found: %+v", events)
	}
}

func TestPipelineShortCircuitsAndRecordsFailedStage(t *testing.T) {
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	thread, err := svc.CreateThread(context.Background(), ident, productdata.CreateThreadInput{Title: "Pipeline failure", Mode: productdata.ThreadModeChat})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.StartRun(context.Background(), ident, thread.ID, productdata.StartRunInput{})
	if err != nil {
		t.Fatal(err)
	}
	job, _, ok, err := svc.ClaimBackgroundJob(context.Background(), ident, productdata.ClaimBackgroundJobInput{WorkerID: "worker_pipeline", LeaseSeconds: 5})
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("claim ok = false")
	}
	ctxData, err := svc.PrepareRunContext(context.Background(), ident, job)
	if err != nil {
		t.Fatal(err)
	}
	state := &PipelineState{RunContext: ctxData}
	pipeline := Pipeline{Recorder: PipelineRecorder{Service: svc}}

	err = pipeline.Execute(context.Background(), state, []PipelineStage{
		{Name: productdata.PipelineStepPrepareContext, Run: func(_ context.Context, _ *PipelineState) error {
			return productdata.NewError(productdata.CodeInvalidRequest, "missing context")
		}},
		{Name: productdata.PipelineStepFinalize, Run: func(_ context.Context, _ *PipelineState) error {
			t.Fatal("finalize should not run")
			return nil
		}},
	})
	if err == nil {
		t.Fatal("Execute() err = nil")
	}
	events, err := svc.ListRunEvents(context.Background(), ident, run.ID, 0)
	if err != nil {
		t.Fatal(err)
	}
	last := events[len(events)-1]
	if last.Type != productdata.EventPipelineStepFailed || last.Metadata["step"] != string(productdata.PipelineStepPrepareContext) {
		t.Fatalf("last event = %+v", last)
	}
}

func TestPipelineRecorderReturnsFalseWithoutService(t *testing.T) {
	recorder := PipelineRecorder{}

	if recorder.StepStarted(context.Background(), "run_missing", productdata.PipelineStepInvokeRuntime, nil) {
		t.Fatal("StepStarted() = true, want false")
	}
}
