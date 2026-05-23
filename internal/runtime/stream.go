package runtime

import (
	"context"
	"sync"

	"github.com/sheridiany/loomi/internal/productdata"
)

type Broadcaster struct {
	mu          sync.Mutex
	subscribers map[string]map[chan productdata.RunEvent]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subscribers: map[string]map[chan productdata.RunEvent]struct{}{}}
}

func (b *Broadcaster) Subscribe(ctx context.Context, runID string) <-chan productdata.RunEvent {
	ch := make(chan productdata.RunEvent, 8)
	b.mu.Lock()
	if b.subscribers[runID] == nil {
		b.subscribers[runID] = map[chan productdata.RunEvent]struct{}{}
	}
	b.subscribers[runID][ch] = struct{}{}
	b.mu.Unlock()
	go func() {
		<-ctx.Done()
		b.mu.Lock()
		delete(b.subscribers[runID], ch)
		if len(b.subscribers[runID]) == 0 {
			delete(b.subscribers, runID)
		}
		b.mu.Unlock()
		close(ch)
	}()
	return ch
}

func (b *Broadcaster) Publish(event productdata.RunEvent) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers[event.RunID] {
		select {
		case ch <- event:
		default:
		}
	}
}

func CollectHistoryThenLive(ctx context.Context, history []productdata.RunEvent, live <-chan productdata.RunEvent, afterSequence int) []productdata.RunEvent {
	merged := make([]productdata.RunEvent, 0, len(history))
	seen := map[string]struct{}{}
	for _, event := range history {
		if event.Sequence <= afterSequence {
			continue
		}
		merged = append(merged, event)
		seen[event.ID] = struct{}{}
	}
	for {
		select {
		case <-ctx.Done():
			return merged
		case event, ok := <-live:
			if !ok {
				return merged
			}
			if event.Sequence <= afterSequence {
				continue
			}
			if _, ok := seen[event.ID]; ok {
				continue
			}
			merged = append(merged, event)
			seen[event.ID] = struct{}{}
		}
	}
}
