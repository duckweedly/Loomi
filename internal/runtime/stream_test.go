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
