package runtime

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultSandboxExecTimeoutMS = 5_000
	maxSandboxExecTimeoutMS     = 30_000
	defaultSandboxOutputBytes   = 16 * 1024
	maxSandboxOutputBytes       = 64 * 1024
)

type SandboxToolExecutor struct {
	Root string
}

func SandboxToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameSandboxExecCommand, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetySandboxCommand, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

type boundedOutput struct {
	buf   bytes.Buffer
	limit int
	total int
}

func (e SandboxToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if invocation.ToolName != productdata.ToolNameSandboxExecCommand {
		return nil, errors.New("sandbox tool is not supported")
	}
	scope, err := newWorkspaceScope(e.Root)
	if err != nil {
		return nil, err
	}
	return scope.execCommand(ctx, invocation.ArgumentsSummary)
}

func (s workspaceScope) execCommand(ctx context.Context, args map[string]any) (map[string]any, error) {
	for key := range args {
		if key != "argv" && key != "cwd" && key != "timeout_ms" && key != "max_output_bytes" {
			return nil, errors.New("sandbox exec argument is not supported")
		}
	}
	argv, err := sandboxArgv(args["argv"])
	if err != nil {
		return nil, err
	}
	if !allowedReadOnlyCommand(argv) {
		return nil, errors.New("sandbox exec command is not allowed")
	}
	cwdArg := strings.TrimSpace(stringArg(args, "cwd", "."))
	if cwdArg == "" {
		cwdArg = "."
	}
	cwdPath, cwdRel, err := s.resolveDir(cwdArg)
	if err != nil {
		return nil, err
	}
	timeoutMS := boundedInt(args, "timeout_ms", defaultSandboxExecTimeoutMS, maxSandboxExecTimeoutMS)
	outputLimit := boundedInt(args, "max_output_bytes", defaultSandboxOutputBytes, maxSandboxOutputBytes)
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	defer cancel()

	command := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	command.Dir = cwdPath
	var stdout boundedOutput
	var stderr boundedOutput
	stdout.limit = outputLimit
	stderr.limit = outputLimit
	command.Stdout = &stdout
	command.Stderr = &stderr

	err = command.Run()
	timedOut := runCtx.Err() == context.DeadlineExceeded
	exitCode := 0
	if command.ProcessState != nil {
		exitCode = command.ProcessState.ExitCode()
	}
	if err != nil && !timedOut {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, errors.New("sandbox command could not be executed")
		}
	}
	return map[string]any{
		"tool":              productdata.ToolNameSandboxExecCommand,
		"scope":             "bounded_read_only_command",
		"operation":         "exec_command",
		"argv":              append([]string(nil), argv...),
		"cwd":               cwdRel,
		"exit_code":         exitCode,
		"timed_out":         timedOut,
		"stdout":            sandboxOutputPreview(stdout.String(), s.root),
		"stderr":            sandboxOutputPreview(stderr.String(), s.root),
		"stdout_bytes":      stdout.total,
		"stderr_bytes":      stderr.total,
		"stdout_truncated":  stdout.Truncated(),
		"stderr_truncated":  stderr.Truncated(),
		"redaction_applied": false,
	}, nil
}

func sandboxArgv(value any) ([]string, error) {
	raw := []string(nil)
	switch typed := value.(type) {
	case []string:
		raw = typed
	case []any:
		raw = make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok {
				return nil, errors.New("sandbox exec argv must contain strings")
			}
			raw = append(raw, text)
		}
	default:
		return nil, errors.New("sandbox exec argv is required")
	}
	if len(raw) == 0 {
		return nil, errors.New("sandbox exec argv is required")
	}
	argv := make([]string, 0, len(raw))
	for _, item := range raw {
		text := strings.TrimSpace(item)
		if text == "" {
			return nil, errors.New("sandbox exec argv cannot contain empty values")
		}
		argv = append(argv, text)
	}
	if strings.ContainsAny(argv[0], `/\`) || filepath.Base(argv[0]) != argv[0] {
		return nil, errors.New("sandbox exec command must be argv-form with a bare command name")
	}
	return argv, nil
}

func allowedReadOnlyCommand(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	command := strings.ToLower(filepath.Base(strings.TrimSpace(argv[0])))
	switch command {
	case "pwd":
		return len(argv) == 1
	case "ls":
		return len(argv) == 1 || (len(argv) == 2 && argv[1] == ".")
	case "git":
		return gitReadOnlyArgsAllowed(argv[1:])
	default:
		return false
	}
}

func gitReadOnlyArgsAllowed(args []string) bool {
	if len(args) != 1 {
		return false
	}
	switch args[0] {
	case "status":
		return true
	default:
		return false
	}
}

func sandboxOutputPreview(content string, root string) string {
	content = strings.ToValidUTF8(content, "")
	content = strings.ReplaceAll(content, root, ".")
	if strings.ContainsRune(content, 0) || !utf8.ValidString(content) {
		return ""
	}
	return content
}

func (w *boundedOutput) Write(p []byte) (int, error) {
	written := len(p)
	w.total += len(p)
	remaining := w.limit - w.buf.Len()
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = w.buf.Write(p)
	}
	return written, nil
}

func (w *boundedOutput) String() string {
	return w.buf.String()
}

func (w *boundedOutput) Truncated() bool {
	return w.total > w.limit
}
