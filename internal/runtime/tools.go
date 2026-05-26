package runtime

import (
	"errors"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

type ToolApprovalPolicy string

type ToolSafetyClass string

const (
	ToolApprovalAlwaysRequired ToolApprovalPolicy = "always_required"
	ToolApprovalNotRequired    ToolApprovalPolicy = "not_required"

	ToolSafetyNoSideEffectInternal ToolSafetyClass = "no_side_effect_internal"
	ToolSafetyWorkspaceMutation    ToolSafetyClass = "workspace_mutation"
	ToolSafetySandboxCommand       ToolSafetyClass = "sandbox_command"
	ToolSafetyPublicNetworkRead    ToolSafetyClass = "public_network_read"
)

type ToolArguments struct {
	Timezone string
}

type ToolDefinition struct {
	Name           string
	ApprovalPolicy ToolApprovalPolicy
	SafetyClass    ToolSafetyClass
	Source         ToolSource
	ExecutionState ToolExecutionState
}

func CurrentTimeToolDefinition() ToolDefinition {
	return ToolDefinition{Name: "runtime.get_current_time", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted}
}

func ToolResolutionsForPersona(allowedToolNames []string) []productdata.ToolResolution {
	resolutions := make([]productdata.ToolResolution, 0, len(allowedToolNames))
	for _, name := range allowedToolNames {
		if productdata.IsDiscoveryToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalNotRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupDiscovery), RiskLevel: string(productdata.ToolRiskLow)})
			continue
		}
		if productdata.IsWorkspaceToolName(name) {
			riskLevel := productdata.ToolRiskLow
			if name == productdata.ToolNameWorkspaceWriteFile || name == productdata.ToolNameWorkspaceEdit || name == productdata.ToolNameWorkspacePatchPreview || name == productdata.ToolNameWorkspacePatchApply {
				riskLevel = productdata.ToolRiskHigh
			}
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupWorkspace), RiskLevel: string(riskLevel)})
			continue
		}
		if productdata.IsSandboxToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupSandbox), RiskLevel: string(productdata.ToolRiskHigh)})
			continue
		}
		if productdata.IsLSPToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupLSP), RiskLevel: string(productdata.ToolRiskLow)})
			continue
		}
		if productdata.IsWebToolName(name) {
			approvalPolicy := string(ToolApprovalAlwaysRequired)
			if name == productdata.ToolNameWebSearch {
				approvalPolicy = string(ToolApprovalNotRequired)
			}
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: approvalPolicy, ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupWeb), RiskLevel: string(productdata.ToolRiskMedium)})
			continue
		}
		if productdata.IsBrowserToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupBrowser), RiskLevel: string(productdata.ToolRiskMedium)})
			continue
		}
		if productdata.IsArtifactToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupArtifact), RiskLevel: string(productdata.ToolRiskMedium)})
			continue
		}
		if productdata.IsAgentToolName(name) {
			resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: string(productdata.ToolExecutionStateExecutable), Source: string(productdata.ToolCatalogSourceBuiltin), Group: string(productdata.ToolCatalogGroupAgent), RiskLevel: string(productdata.ToolRiskMedium)})
			continue
		}
		if name != productdata.ToolNameCurrentTime {
			if IsMCPToolName(name) {
				resolutions = append(resolutions, productdata.ToolResolution{Name: name, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: "discovered_non_executable"})
			}
			continue
		}
		resolutions = append(resolutions, productdata.ToolResolution{Name: productdata.ToolNameCurrentTime, ApprovalPolicy: string(ToolApprovalAlwaysRequired), ExecutionState: "allowlisted"})
	}
	return resolutions
}

func (d ToolDefinition) Execute(arguments ToolArguments) (map[string]any, error) {
	if d.Name != "runtime.get_current_time" {
		return nil, errors.New("tool is not supported")
	}
	if arguments.Timezone != "UTC" {
		return nil, errors.New("timezone must be UTC")
	}
	return map[string]any{"iso_time": time.Now().UTC().Format(time.RFC3339Nano), "timezone": "UTC", "source": "runtime"}, nil
}

func (d ToolDefinition) NormalizeArguments(arguments map[string]any) (ToolArguments, error) {
	if d.Name != "runtime.get_current_time" {
		return ToolArguments{}, errors.New("tool is not supported")
	}
	for key := range arguments {
		if key != "timezone" {
			return ToolArguments{}, errors.New("tool argument is not supported")
		}
	}
	value, ok := arguments["timezone"]
	if !ok || value == nil {
		return ToolArguments{Timezone: "UTC"}, nil
	}
	timezone, ok := value.(string)
	if !ok || timezone != "UTC" {
		return ToolArguments{}, errors.New("timezone must be UTC")
	}
	return ToolArguments{Timezone: "UTC"}, nil
}
