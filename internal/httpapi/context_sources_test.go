package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/productdata"
)

func TestThreadContextSourcesEndpointRegistersSafeSources(t *testing.T) {
	srv := NewServerWithProduct(config.Config{AppEnv: "local"}, fakeChecker{}, productdata.NewMemoryService())
	threadRes := requestJSON(t, srv, http.MethodPost, "/v1/threads", `{"title":"Sources","mode":"work"}`)
	if threadRes.Code != http.StatusCreated {
		t.Fatalf("thread status=%d body=%s", threadRes.Code, threadRes.Body.String())
	}
	threadID := decodeThreadID(t, threadRes.Body.Bytes())

	create := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/sources", `{"kind":"url","title":"Docs","locator":"https://example.com/docs?token=secret","summary":"public docs"}`)
	if create.Code != http.StatusCreated {
		t.Fatalf("create status=%d body=%s", create.Code, create.Body.String())
	}
	if body := create.Body.String(); !strings.Contains(body, `"locator":"https://example.com/docs"`) || strings.Contains(body, "token=secret") {
		t.Fatalf("unsafe source body: %s", body)
	}

	list := requestJSON(t, srv, http.MethodGet, "/v1/threads/"+threadID+"/sources?limit=10", "")
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), `"title":"Docs"`) {
		t.Fatalf("list status=%d body=%s", list.Code, list.Body.String())
	}

	badURL := requestJSON(t, srv, http.MethodPost, "/v1/threads/"+threadID+"/sources", `{"kind":"url","title":"Bad","locator":"http://127.0.0.1/private?token=secret"}`)
	if badURL.Code != http.StatusBadRequest || strings.Contains(badURL.Body.String(), "token=secret") {
		t.Fatalf("bad url status=%d body=%s", badURL.Code, badURL.Body.String())
	}
}
