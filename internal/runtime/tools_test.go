package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentTimeToolDefinitionValidatesTimezone(t *testing.T) {
	tool := CurrentTimeToolDefinition()
	if tool.Name != "runtime.get_current_time" {
		t.Fatalf("tool.Name = %q", tool.Name)
	}
	if tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", tool.ApprovalPolicy)
	}
	if tool.SafetyClass != ToolSafetyNoSideEffectInternal {
		t.Fatalf("SafetyClass = %q", tool.SafetyClass)
	}
	if got, err := tool.NormalizeArguments(map[string]any{}); err != nil || got.Timezone != "UTC" {
		t.Fatalf("NormalizeArguments(empty) = %+v, %v", got, err)
	}
	if got, err := tool.NormalizeArguments(map[string]any{"timezone": "UTC"}); err != nil || got.Timezone != "UTC" {
		t.Fatalf("NormalizeArguments(UTC) = %+v, %v", got, err)
	}
	if _, err := tool.NormalizeArguments(map[string]any{"timezone": "Asia/Shanghai"}); err == nil {
		t.Fatal("NormalizeArguments(Asia/Shanghai) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"shell": "pwd"}); err == nil {
		t.Fatal("NormalizeArguments(shell) error = nil, want error")
	}
}

func TestCurrentTimeToolExecutesSafeResult(t *testing.T) {
	tool := CurrentTimeToolDefinition()
	result, err := tool.Execute(ToolArguments{Timezone: "UTC"})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result["timezone"] != "UTC" || result["source"] != "runtime" || result["iso_time"] == "" {
		t.Fatalf("result = %+v", result)
	}
}

func TestWorkspaceReadToolsValidateRootAndSensitivePaths(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(main.go) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("SECRET=sk-live-123\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(.env) error = %v", err)
	}

	readTool := WorkspaceReadFileToolDefinition(root)
	if readTool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", readTool.ApprovalPolicy)
	}
	if readTool.SafetyClass != ToolSafetyWorkspaceReadOnly {
		t.Fatalf("SafetyClass = %q", readTool.SafetyClass)
	}
	if _, err := readTool.NormalizeArguments(map[string]any{"path": "../outside.txt"}); err == nil {
		t.Fatal("NormalizeArguments(path traversal) error = nil, want error")
	}
	if _, err := readTool.NormalizeArguments(map[string]any{"path": ".env"}); err == nil {
		t.Fatal("NormalizeArguments(.env) error = nil, want error")
	}
	if args, err := readTool.NormalizeArguments(map[string]any{"path": "main.go", "max_bytes": 12}); err != nil || args.Path != "main.go" || args.MaxBytes != 12 {
		t.Fatalf("NormalizeArguments(main.go) = %+v, %v", args, err)
	}
}

func TestWorkspaceGlobGrepAndReadReturnBoundedRelativeResults(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "internal"), 0o700); err != nil {
		t.Fatalf("Mkdir(internal) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "tool.go"), []byte("package internal\nfunc Tool() {}\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(tool.go) error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# Loomi\nworkspace read tools\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(README.md) error = %v", err)
	}

	globTool := WorkspaceGlobToolDefinition(root)
	globArgs, err := globTool.NormalizeArguments(map[string]any{"pattern": "**/*.go", "limit": 5})
	if err != nil {
		t.Fatalf("NormalizeArguments(glob) error = %v", err)
	}
	globResult, err := globTool.Execute(globArgs)
	if err != nil {
		t.Fatalf("Execute(glob) error = %v", err)
	}
	matches, ok := globResult["matches"].([]string)
	if !ok || len(matches) != 1 || matches[0] != "internal/tool.go" {
		t.Fatalf("glob matches = %#v", globResult["matches"])
	}

	grepTool := WorkspaceGrepToolDefinition(root)
	grepArgs, err := grepTool.NormalizeArguments(map[string]any{"query": "workspace", "path": ".", "limit": 5})
	if err != nil {
		t.Fatalf("NormalizeArguments(grep) error = %v", err)
	}
	grepResult, err := grepTool.Execute(grepArgs)
	if err != nil {
		t.Fatalf("Execute(grep) error = %v", err)
	}
	grepMatches, ok := grepResult["matches"].([]WorkspaceGrepMatch)
	if !ok || len(grepMatches) != 1 || grepMatches[0].Path != "README.md" || grepMatches[0].Line != 2 {
		t.Fatalf("grep matches = %#v", grepResult["matches"])
	}

	readTool := WorkspaceReadFileToolDefinition(root)
	readArgs, err := readTool.NormalizeArguments(map[string]any{"path": "README.md", "max_bytes": 8})
	if err != nil {
		t.Fatalf("NormalizeArguments(read) error = %v", err)
	}
	readResult, err := readTool.Execute(readArgs)
	if err != nil {
		t.Fatalf("Execute(read) error = %v", err)
	}
	if readResult["path"] != "README.md" || readResult["truncated"] != true || strings.Contains(readResult["preview"].(string), "tools") {
		t.Fatalf("read result = %+v", readResult)
	}
}

