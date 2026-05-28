package runtime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/sheridiany/loomi/internal/productdata"
)

type failingSandboxProcessRepository struct {
	saveErr error
	listErr error
	saves   int
	records map[string]SandboxProcessRecord
}

func (r *failingSandboxProcessRepository) SaveSandboxProcess(_ context.Context, record SandboxProcessRecord) error {
	r.saves++
	if r.saveErr != nil {
		return r.saveErr
	}
	if r.records == nil {
		r.records = map[string]SandboxProcessRecord{}
	}
	r.records[record.ProcessID] = record
	return nil
}

func (r *failingSandboxProcessRepository) ListSandboxProcesses(context.Context) ([]SandboxProcessRecord, error) {
	if r.listErr != nil {
		return nil, r.listErr
	}
	records := make([]SandboxProcessRecord, 0, len(r.records))
	for _, record := range r.records {
		records = append(records, record)
	}
	return records, nil
}

func (r *failingSandboxProcessRepository) DeleteSandboxProcessesUpdatedBefore(context.Context, time.Time) (int, error) {
	return 0, nil
}

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

func TestSandboxExecCommandRedactsSecretLookingOutput(t *testing.T) {
	root := t.TempDir()
	executor := SandboxToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameSandboxExecCommand, ArgumentsSummary: map[string]any{"argv": []any{"go", "test", "./.../token=secret", "/Users/xuean/private"}}})
	if err == nil {
		t.Fatalf("unsafe validation args unexpectedly executed: %+v", result)
	}

	leaky := boundedOutput{limit: 1024}
	_, _ = leaky.Write([]byte("token=secret\n/Users/xuean/private\n" + root + "/src\n"))
	preview, redacted := sandboxSafeOutputPreview(leaky.String(), root)
	if !redacted || strings.Contains(preview, "token=secret") || strings.Contains(preview, "/Users/") || strings.Contains(preview, root) {
		t.Fatalf("preview=%q redacted=%v", preview, redacted)
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
	if terminated["operation"] != "terminate_process" || (terminated["status"] != "terminated" && terminated["status"] != "exited") || terminated["process_id"] != processID {
		t.Fatalf("terminated = %+v", terminated)
	}
	if terminated["terminal_summary"] == "" {
		t.Fatalf("terminated without terminal summary: %+v", terminated)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "other_run", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}}); err == nil {
		t.Fatal("cross-run continue err = nil")
	}
}

func TestSandboxProcessContinueReturnsPromptlyWithoutOutput(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_prompt", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	started := time.Now()
	continued, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_prompt", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": 0}})
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(started) > 100*time.Millisecond || continued["status"] != "running" || continued["stdout"] != "" {
		t.Fatalf("continue blocked or returned unexpected output after %s: %+v", time.Since(started), continued)
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_prompt", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
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

func TestSandboxProcessContinueCursorReadsBoundedLongOutput(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_long_output", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000, "max_output_bytes": 12}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)

	first, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_long_output", ToolCallID: "tc_continue_1", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "alpha-0000\n", "input_seq": 1, "cursor": 0}})
	if err != nil {
		t.Fatal(err)
	}
	first = waitForSandboxStdout(t, executor, "run_long_output", processID, 0, "alpha")
	cursor := first["next_cursor"].(int)
	if cursor <= 0 || first["stdout_truncated"] != false {
		t.Fatalf("first = %+v", first)
	}

	second, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_long_output", ToolCallID: "tc_continue_2", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "beta-1111\n", "input_seq": 2, "cursor": cursor}})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && !strings.Contains(second["stdout"].(string), "beta") {
		second, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_long_output", ToolCallID: "tc_continue_poll_2", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": cursor}})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(second["stdout"].(string), "beta") || strings.Contains(second["stdout"].(string), "alpha") {
		t.Fatalf("second = %+v", second)
	}
	if second["next_cursor"].(int) <= cursor {
		t.Fatalf("cursor did not advance: first=%+v second=%+v", first, second)
	}
	if second["stdout_bytes"].(int) <= 12 || second["stdout_truncated"] != true {
		t.Fatalf("long output was not bounded: %+v", second)
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_long_output", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
}

