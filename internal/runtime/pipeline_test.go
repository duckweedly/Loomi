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

func TestPipelineRecorderReturnsFalseWithoutService(t *testing.T) {
	recorder := PipelineRecorder{}

	if recorder.StepStarted(context.Background(), "run_missing", productdata.PipelineStepInvokeRuntime, nil) {
		t.Fatal("StepStarted() = true, want false")
	}
}