func TestWorkspaceWriteFileValidatesAndWritesTextOnlyInsideRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "internal"), 0o700); err != nil {
		t.Fatalf("Mkdir(internal) error = %v", err)
	}
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatalf("Symlink(escape) error = %v", err)
	}

	tool := WorkspaceWriteFileToolDefinition(root)
	if tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", tool.ApprovalPolicy)
	}
	if tool.SafetyClass != ToolSafetyWorkspaceWrite {
		t.Fatalf("SafetyClass = %q", tool.SafetyClass)
	}
	if _, err := tool.NormalizeArguments(map[string]any{"path": "../outside.txt", "content": "x"}); err == nil {
		t.Fatal("NormalizeArguments(path traversal) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"path": ".env", "content": "SECRET=sk-live"}); err == nil {
		t.Fatal("NormalizeArguments(.env) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"path": "missing/file.txt", "content": "x"}); err == nil {
		t.Fatal("NormalizeArguments(missing parent) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"path": "escape/file.txt", "content": "x"}); err == nil {
		t.Fatal("NormalizeArguments(symlink escape) error = nil, want error")
	}

	args, err := tool.NormalizeArguments(map[string]any{"path": "internal/generated.txt", "content": "hello Loomi\n"})
	if err != nil {
		t.Fatalf("NormalizeArguments(write) error = %v", err)
	}
	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute(write) error = %v", err)
	}
	if result["path"] != "internal/generated.txt" || result["bytes_written"] != len("hello Loomi\n") || result["created"] != true {
		t.Fatalf("write result = %+v", result)
	}
	content, err := os.ReadFile(filepath.Join(root, "internal", "generated.txt"))
	if err != nil {
		t.Fatalf("ReadFile(generated) error = %v", err)
	}
	if string(content) != "hello Loomi\n" {
		t.Fatalf("generated content = %q", content)
	}
}

func TestWorkspaceEditReplacesExactlyOnceAndDoesNotMutateOnFailure(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "notes.txt")
	original := "alpha\nbeta\ngamma\n"
	if err := os.WriteFile(target, []byte(original), 0o600); err != nil {
		t.Fatalf("WriteFile(notes) error = %v", err)
	}

	tool := WorkspaceEditToolDefinition(root)
	args, err := tool.NormalizeArguments(map[string]any{"path": "notes.txt", "old_text": "beta\n", "new_text": "delta\n"})
	if err != nil {
		t.Fatalf("NormalizeArguments(edit) error = %v", err)
	}
	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute(edit) error = %v", err)
	}
	if result["path"] != "notes.txt" || result["replacements"] != 1 {
		t.Fatalf("edit result = %+v", result)
	}
	content, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile(notes) error = %v", err)
	}
	if string(content) != "alpha\ndelta\ngamma\n" {
		t.Fatalf("edited content = %q", content)
	}

	beforeMissing := string(content)
	missingArgs, err := tool.NormalizeArguments(map[string]any{"path": "notes.txt", "old_text": "missing", "new_text": "x"})
	if err != nil {
		t.Fatalf("NormalizeArguments(missing) error = %v", err)
	}
	if _, err := tool.Execute(missingArgs); err == nil {
		t.Fatal("Execute(missing old_text) error = nil, want error")
	}
	afterMissing, _ := os.ReadFile(target)
	if string(afterMissing) != beforeMissing {
		t.Fatalf("missing edit mutated file: %q", afterMissing)
	}

	if err := os.WriteFile(target, []byte("same\nsame\n"), 0o600); err != nil {
		t.Fatalf("WriteFile(duplicate) error = %v", err)
	}
	duplicateArgs, err := tool.NormalizeArguments(map[string]any{"path": "notes.txt", "old_text": "same", "new_text": "once"})
	if err != nil {
		t.Fatalf("NormalizeArguments(duplicate) error = %v", err)
	}
	if _, err := tool.Execute(duplicateArgs); err == nil {
		t.Fatal("Execute(duplicate old_text) error = nil, want error")
	}
	afterDuplicate, _ := os.ReadFile(target)
	if string(afterDuplicate) != "same\nsame\n" {
		t.Fatalf("duplicate edit mutated file: %q", afterDuplicate)
	}
}

