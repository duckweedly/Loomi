package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
)

type ToolSource string

type ToolExecutionState string

const (
	ToolSourceInternal ToolSource = "internal"
	ToolSourceMCP      ToolSource = "mcp"

	ToolExecutionAllowlisted ToolExecutionState = "allowlisted"
	ToolExecutionDisabled    ToolExecutionState = "disabled"
)

type ToolSpec struct {
	Name           string
	Source         ToolSource
	ServerSlug     string
	ApprovalPolicy ToolApprovalPolicy
	ExecutionState ToolExecutionState
	SchemaHash     string
}

type MCPToolCandidate struct {
	ServerSlug       string
	MCPName          string
	Name             string
	Description      string
	SchemaHash       string
	ExecutionEnabled bool
}

var safeToolNamePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_-]{0,63}$`)

func MapMCPToolCandidate(serverSlug string, toolName string, description string, schema map[string]any) (MCPToolCandidate, error) {
	serverSlug = strings.TrimSpace(serverSlug)
	toolName = strings.TrimSpace(toolName)
	if !safeSlugPattern.MatchString(serverSlug) {
		return MCPToolCandidate{}, errors.New("mcp server slug is invalid")
	}
	if !safeToolNamePattern.MatchString(toolName) {
		return MCPToolCandidate{}, errors.New("mcp tool name is invalid")
	}
	if toolName == "runtime.get_current_time" {
		return MCPToolCandidate{}, errors.New("mcp tool cannot override internal tool")
	}
	rawSchema, err := json.Marshal(schema)
	if err != nil {
		return MCPToolCandidate{}, err
	}
	sum := sha256.Sum256(rawSchema)
	return MCPToolCandidate{
		ServerSlug:       serverSlug,
		MCPName:          toolName,
		Name:             "mcp." + serverSlug + "." + toolName,
		Description:      RedactMCPText(description),
		SchemaHash:       "sha256:" + hex.EncodeToString(sum[:]),
		ExecutionEnabled: false,
	}, nil
}

func MCPToolSpecs(candidates []MCPToolCandidate) []ToolSpec {
	specs := make([]ToolSpec, 0, len(candidates))
	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.Name) == "" {
			continue
		}
		specs = append(specs, ToolSpec{
			Name:           candidate.Name,
			Source:         ToolSourceMCP,
			ServerSlug:     candidate.ServerSlug,
			ApprovalPolicy: ToolApprovalAlwaysRequired,
			ExecutionState: ToolExecutionDisabled,
			SchemaHash:     candidate.SchemaHash,
		})
	}
	return specs
}

func IsMCPToolName(name string) bool {
	parts := strings.Split(strings.TrimSpace(name), ".")
	return len(parts) == 3 && parts[0] == "mcp" && safeSlugPattern.MatchString(parts[1]) && safeToolNamePattern.MatchString(parts[2])
}
