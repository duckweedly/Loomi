package runtime

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type ToolApprovalPolicy string

type ToolSafetyClass string

const (
	ToolApprovalAlwaysRequired ToolApprovalPolicy = "always_required"
	ToolApprovalNotRequired    ToolApprovalPolicy = "not_required"

	ToolSafetyNoSideEffectInternal ToolSafetyClass = "no_side_effect_internal"
	ToolSafetyMCPBridge            ToolSafetyClass = "mcp_bridge"
	ToolSafetyWorkspaceReadOnly    ToolSafetyClass = "workspace_read_only"
	ToolSafetyWorkspaceWrite       ToolSafetyClass = "workspace_write"
	ToolSafetyWorkspaceExec        ToolSafetyClass = "workspace_exec"
)

type TodoItem struct {
	Title  string
	Status string
}

type ToolArguments struct {
	Timezone   string
	TodoItems  []TodoItem
	MCPServer  string
	MCPTool    string
	MCPMessage string
	Pattern    string
	Query      string
	Path       string
	Content    string
	OldText    string
	NewText    string
	Command    []string
	Cwd        string
	Timeout    int
	Limit      int
	MaxBytes   int
}

type WorkspaceGrepMatch struct {
	Path    string `json:"path"`
	Line    int    `json:"line"`
	Preview string `json:"preview"`
}

type ToolCatalogEntry struct {
	Name           string `json:"name"`
	Label          string `json:"label"`
	Group          string `json:"group"`
	Capability     string `json:"capability"`
	ApprovalPolicy string `json:"approval_policy"`
	SafetyClass    string `json:"safety_class"`
	RiskLevel      string `json:"risk_level"`
	SideEffect     string `json:"side_effect"`
	Enabled        bool   `json:"enabled"`
	Description    string `json:"description"`
}

type ToolCatalogSnapshot struct {
	Tools []ToolCatalogEntry `json:"tools"`
}

type ToolDefinition struct {
	Name           string
	ApprovalPolicy ToolApprovalPolicy
	SafetyClass    ToolSafetyClass
	WorkspaceRoot  string
}

func ToolCatalog() ToolCatalogSnapshot {
	return ToolCatalogSnapshot{Tools: []ToolCatalogEntry{
		{Name: "runtime.get_current_time", Label: "Current time", Group: "runtime", Capability: "time", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyNoSideEffectInternal), RiskLevel: "low", SideEffect: "none", Enabled: true, Description: "Return the current UTC time."},
		{Name: "runtime.todo_write", Label: "Write todo plan", Group: "runtime", Capability: "plan", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyNoSideEffectInternal), RiskLevel: "low", SideEffect: "none", Enabled: true, Description: "Publish a bounded structured plan for the current run."},
		{Name: "mcp.call_tool", Label: "Call MCP tool", Group: "mcp", Capability: "call_tool", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyMCPBridge), RiskLevel: "medium", SideEffect: "mcp", Enabled: true, Description: "Call one allowlisted local MCP-style tool."},
		{Name: "workspace.glob", Label: "Glob files", Group: "workspace", Capability: "read", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceReadOnly), RiskLevel: "medium", SideEffect: "read", Enabled: true, Description: "List bounded relative file matches inside the workspace."},
		{Name: "workspace.grep", Label: "Search files", Group: "workspace", Capability: "read", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceReadOnly), RiskLevel: "medium", SideEffect: "read", Enabled: true, Description: "Search bounded text matches inside the workspace."},
		{Name: "workspace.read_file", Label: "Read file", Group: "workspace", Capability: "read", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceReadOnly), RiskLevel: "medium", SideEffect: "read", Enabled: true, Description: "Read a bounded UTF-8 preview from one workspace file."},
		{Name: "workspace.write_file", Label: "Write file", Group: "workspace", Capability: "write", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceWrite), RiskLevel: "high", SideEffect: "write", Enabled: true, Description: "Write bounded UTF-8 text to an existing workspace directory."},
		{Name: "workspace.edit", Label: "Edit file", Group: "workspace", Capability: "write", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceWrite), RiskLevel: "high", SideEffect: "write", Enabled: true, Description: "Replace one exact text match inside a workspace file."},
		{Name: "workspace.exec_command", Label: "Exec command", Group: "workspace", Capability: "exec", ApprovalPolicy: string(ToolApprovalAlwaysRequired), SafetyClass: string(ToolSafetyWorkspaceExec), RiskLevel: "high", SideEffect: "process", Enabled: true, Description: "Run one bounded argv command inside the workspace."},
	}}
}

