package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type MCPDiscoveryRunner interface {
	ListTools(context.Context, MCPServerConfig) ([]byte, error)
}

type StdioMCPDiscoveryRunner struct{}

const (
	MaxMCPDiscoveryTools = 64
	MaxMCPSchemaBytes    = 8192
)

type mcpListToolsEnvelope struct {
	Result struct {
		Tools []struct {
			Name        string         `json:"name"`
			Description string         `json:"description"`
			InputSchema map[string]any `json:"inputSchema"`
		} `json:"tools"`
	} `json:"result"`
	Tools []struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		InputSchema map[string]any `json:"inputSchema"`
	} `json:"tools"`
}

func ParseMCPListToolsResponse(serverSlug string, payload []byte) (MCPDiscoveryResult, error) {
	var envelope mcpListToolsEnvelope
	if err := json.Unmarshal(payload, &envelope); err != nil {
		return MCPDiscoveryResult{}, err
	}
	tools := envelope.Result.Tools
	if len(tools) == 0 {
		tools = envelope.Tools
	}
	if len(tools) > MaxMCPDiscoveryTools {
		return MCPDiscoveryResult{}, errors.New("mcp tool list is too large")
	}
	seen := map[string]bool{}
	candidates := make([]MCPToolCandidate, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			return MCPDiscoveryResult{}, errors.New("mcp tool name is required")
		}
		if seen[name] {
			return MCPDiscoveryResult{}, errors.New("duplicate mcp tool name")
		}
		seen[name] = true
		schema := tool.InputSchema
		if schema == nil {
			schema = map[string]any{}
		}
		if err := validateMCPInputSchema(schema); err != nil {
			return MCPDiscoveryResult{}, err
		}
		candidate, err := MapMCPToolCandidate(serverSlug, name, tool.Description, schema)
		if err != nil {
			return MCPDiscoveryResult{}, err
		}
		candidates = append(candidates, candidate)
	}
	return MCPDiscoveryResult{ServerSlug: strings.TrimSpace(serverSlug), Status: MCPDiscoverySucceeded, Candidates: candidates}, nil
}

func validateMCPInputSchema(schema map[string]any) error {
	raw, err := json.Marshal(schema)
	if err != nil {
		return err
	}
	if len(raw) > MaxMCPSchemaBytes {
		return errors.New("mcp tool schema is too large")
	}
	if len(schema) == 0 {
		return nil
	}
	if value, ok := schema["type"]; ok && value != "object" {
		return errors.New("mcp tool schema type must be object")
	}
	if value, ok := schema["properties"]; ok {
		if _, ok := value.(map[string]any); !ok {
			return errors.New("mcp tool schema properties must be an object")
		}
	}
	if value, ok := schema["required"]; ok {
		items, ok := value.([]any)
		if !ok {
			return errors.New("mcp tool schema required must be an array")
		}
		for _, item := range items {
			if _, ok := item.(string); !ok {
				return errors.New("mcp tool schema required entries must be strings")
			}
		}
	}
	return nil
}

func DiscoverMCPTools(runner any, config MCPServerConfig) (MCPDiscoveryResult, error) {
	config, err := ValidateMCPServerConfig(config)
	if err != nil {
		return MCPDiscoveryResult{ServerSlug: config.Slug, Status: MCPDiscoveryRejected, ErrorCode: MCPDiscoveryInvalidConfig, Message: RedactMCPText(err.Error())}, err
	}
	if !config.Enabled {
		return MCPDiscoveryResult{ServerSlug: config.Slug, Status: MCPDiscoveryDisabled}, nil
	}
	discoveryRunner, ok := runner.(MCPDiscoveryRunner)
	if !ok || discoveryRunner == nil {
		discoveryRunner = StdioMCPDiscoveryRunner{}
	}
	payload, err := discoveryRunner.ListTools(context.Background(), config)
	if err != nil {
		return MCPDiscoveryFailure(config.Slug, MCPDiscoveryInvalidResponse, err.Error(), true), nil
	}
	result, err := ParseMCPListToolsResponse(config.Slug, payload)
	if err != nil {
		return MCPDiscoveryFailure(config.Slug, MCPDiscoveryInvalidResponse, err.Error(), false), err
	}
	return result, nil
}

func (StdioMCPDiscoveryRunner) ListTools(ctx context.Context, config MCPServerConfig) ([]byte, error) {
	timeout := time.Duration(config.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, config.Command, config.Args...)
	cmd.Env = append(cmd.Environ(), envPairs(config.Env)...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	_, _ = stdin.Write(mcpFrame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"loomi","version":"m11"}}}`))
	_, _ = stdin.Write(mcpFrame(`{"jsonrpc":"2.0","method":"notifications/initialized","params":{}}`))
	_, _ = stdin.Write(mcpFrame(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`))
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("%s: %w", MCPDiscoveryTimeout, ctx.Err())
		}
		return nil, fmt.Errorf("mcp discovery process failed: %s", RedactMCPText(stderr.String()))
	}
	if frame := lastJSONFrameWithTools(stdout.Bytes()); len(frame) > 0 {
		return frame, nil
	}
	return stdout.Bytes(), nil
}

func mcpFrame(payload string) []byte {
	return []byte(fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(payload), payload))
}

func envPairs(env map[string]string) []string {
	pairs := make([]string, 0, len(env))
	for key, value := range env {
		pairs = append(pairs, key+"="+value)
	}
	return pairs
}

func lastJSONFrameWithTools(output []byte) []byte {
	parts := bytes.Split(output, []byte("\r\n\r\n"))
	for i := len(parts) - 1; i >= 0; i-- {
		part := bytes.TrimSpace(parts[i])
		if len(part) == 0 || !bytes.Contains(part, []byte(`"tools"`)) {
			continue
		}
		if idx := bytes.IndexByte(part, '{'); idx >= 0 {
			return part[idx:]
		}
	}
	return nil
}
