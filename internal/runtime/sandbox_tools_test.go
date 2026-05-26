package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestSandboxExecCommandRunsArgvWithinWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	executor := SandboxToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}, "cwd": "."}})
	if err != nil {
		t.Fatal(err)
	}

	if result["tool"] != productdata.ToolNameSandboxExecCommand || result["scope"] != "bounded_read_only_command" || result["operation"] != "exec_command" || result["cwd"] != "." || result["exit_code"] != 0 || result["timed_out"] != false || result["stderr"] != "" {
		t.Fatalf("result = %+v", result)
	}
	if result["stdout"] != "."+"\n" && strings.TrimSpace(result["stdout"].(string)) != "." {
		t.Fatalf("stdout = %q", result["stdout"])
	}
	if strings.Contains(strings.Join(stringSliceResult(t, result["argv"]), " "), root) || strings.Contains(strings.Join(stringSliceResult(t, result["argv"]), " "), "/Users/") {
		t.Fatalf("result leaked host path: %+v", result)
	}
}

func TestSandboxExecCommandReportsNonZeroExitWithoutWorkerFailure(t *testing.T) {
	root := t.TempDir()
	executor := SandboxToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"git", "status"}}})
	if err != nil {
		t.Fatal(err)
	}

	if result["exit_code"] == 0 || result["timed_out"] != false {
		t.Fatalf("result = %+v", result)
	}
}

func TestSandboxExecCommandBoundsOutput(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"alpha.txt", "beta.txt", "gamma.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executor := SandboxToolExecutor{Root: root}

	truncated, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"ls"}, "max_output_bytes": 3}})
	if err != nil {
		t.Fatal(err)
	}
	if truncated["stdout_truncated"] != true {
		t.Fatalf("truncated = %+v", truncated)
	}
	if truncated["stdout_bytes"].(int) <= 3 {
		t.Fatalf("stdout_bytes = %+v", truncated)
	}
}

func TestSandboxExecCommandUsesRelativeWorkspaceCwd(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	executor := SandboxToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}, "cwd": "src"}})
	if err != nil {
		t.Fatal(err)
	}
	if result["cwd"] != "src" || strings.Contains(result["stdout"].(string), root) {
		t.Fatalf("result leaked absolute cwd: %+v", result)
	}
}

func TestSandboxExecCommandRejectsUnsafeRequestsBeforeSpawn(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "secrets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := SandboxToolExecutor{Root: root}

	cases := []map[string]any{
		{"argv": []any{}},
		{"argv": "cat keep.txt"},
		{"argv": []any{"sh", "-c", "touch shell-created.txt"}},
		{"argv": []any{"rm", "keep.txt"}},
		{"argv": []any{"touch", "created.txt"}},
		{"argv": []any{"python", "-c", "print(1)"}},
		{"argv": []any{"node", "-e", "console.log(1)"}},
		{"argv": []any{"curl", "https://example.test"}},
		{"argv": []any{"chmod", "777", "keep.txt"}},
		{"argv": []any{"cat", "keep.txt"}},
		{"argv": []any{"cat", ".env"}},
		{"argv": []any{"head", "keep.txt"}},
		{"argv": []any{"tail", "keep.txt"}},
		{"argv": []any{"sed", "-n", "1p", "keep.txt"}},
		{"argv": []any{"wc", "keep.txt"}},
		{"argv": []any{"rg", "token"}},
		{"argv": []any{"ls", "secrets"}},
		{"argv": []any{"ls", "-la"}},
		{"argv": []any{"git", "reset", "--hard"}},
		{"argv": []any{"git", "push"}},
		{"argv": []any{"git", "clean", "-fd"}},
		{"argv": []any{"git", "diff"}},
		{"argv": []any{"git", "log"}},
		{"argv": []any{"cat", "../keep.txt"}},
		{"argv": []any{"cat", "/etc/passwd"}},
		{"argv": []any{"cat", "keep.txt"}, "cwd": "../"},
		{"argv": []any{"cat", "keep.txt"}, "cwd": root},
		{"argv": []any{"cat", "keep.txt"}, "cwd": "secrets"},
		{"argv": []any{"cat", "keep.txt"}, "env": map[string]any{"TOKEN": "secret"}},
	}
	for _, args := range cases {
		if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: args}); err == nil {
			t.Fatalf("exec(%+v) err = nil", args)
		}
	}
	content, err := os.ReadFile(filepath.Join(root, "keep.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "keep" {
		t.Fatalf("destructive command changed file: %q", string(content))
	}
	if _, err := os.Stat(filepath.Join(root, "shell-created.txt")); err == nil {
		t.Fatal("shell-form command executed")
	}
}

func stringSliceResult(t *testing.T, value any) []string {
	t.Helper()
	items, ok := value.([]string)
	if !ok {
		t.Fatalf("value is not []string: %+v", value)
	}
	return items
}