func CurrentTimeToolDefinition() ToolDefinition {
	return ToolDefinition{Name: "runtime.get_current_time", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal}
}

func TodoWriteToolDefinition() ToolDefinition {
	return ToolDefinition{Name: "runtime.todo_write", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal}
}

func MCPCallToolDefinition() ToolDefinition {
	return ToolDefinition{Name: "mcp.call_tool", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyMCPBridge}
}

func WorkspaceGlobToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.glob", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceReadOnly, WorkspaceRoot: root}
}

func WorkspaceGrepToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.grep", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceReadOnly, WorkspaceRoot: root}
}

func WorkspaceReadFileToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.read_file", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceReadOnly, WorkspaceRoot: root}
}

func WorkspaceWriteFileToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.write_file", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceWrite, WorkspaceRoot: root}
}

func WorkspaceEditToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.edit", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceWrite, WorkspaceRoot: root}
}

func WorkspaceExecCommandToolDefinition(root string) ToolDefinition {
	return ToolDefinition{Name: "workspace.exec_command", ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceExec, WorkspaceRoot: root}
}

func WorkspaceToolDefinition(name string, root string) (ToolDefinition, error) {
	switch name {
	case "workspace.glob":
		return WorkspaceGlobToolDefinition(root), nil
	case "workspace.grep":
		return WorkspaceGrepToolDefinition(root), nil
	case "workspace.read_file":
		return WorkspaceReadFileToolDefinition(root), nil
	case "workspace.write_file":
		return WorkspaceWriteFileToolDefinition(root), nil
	case "workspace.edit":
		return WorkspaceEditToolDefinition(root), nil
	case "workspace.exec_command":
		return WorkspaceExecCommandToolDefinition(root), nil
	default:
		return ToolDefinition{}, errors.New("tool is not supported")
	}
}

func (d ToolDefinition) Execute(arguments ToolArguments) (map[string]any, error) {
	switch d.Name {
	case "runtime.get_current_time":
		if arguments.Timezone != "UTC" {
			return nil, errors.New("timezone must be UTC")
		}
		return map[string]any{"iso_time": time.Now().UTC().Format(time.RFC3339Nano), "timezone": "UTC", "source": "runtime"}, nil
	case "runtime.todo_write":
		return executeTodoWrite(arguments)
	case "mcp.call_tool":
		return executeMCPCallTool(arguments)
	case "workspace.glob":
		return d.executeWorkspaceGlob(arguments)
	case "workspace.grep":
		return d.executeWorkspaceGrep(arguments)
	case "workspace.read_file":
		return d.executeWorkspaceReadFile(arguments)
	case "workspace.write_file":
		return d.executeWorkspaceWriteFile(arguments)
	case "workspace.edit":
		return d.executeWorkspaceEdit(arguments)
	case "workspace.exec_command":
		return d.executeWorkspaceExecCommand(arguments)
	default:
		return nil, errors.New("tool is not supported")
	}
}

func (d ToolDefinition) NormalizeArguments(arguments map[string]any) (ToolArguments, error) {
	switch d.Name {
	case "runtime.get_current_time":
		return normalizeCurrentTimeArguments(arguments)
	case "runtime.todo_write":
		return normalizeTodoWriteArguments(arguments)
	case "mcp.call_tool":
		return normalizeMCPCallToolArguments(arguments)
	case "workspace.glob":
		return d.normalizeWorkspaceGlobArguments(arguments)
	case "workspace.grep":
		return d.normalizeWorkspaceGrepArguments(arguments)
	case "workspace.read_file":
		return d.normalizeWorkspaceReadFileArguments(arguments)
	case "workspace.write_file":
		return d.normalizeWorkspaceWriteFileArguments(arguments)
	case "workspace.edit":
		return d.normalizeWorkspaceEditArguments(arguments)
	case "workspace.exec_command":
		return d.normalizeWorkspaceExecCommandArguments(arguments)
	default:
		return ToolArguments{}, errors.New("tool is not supported")
	}
}

