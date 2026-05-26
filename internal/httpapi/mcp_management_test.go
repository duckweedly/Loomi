package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
	productruntime "github.com/sheridiany/loomi/internal/runtime"
)

func TestM25MCPServersHandlerReturnsSafeReadOnlyStatus(t *testing.T) {
	t.Setenv("LOOMI_MCP_SERVERS_JSON", `[{"slug":"local-smoke","display_name":"Local Smoke","enabled":true,"transport":"stdio","command":"/Users/xuean/private/bin/mcp","args":["--token=SECRET_CANARY_ARG"],"env":{"TOKEN":"SECRET_CANARY_ENV"},"timeout_ms":5000}]`)
	svc := productdata.NewMemoryService()
	ident := identity.LocalDevIdentity()
	run := createToolsAPIRun(t, svc, ident)
	discovery := productruntime.MCPDiscoveryResult{
		ServerSlug: "local-smoke",
		Status:     productruntime.MCPDiscoverySucceeded,
		Candidates: []productruntime.MCPToolCandidate{{
			Name:       "mcp.local-smoke.echo",
			MCPName:    "echo",
			SchemaHash: "sha256:test-schema",
		}},
	}
	if _, err := svc.AppendRunEvent(context.Background(), ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: "mcp_discovery_succeeded", Summary: "MCP discovery succeeded", Metadata: productruntime.MCPDiscoveryEventMetadata(discovery)}); err != nil {
		t.Fatal(err)
	}
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	res := requestJSON(t, srv, http.MethodGet, "/v1/mcp/servers", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	body := res.Body.String()
	for _, expected := range []string{`"server_slug":"local-smoke"`, `"display_name":"Local Smoke"`, `"transport":"stdio"`, `"enabled":true`, `"discovery_status":"succeeded"`, `"candidate_names":["mcp.local-smoke.echo"]`, `"execution_mode":"approval_gated"`} {
		if !strings.Contains(body, expected) {
			t.Fatalf("body missing %q: %s", expected, body)
		}
	}
	for _, forbidden := range []string{"SECRET_CANARY", "/Users/xuean", "command", "args", "env", "TOKEN"} {
		if strings.Contains(body, forbidden) {
			t.Fatalf("body leaked %q: %s", forbidden, body)
		}
	}
	var parsed struct {
		Servers []productruntime.MCPServerStatus `json:"servers"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &parsed); err != nil {
		t.Fatal(err)
	}
	if len(parsed.Servers) != 1 || parsed.Servers[0].CandidateCount != 1 {
		t.Fatalf("servers = %+v", parsed.Servers)
	}
}

func TestMCPServersHandlerSavesDiscoversAndDeletesConfig(t *testing.T) {
	t.Setenv("LOOMI_MCP_SERVERS_JSON", "")
	svc := productdata.NewMemoryService()
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, svc)

	saveBody := `{"slug":"local-disabled","display_name":"Local Disabled","enabled":false,"transport":"stdio","command":"","args":["--secret=SECRET_CANARY_ARG"],"env":{"TOKEN":"SECRET_CANARY_ENV"},"timeout_ms":5000}`
	save := requestJSON(t, srv, http.MethodPost, "/v1/mcp/servers", saveBody)
	if save.Code != http.StatusOK {
		t.Fatalf("save status = %d body=%s", save.Code, save.Body.String())
	}
	if body := save.Body.String(); !strings.Contains(body, `"server_slug":"local-disabled"`) || !strings.Contains(body, `"discovery_status":"disabled"`) {
		t.Fatalf("save body missing saved disabled status: %s", body)
	}
	if body := save.Body.String(); strings.Contains(body, "SECRET_CANARY") || strings.Contains(body, `"command"`) || strings.Contains(body, `"env"`) {
		t.Fatalf("save body leaked config: %s", body)
	}

	discover := requestJSON(t, srv, http.MethodPost, "/v1/mcp/servers/local-disabled/discover", "")
	if discover.Code != http.StatusOK {
		t.Fatalf("discover status = %d body=%s", discover.Code, discover.Body.String())
	}
	if body := discover.Body.String(); !strings.Contains(body, `"discovery_status":"disabled"`) || !strings.Contains(body, `"last_discovered_at"`) {
		t.Fatalf("discover body missing disabled event: %s", body)
	}

	deleted := requestJSON(t, srv, http.MethodDelete, "/v1/mcp/servers/local-disabled", "")
	if deleted.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", deleted.Code, deleted.Body.String())
	}
	if strings.Contains(deleted.Body.String(), "local-disabled") {
		t.Fatalf("delete body still contains server: %s", deleted.Body.String())
	}
}