func TestWorkspaceExecCommandRunsArgvInsideWorkspaceWithBoundedOutput(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "internal"), 0o700); err != nil {
		t.Fatalf("Mkdir(internal) error = %v", err)
	}

	tool := WorkspaceExecCommandToolDefinition(root)
	if tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", tool.ApprovalPolicy)
	}
	if tool.SafetyClass != ToolSafetyWorkspaceExec {
		t.Fatalf("SafetyClass = %q", tool.SafetyClass)
	}
	if _, err := tool.NormalizeArguments(map[string]any{"command": []any{"sh", "-c", "echo no"}}); err == nil {
		t.Fatal("NormalizeArguments(shell wrapper) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"command": []any{"rm", "-rf", "internal"}}); err == nil {
		t.Fatal("NormalizeArguments(dangerous command) error = nil, want error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"command": []any{"printf", "x"}, "cwd": "../"}); err == nil {
		t.Fatal("NormalizeArguments(cwd escape) error = nil, want error")
	}

	args, err := tool.NormalizeArguments(map[string]any{"command": []any{"printf", "hello"}, "cwd": "internal", "timeout_seconds": 5})
	if err != nil {
		t.Fatalf("NormalizeArguments(exec) error = %v", err)
	}
	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute(exec) error = %v", err)
	}
	if result["cwd"] != "internal" || result["exit_code"] != 0 || result["stdout"] != "hello" || result["stderr"] != "" || result["timed_out"] != false {
		t.Fatalf("exec result = %+v", result)
	}
}

func TestWorkspaceExecCommandTimesOutAndTruncatesOutput(t *testing.T) {
	root := t.TempDir()
	tool := WorkspaceExecCommandToolDefinition(root)
	timeoutArgs, err := tool.NormalizeArguments(map[string]any{"command": []any{"sleep", "2"}, "timeout_seconds": 1})
	if err != nil {
		t.Fatalf("NormalizeArguments(timeout) error = %v", err)
	}
	timeoutResult, err := tool.Execute(timeoutArgs)
	if err != nil {
		t.Fatalf("Execute(timeout) error = %v", err)
	}
	if timeoutResult["timed_out"] != true || timeoutResult["exit_code"] != -1 {
		t.Fatalf("timeout result = %+v", timeoutResult)
	}

	outputArgs, err := tool.NormalizeArguments(map[string]any{"command": []any{"printf", strings.Repeat("x", 9000)}, "timeout_seconds": 5})
	if err != nil {
		t.Fatalf("NormalizeArguments(output) error = %v", err)
	}
	outputResult, err := tool.Execute(outputArgs)
	if err != nil {
		t.Fatalf("Execute(output) error = %v", err)
	}
	if outputResult["stdout_truncated"] != true || len(outputResult["stdout"].(string)) > 4096 {
		t.Fatalf("output result = %+v", outputResult)
	}
}

