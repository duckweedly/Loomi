package runtime

import (
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
)

type MCPTransport string

type MCPDiscoveryStatus string

type MCPDiscoveryErrorCode string

const (
	MCPTransportStdio MCPTransport = "stdio"

	MCPDiscoverySucceeded MCPDiscoveryStatus = "succeeded"
	MCPDiscoveryFailed    MCPDiscoveryStatus = "failed"
	MCPDiscoveryDisabled  MCPDiscoveryStatus = "disabled"
	MCPDiscoveryRejected  MCPDiscoveryStatus = "rejected"

	MCPDiscoveryTimeout           MCPDiscoveryErrorCode = "mcp_discovery_timeout"
	MCPDiscoveryInvalidConfig     MCPDiscoveryErrorCode = "mcp_config_invalid"
	MCPDiscoveryInvalidResponse   MCPDiscoveryErrorCode = "mcp_discovery_invalid_response"
	MCPDiscoveryUnsupportedSchema MCPDiscoveryErrorCode = "mcp_tool_schema_unsupported"
)

type MCPServerConfig struct {
	Slug        string            `json:"slug"`
	DisplayName string            `json:"display_name"`
	Enabled     bool              `json:"enabled"`
	Transport   MCPTransport      `json:"transport"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	Env         map[string]string `json:"env"`
	TimeoutMS   int               `json:"timeout_ms"`
}

type MCPDiscoveryResult struct {
	ServerSlug string
	Status     MCPDiscoveryStatus
	Candidates []MCPToolCandidate
	ErrorCode  MCPDiscoveryErrorCode
	Message    string
	Retryable  bool
}

var safeSlugPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

func ValidateMCPServerConfig(config MCPServerConfig) (MCPServerConfig, error) {
	config.Slug = strings.TrimSpace(config.Slug)
	config.DisplayName = strings.TrimSpace(config.DisplayName)
	config.Command = strings.TrimSpace(config.Command)
	if config.Slug == "" || config.DisplayName == "" {
		return MCPServerConfig{}, errors.New("mcp server slug and display name are required")
	}
	if !safeSlugPattern.MatchString(config.Slug) {
		return MCPServerConfig{}, errors.New("mcp server slug is invalid")
	}
	if config.Transport != MCPTransportStdio {
		return MCPServerConfig{}, errors.New("only local stdio MCP servers are supported")
	}
	if config.Enabled && config.Command == "" {
		return MCPServerConfig{}, errors.New("mcp server command is required")
	}
	if looksRemote(config.Command) {
		return MCPServerConfig{}, errors.New("remote MCP endpoints are not supported")
	}
	if config.TimeoutMS <= 0 {
		config.TimeoutMS = 5000
	} else if config.TimeoutMS < 100 {
		config.TimeoutMS = 100
	} else if config.TimeoutMS > 60000 {
		config.TimeoutMS = 60000
	}
	return config, nil
}

func MCPServerConfigsFromEnv() (map[string]MCPServerConfig, error) {
	raw := strings.TrimSpace(os.Getenv("LOOMI_MCP_SERVERS_JSON"))
	if raw == "" {
		return map[string]MCPServerConfig{}, nil
	}
	var configs []MCPServerConfig
	if err := json.Unmarshal([]byte(raw), &configs); err != nil {
		return nil, errors.New("mcp config json is invalid")
	}
	bySlug := make(map[string]MCPServerConfig, len(configs))
	for _, config := range configs {
		validated, err := ValidateMCPServerConfig(config)
		if err != nil {
			return nil, err
		}
		if _, exists := bySlug[validated.Slug]; exists {
			return nil, errors.New("mcp server slug is duplicated")
		}
		bySlug[validated.Slug] = validated
	}
	return bySlug, nil
}

func (c MCPServerConfig) SafeSummary() map[string]any {
	return map[string]any{
		"mcp_server_slug":    c.Slug,
		"mcp_display_name":   c.DisplayName,
		"mcp_enabled":        c.Enabled,
		"mcp_transport":      string(c.Transport),
		"mcp_timeout_ms":     c.TimeoutMS,
		"mcp_has_args":       len(c.Args) > 0,
		"mcp_has_env":        len(c.Env) > 0,
		"mcp_config_source":  "local",
		"mcp_execution_mode": "disabled",
	}
}

func MCPDiscoveryFailure(serverSlug string, code MCPDiscoveryErrorCode, message string, retryable bool) MCPDiscoveryResult {
	return MCPDiscoveryResult{ServerSlug: strings.TrimSpace(serverSlug), Status: MCPDiscoveryFailed, ErrorCode: code, Message: RedactMCPText(message), Retryable: retryable}
}

func RedactMCPText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	lower := strings.ToLower(trimmed)
	for _, marker := range []string{"api_key", "authorization", "password", "secret", "token", "credential", "bearer ", "sk-", ".ssh", "id_ed25519", "id_rsa", ".env"} {
		if strings.Contains(lower, marker) {
			return "[redacted]"
		}
	}
	if strings.Contains(trimmed, "/Users/") || strings.Contains(trimmed, "/.ssh/") || strings.Contains(trimmed, "\\Users\\") {
		return "[redacted]"
	}
	return trimmed
}

func looksRemote(command string) bool {
	lower := strings.ToLower(strings.TrimSpace(command))
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") || strings.HasPrefix(lower, "ws://") || strings.HasPrefix(lower, "wss://")
}
