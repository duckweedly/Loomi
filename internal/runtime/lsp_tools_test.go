package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestLSPReadOnlyToolsExecuteWithinFixtureRoot(t *testing.T) {
	root := createLSPFixture(t)
	executor := LSPToolExecutor{Root: root}

	symbols, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go", "query": "Tool", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if symbols["operation"] != "symbols" || symbols["scope"] != "lsp" || symbols["path"] != "src/main.go" || symbols["count"] != 1 {
		t.Fatalf("symbols = %+v", symbols)
	}
	if strings.Contains(fmt.Sprint(symbols), root) {
		t.Fatalf("symbols leaked root: %+v", symbols)
	}

	refs, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPReferences, ArgumentsSummary: map[string]any{"path": "src/main.go", "line": 3, "column": 6, "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if refs["operation"] != "references" || refs["count"] != 2 {
		t.Fatalf("references = %+v", refs)
	}

	diagnostics, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPDiagnostics, ArgumentsSummary: map[string]any{"path": "src/bad.go", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if diagnostics["operation"] != "diagnostics" || diagnostics["count"] != 1 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
}

func TestLSPReadOnlyToolsRejectUnsafePaths(t *testing.T) {
	root := createLSPFixture(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.go"), []byte("package secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.go"), filepath.Join(root, "src", "outside-link.go")); err != nil {
		t.Fatal(err)
	}
	executor := LSPToolExecutor{Root: root}
	for _, args := range []map[string]any{
		{"path": "../outside.go"},
		{"path": filepath.Join(outside, "secret.go")},
		{"path": ".env"},
		{"path": "secrets/token.go"},
		{"path": "src/outside-link.go"},
	} {
		_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: args})
		if err == nil {
			t.Fatalf("symbols(%+v) err = nil", args)
		}
		if strings.Contains(err.Error(), "secret") || strings.Contains(err.Error(), outside) {
			t.Fatalf("error leaked sensitive detail: %v", err)
		}
	}
}

func TestLSPReadOnlyToolsRejectInvalidArgumentsAndBoundResults(t *testing.T) {
	root := createLSPFixture(t)
	executor := LSPToolExecutor{Root: root}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPSymbols, ArgumentsSummary: map[string]any{"path": "src/main.go", "api_key": "sk-secret"}}); err == nil {
		t.Fatal("symbols unsupported argument err = nil")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPReferences, ArgumentsSummary: map[string]any{"path": "src/main.go"}}); err == nil {
		t.Fatal("references missing position err = nil")
	}
	refs, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameLSPReferences, ArgumentsSummary: map[string]any{"path": "src/main.go", "line": 3, "column": 6, "limit": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if refs["count"] != 1 || refs["truncated"] != true {
		t.Fatalf("bounded refs = %+v", refs)
	}
}

func createLSPFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n\ntype ToolBroker struct{}\n\nfunc useToolBroker() { _ = ToolBroker{} }\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "bad.go"), []byte("package main\n\nfunc broken(\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}