func normalizeCurrentTimeArguments(arguments map[string]any) (ToolArguments, error) {
	for key := range arguments {
		if key != "timezone" {
			return ToolArguments{}, errors.New("tool argument is not supported")
		}
	}
	value, ok := arguments["timezone"]
	if !ok || value == nil {
		return ToolArguments{Timezone: "UTC"}, nil
	}
	timezone, ok := value.(string)
	if !ok || timezone != "UTC" {
		return ToolArguments{}, errors.New("timezone must be UTC")
	}
	return ToolArguments{Timezone: "UTC"}, nil
}

func normalizeTodoWriteArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "items"); err != nil {
		return ToolArguments{}, err
	}
	rawItems, ok := arguments["items"].([]any)
	if !ok || len(rawItems) == 0 || len(rawItems) > 20 {
		return ToolArguments{}, errors.New("todo items are invalid")
	}
	items := make([]TodoItem, 0, len(rawItems))
	for _, raw := range rawItems {
		item, ok := raw.(map[string]any)
		if !ok {
			return ToolArguments{}, errors.New("todo item is invalid")
		}
		title, ok := item["title"].(string)
		title = strings.TrimSpace(title)
		if !ok || title == "" || len(title) > 160 {
			return ToolArguments{}, errors.New("todo item title is invalid")
		}
		status := "pending"
		if rawStatus, ok := item["status"]; ok && rawStatus != nil {
			statusValue, ok := rawStatus.(string)
			if !ok {
				return ToolArguments{}, errors.New("todo item status is invalid")
			}
			status = strings.TrimSpace(statusValue)
		}
		if !validTodoStatus(status) {
			return ToolArguments{}, errors.New("todo item status is invalid")
		}
		items = append(items, TodoItem{Title: title, Status: status})
	}
	return ToolArguments{TodoItems: items}, nil
}

func normalizeMCPCallToolArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "server", "tool", "arguments"); err != nil {
		return ToolArguments{}, err
	}
	server, ok := arguments["server"].(string)
	if !ok || strings.TrimSpace(server) != "local" {
		return ToolArguments{}, errors.New("mcp server is invalid")
	}
	tool, ok := arguments["tool"].(string)
	if !ok || strings.TrimSpace(tool) != "echo" {
		return ToolArguments{}, errors.New("mcp tool is invalid")
	}
	rawArguments, ok := arguments["arguments"].(map[string]any)
	if !ok {
		return ToolArguments{}, errors.New("mcp arguments are invalid")
	}
	if err := requireOnlyKeys(rawArguments, "message"); err != nil {
		return ToolArguments{}, err
	}
	message, ok := rawArguments["message"].(string)
	message = strings.TrimSpace(message)
	if !ok || message == "" || len(message) > 500 || secretLookingText(message) {
		return ToolArguments{}, errors.New("mcp message is invalid")
	}
	return ToolArguments{MCPServer: "local", MCPTool: "echo", MCPMessage: message}, nil
}

func (d ToolDefinition) normalizeWorkspaceGlobArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "pattern", "limit"); err != nil {
		return ToolArguments{}, err
	}
	pattern, ok := arguments["pattern"].(string)
	if !ok || strings.TrimSpace(pattern) == "" {
		return ToolArguments{}, errors.New("pattern is required")
	}
	pattern = strings.TrimSpace(pattern)
	if _, err := d.resolveWorkspacePath(pattern); err != nil {
		return ToolArguments{}, err
	}
	limit, err := optionalPositiveInt(arguments["limit"], 50, 200)
	if err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Pattern: pattern, Limit: limit}, nil
}

func (d ToolDefinition) normalizeWorkspaceGrepArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "query", "path", "limit"); err != nil {
		return ToolArguments{}, err
	}
	query, ok := arguments["query"].(string)
	if !ok || strings.TrimSpace(query) == "" {
		return ToolArguments{}, errors.New("query is required")
	}
	pathValue := "."
	if value, ok := arguments["path"]; ok && value != nil {
		pathString, ok := value.(string)
		if !ok || strings.TrimSpace(pathString) == "" {
			return ToolArguments{}, errors.New("path is invalid")
		}
		pathValue = strings.TrimSpace(pathString)
	}
	if _, err := d.resolveWorkspacePath(pathValue); err != nil {
		return ToolArguments{}, err
	}
	limit, err := optionalPositiveInt(arguments["limit"], 50, 200)
	if err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Query: strings.TrimSpace(query), Path: pathValue, Limit: limit}, nil
}