func TestTodoWriteToolNormalizesAndCountsItems(t *testing.T) {
	tool := TodoWriteToolDefinition()
	args, err := tool.NormalizeArguments(map[string]any{"items": []any{map[string]any{"title": " Inspect current tools ", "status": "completed"}, map[string]any{"title": "Implement todo_write", "status": "in_progress"}, map[string]any{"title": "Run validation"}}})
	if err != nil {
		t.Fatalf("NormalizeArguments() error = %v", err)
	}
	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result["total"] != 3 || result["completed_count"] != 1 || result["in_progress_count"] != 1 || result["pending_count"] != 1 {
		t.Fatalf("result = %+v", result)
	}
	items := result["items"].([]map[string]string)
	if items[0]["title"] != "Inspect current tools" || items[2]["status"] != "pending" {
		t.Fatalf("items = %+v", items)
	}

	if _, err := tool.NormalizeArguments(map[string]any{"items": []any{}}); err == nil {
		t.Fatal("expected empty items error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"items": []any{map[string]any{"title": "bad", "status": "blocked"}}}); err == nil {
		t.Fatal("expected invalid status error")
	}
}

func TestMCPCallToolNormalizesAndEchoesLocalMessage(t *testing.T) {
	tool := MCPCallToolDefinition()
	if tool.ApprovalPolicy != ToolApprovalAlwaysRequired {
		t.Fatalf("ApprovalPolicy = %q", tool.ApprovalPolicy)
	}
	if tool.SafetyClass != ToolSafetyMCPBridge {
		t.Fatalf("SafetyClass = %q", tool.SafetyClass)
	}
	args, err := tool.NormalizeArguments(map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": " hello mcp "}})
	if err != nil {
		t.Fatalf("NormalizeArguments() error = %v", err)
	}
	if args.MCPServer != "local" || args.MCPTool != "echo" || args.MCPMessage != "hello mcp" {
		t.Fatalf("args = %+v", args)
	}
	result, err := tool.Execute(args)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result["server"] != "local" || result["tool"] != "echo" || result["message"] != "hello mcp" || result["side_effect"] != "none" {
		t.Fatalf("result = %+v", result)
	}

	if _, err := tool.NormalizeArguments(map[string]any{"server": "remote", "tool": "echo", "arguments": map[string]any{"message": "hello"}}); err == nil {
		t.Fatal("expected unknown server error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"server": "local", "tool": "shell", "arguments": map[string]any{"message": "hello"}}); err == nil {
		t.Fatal("expected unknown tool error")
	}
	if _, err := tool.NormalizeArguments(map[string]any{"server": "local", "tool": "echo", "arguments": map[string]any{"message": "secret=sk-live"}}); err == nil {
		t.Fatal("expected secret-looking message error")
	}
}

func TestToolCatalogListsAllowlistedToolsWithRiskMetadata(t *testing.T) {
	catalog := ToolCatalog()
	names := make([]string, 0, len(catalog.Tools))
	for _, tool := range catalog.Tools {
		names = append(names, tool.Name)
		if tool.ApprovalPolicy != string(ToolApprovalAlwaysRequired) {
			t.Fatalf("tool %s ApprovalPolicy = %q", tool.Name, tool.ApprovalPolicy)
		}
		if tool.Enabled != true || tool.Description == "" || tool.RiskLevel == "" || tool.SideEffect == "" {
			t.Fatalf("tool = %+v", tool)
		}
	}
	want := []string{"runtime.get_current_time", "runtime.todo_write", "mcp.call_tool", "workspace.glob", "workspace.grep", "workspace.read_file", "workspace.write_file", "workspace.edit", "workspace.exec_command"}
	if strings.Join(names, ",") != strings.Join(want, ",") {
		t.Fatalf("names = %+v, want %+v", names, want)
	}
	todoTool := catalog.Tools[1]
	if todoTool.Name != "runtime.todo_write" || todoTool.Capability != "plan" || todoTool.RiskLevel != "low" || todoTool.SideEffect != "none" {
		t.Fatalf("todo tool = %+v", todoTool)
	}
	mcpTool := catalog.Tools[2]
	if mcpTool.Name != "mcp.call_tool" || mcpTool.Capability != "call_tool" || mcpTool.RiskLevel != "medium" || mcpTool.SideEffect != "mcp" || mcpTool.SafetyClass != string(ToolSafetyMCPBridge) {
		t.Fatalf("mcp tool = %+v", mcpTool)
	}
	execTool := catalog.Tools[len(catalog.Tools)-1]
	if execTool.RiskLevel != "high" || execTool.SideEffect != "process" || execTool.SafetyClass != string(ToolSafetyWorkspaceExec) {
		t.Fatalf("exec tool = %+v", execTool)
	}
}
