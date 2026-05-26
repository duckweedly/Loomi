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

func TestWorkspaceReadToolsExecuteWithinFixtureRoot(t *testing.T) {
	root := createWorkspaceFixture(t)
	executor := WorkspaceToolExecutor{Root: root}

	glob, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGlob, ArgumentsSummary: map[string]any{"pattern": "**/*.txt", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if glob["match_count"] != 2 || glob["scope"] != "workspace" {
		t.Fatalf("glob = %+v", glob)
	}

	grep, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGrep, ArgumentsSummary: map[string]any{"query": "needle", "path": "src", "include": "*.txt", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if grep["match_count"] != 1 || strings.Contains(fmt.Sprintf("%+v", grep), root) {
		t.Fatalf("grep = %+v", grep)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "limit": 6}})
	if err != nil {
		t.Fatal(err)
	}
	if read["content"] != "needle" || read["truncated"] != true || read["path"] != "src/notes.txt" {
		t.Fatalf("read = %+v", read)
	}
}

func TestWorkspaceReadToolsRejectTraversalSensitiveAndSymlinkEscape(t *testing.T) {
	root := createWorkspaceFixture(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("outside secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "src", "outside-link.txt")); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	cases := []map[string]any{
		{"path": "../outside.txt"},
		{"path": filepath.Join(outside, "secret.txt")},
		{"path": ".env"},
		{"path": "secrets/token.txt"},
		{"path": "src/outside-link.txt"},
	}
	for _, args := range cases {
		_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: args})
		if err == nil {
			t.Fatalf("read(%+v) err = nil", args)
		}
		if strings.Contains(err.Error(), "outside secret") || strings.Contains(err.Error(), "fixture-secret") {
			t.Fatalf("error leaked content: %v", err)
		}
	}
}

func TestWorkspaceReadToolsSummarizeBinaryAndInvalidUTF8(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "bin.dat"), []byte{0, 1, 2}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "bad.txt"), []byte{0xff, 'a'}, 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	binary, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "bin.dat"}})
	if err != nil {
		t.Fatal(err)
	}
	if binary["utf8_valid"] != false || binary["content"] != "" {
		t.Fatalf("binary = %+v", binary)
	}
	invalid, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "bad.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if invalid["utf8_valid"] != false || invalid["content"] != "a" {
		t.Fatalf("invalid = %+v", invalid)
	}
}

func TestWorkspaceWriteFileCreatesNewTextFileWithinFixtureRoot(t *testing.T) {
	root := createWorkspaceFixture(t)
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceWriteFile, ArgumentsSummary: map[string]any{"path": "src/generated.txt", "content": "hello\nworld\n"}})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "src", "generated.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "hello\nworld\n" {
		t.Fatalf("content = %q", string(content))
	}
	if result["tool"] != productdata.ToolNameWorkspaceWriteFile || result["scope"] != "workspace" || result["operation"] != "write_file" || result["path"] != "src/generated.txt" || result["changed"] != true || result["bytes_written"] != len(content) || result["line_count_after"] != 2 {
		t.Fatalf("result = %+v", result)
	}
	if strings.Contains(fmt.Sprintf("%+v", result), root) || strings.Contains(fmt.Sprintf("%+v", result), "hello") {
		t.Fatalf("result leaked root or content: %+v", result)
	}
}

func TestWorkspaceWriteFileRejectsExistingTraversalSensitiveSymlinkInvalidUTF8AndTooLarge(t *testing.T) {
	root := createWorkspaceFixture(t)
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(root, "src", "outside-dir")); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}
	before, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}

	cases := []map[string]any{
		{"path": "src/notes.txt", "content": "overwrite"},
		{"path": "../outside.txt", "content": "escape"},
		{"path": ".env", "content": "secret"},
		{"path": "secrets/token.txt", "content": "secret"},
		{"path": "src/outside-dir/new.txt", "content": "escape"},
		{"path": "src/invalid.txt", "content": string([]byte{0xff, 'x'})},
		{"path": "src/large.txt", "content": strings.Repeat("x", 16), "max_bytes": 8},
	}
	for _, args := range cases {
		if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceWriteFile, ArgumentsSummary: args}); err == nil {
			t.Fatalf("write_file(%+v) err = nil", args)
		}
	}
	after, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("existing file changed: %q", string(after))
	}
	if _, err := os.Stat(filepath.Join(outside, "new.txt")); err == nil {
		t.Fatal("symlink escape wrote outside file")
	}
}

func TestWorkspaceEditReplacesExactTextOnceWithinFixtureRoot(t *testing.T) {
	root := createWorkspaceFixture(t)
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "thread\nsecond line\n" {
		t.Fatalf("content = %q", string(content))
	}
	if result["tool"] != productdata.ToolNameWorkspaceEdit || result["scope"] != "workspace" || result["operation"] != "edit" || result["path"] != "src/notes.txt" || result["changed"] != true || result["bytes_before"] != len("needle\nsecond line\n") || result["bytes_after"] != len(content) || result["line_count_before"] != 2 || result["line_count_after"] != 2 {
		t.Fatalf("result = %+v", result)
	}
	if strings.Contains(fmt.Sprintf("%+v", result), root) || strings.Contains(fmt.Sprintf("%+v", result), "thread") {
		t.Fatalf("result leaked root or content: %+v", result)
	}
}

func TestWorkspaceEditRejectsMissingDuplicateSensitiveBinaryInvalidUTF8TooLargeAndSymlinkEscape(t *testing.T) {
	root := createWorkspaceFixture(t)
	outside := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "src", "dupe.txt"), []byte("same\nsame\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "binary.txt"), []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "invalid.txt"), []byte{0xff, 'x'}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(outside, "secret.txt"), filepath.Join(root, "src", "outside-link.txt")); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}
	before, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}

	cases := []map[string]any{
		{"path": "src/notes.txt", "old_text": "missing", "new_text": "x"},
		{"path": "src/dupe.txt", "old_text": "same", "new_text": "x"},
		{"path": "../outside.txt", "old_text": "x", "new_text": "y"},
		{"path": ".env", "old_text": "fixture", "new_text": "x"},
		{"path": "secrets/token.txt", "old_text": "fixture", "new_text": "x"},
		{"path": "src/binary.txt", "old_text": "a", "new_text": "x"},
		{"path": "src/invalid.txt", "old_text": "x", "new_text": "y"},
		{"path": "src/notes.txt", "old_text": string([]byte{0xff, 'x'}), "new_text": "x"},
		{"path": "src/notes.txt", "old_text": "needle", "new_text": strings.Repeat("x", 16), "max_bytes": 8},
		{"path": "src/outside-link.txt", "old_text": "outside", "new_text": "changed"},
	}
	for _, args := range cases {
		if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: args}); err == nil {
			t.Fatalf("edit(%+v) err = nil", args)
		}
	}
	after, err := os.ReadFile(filepath.Join(root, "src", "notes.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != string(before) {
		t.Fatalf("notes changed: %q", string(after))
	}
	outsideContent, err := os.ReadFile(filepath.Join(outside, "secret.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(outsideContent) != "outside" {
		t.Fatalf("symlink escape changed outside file: %q", string(outsideContent))
	}
}

func createWorkspaceFixture(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "secrets"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "notes.txt"), []byte("needle\nsecond line\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "other.txt"), []byte("haystack\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("fixture-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "secrets", "token.txt"), []byte("fixture-secret"), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}