func (d ToolDefinition) normalizeWorkspaceReadFileArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "path", "max_bytes"); err != nil {
		return ToolArguments{}, err
	}
	pathValue, ok := arguments["path"].(string)
	if !ok || strings.TrimSpace(pathValue) == "" {
		return ToolArguments{}, errors.New("path is required")
	}
	pathValue = strings.TrimSpace(pathValue)
	if _, err := d.resolveWorkspacePath(pathValue); err != nil {
		return ToolArguments{}, err
	}
	maxBytes, err := optionalPositiveInt(arguments["max_bytes"], 4096, 16384)
	if err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Path: pathValue, MaxBytes: maxBytes}, nil
}

func (d ToolDefinition) normalizeWorkspaceWriteFileArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "path", "content"); err != nil {
		return ToolArguments{}, err
	}
	pathValue, ok := arguments["path"].(string)
	if !ok || strings.TrimSpace(pathValue) == "" {
		return ToolArguments{}, errors.New("path is required")
	}
	content, ok := arguments["content"].(string)
	if !ok {
		return ToolArguments{}, errors.New("content is required")
	}
	if len(content) > 65536 {
		return ToolArguments{}, errors.New("content is too large")
	}
	pathValue = strings.TrimSpace(pathValue)
	if _, err := d.resolveWorkspaceMutationPath(pathValue, true); err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Path: pathValue, Content: content}, nil
}

func (d ToolDefinition) normalizeWorkspaceEditArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "path", "old_text", "new_text"); err != nil {
		return ToolArguments{}, err
	}
	pathValue, ok := arguments["path"].(string)
	if !ok || strings.TrimSpace(pathValue) == "" {
		return ToolArguments{}, errors.New("path is required")
	}
	oldText, ok := arguments["old_text"].(string)
	if !ok || oldText == "" {
		return ToolArguments{}, errors.New("old_text is required")
	}
	newText, ok := arguments["new_text"].(string)
	if !ok {
		return ToolArguments{}, errors.New("new_text is required")
	}
	if len(oldText) > 65536 || len(newText) > 65536 {
		return ToolArguments{}, errors.New("edit content is too large")
	}
	pathValue = strings.TrimSpace(pathValue)
	if _, err := d.resolveWorkspaceMutationPath(pathValue, false); err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Path: pathValue, OldText: oldText, NewText: newText}, nil
}

func (d ToolDefinition) normalizeWorkspaceExecCommandArguments(arguments map[string]any) (ToolArguments, error) {
	if err := requireOnlyKeys(arguments, "command", "cwd", "timeout_seconds"); err != nil {
		return ToolArguments{}, err
	}
	command, err := normalizeCommandArgv(arguments["command"])
	if err != nil {
		return ToolArguments{}, err
	}
	if dangerousCommand(command) {
		return ToolArguments{}, errors.New("command is not allowed")
	}
	cwd := "."
	if value, ok := arguments["cwd"]; ok && value != nil {
		cwdValue, ok := value.(string)
		if !ok || strings.TrimSpace(cwdValue) == "" {
			return ToolArguments{}, errors.New("cwd is invalid")
		}
		cwd = strings.TrimSpace(cwdValue)
	}
	if _, err := d.resolveWorkspacePath(cwd); err != nil {
		return ToolArguments{}, err
	}
	timeout, err := optionalPositiveInt(arguments["timeout_seconds"], 30, 120)
	if err != nil {
		return ToolArguments{}, err
	}
	return ToolArguments{Command: command, Cwd: cwd, Timeout: timeout}, nil
}

