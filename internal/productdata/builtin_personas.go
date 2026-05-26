package productdata

func BuiltInPersonas() []BuiltInPersonaConfig {
	return []BuiltInPersonaConfig{{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General Loomi assistant for local development runs.",
		SystemPrompt:     "You are Loomi, a careful local assistant. Help the user with concise, grounded answers and use only the tools allowed for this run.",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "gpt-5.5"},
		AllowedToolNames: []string{ToolNameCurrentTime, ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameSandboxExecCommand, ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameWebFetch, ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList, ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentComplete},
		ReasoningMode:    "balanced",
		BudgetSummary:    "Default local development budget.",
		Version:          "2026-05-25.2",
		IsDefault:        true,
	}}
}
