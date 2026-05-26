package runtime

import (
	"strings"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestMCPServerStatusesMergeLocalConfigAndSafeDiscoveryEvents(t *testing.T) {
	configs := map[string]MCPServerConfig{
		"failed-smoke": {Slug: "failed-smoke", DisplayName: "Failed Smoke", Enabled: true, Transport: MCPTransportStdio, Command: "/Users/xuean/private/bin/mcp", Args: []string{"--token=secret"}, Env: map[string]string{"TOKEN": "secret"}, TimeoutMS: 5000},
		"local-smoke":  {Slug: "local-smoke", DisplayName: "Local Smoke", Enabled: true, Transport: MCPTransportStdio, Command: "mcp", TimeoutMS: 5000},
	}
	events := []productdata.RunEvent{
		{ID: "evt_1", Type: "mcp_discovery_failed", Sequence: 1, CreatedAt: time.Date(2026, 5, 26, 1, 0, 0, 0, time.UTC), Metadata: map[string]any{"server_slug": "failed-smoke", "status": "failed", "error_code": "mcp_discovery_timeout", "message": "SECRET_CANARY"}},
		{ID: "evt_2", Type: "mcp_discovery_succeeded", Sequence: 2, CreatedAt: time.Date(2026, 5, 26, 1, 1, 0, 0, time.UTC), Metadata: map[string]any{"server_slug": "local-smoke", "status": "succeeded", "candidate_names": []string{"mcp.local-smoke.echo", "mcp.other.echo"}}},
	}

	statuses := MCPServerStatuses(configs, events)

	if len(statuses) != 2 {
		t.Fatalf("statuses = %+v", statuses)
	}
	if statuses[1].ServerSlug != "local-smoke" || statuses[1].ExecutionMode != "approval_gated" || statuses[1].CandidateCount != 1 || statuses[1].CandidateNames[0] != "mcp.local-smoke.echo" {
		t.Fatalf("local status = %+v", statuses[1])
	}
	encoded := strings.Join([]string{statuses[0].DisplayName, statuses[0].RedactedErrorCode, statuses[1].DisplayName}, " ")
	for _, forbidden := range []string{"SECRET_CANARY", "/Users/xuean", "TOKEN", "secret"} {
		if strings.Contains(encoded, forbidden) {
			t.Fatalf("status leaked %q: %+v", forbidden, statuses)
		}
	}
}