func (d ToolDefinition) executeWorkspaceGlob(arguments ToolArguments) (map[string]any, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	matches := []string{}
	truncated := false
	err = filepath.WalkDir(root, func(fullPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			if sensitiveWorkspacePath(entry.Name()) && fullPath != root {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, fullPath)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if sensitiveWorkspacePath(rel) || !workspaceGlobMatches(arguments.Pattern, rel) {
			return nil
		}
		if len(matches) >= arguments.Limit {
			truncated = true
			return filepath.SkipAll
		}
		matches = append(matches, rel)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	return map[string]any{"matches": matches, "match_count": len(matches), "truncated": truncated}, nil
}

func (d ToolDefinition) executeWorkspaceGrep(arguments ToolArguments) (map[string]any, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	start, err := d.resolveWorkspacePath(arguments.Path)
	if err != nil {
		return nil, err
	}
	matches := []WorkspaceGrepMatch{}
	truncated := false
	visit := func(fullPath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if entry.IsDir() {
			if sensitiveWorkspacePath(entry.Name()) && fullPath != root {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, fullPath)
		if err != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if sensitiveWorkspacePath(rel) {
			return nil
		}
		fileMatches, err := grepFile(fullPath, rel, arguments.Query, arguments.Limit-len(matches))
		if err != nil {
			return nil
		}
		matches = append(matches, fileMatches...)
		if len(matches) >= arguments.Limit {
			truncated = true
			return filepath.SkipAll
		}
		return nil
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		err = filepath.WalkDir(start, visit)
	} else {
		err = visit(start, fileInfoDirEntry{info: info}, nil)
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"matches": matches, "match_count": len(matches), "truncated": truncated}, nil
}

func (d ToolDefinition) executeWorkspaceReadFile(arguments ToolArguments) (map[string]any, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	fullPath, err := d.resolveWorkspacePath(arguments.Path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("path is a directory")
	}
	readLimit := arguments.MaxBytes + 1
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	if len(data) > readLimit {
		data = data[:readLimit]
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, errors.New("binary files are not supported")
	}
	truncated := len(data) > arguments.MaxBytes
	if truncated {
		data = data[:arguments.MaxBytes]
	}
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return nil, err
	}
	return map[string]any{"path": filepath.ToSlash(rel), "size_bytes": info.Size(), "preview": string(data), "truncated": truncated}, nil
}

func (d ToolDefinition) executeWorkspaceWriteFile(arguments ToolArguments) (map[string]any, error) {
	fullPath, err := d.resolveWorkspaceMutationPath(arguments.Path, true)
	if err != nil {
		return nil, err
	}
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	created := false
	if _, err := os.Stat(fullPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		created = true
	} else if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return nil, errors.New("path is a directory")
	}
	if err := os.WriteFile(fullPath, []byte(arguments.Content), 0o600); err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return nil, err
	}
	return map[string]any{"path": filepath.ToSlash(rel), "bytes_written": len([]byte(arguments.Content)), "created": created, "truncated": false}, nil
}

func (d ToolDefinition) executeWorkspaceEdit(arguments ToolArguments) (map[string]any, error) {
	fullPath, err := d.resolveWorkspaceMutationPath(arguments.Path, false)
	if err != nil {
		return nil, err
	}
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("path is a directory")
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(data, 0) >= 0 || !utf8.Valid(data) {
		return nil, errors.New("file is not valid UTF-8 text")
	}
	content := string(data)
	count := strings.Count(content, arguments.OldText)
	if count != 1 {
		return nil, errors.New("old_text must match exactly once")
	}
	next := strings.Replace(content, arguments.OldText, arguments.NewText, 1)
	if err := os.WriteFile(fullPath, []byte(next), info.Mode().Perm()); err != nil {
		return nil, err
	}
	rel, err := filepath.Rel(root, fullPath)
	if err != nil {
		return nil, err
	}
	return map[string]any{"path": filepath.ToSlash(rel), "replacements": 1, "bytes_before": len(data), "bytes_after": len([]byte(next))}, nil
}

func (d ToolDefinition) executeWorkspaceExecCommand(arguments ToolArguments) (map[string]any, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return nil, err
	}
	cwd, err := d.resolveWorkspacePath(arguments.Cwd)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(cwd)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("cwd must be a directory")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(arguments.Timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, arguments.Command[0], arguments.Command[1:]...)
	cmd.Dir = cwd
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	timedOut := ctx.Err() == context.DeadlineExceeded
	exitCode := 0
	if err != nil {
		exitCode = -1
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && !timedOut {
			exitCode = exitErr.ExitCode()
		}
	}
	rel, err := filepath.Rel(root, cwd)
	if err != nil {
		return nil, err
	}
	stdoutPreview, stdoutTruncated := boundedOutput(stdout.String(), 4096)
	stderrPreview, stderrTruncated := boundedOutput(stderr.String(), 4096)
	return map[string]any{"cwd": filepath.ToSlash(rel), "exit_code": exitCode, "stdout": stdoutPreview, "stderr": stderrPreview, "timed_out": timedOut, "stdout_truncated": stdoutTruncated, "stderr_truncated": stderrTruncated}, nil
}

