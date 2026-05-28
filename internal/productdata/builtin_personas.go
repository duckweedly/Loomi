package productdata

func BuiltInPersonas() []BuiltInPersonaConfig {
	return []BuiltInPersonaConfig{{
		Slug:             "loomi-default",
		Name:             "Loomi Default",
		Description:      "General Loomi assistant for local development runs.",
		SystemPrompt:     "You are Loomi, a careful local assistant. Answer first with the result or blocker. Keep final answers brief: usually 1-3 short paragraphs or a compact list when useful. Skip prefaces, avoid repeating the user's request, and do not pad with generic caveats. Use only the tools allowed for this run. Do not expose hidden chain-of-thought; share only concise progress or a safe summary when it helps.",
		ModelRoute:       PersonaModelRoute{ProviderID: "custom", Model: "gpt-5.5"},
		AllowedToolNames: []string{ToolNameCurrentTime, ToolNameLoadTools, ToolNameLoadSkill, ToolNameWorkspaceGlob, ToolNameWorkspaceGrep, ToolNameWorkspaceRead, ToolNameWorkspaceListDirectory, ToolNameWorkspaceTreeSummary, ToolNameWorkspaceWriteFile, ToolNameWorkspaceEdit, ToolNameWorkspacePatchPreview, ToolNameWorkspacePatchApply, ToolNameSandboxExecCommand, ToolNameSandboxStartProcess, ToolNameSandboxContinueProcess, ToolNameSandboxTerminateProcess, ToolNameLSPDiagnostics, ToolNameLSPSymbols, ToolNameLSPReferences, ToolNameLSPDefinition, ToolNameLSPHover, ToolNameWebFetch, ToolNameWebSearch, ToolNameBrowserOpen, ToolNameBrowserSnapshot, ToolNameBrowserClickLink, ToolNameBrowserScreenshot, ToolNameBrowserType, ToolNameBrowserPress, ToolNameArtifactCreateText, ToolNameArtifactRead, ToolNameArtifactList, ToolNameAgentSpawn, ToolNameAgentList, ToolNameAgentComplete, ToolNameMemorySearch, ToolNameMemoryList, ToolNameMemoryRead, ToolNameMemoryWrite, ToolNameMemoryEdit, ToolNameMemoryForget, ToolNameMemoryContext, ToolNameMemoryTimeline, ToolNameMemoryConnections, ToolNameMemoryThreadSearch, ToolNameMemoryThreadFetch, ToolNameMemoryStatus, ToolNameNotebookRead, ToolNameNotebookWrite, ToolNameNotebookEdit, ToolNameNotebookForget, ToolNameTodoWrite},
		ReasoningMode:    "balanced",
		BudgetSummary:    "Default local development budget.",
		Version:          "2026-05-27.2",
		IsDefault:        true,
	}}
}
