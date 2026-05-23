package runtime

import "github.com/sheridiany/loomi/internal/productdata"

const DefaultScriptName = "m4_smoke"

type Step struct {
	Category productdata.RunEventCategory
	Type     string
	Summary  string
	Content  *string
	Metadata map[string]any
}

type Simulator struct {
	ScriptName string
}

func NewSimulator(scriptName string) Simulator {
	if scriptName == "" {
		scriptName = DefaultScriptName
	}
	return Simulator{ScriptName: scriptName}
}

func (s Simulator) Source() productdata.RunSource {
	return productdata.RunSourceLocalSimulated
}

func (s Simulator) Steps() []Step {
	content := "Local simulated response is ready."
	return []Step{
		{Category: productdata.RunEventCategoryLifecycle, Type: "run_started", Summary: "Run started", Metadata: map[string]any{"script_name": s.ScriptName}},
		{Category: productdata.RunEventCategoryProgress, Type: "context_loaded", Summary: "Context loaded", Metadata: map[string]any{"script_name": s.ScriptName}},
		{Category: productdata.RunEventCategoryProgress, Type: "drafting", Summary: "Drafting response", Metadata: map[string]any{"script_name": s.ScriptName}},
		{Category: productdata.RunEventCategoryMessage, Type: "assistant_message", Summary: "Simulated response", Content: &content, Metadata: map[string]any{"script_name": s.ScriptName}},
		{Category: productdata.RunEventCategoryFinal, Type: "run_completed", Summary: "Run completed", Metadata: map[string]any{"script_name": s.ScriptName}},
	}
}
