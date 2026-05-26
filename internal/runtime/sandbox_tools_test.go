package runtime

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestSandboxExecCommandRunsArgvWithinWorkspaceRoot(t *testing.T) {
	root := t.TempDir()
	executor := SandboxToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"pwd"}, "cwd": "."}})
	if err != nil {
		t.Fatal(err)
	}

	if result["tool"] != productdata.ToolNameSandboxExecCommand || result["scope"] != "bounded_command" || result["operation"] != "exec_command" || result["cwd"] != "." || result["exit_code"] != 0 || result["timed_out"] != false || result["stderr"] != "" {
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

func TestSandboxExecCommandAllowsCodeAgentReadAndValidationCommands(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "keep.txt"), []byte("needle\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := SandboxToolExecutor{Root: root}

	cases := []map[string]any{
		{"argv": []any{"cat", "keep.txt"}},
		{"argv": []any{"head", "keep.txt"}},
		{"argv": []any{"tail", "keep.txt"}},
		{"argv": []any{"sed", "-n", "1p", "keep.txt"}},
		{"argv": []any{"wc", "keep.txt"}},
		{"argv": []any{"rg", "needle", "."}},
		{"argv": []any{"git", "diff"}},
		{"argv": []any{"git", "log"}},
		{"argv": []any{"git", "show", "--stat"}},
		{"argv": []any{"go", "test", "./..."}},
	}
	for _, args := range cases {
		if result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: args}); err != nil {
			t.Fatalf("exec(%+v) err = %v", args, err)
		} else if result["operation"] != "exec_command" {
			t.Fatalf("result = %+v", result)
		}
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
		{"argv": []any{"cat", ".env"}},
		{"argv": []any{"ls", "secrets"}},
		{"argv": []any{"ls", "-la"}},
		{"argv": []any{"git", "reset", "--hard"}},
		{"argv": []any{"git", "push"}},
		{"argv": []any{"git", "clean", "-fd"}},
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

func TestSandboxProcessStartContinueAndTerminate(t *testing.T) {
	root := t.TempDir()
	fifo := filepath.Join(root, "stream.txt")
	if err := syscall.Mkfifo(fifo, 0o600); err != nil {
		t.Fatal(err)
	}
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}
	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_process", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat", "stream.txt"}, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID, _ := start["process_id"].(string)
	if processID == "" || start["operation"] != "start_process" || start["status"] != "running" || start["scope"] != "bounded_process" {
		t.Fatalf("start = %+v", start)
	}
	go func() {
		file, err := os.OpenFile(fifo, os.O_WRONLY, 0)
		if err == nil {
			_, _ = file.WriteString("hello process\n")
			_ = file.Close()
		}
	}()
	deadline := time.Now().Add(2 * time.Second)
	var continued map[string]any
	for time.Now().Before(deadline) {
		continued, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_process", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(continued["stdout"].(string), "hello process") {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(continued["stdout"].(string), "hello process") {
		t.Fatalf("continued = %+v", continued)
	}
	terminated, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_process", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
	if err != nil {
		t.Fatal(err)
	}
	if terminated["operation"] != "terminate_process" || terminated["status"] != "terminated" || terminated["process_id"] != processID {
		t.Fatalf("terminated = %+v", terminated)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "other_run", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}}); err == nil {
		t.Fatal("cross-run continue err = nil")
	}
}

func TestSandboxProcessContinueSupportsCursorAndStdin(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stdin", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID, _ := start["process_id"].(string)
	if processID == "" || start["stdin_open"] != true || start["next_cursor"] != 0 {
		t.Fatalf("start = %+v", start)
	}

	first, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stdin", ToolCallID: "tc_continue_1", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "alpha\n", "input_seq": 1, "cursor": 0}})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(first["stdout"].(string), "alpha") {
		first, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_stdin", ToolCallID: "tc_continue_poll_1", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": 0}})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(first["stdout"].(string), "alpha") {
		t.Fatalf("first = %+v", first)
	}
	cursor, ok := first["next_cursor"].(int)
	if !ok || cursor <= 0 {
		t.Fatalf("cursor = %+v result=%+v", first["next_cursor"], first)
	}

	second, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stdin", ToolCallID: "tc_continue_2", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "beta\n", "input_seq": 2, "cursor": cursor, "close_stdin": true}})
	if err != nil {
		t.Fatal(err)
	}
	deadline = time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(second["stdout"].(string), "beta") {
		second, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_stdin", ToolCallID: "tc_continue_poll_2", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": cursor}})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if strings.Contains(second["stdout"].(string), "alpha") || !strings.Contains(second["stdout"].(string), "beta") || second["stdin_open"] != false {
		t.Fatalf("second = %+v", second)
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
