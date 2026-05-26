package runtime

import (
	"context"
	"errors"

	"github.com/sheridiany/loomi/internal/productdata"
)

type ToolInvocation struct {
	ThreadID            string
	RunID               string
	ToolCallID          string
	ToolName            string
	CandidateSchemaHash string
	ArgumentsSummary    map[string]any
	ApprovalStatus      productdata.ToolCallApprovalStatus
	ExecutionStatus     productdata.ToolCallExecutionStatus
	Catalog             []productdata.ToolCatalogEntry
	EnabledTools        []productdata.ToolResolution
}

type ToolResult struct {
	ToolName      string
	ToolCallID    string
	ResultSummary map[string]any
}

type ToolExecutor interface {
	ExecuteTool(context.Context, ToolInvocation) (ToolResult, error)
}

type ToolBroker struct {
	Executor ToolExecutor
}

type DefaultToolExecutor struct {
	DiscoveryExecutor DiscoveryToolExecutor
	MCPExecutor       MCPToolExecutor
	WorkspaceExecutor WorkspaceToolExecutor
	SandboxExecutor   SandboxToolExecutor
	LSPExecutor       LSPToolExecutor
	WebExecutor       WebToolExecutor
	BrowserExecutor   BrowserToolExecutor
	ArtifactExecutor  ArtifactToolExecutor
	AgentExecutor     AgentToolExecutor
}

func ToolInvocationFromCall(call productdata.ToolCall, catalog []productdata.ToolCatalogEntry, enabledTools []productdata.ToolResolution) ToolInvocation {
	return ToolInvocation{
		ThreadID:            call.ThreadID,
		RunID:               call.RunID,
		ToolCallID:          call.ToolCallID,
		ToolName:            call.ToolName,
		CandidateSchemaHash: call.CandidateSchemaHash,
		ArgumentsSummary:    call.ArgumentsSummary,
		ApprovalStatus:      call.ApprovalStatus,
		ExecutionStatus:     call.ExecutionStatus,
		Catalog:             catalog,
		EnabledTools:        enabledTools,
	}
}

func (b ToolBroker) Execute(ctx context.Context, invocation ToolInvocation) (ToolResult, error) {
	if invocation.ApprovalStatus != productdata.ToolCallApprovalApproved || invocation.ExecutionStatus != productdata.ToolCallExecutionExecuting {
		return ToolResult{}, errors.New("tool invocation is not approved for execution")
	}
	entry, ok := catalogEntryByName(invocation.Catalog, invocation.ToolName)
	if !ok {
		return ToolResult{}, errors.New("tool is not in catalog")
	}
	if !entry.Enabled || entry.ExecutionState != productdata.ToolExecutionStateExecutable {
		return ToolResult{}, errors.New("tool is disabled")
	}
	resolution, ok := toolResolutionByName(invocation.EnabledTools, invocation.ToolName)
	if !ok || resolution.ExecutionState != string(productdata.ToolExecutionStateExecutable) {
		return ToolResult{}, errors.New("tool is not allowed for this run")
	}
	if productdata.IsMCPToolName(invocation.ToolName) {
		if invocation.CandidateSchemaHash == "" || entry.InputSchemaHash == "" || resolution.InputSchemaHash == "" {
			return ToolResult{}, errors.New("mcp tool schema hash is missing")
		}
		if invocation.CandidateSchemaHash != entry.InputSchemaHash || invocation.CandidateSchemaHash != resolution.InputSchemaHash {
			return ToolResult{}, errors.New("mcp tool schema hash mismatch")
		}
	}
	if b.Executor == nil {
		return ToolResult{}, errors.New("tool executor is unavailable")
	}
	result, err := b.Executor.ExecuteTool(ctx, invocation)
	if err != nil {
		return ToolResult{}, err
	}
	result.ToolName = invocation.ToolName
	result.ToolCallID = invocation.ToolCallID
	result.ResultSummary = safeToolResultSummary(result.ResultSummary)
	return result, nil
}

func (e DefaultToolExecutor) ExecuteTool(ctx context.Context, invocation ToolInvocation) (ToolResult, error) {
	if productdata.IsDiscoveryToolName(invocation.ToolName) {
		result, err := e.DiscoveryExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsMCPToolName(invocation.ToolName) {
		if e.MCPExecutor == nil {
			return ToolResult{}, errors.New("mcp executor is unavailable")
		}
		result, err := e.MCPExecutor.ExecuteMCPTool(ctx, productdata.ToolCall{
			ThreadID:            invocation.ThreadID,
			RunID:               invocation.RunID,
			ToolCallID:          invocation.ToolCallID,
			ToolName:            invocation.ToolName,
			CandidateSchemaHash: invocation.CandidateSchemaHash,
			ArgumentsSummary:    invocation.ArgumentsSummary,
			ApprovalStatus:      invocation.ApprovalStatus,
			ExecutionStatus:     invocation.ExecutionStatus,
		})
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: RedactMCPSummary(result)}, nil
	}
	if productdata.IsWorkspaceToolName(invocation.ToolName) {
		result, err := e.WorkspaceExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsSandboxToolName(invocation.ToolName) {
		result, err := e.SandboxExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsLSPToolName(invocation.ToolName) {
		result, err := e.LSPExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsWebToolName(invocation.ToolName) {
		result, err := e.WebExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsBrowserToolName(invocation.ToolName) {
		result, err := e.BrowserExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsArtifactToolName(invocation.ToolName) {
		result, err := e.ArtifactExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsAgentToolName(invocation.ToolName) {
		result, err := e.AgentExecutor.Execute(ctx, invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	if productdata.IsTodoToolName(invocation.ToolName) {
		result, err := ExecuteTodoWrite(invocation)
		if err != nil {
			return ToolResult{}, err
		}
		return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
	}
	tool := CurrentTimeToolDefinition()
	if invocation.ToolName != tool.Name {
		return ToolResult{}, errors.New("tool is not supported")
	}
	args, err := tool.NormalizeArguments(invocation.ArgumentsSummary)
	if err != nil {
		return ToolResult{}, err
	}
	result, err := tool.Execute(args)
	if err != nil {
		return ToolResult{}, err
	}
	return ToolResult{ToolName: invocation.ToolName, ToolCallID: invocation.ToolCallID, ResultSummary: result}, nil
}

func catalogEntryByName(catalog []productdata.ToolCatalogEntry, name string) (productdata.ToolCatalogEntry, bool) {
	for _, entry := range catalog {
		if entry.Name == name {
			return entry, true
		}
	}
	return productdata.ToolCatalogEntry{}, false
}

func toolResolutionByName(tools []productdata.ToolResolution, name string) (productdata.ToolResolution, bool) {
	for _, tool := range tools {
		if tool.Name == name {
			return tool, true
		}
	}
	return productdata.ToolResolution{}, false
}

func safeToolResultSummary(result map[string]any) map[string]any {
	redacted := productdata.RedactEventMetadata(result)
	safe := map[string]any{}
	for key, value := range redacted {
		if text, ok := value.(string); ok && text == "[redacted]" {
			continue
		}
		safe[key] = value
	}
	return safe
}
