package runtime

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

func TestWorkspaceReadToolsRequireAuthorizedRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("LOOMI_WORKSPACE_ROOT", "")

	executor := WorkspaceToolExecutor{}
	_, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGlob, ArgumentsSummary: map[string]any{"pattern": "**/*.txt", "path": ".", "limit": 10}})
	if err == nil || !strings.Contains(err.Error(), "workspace root is not authorized") {
		t.Fatalf("err = %v", err)
	}
}

func TestWorkspaceGrepNoMatchIsBoundedByScannedFiles(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < maxWorkspaceGrepFiles+5; i++ {
		path := filepath.Join(root, fmt.Sprintf("file-%03d.txt", i))
		if err := os.WriteFile(path, []byte("ordinary content\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executor := WorkspaceToolExecutor{Root: root}

	grep, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGrep, ArgumentsSummary: map[string]any{"query": "unlikely_query_to_list_nothing", "path": ".", "limit": 20}})
	if err != nil {
		t.Fatal(err)
	}

	if grep["match_count"] != 0 || grep["truncated"] != true || grep["scanned_file_count"] != maxWorkspaceGrepFiles {
		t.Fatalf("grep = %+v", grep)
	}
}

func TestWorkspaceGlobSkipsGeneratedDirectories(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"node_modules/pkg", "dist/assets", ".claude/worktrees/run"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, dir, "generated.txt"), []byte("generated"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "docs", "keep.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGlob, ArgumentsSummary: map[string]any{"pattern": "**/*", "path": ".", "limit": 20}})
	if err != nil {
		t.Fatal(err)
	}

	matches := result["matches"].([]map[string]any)
	if len(matches) != 1 || matches[0]["path"] != "docs/keep.txt" {
		t.Fatalf("glob matches = %+v", matches)
	}
	if result["skipped_dir_count"] != 3 {
		t.Fatalf("glob result = %+v", result)
	}
}

func TestWorkspaceReadToolsTreatSelectedRootNameAsRootAlias(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "Downloads")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "receipt.txt"), []byte("downloaded file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	glob, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceGlob, ArgumentsSummary: map[string]any{"pattern": "*.txt", "path": "downloads", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}
	if glob["match_count"] != 1 || !strings.Contains(fmt.Sprintf("%+v", glob), "receipt.txt") {
		t.Fatalf("glob = %+v", glob)
	}

	read, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "Downloads/receipt.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if read["path"] != "receipt.txt" || read["content"] != "downloaded file\n" {
		t.Fatalf("read = %+v", read)
	}
}

func TestWorkspaceReadDirectoryReturnsSafeSummary(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "receipt.txt"), []byte("downloaded file\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	read, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": ".", "limit": 10}})
	if err != nil {
		t.Fatal(err)
	}

	if read["path"] != "." || read["kind"] != "directory" || read["content"] != "" || read["entry_count"] != 2 {
		t.Fatalf("read directory = %+v", read)
	}
	if strings.Contains(fmt.Sprintf("%+v", read), root) || !strings.Contains(fmt.Sprintf("%+v", read), "receipt.txt") {
		t.Fatalf("read directory leaked root or missed entries: %+v", read)
	}
}

func TestWorkspaceListDirectoryReturnsBoundedEntries(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "src", "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("# docs\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "nested", "deep.txt"), []byte("deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceListDirectory, ArgumentsSummary: map[string]any{"path": ".", "max_entries": 20, "depth": 2}})
	if err != nil {
		t.Fatal(err)
	}

	entries := result["entries"].([]map[string]any)
	paths := workspaceTestEntryPaths(entries)
	if result["tool"] != productdata.ToolNameWorkspaceListDirectory || result["path"] != "." || result["returned_entries"] != 4 || result["truncated"] != false {
		t.Fatalf("list result = %+v", result)
	}
	for _, expected := range []string{"README.md", "src", "src/main.go", "src/nested"} {
		if !workspaceTestHasPath(paths, expected) {
			t.Fatalf("paths missing %s: %+v", expected, entries)
		}
	}
	if workspaceTestHasPath(paths, "src/nested/deep.txt") || strings.Contains(fmt.Sprintf("%+v", result), root) {
		t.Fatalf("list leaked root or exceeded depth: %+v", result)
	}
}

func TestWorkspaceListDirectoryBoundsDepthSkipsGeneratedAndHidden(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"src/deep/third", "node_modules/pkg", "dist/assets", ".git/objects"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{"src/deep/third/file.go", "node_modules/pkg/generated.js", "dist/assets/app.js", ".hidden.txt"} {
		if err := os.WriteFile(filepath.Join(root, file), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceListDirectory, ArgumentsSummary: map[string]any{"path": ".", "max_entries": 50, "depth": 99}})
	if err != nil {
		t.Fatal(err)
	}

	paths := workspaceTestEntryPaths(result["entries"].([]map[string]any))
	if workspaceTestHasPath(paths, ".hidden.txt") || workspaceTestHasPath(paths, "node_modules/pkg/generated.js") || workspaceTestHasPath(paths, "dist/assets/app.js") || workspaceTestHasPath(paths, "src/deep/third/file.go") {
		t.Fatalf("unexpected paths with default hidden/depth/generated rules: %+v", paths)
	}
	if result["depth"] != 3 || result["skipped_dir_count"] != 3 {
		t.Fatalf("bounded result = %+v", result)
	}

	withHidden, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceListDirectory, ArgumentsSummary: map[string]any{"path": ".", "include_hidden": true, "max_entries": 50}})
	if err != nil {
		t.Fatal(err)
	}
	if !workspaceTestHasPath(workspaceTestEntryPaths(withHidden["entries"].([]map[string]any)), ".hidden.txt") {
		t.Fatalf("hidden entries missing when requested: %+v", withHidden)
	}
}

func TestWorkspaceListDirectoryTruncatesLargeDirectoriesAndRejectsEscape(t *testing.T) {
	root := t.TempDir()
	for i := 0; i < 12; i++ {
		if err := os.WriteFile(filepath.Join(root, fmt.Sprintf("file-%02d.txt", i)), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceListDirectory, ArgumentsSummary: map[string]any{"path": ".", "max_entries": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if result["returned_entries"] != 5 || result["truncated"] != true || result["total_entries_seen"] != 6 {
		t.Fatalf("truncated result = %+v", result)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceListDirectory, ArgumentsSummary: map[string]any{"path": "../outside"}}); err == nil {
		t.Fatal("root escape err = nil")
	}
}

func TestWorkspaceTreeSummaryClassifiesDirectoryContents(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"src/main.go":       "package main\n",
		"docs/spec.pdf":     "pdf",
		"images/photo.jpg":  "jpg",
		"video/demo.mp4":    "mp4",
		"archive/data.zip":  "zip",
		"App.app/Info.txt":  "app",
		"secret-token.txt":  "secret",
		"dist/generated.js": "generated",
	}
	for name, content := range files {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(root, name)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(root, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	executor := WorkspaceToolExecutor{Root: root}

	result, err := executor.Execute(context.Background(), ToolInvocation{ToolName: productdata.ToolNameWorkspaceTreeSummary, ArgumentsSummary: map[string]any{"path": ".", "max_entries": 50, "depth": 3}})
	if err != nil {
		t.Fatal(err)
	}

	byKind := result["by_kind"].(map[string]int)
	byExt := result["by_extension"].(map[string]int)
	if result["tool"] != productdata.ToolNameWorkspaceTreeSummary || result["files_count"] != 7 || result["directories_count"] != 6 || result["skipped_dir_count"] != 1 {
		t.Fatalf("summary = %+v", result)
	}
	for _, kind := range []string{"code", "document", "image", "video", "archive", "app"} {
		if byKind[kind] == 0 {
			t.Fatalf("kind %s missing: %+v", kind, result)
		}
	}
	if byExt[".go"] != 1 || byExt[".pdf"] != 1 {
		t.Fatalf("extensions = %+v", byExt)
	}
	if strings.Contains(fmt.Sprintf("%+v", result), "secret-token") || strings.Contains(fmt.Sprintf("%+v", result), root) {
		t.Fatalf("summary leaked sensitive-looking name or root: %+v", result)
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

	read, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_success", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}})
	if err != nil {
		t.Fatal(err)
	}
	if read["path"] != "src/notes.txt" {
		t.Fatalf("read = %+v", read)
	}
	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_success", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}})
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
	if result["diff"] == "" || result["snippet"] == "" || !strings.Contains(fmt.Sprintf("%+v", result), "-needle") || !strings.Contains(fmt.Sprintf("%+v", result), "+thread") {
		t.Fatalf("result missing diff/snippet: %+v", result)
	}
	if strings.Contains(fmt.Sprintf("%+v", result), root) {
		t.Fatalf("result leaked root or content: %+v", result)
	}
}

func TestWorkspaceEditNormalizesMatchAndPreservesCRLF(t *testing.T) {
	root := createWorkspaceFixture(t)
	path := filepath.Join(root, "src", "crlf.txt")
	if err := os.WriteFile(path, []byte("alpha\r\nbeta\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_crlf", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/crlf.txt"}}); err != nil {
		t.Fatal(err)
	}
	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_crlf", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/crlf.txt", "old_text": "beta\n", "new_text": "gamma\n"}})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "alpha\r\ngamma\r\n" {
		t.Fatalf("content = %q", string(content))
	}
	if result["line_endings_preserved"] != true || result["match_strategy"] != "normalized_line_endings" {
		t.Fatalf("result = %+v", result)
	}
}

func TestWorkspaceEditPreservesIndentationAndStripsTrailingWhitespace(t *testing.T) {
	root := createWorkspaceFixture(t)
	path := filepath.Join(root, "src", "indent.go")
	if err := os.WriteFile(path, []byte("func main() {\n\told()\n}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	executor := WorkspaceToolExecutor{Root: root}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_indent", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/indent.go"}}); err != nil {
		t.Fatal(err)
	}
	result, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_indent", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/indent.go", "old_text": "\told()\n", "new_text": "first()  \nsecond()\t\n"}})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "func main() {\n\tfirst()\n\tsecond()\n}\n" {
		t.Fatalf("content = %q", string(content))
	}
	snippet, _ := result["snippet"].(string)
	if result["indentation_preserved"] != true || result["trailing_whitespace_stripped"] != true || !strings.Contains(snippet, "\tfirst()\n\tsecond()") {
		t.Fatalf("result = %+v", result)
	}
}

func TestWorkspaceEditRequiresReadAndRejectsStaleRead(t *testing.T) {
	root := createWorkspaceFixture(t)
	executor := WorkspaceToolExecutor{Root: root}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_unread", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}}); err == nil {
		t.Fatal("edit before read err = nil")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_stale", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "src", "notes.txt"), []byte("needle\nchanged elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_edit_stale", ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}}); err == nil {
		t.Fatal("stale edit err = nil")
	}
}

func TestWorkspacePatchPreviewAndApplyRequireReadAndPreview(t *testing.T) {
	root := createWorkspaceFixture(t)
	path := filepath.Join(root, "src", "notes.txt")
	executor := WorkspaceToolExecutor{Root: root}
	args := map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_patch", ToolName: productdata.ToolNameWorkspacePatchPreview, ArgumentsSummary: args}); err == nil {
		t.Fatal("preview before read err = nil")
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_patch", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	preview, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_patch", ToolName: productdata.ToolNameWorkspacePatchPreview, ArgumentsSummary: args})
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "needle\nsecond line\n" {
		t.Fatalf("preview changed file: %q", string(content))
	}
	if preview["tool"] != productdata.ToolNameWorkspacePatchPreview || preview["operation"] != "patch_preview" || preview["changed"] != false || preview["preview_id"] == "" || !strings.Contains(fmt.Sprintf("%+v", preview), "-needle") || !strings.Contains(fmt.Sprintf("%+v", preview), "+thread") {
		t.Fatalf("preview = %+v", preview)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_without_preview", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_without_preview", ToolName: productdata.ToolNameWorkspacePatchApply, ArgumentsSummary: args}); err == nil {
		t.Fatal("apply without preview err = nil")
	}
	apply, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_patch", ToolName: productdata.ToolNameWorkspacePatchApply, ArgumentsSummary: args})
	if err != nil {
		t.Fatal(err)
	}
	content, err = os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "thread\nsecond line\n" {
		t.Fatalf("content = %q", string(content))
	}
	if apply["tool"] != productdata.ToolNameWorkspacePatchApply || apply["operation"] != "patch_apply" || apply["changed"] != true || apply["preview_id"] != preview["preview_id"] {
		t.Fatalf("apply = %+v", apply)
	}
}

func TestWorkspacePatchApplyRejectsStalePreview(t *testing.T) {
	root := createWorkspaceFixture(t)
	path := filepath.Join(root, "src", "notes.txt")
	executor := WorkspaceToolExecutor{Root: root}
	args := map[string]any{"path": "src/notes.txt", "old_text": "needle\n", "new_text": "thread\n"}

	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stale_patch", ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": "src/notes.txt"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stale_patch", ToolName: productdata.ToolNameWorkspacePatchPreview, ArgumentsSummary: args}); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Millisecond)
	if err := os.WriteFile(path, []byte("needle\nchanged elsewhere\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: "run_stale_patch", ToolName: productdata.ToolNameWorkspacePatchApply, ArgumentsSummary: args}); err == nil {
		t.Fatal("stale patch apply err = nil")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "needle\nchanged elsewhere\n" {
		t.Fatalf("stale apply changed file: %q", string(content))
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
		runID := "run_edit_reject_" + fmt.Sprint(len(fmt.Sprint(args)))
		_, _ = executor.Execute(context.Background(), ToolInvocation{RunID: runID, ToolName: productdata.ToolNameWorkspaceRead, ArgumentsSummary: map[string]any{"path": args["path"]}})
		if _, err := executor.Execute(context.Background(), ToolInvocation{RunID: runID, ToolName: productdata.ToolNameWorkspaceEdit, ArgumentsSummary: args}); err == nil {
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

func workspaceTestEntryPaths(entries []map[string]any) []string {
	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if path, ok := entry["path"].(string); ok {
			paths = append(paths, path)
		}
	}
	return paths
}

func workspaceTestHasPath(paths []string, expected string) bool {
	for _, path := range paths {
		if path == expected {
			return true
		}
	}
	return false
}