func TestSandboxProcessContinueAfterExitReturnsTerminalSummary(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_exit", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	continued, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_exit", ToolCallID: "tc_continue_1", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "done\n", "input_seq": 1, "close_stdin": true}})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) && continued["status"] == "running" {
		continued, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_exit", ToolCallID: "tc_continue_poll", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": continued["next_cursor"]}})
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if continued["status"] != "exited" || continued["exit_code"] != 0 || !strings.Contains(continued["terminal_summary"].(string), "exited exit_code=0") {
		t.Fatalf("continued = %+v", continued)
	}
}

func TestSandboxProcessContinueAfterTerminateOnlyReturnsSafeState(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_terminated_continue", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	terminated, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_terminated_continue", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
	if err != nil {
		t.Fatal(err)
	}
	continued, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_terminated_continue", ToolCallID: "tc_continue_after_terminate", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "should-not-write\n", "input_seq": 1, "close_stdin": true, "cursor": terminated["next_cursor"]}})
	if err == nil {
		t.Fatalf("terminal process accepted stdin: %+v", continued)
	}
	continued, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_terminated_continue", ToolCallID: "tc_continue_after_terminate_read", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": terminated["next_cursor"]}})
	if err != nil {
		t.Fatal(err)
	}
	if continued["status"] != "terminated" || continued["operation"] != "continue_process" || continued["input_seq"] != 0 || continued["stdin_open"] != false {
		t.Fatalf("continued = %+v", continued)
	}
	if strings.Contains(continued["stdout"].(string), "should-not-write") {
		t.Fatalf("continue wrote to terminated process: %+v", continued)
	}
}

func TestSandboxProcessOutputRedactionCoversPathsAndSecrets(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_redaction", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	secretText := "token=secret\n" + root + "/src\n/Users/xuean/private\n"
	_, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_redaction", ToolCallID: "tc_continue_1", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": secretText, "input_seq": 1, "cursor": 0}})
	if err != nil {
		t.Fatal(err)
	}
	continued := waitForSandboxStdout(t, executor, "run_redaction", processID, 0, "[redacted]")
	if continued["redaction_applied"] != true || strings.Contains(continued["stdout"].(string), "token=secret") || strings.Contains(continued["stdout"].(string), root) || strings.Contains(continued["stdout"].(string), "/Users/") {
		t.Fatalf("continued = %+v", continued)
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_redaction", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
}

func TestSandboxProcessRegistryRebuildKeepsTerminalProcessReadableButNotWritable(t *testing.T) {
	root := t.TempDir()
	repo := NewMemorySandboxProcessRepository()
	store := NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_resume_terminal", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	_, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_resume_terminal", ToolCallID: "tc_continue_close", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "persisted tail\n", "input_seq": 1, "close_stdin": true}})
	if err != nil {
		t.Fatal(err)
	}
	terminal := waitForSandboxStatus(t, executor, "run_resume_terminal", processID, "exited")

	resumed := SandboxToolExecutor{Root: root, Store: NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})}
	read, err := resumed.Execute(context.Background(), ToolInvocation{RunID: "run_resume_terminal", ToolCallID: "tc_continue_read", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": 0}})
	if err != nil {
		t.Fatal(err)
	}
	if read["status"] != "exited" || read["next_cursor"] != terminal["next_cursor"] || !strings.Contains(read["stdout"].(string), "persisted tail") {
		t.Fatalf("read = %+v terminal=%+v", read, terminal)
	}
	if _, err := resumed.Execute(context.Background(), ToolInvocation{RunID: "run_resume_terminal", ToolCallID: "tc_continue_write", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "must-not-write\n", "input_seq": 2}}); err == nil {
		t.Fatal("terminal process accepted stdin after registry rebuild")
	}
}

func TestSandboxProcessStartFailsWhenDurableSaveFails(t *testing.T) {
	root := t.TempDir()
	repo := &failingSandboxProcessRepository{saveErr: errors.New("database unavailable")}
	executor := SandboxToolExecutor{Root: root, Store: NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})}

	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_save_start", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})

	if err == nil || !strings.Contains(err.Error(), "durable state could not be saved") {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	if repo.saves == 0 {
		t.Fatal("start did not attempt durable save")
	}
}

