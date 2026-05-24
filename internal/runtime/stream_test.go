package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestBroadcasterDeliversLiveEventsToRunSubscribers(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := NewBroadcaster()
	events := b.Subscribe(ctx, "run_1")
	b.Publish(productdata.RunEvent{ID: "evt_1", RunID: "run_1", Sequence: 1})
	select {
	case got := <-events:
		if got.ID != "evt_1" || got.Sequence != 1 {
			t.Fatalf("event = %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for live event")
	}
}

func TestMergeHistoryThenLiveSkipsDeliveredSequences(t *testing.T) {
	history := []productdata.RunEvent{
		{ID: "evt_1", RunID: "run_1", Sequence: 1},
		{ID: "evt_2", RunID: "run_1", Sequence: 2},
	}
	live := make(chan productdata.RunEvent, 1)
	live <- productdata.RunEvent{ID: "evt_3", RunID: "run_1", Sequence: 3}
	close(live)
	merged := CollectHistoryThenLive(context.Background(), history, live, 1)
	if len(merged) != 2 || merged[0].ID != "evt_2" || merged[1].ID != "evt_3" {
		t.Fatalf("merged = %+v", merged)
	}
}

func TestMergeHistoryThenLiveKeepsM6WorkerEventOrder(t *testing.T) {
	history := []productdata.RunEvent{
		{ID: "evt_queued", RunID: "run_1", Sequence: 2, Type: productdata.EventRunQueued},
		{ID: "evt_claimed", RunID: "run_1", Sequence: 3, Type: productdata.EventJobClaimed},
		{ID: "evt_step_started", RunID: "run_1", Sequence: 4, Type: productdata.EventPipelineStepStarted},
	}
	live := make(chan productdata.RunEvent, 2)
	live <- productdata.RunEvent{ID: "evt_step_completed", RunID: "run_1", Sequence: 5, Type: productdata.EventPipelineStepCompleted}
	live <- productdata.RunEvent{ID: "evt_step_started", RunID: "run_1", Sequence: 4, Type: productdata.EventPipelineStepStarted}
	close(live)

	merged := CollectHistoryThenLive(context.Background(), history, live, 1)
	got := make([]string, 0, len(merged))
	for _, event := range merged {
		got = append(got, event.Type)
	}
	want := []string{productdata.EventRunQueued, productdata.EventJobClaimed, productdata.EventPipelineStepStarted, productdata.EventPipelineStepCompleted}
	if len(got) != len(want) {
		t.Fatalf("merged = %+v", merged)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got = %+v want = %+v", got, want)
		}
	}
}
