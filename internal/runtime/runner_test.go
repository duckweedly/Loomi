package runtime

import "testing"

func TestM5NormalizedEventConstants(t *testing.T) {
	want := []string{
		EventModelRequestStarted,
		EventModelOutputDelta,
		EventModelOutputCompleted,
		EventModelRefusal,
		EventToolCallBlocked,
		EventProviderError,
		EventProviderTimeout,
		EventProviderRateLimited,
	}
	for _, eventType := range want {
		if eventType == "" {
			t.Fatalf("event constants = %+v", want)
		}
	}
}