func executeTodoWrite(arguments ToolArguments) (map[string]any, error) {
	if len(arguments.TodoItems) == 0 || len(arguments.TodoItems) > 20 {
		return nil, errors.New("todo items are invalid")
	}
	items := make([]map[string]string, 0, len(arguments.TodoItems))
	pendingCount := 0
	inProgressCount := 0
	completedCount := 0
	for _, item := range arguments.TodoItems {
		if item.Title == "" || !validTodoStatus(item.Status) {
			return nil, errors.New("todo item is invalid")
		}
		switch item.Status {
		case "pending":
			pendingCount++
		case "in_progress":
			inProgressCount++
		case "completed":
			completedCount++
		}
		items = append(items, map[string]string{"title": item.Title, "status": item.Status})
	}
	return map[string]any{"total": len(items), "pending_count": pendingCount, "in_progress_count": inProgressCount, "completed_count": completedCount, "items": items}, nil
}

func executeMCPCallTool(arguments ToolArguments) (map[string]any, error) {
	if arguments.MCPServer != "local" || arguments.MCPTool != "echo" || arguments.MCPMessage == "" {
		return nil, errors.New("mcp call is invalid")
	}
	return map[string]any{"server": "local", "tool": "echo", "message": arguments.MCPMessage, "side_effect": "none"}, nil
}

func validTodoStatus(status string) bool {
	return status == "pending" || status == "in_progress" || status == "completed"
}

func secretLookingText(value string) bool {
	lower := strings.ToLower(value)
	for _, marker := range []string{"postgres://", "postgresql://", "password=", "api_key", "bearer ", "secret", "token", "sk-"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func requireOnlyKeys(arguments map[string]any, allowed ...string) error {
	allowedSet := map[string]struct{}{}
	for _, key := range allowed {
		allowedSet[key] = struct{}{}
	}
	for key := range arguments {
		if _, ok := allowedSet[key]; !ok {
			return errors.New("tool argument is not supported")
		}
	}
	return nil
}

func normalizeCommandArgv(value any) ([]string, error) {
	items, ok := value.([]any)
	if !ok {
		if typed, typedOK := value.([]string); typedOK {
			items = make([]any, len(typed))
			for index, item := range typed {
				items[index] = item
			}
		} else {
			return nil, errors.New("command must be an argv array")
		}
	}
	if len(items) == 0 || len(items) > 32 {
		return nil, errors.New("command argv length is invalid")
	}
	command := make([]string, 0, len(items))
	for _, item := range items {
		part, ok := item.(string)
		if !ok || strings.TrimSpace(part) == "" || len(part) > 65536 {
			return nil, errors.New("command argv item is invalid")
		}
		command = append(command, strings.TrimSpace(part))
	}
	return command, nil
}

func dangerousCommand(command []string) bool {
	if len(command) == 0 {
		return true
	}
	base := strings.ToLower(filepath.Base(command[0]))
	switch base {
	case "sh", "bash", "zsh", "fish", "rm", "dd", "mkfs", "chmod", "chown", "kill", "killall", "shutdown", "reboot", "sudo", "su":
		return true
	case "git":
		if len(command) > 1 {
			subcommand := strings.ToLower(command[1])
			return subcommand == "push" || subcommand == "reset" || subcommand == "clean" || subcommand == "checkout"
		}
	}
	return false
}

func boundedOutput(value string, limit int) (string, bool) {
	if len(value) <= limit {
		return value, false
	}
	return value[:limit], true
}

func optionalPositiveInt(value any, defaultValue int, maxValue int) (int, error) {
	if value == nil {
		return defaultValue, nil
	}
	var number int
	switch typed := value.(type) {
	case int:
		number = typed
	case int64:
		number = int(typed)
	case float64:
		number = int(typed)
	default:
		return 0, errors.New("limit must be a number")
	}
	if number <= 0 || number > maxValue {
		return 0, errors.New("limit is out of range")
	}
	return number, nil
}

func (d ToolDefinition) workspaceRoot() (string, error) {
	root := strings.TrimSpace(d.WorkspaceRoot)
	if root == "" {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		root = wd
	}
	return filepath.Abs(root)
}

func (d ToolDefinition) resolveWorkspacePath(value string) (string, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return "", err
	}
	normalized := strings.TrimSpace(strings.ReplaceAll(value, "\\", "/"))
	if normalized == "" || strings.HasPrefix(normalized, "/") || strings.HasPrefix(normalized, "..") || strings.Contains(normalized, "../") || sensitiveWorkspacePath(normalized) {
		return "", errors.New("workspace path is invalid")
	}
	candidate := filepath.Clean(filepath.Join(root, filepath.FromSlash(normalized)))
	relative, err := filepath.Rel(root, candidate)
	if err != nil || strings.HasPrefix(relative, "..") || filepath.IsAbs(relative) {
		return "", errors.New("workspace path escapes root")
	}
	return candidate, nil
}

func (d ToolDefinition) resolveWorkspaceMutationPath(value string, allowMissingFile bool) (string, error) {
	root, err := d.workspaceRoot()
	if err != nil {
		return "", err
	}
	candidate, err := d.resolveWorkspacePath(value)
	if err != nil {
		return "", err
	}
	parent := filepath.Dir(candidate)
	parentInfo, err := os.Stat(parent)
	if err != nil {
		return "", errors.New("parent directory is required")
	}
	if !parentInfo.IsDir() {
		return "", errors.New("parent path is not a directory")
	}
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", err
	}
	realParent, err := filepath.EvalSymlinks(parent)
	if err != nil {
		return "", err
	}
	if !pathWithinRoot(realRoot, realParent) {
		return "", errors.New("workspace path escapes root")
	}
	info, err := os.Lstat(candidate)
	if err != nil {
		if allowMissingFile && errors.Is(err, os.ErrNotExist) {
			return candidate, nil
		}
		return "", err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		realTarget, err := filepath.EvalSymlinks(candidate)
		if err != nil {
			return "", err
		}
		if !pathWithinRoot(realRoot, realTarget) {
			return "", errors.New("workspace path escapes root")
		}
	}
	return candidate, nil
}

