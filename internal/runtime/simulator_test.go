package runtime

import (
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestSimulatorProducesDeterministicLocalSteps(t *testing.T) {
	sim := NewSimulator("m4_smoke")
	steps := sim.Steps()
	if sim.Source() != productdata.RunSourceLocalSimulated {
		t.Fatalf("source = %q", sim.Source())
	}
	if len(steps) != 5 {
		t.Fatalf("len(steps) = %d", len(steps))
	}
	want := []productdata.RunEventCategory{
		productdata.RunEventCategoryLifecycle,
		productdata.RunEventCategoryProgress,
		productdata.RunEventCategoryProgress,
		productdata.RunEventCategoryMessage,
		productdata.RunEventCategoryFinal,
	}
	for i, category := range want {
		if steps[i].Category != category {
			t.Fatalf("step %d category = %q, want %q", i, steps[i].Category, category)
		}
	}
	if steps[4].Type != "run_completed" {
		t.Fatalf("final type = %q", steps[4].Type)
	}
}