func TestSandboxProcessContinueFailsWhenDurableSaveFails(t *testing.T) {
	root := t.TempDir()
	repo := &failingSandboxProcessRepository{}
	store := NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})
	executor := SandboxToolExecutor{Root: root, Store: store}
	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_save_continue", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	repo.saveErr = errors.New("database unavailable")

	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_save_continue", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "hello\n", "input_seq": 1}})

	if err == nil || !strings.Contains(err.Error(), "durable state could not be saved") {
		t.Fatalf("result=%+v err=%v", result, err)
	}
	cleanupSandboxProcess(t, store, processID)
}

func TestSandboxProcessCompletionSaveFailureSurfacesOnContinue(t *testing.T) {
	root := t.TempDir()
	repo := &failingSandboxProcessRepository{}
	store := NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})
	executor := SandboxToolExecutor{Root: root, Store: store}
	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_save_completion", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	repo.saveErr = errors.New("database unavailable")
	closeSandboxProcessStdin(t, store, processID)
	if !waitForSandboxProcessDone(store, processID, time.Second) {
		t.Fatal("process did not exit")
	}

	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_save_completion", ToolCallID: "tc_poll", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}})

	if err == nil || !strings.Contains(err.Error(), "durable state could not be saved") {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}

func TestSandboxProcessRegistryRebuildMarksMissingRunningProcessFailed(t *testing.T) {
	root := t.TempDir()
	repo := NewMemorySandboxProcessRepository()
	started := time.Now().Add(-time.Minute)
	if err := repo.SaveSandboxProcess(context.Background(), SandboxProcessRecord{
		RunID: "run_missing", ProcessID: "sp_missing", ArgvSummary: []string{"cat"}, CwdAlias: ".", Status: SandboxProcessStatusRunning, Cursor: 5, StartedAt: started, UpdatedAt: started, StdoutTail: "hello", StdoutCursor: 5,
	}); err != nil {
		t.Fatal(err)
	}
	executor := SandboxToolExecutor{Root: root, Store: NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{})}

	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_missing", ToolCallID: "tc_continue_missing", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": "sp_missing"}})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "lost" || !strings.Contains(result["terminal_summary"].(string), "process missing after registry restore") {
		t.Fatalf("result = %+v", result)
	}
}

func TestSandboxProcessTTLCleanupExpiresOldRunningProcessWithoutKillingForeignProcesses(t *testing.T) {
	root := t.TempDir()
	now := time.Now()
	repo := NewMemorySandboxProcessRepository()
	if err := repo.SaveSandboxProcess(context.Background(), SandboxProcessRecord{
		RunID: "run_ttl", ProcessID: "sp_ttl", ArgvSummary: []string{"cat"}, CwdAlias: ".", Status: SandboxProcessStatusRunning, Cursor: 0, StartedAt: now.Add(-2 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour),
	}); err != nil {
		t.Fatal(err)
	}
	store := NewSandboxProcessStoreWithRepository(repo, SandboxProcessStoreOptions{Now: func() time.Time { return now }, MaxLifetime: time.Hour, IdleTimeout: 30 * time.Minute})
	executor := SandboxToolExecutor{Root: root, Store: store}

	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_ttl", ToolCallID: "tc_continue_expired", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": "sp_ttl"}})
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "expired" || !strings.Contains(result["terminal_summary"].(string), "expired") {
		t.Fatalf("result = %+v", result)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "other_run", ToolCallID: "tc_terminate_cross", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": "sp_ttl"}}); err == nil {
		t.Fatal("cross-run terminate accepted expired handle")
	}
}

func TestSandboxProcessCrossRunContinueAndTerminateRejected(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_owner", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_other", ToolCallID: "tc_continue_cross", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}}); err == nil {
		t.Fatal("cross-run continue accepted")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_other", ToolCallID: "tc_terminate_cross", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}}); err == nil {
		t.Fatal("cross-run terminate accepted")
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_owner", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
}

func TestSandboxProcessTerminalRunRejectsStartAndContinueMutation(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_done", ToolCallID: "tc_start_terminal", ToolName: productdata.ToolNameSandboxStartProcess, RunStatus: productdata.RunStatusCompleted, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true}}); err == nil {
		t.Fatal("terminal run accepted start_process")
	}
	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_done", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_done", ToolCallID: "tc_continue_terminal_write", ToolName: productdata.ToolNameSandboxContinueProcess, RunStatus: productdata.RunStatusCompleted, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "blocked\n", "input_seq": 1}}); err == nil {
		t.Fatal("terminal run accepted continue mutation")
	}
	read, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_done", ToolCallID: "tc_continue_terminal_read", ToolName: productdata.ToolNameSandboxContinueProcess, RunStatus: productdata.RunStatusCompleted, ArgumentsSummary: map[string]any{"process_id": processID}})
	if err != nil {
		t.Fatal(err)
	}
	if read["operation"] != "continue_process" {
		t.Fatalf("read = %+v", read)
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_done", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
}

