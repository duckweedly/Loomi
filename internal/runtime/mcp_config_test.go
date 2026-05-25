package runtime

import "testing"

func TestValidateMCPServerConfigAcceptsOnlyLocalStdio(t *testing.T) {
	config := MCPServerConfig{
		Slug:        "local-search",
		DisplayName: "Local Search",
		Enabled:     true,
		Transport:   MCPTransportStdio,
		Command:     "loomi-test-mcp",
		TimeoutMS:   5000,
	}

	validated, err := ValidateMCPServerConfig(config)
	if err != nil {
		t.Fatalf("ValidateMCPServerConfig() error = %v", err)
	}
	if validated.Slug != "local-search" || validated.Transport != MCPTransportStdio {
		t.Fatalf("validated = %+v", validated)
	}
}

func TestValidateMCPServerConfigRejectsRemoteAndUnsupportedTransports(t *testing.T) {
	for _, config := range []MCPServerConfig{
		{Slug: "http", DisplayName: "HTTP", Enabled: true, Transport: "http", Command: "https://example.com/mcp", TimeoutMS: 5000},
		{Slug: "sse", DisplayName: "SSE", Enabled: true, Transport: "sse", Command: "https://example.com/sse", TimeoutMS: 5000},
		{Slug: "oauth", DisplayName: "OAuth", Enabled: true, Transport: "oauth", Command: "oauth-client", TimeoutMS: 5000},
		{Slug: "remote", DisplayName: "Remote", Enabled: true, Transport: MCPTransportStdio, Command: "https://example.com/mcp", TimeoutMS: 5000},
	} {
		if _, err := ValidateMCPServerConfig(config); err == nil {
			t.Fatalf("ValidateMCPServerConfig(%+v) error = nil", config)
		}
	}
}

func TestMCPRedactionRemovesSensitiveConfigAndProcessOutput(t *testing.T) {
	config := MCPServerConfig{
		Slug:        "local-search",
		DisplayName: "Local Search",
		Enabled:     true,
		Transport:   MCPTransportStdio,
		Command:     "/Users/xuean/.ssh/id_ed25519",
		Args:        []string{"--token", "sk-live-123"},
		Env:         map[string]string{"API_TOKEN": "sk-live-123"},
		TimeoutMS:   5000,
	}
	summary := config.SafeSummary()
	if _, ok := summary["command"]; ok {
		t.Fatalf("command leaked in summary = %+v", summary)
	}
	if _, ok := summary["args"]; ok {
		t.Fatalf("args leaked in summary = %+v", summary)
	}
	if _, ok := summary["env"]; ok {
		t.Fatalf("env leaked in summary = %+v", summary)
	}

	failure := MCPDiscoveryFailure("local-search", MCPDiscoveryTimeout, "stderr: token sk-live-123 in /Users/xuean/.ssh/id_ed25519", true)
	if failure.Message != "[redacted]" {
		t.Fatalf("failure message = %q", failure.Message)
	}
}