func pathWithinRoot(root string, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && !filepath.IsAbs(relative)
}

func sensitiveWorkspacePath(value string) bool {
	value = strings.ToLower(strings.ReplaceAll(value, "\\", "/"))
	for _, segment := range strings.Split(value, "/") {
		if segment == "" {
			continue
		}
		if segment == ".ssh" || segment == ".aws" || segment == "secrets" || segment == "credentials" || strings.HasPrefix(segment, ".env") || strings.HasPrefix(segment, "id_rsa") || strings.HasPrefix(segment, "id_ed25519") || strings.HasSuffix(segment, ".pem") {
			return true
		}
	}
	return false
}

func workspaceGlobMatches(pattern string, rel string) bool {
	pattern = path.Clean(strings.ReplaceAll(pattern, "\\", "/"))
	rel = path.Clean(rel)
	if ok, _ := path.Match(pattern, rel); ok {
		return true
	}
	if strings.HasPrefix(pattern, "**/") {
		suffix := strings.TrimPrefix(pattern, "**/")
		if ok, _ := path.Match(suffix, path.Base(rel)); ok {
			return true
		}
		if strings.Contains(suffix, "*") && strings.HasSuffix(rel, strings.TrimPrefix(suffix, "*")) {
			return true
		}
	}
	return false
}

func grepFile(fullPath string, rel string, query string, limit int) ([]WorkspaceGrepMatch, error) {
	if limit <= 0 {
		return nil, nil
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, err
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return nil, nil
	}
	lines := strings.Split(string(data), "\n")
	matches := []WorkspaceGrepMatch{}
	for index, line := range lines {
		if strings.Contains(line, query) {
			matches = append(matches, WorkspaceGrepMatch{Path: rel, Line: index + 1, Preview: boundedPreview(line, 240)})
			if len(matches) >= limit {
				break
			}
		}
	}
	return matches, nil
}

func boundedPreview(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}

type fileInfoDirEntry struct {
	info os.FileInfo
}

func (e fileInfoDirEntry) Name() string               { return e.info.Name() }
func (e fileInfoDirEntry) IsDir() bool                { return e.info.IsDir() }
func (e fileInfoDirEntry) Type() fs.FileMode          { return e.info.Mode().Type() }
func (e fileInfoDirEntry) Info() (os.FileInfo, error) { return e.info, nil }