func TestSandboxProcessEventMetadataDoesNotLeakPathsOrSecrets(t *testing.T) {
	root := t.TempDir()
	store := NewSandboxProcessStore()
	executor := SandboxToolExecutor{Root: root, Store: store}

	start, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_safe_metadata", ToolCallID: "tc_start", ToolName: productdata.ToolNameSandboxStartProcess, ArgumentsSummary: map[string]any{"argv": []any{"cat"}, "stdin": true, "timeout_ms": 100000}})
	if err != nil {
		t.Fatal(err)
	}
	processID := start["process_id"].(string)
	_, err = executor.Execute(context.Background(), ToolInvocation{RunID: "run_safe_metadata", ToolCallID: "tc_continue", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "stdin_text": "token=secret\n" + root + "/private\n/Users/xuean/private\n", "input_seq": 1}})
	if err != nil {
		t.Fatal(err)
	}
	result := waitForSandboxStdout(t, executor, "run_safe_metadata", processID, 0, "[redacted]")
	rendered := productdata.RedactEventMetadata(result)
	body := strings.Join([]string{stringValue(rendered, "argv_summary"), stringValue(rendered, "stdout"), stringValue(rendered, "terminal_summary")}, "\n")
	if strings.Contains(body, root) || strings.Contains(body, "/Users/") || strings.Contains(body, "token=secret") {
		t.Fatalf("metadata leaked: %+v", rendered)
	}
	_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: "run_safe_metadata", ToolCallID: "tc_terminate", ToolName: productdata.ToolNameSandboxTerminateProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
}

func waitForSandboxStdout(t *testing.T, executor SandboxToolExecutor, runID string, processID string, cursor int, want string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var result map[string]any
	var err error
	for time.Now().Before(deadline) {
		result, err = executor.Execute(context.Background(), ToolInvocation{RunID: runID, ToolCallID: "tc_continue_wait", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID, "cursor": cursor}})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(result["stdout"].(string), want) {
			return result
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("stdout never contained %q: %+v", want, result)
	return nil
}

func waitForSandboxStatus(t *testing.T, executor SandboxToolExecutor, runID string, processID string, want string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	var result map[string]any
	var err error
	for time.Now().Before(deadline) {
		result, err = executor.Execute(context.Background(), ToolInvocation{RunID: runID, ToolCallID: "tc_continue_status_wait", ToolName: productdata.ToolNameSandboxContinueProcess, ArgumentsSummary: map[string]any{"process_id": processID}})
		if err != nil {
			t.Fatal(err)
		}
		if result["status"] == want {
			return result
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("status never became %q: %+v", want, result)
	return nil
}

func waitForSandboxProcessDone(store *SandboxProcessStore, processID string, timeout time.Duration) bool {
	if store == nil {
		return false
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		store.mu.Lock()
		process := store.processes[processID]
		store.mu.Unlock()
		if process == nil {
			return false
		}
		select {
		case <-process.done:
			return true
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
	return false
}

func closeSandboxProcessStdin(t *testing.T, store *SandboxProcessStore, processID string) {
	t.Helper()
	store.mu.Lock()
	process := store.processes[processID]
	store.mu.Unlock()
	if process == nil {
		t.Fatalf("process %s not found", processID)
	}
	process.mu.Lock()
	stdin := process.stdin
	process.stdinOpen = false
	process.mu.Unlock()
	if stdin != nil {
		if err := stdin.Close(); err != nil {
			t.Fatal(err)
		}
	}
}

func cleanupSandboxProcess(t *testing.T, store *SandboxProcessStore, processID string) {
	t.Helper()
	store.mu.Lock()
	process := store.processes[processID]
	store.mu.Unlock()
	if process == nil {
		return
	}
	process.mu.Lock()
	cancel := process.cancel
	command := process.command
	process.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if command != nil && command.Process != nil {
		_ = killSandboxProcessGroup(command.Process, syscall.SIGTERM)
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
