package runtime

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestStdioMCPToolExecutorCallsToolAndRedactsResult(t *testing.T) {
	if os.Getenv("LOOMI_MCP_TOOL_FIXTURE") == "1" {
		runMCPToolFixture(t)
		return
	}
	executor := StdioMCPToolExecutor{Configs: map[string]MCPServerConfig{"local-search": {
		Slug:        "local-search",
		DisplayName: "Local Search",
		Enabled:     true,
		Transport:   MCPTransportStdio,
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestStdioMCPToolExecutorCallsToolAndRedactsResult"},
		Env:         map[string]string{"LOOMI_MCP_TOOL_FIXTURE": "1", "TOKEN": "sk-secret"},
		TimeoutMS:   1000,
	}}}
	result, err := executor.ExecuteMCPTool(context.Background(), productdata.ToolCall{ToolCallID: "tc_mcp_1", ToolName: "mcp.local-search.search", ArgumentsSummary: map[string]any{"query": "status"}})
	if err != nil {
		t.Fatal(err)
	}
	if result["summary"] != "ok" {
		t.Fatalf("result = %+v", result)
	}
	if result["path"] == "/Users/xuean/.ssh/id_ed25519" || result["token"] == "sk-secret" {
		t.Fatalf("result leaked sensitive fields: %+v", result)
	}
}

func TestStdioMCPToolExecutorClassifiesTimeoutWithoutRawStderr(t *testing.T) {
	if os.Getenv("LOOMI_MCP_TOOL_SLEEP_FIXTURE") == "1" {
		time.Sleep(2 * time.Second)
		return
	}
	executor := StdioMCPToolExecutor{Configs: map[string]MCPServerConfig{"local-search": {
		Slug:        "local-search",
		DisplayName: "Local Search",
		Enabled:     true,
		Transport:   MCPTransportStdio,
		Command:     os.Args[0],
		Args:        []string{"-test.run=TestStdioMCPToolExecutorClassifiesTimeoutWithoutRawStderr"},
		Env:         map[string]string{"LOOMI_MCP_TOOL_SLEEP_FIXTURE": "1"},
		TimeoutMS:   1,
	}}}
	_, err := executor.ExecuteMCPTool(context.Background(), productdata.ToolCall{ToolCallID: "tc_mcp_1", ToolName: "mcp.local-search.search"})
	if err == nil {
		t.Fatal("ExecuteMCPTool() error = nil")
	}
	if err.Error() != "mcp_stdio_timeout" {
		t.Fatalf("err = %v", err)
	}
}

func runMCPToolFixture(t *testing.T) {
	t.Helper()
	reader := bufio.NewReader(os.Stdin)
	for {
		frame, ok := readMCPFixtureFrame(t, reader)
		if !ok {
			return
		}
		method, _ := frame["method"].(string)
		id := frame["id"]
		switch method {
		case "initialize":
			_, _ = os.Stdout.Write(mcpFrame(`{"jsonrpc":"2.0","id":1,"result":{"protocolVersion":"2024-11-05"}}`))
		case "tools/call":
			_, _ = os.Stdout.Write(mcpFrame(`{"jsonrpc":"2.0","id":2,"result":{"summary":"ok","path":"/Users/xuean/.ssh/id_ed25519","token":"sk-secret"}}`))
			return
		default:
			if id != nil {
				t.Fatalf("unexpected request method %q", method)
			}
		}
	}
}

func readMCPFixtureFrame(t *testing.T, reader *bufio.Reader) (map[string]any, bool) {
	t.Helper()
	contentLength := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				return nil, false
			}
			t.Fatal(err)
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			break
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if ok && strings.EqualFold(strings.TrimSpace(key), "Content-Length") {
			parsed, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				t.Fatal(err)
			}
			contentLength = parsed
		}
	}
	if contentLength <= 0 {
		t.Fatal("missing content length")
	}
	payload := make([]byte, contentLength)
	if _, err := io.ReadFull(reader, payload); err != nil {
		t.Fatal(err)
	}
	var frame map[string]any
	if err := json.Unmarshal(payload, &frame); err != nil {
		t.Fatal(err)
	}
	return frame, true
}

func TestStdioMCPToolExecutorRejectsMissingConfig(t *testing.T) {
	_, err := (StdioMCPToolExecutor{}).ExecuteMCPTool(context.Background(), productdata.ToolCall{ToolCallID: "tc_mcp_1", ToolName: "mcp.missing.search"})
	if err == nil {
		t.Fatal("ExecuteMCPTool() error = nil")
	}
}
