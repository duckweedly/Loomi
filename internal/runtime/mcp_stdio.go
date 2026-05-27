package runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

type StdioMCPToolExecutor struct {
	Configs      map[string]MCPServerConfig
	ConfigLoader func(context.Context) (map[string]MCPServerConfig, error)
}

func (e StdioMCPToolExecutor) ExecuteMCPTool(ctx context.Context, call productdata.ToolCall) (map[string]any, error) {
	serverSlug := mcpServerSlug(call.ToolName)
	configs := e.Configs
	if e.ConfigLoader != nil {
		loaded, err := e.ConfigLoader(ctx)
		if err != nil {
			return nil, errors.New("mcp_config_unavailable")
		}
		configs = loaded
	}
	config, ok := configs[serverSlug]
	if !ok {
		return nil, errors.New("mcp_config_unavailable")
	}
	validated, err := ValidateMCPServerConfig(config)
	if err != nil {
		return nil, errors.New("mcp_config_invalid")
	}
	timeout := time.Duration(validated.TimeoutMS) * time.Millisecond
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	command := exec.CommandContext(execCtx, validated.Command, validated.Args...)
	command.Env = os.Environ()
	for key, value := range validated.Env {
		command.Env = append(command.Env, key+"="+value)
	}
	stdin, err := command.StdinPipe()
	if err != nil {
		return nil, errors.New("mcp_stdio_unavailable")
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		return nil, errors.New("mcp_stdio_unavailable")
	}
	_ = writeMCPRequest(stdin, map[string]any{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": map[string]any{"protocolVersion": "2024-11-05", "capabilities": map[string]any{}, "clientInfo": map[string]any{"name": "loomi", "version": "m12"}}})
	_ = writeMCPRequest(stdin, map[string]any{"jsonrpc": "2.0", "method": "notifications/initialized", "params": map[string]any{}})
	_ = writeMCPRequest(stdin, map[string]any{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": map[string]any{"name": mcpToolName(call.ToolName), "arguments": call.ArgumentsSummary}})
	_ = stdin.Close()
	err = command.Wait()
	if execCtx.Err() == context.DeadlineExceeded {
		return nil, errors.New("mcp_stdio_timeout")
	}
	if err != nil {
		return nil, errors.New("mcp_stdio_exit")
	}
	result, err := parseMCPToolCallResult(stdout.Bytes())
	if err != nil {
		return nil, err
	}
	return RedactMCPSummary(result), nil
}

func parseMCPToolCallResult(output []byte) (map[string]any, error) {
	for _, payload := range mcpJSONFrames(output) {
		var frame map[string]any
		if err := json.Unmarshal(payload, &frame); err != nil {
			return nil, errors.New("mcp_stdio_invalid_response")
		}
		id, _ := frame["id"].(float64)
		if int(id) != 2 {
			continue
		}
		if errorValue, ok := frame["error"]; ok && errorValue != nil {
			return nil, errors.New("mcp_tool_error")
		}
		result, ok := frame["result"].(map[string]any)
		if !ok || len(result) == 0 {
			return nil, errors.New("mcp_stdio_invalid_response")
		}
		return result, nil
	}
	return nil, errors.New("mcp_stdio_invalid_response")
}

func writeMCPRequest(writer io.Writer, payload map[string]any) error {
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = writer.Write(mcpFrame(string(raw)))
	return err
}

func mcpServerSlug(toolName string) string {
	parts := strings.Split(strings.TrimSpace(toolName), ".")
	if len(parts) != 3 {
		return ""
	}
	return parts[1]
}

func mcpToolName(toolName string) string {
	parts := strings.Split(strings.TrimSpace(toolName), ".")
	if len(parts) != 3 {
		return ""
	}
	return parts[2]
}
