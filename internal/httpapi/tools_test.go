package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestToolCatalogHandlerReturnsDeterministicRedactedCatalog(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())

	res := requestJSON(t, srv, http.MethodGet, "/v1/tools/catalog", "")

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", res.Code, res.Body.String())
	}
	var body struct {
		Tools []struct {
			Name           string `json:"name"`
			Group          string `json:"group"`
			Capability     string `json:"capability"`
			ApprovalPolicy string `json:"approval_policy"`
			SafetyClass    string `json:"safety_class"`
			RiskLevel      string `json:"risk_level"`
			SideEffect     string `json:"side_effect"`
			Enabled        bool   `json:"enabled"`
		} `json:"tools"`
		UpdatedAt string `json:"updated_at"`
	}
	if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	names := make([]string, 0, len(body.Tools))
	for _, tool := range body.Tools {
		names = append(names, tool.Name)
		if tool.ApprovalPolicy != "always_required" || !tool.Enabled || tool.RiskLevel == "" || tool.SideEffect == "" {
			t.Fatalf("tool = %+v", tool)
		}
	}
	want := "runtime.get_current_time,runtime.todo_write,mcp.call_tool,workspace.glob,workspace.grep,workspace.read_file,workspace.write_file,workspace.edit,workspace.exec_command"
	if strings.Join(names, ",") != want {
		t.Fatalf("names = %+v", names)
	}
	if body.UpdatedAt == "" {
		t.Fatalf("updated_at missing: %+v", body)
	}
	raw := res.Body.String()
	for _, forbidden := range []string{"sk-", "api_key", "secret", "/Users/", "Authorization"} {
		if strings.Contains(raw, forbidden) {
			t.Fatalf("catalog contains forbidden %q: %s", forbidden, raw)
		}
	}
}
