package productdata

func BuiltInPersonas() []BuiltInPersonaConfig {
	return []BuiltInPersonaConfig{{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General Loomi assistant for local development runs.",
		SystemPrompt:     "You are Loomi, a careful local assistant. Help the user with concise, grounded answers and use only the tools allowed for this run.",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "gpt-5.5"},
		AllowedToolNames: []string{ToolNameCurrentTime, ToolNameLoadTools, ToolNameLoadSkill, ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchPreview, ToolNameWorkspacePatchApply, ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess, ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameLSPDefinition, ToolNameLSPHover, ToolNameWebFetch, ToolNameWebSearch, ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameBrowserScreenshot, ToolNameBrowserType, ToolNameBrowserPress, ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList, ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentComplete, ToolNameTodoWrite},
		ReasoningMode:    "balanced",
		BudgetSummary:    "Default local development budget.",
		Version:          "2026-05-26.1",
		IsDefault:        true,
	}}
}
