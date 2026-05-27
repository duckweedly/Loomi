package runtime

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultSandboxExecTimeoutMS = 5_000
	maxSandboxExecTimeoutMS     = 30_000
	defaultSandboxProcessMS     = 60_000
	maxSandboxProcessMS         = 120_000
	defaultSandboxOutputBytes   = 16 * 1024
	maxSandboxOutputBytes       = 64 * 1024
	defaultSandboxMaxLifetime   = 30 * time.Minute
	defaultSandboxIdleTimeout   = 15 * time.Minute
)

type SandboxProcessStatus string

const (
	SandboxProcessStatusRunning    SandboxProcessStatus = "running"
	SandboxProcessStatusExited     SandboxProcessStatus = "exited"
	SandboxProcessStatusTerminated SandboxProcessStatus = "terminated"
	SandboxProcessStatusFailed     SandboxProcessStatus = "failed"
	SandboxProcessStatusExpired    SandboxProcessStatus = "expired"
)

type SandboxToolExecutor struct {
	Root  string
	Store *SandboxProcessStore
}

func SandboxToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameSandboxExecCommand, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetySandboxCommand, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameSandboxStartProcess, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetySandboxCommand, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameSandboxContinueProcess, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetySandboxCommand, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameSandboxTerminateProcess, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetySandboxCommand, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

type boundedOutput struct {
	mu    sync.Mutex
	buf   []byte
	limit int
	total int
	start int
}

type SandboxProcessRecord struct {
	RunID        string
	ProcessID    string
	Argv         []string
	CwdAlias     string
	Status       SandboxProcessStatus
	Cursor       int
	StartedAt    time.Time
	UpdatedAt    time.Time
	EndedAt      *time.Time
	ExitCode     *int
	StdoutTail   string
	StdoutCursor int
	StderrTail   string
	StderrCursor int
	StdinOpen    bool
	InputSeq     int
	TimedOut     bool
	Summary      string
	OutputLimit  int
}

type SandboxProcessRepository interface {
	SaveSandboxProcess(context.Context, SandboxProcessRecord) error
	ListSandboxProcesses(context.Context) ([]SandboxProcessRecord, error)
}

type MemorySandboxProcessRepository struct {
	mu      sync.Mutex
	records map[string]SandboxProcessRecord
}

type SandboxProcessStoreOptions struct {
	Now         func() time.Time
	MaxLifetime time.Duration
	IdleTimeout time.Duration
}

type SandboxProcessStore struct {
	mu        sync.Mutex
	processes map[string]*sandboxProcess
	repo      SandboxProcessRepository
	now       func() time.Time
	maxLife   time.Duration
	idle      time.Duration
}

type sandboxProcess struct {
	mu         sync.Mutex
	runID      string
	processID  string
	argv       []string
	cwd        string
	status     SandboxProcessStatus
	cancel     context.CancelFunc
	command    *exec.Cmd
	stdin      io.WriteCloser
	stdinOpen  bool
	inputSeq   int
	stdout     *boundedOutput
	stderr     *boundedOutput
	done       chan struct{}
	err        error
	exitCode   int
	timedOut   bool
	terminated bool
	startedAt  time.Time
	updatedAt  time.Time
	endedAt    *time.Time
	summary    string
	restored   bool
}

func NewSandboxProcessStore() *SandboxProcessStore {
	return NewSandboxProcessStoreWithRepository(NewMemorySandboxProcessRepository(), SandboxProcessStoreOptions{})
}

var defaultSandboxProcessStore = NewSandboxProcessStore()

func NewMemorySandboxProcessRepository() *MemorySandboxProcessRepository {
	return &MemorySandboxProcessRepository{records: map[string]SandboxProcessRecord{}}
}

func NewSandboxProcessStoreWithRepository(repo SandboxProcessRepository, options SandboxProcessStoreOptions) *SandboxProcessStore {
	if repo == nil {
		repo = NewMemorySandboxProcessRepository()
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}
	maxLife := options.MaxLifetime
	if maxLife <= 0 {
		maxLife = defaultSandboxMaxLifetime
	}
	idle := options.IdleTimeout
	if idle <= 0 {
		idle = defaultSandboxIdleTimeout
	}
	store := &SandboxProcessStore{processes: map[string]*sandboxProcess{}, repo: repo, now: now, maxLife: maxLife, idle: idle}
	store.restore()
	return store
}

func (r *MemorySandboxProcessRepository) SaveSandboxProcess(_ context.Context, record SandboxProcessRecord) error {
	if r == nil {
		return errors.New("sandbox process repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.records == nil {
		r.records = map[string]SandboxProcessRecord{}
	}
	next := record
	next.Argv = append([]string(nil), record.Argv...)
	r.records[record.ProcessID] = next
	return nil
}

func (r *MemorySandboxProcessRepository) ListSandboxProcesses(_ context.Context) ([]SandboxProcessRecord, error) {
	if r == nil {
		return nil, errors.New("sandbox process repository is unavailable")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	records := make([]SandboxProcessRecord, 0, len(r.records))
	for _, record := range r.records {
		next := record
		next.Argv = append([]string(nil), record.Argv...)
		records = append(records, next)
	}
	return records, nil
}

func (e SandboxToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	root := e.Root
	if root == "" {
		root = invocation.WorkspaceRoot
	}
	scope, err := newWorkspaceScope(root)
	if err != nil {
		return nil, err
	}
	switch invocation.ToolName {
	case productdata.ToolNameSandboxExecCommand:
		return scope.execCommand(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameSandboxStartProcess:
		return e.store().start(ctx, scope, invocation)
	case productdata.ToolNameSandboxContinueProcess:
		return e.store().continueProcess(scope, invocation)
	case productdata.ToolNameSandboxTerminateProcess:
		return e.store().terminateProcess(scope, invocation)
	default:
		return nil, errors.New("sandbox tool is not supported")
	}
}

func (s *SandboxProcessStore) restore() {
	if s == nil || s.repo == nil {
		return
	}
	records, err := s.repo.ListSandboxProcesses(context.Background())
	if err != nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.processes == nil {
		s.processes = map[string]*sandboxProcess{}
	}
	for _, record := range records {
		if record.ProcessID == "" || record.RunID == "" {
			continue
		}
		process := processFromRecord(record)
		if s.processExpired(process, s.nowTime()) {
			process.status = SandboxProcessStatusExpired
			process.stdinOpen = false
			process.summary = "expired by sandbox process TTL"
			now := s.nowTime()
			process.endedAt = &now
			process.updatedAt = now
			_ = s.saveProcessLocked(process)
		}
		s.processes[record.ProcessID] = process
	}
}

func processFromRecord(record SandboxProcessRecord) *sandboxProcess {
	limit := record.OutputLimit
	if limit <= 0 {
		limit = defaultSandboxOutputBytes
	}
	process := &sandboxProcess{
		runID:     record.RunID,
		processID: record.ProcessID,
		argv:      append([]string(nil), record.Argv...),
		cwd:       record.CwdAlias,
		status:    record.Status,
		stdinOpen: record.StdinOpen,
		inputSeq:  record.InputSeq,
		stdout:    outputFromRecord(record.StdoutTail, record.StdoutCursor, limit),
		stderr:    outputFromRecord(record.StderrTail, record.StderrCursor, limit),
		done:      closedProcessDone(),
		exitCode:  -1,
		timedOut:  record.TimedOut,
		startedAt: record.StartedAt,
		updatedAt: record.UpdatedAt,
		endedAt:   record.EndedAt,
		summary:   record.Summary,
		restored:  true,
	}
	if record.ExitCode != nil {
		process.exitCode = *record.ExitCode
	}
	if process.status == "" {
		process.status = SandboxProcessStatusRunning
	}
	if process.status == SandboxProcessStatusRunning {
		process.done = make(chan struct{})
	}
	return process
}

func outputFromRecord(tail string, cursor int, limit int) *boundedOutput {
	output := &boundedOutput{limit: limit}
	output.setTail([]byte(tail), cursor)
	return output
}

func closedProcessDone() chan struct{} {
	done := make(chan struct{})
	close(done)
	return done
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
	configureSandboxCommand(command)
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
	stdoutPreview, stdoutRedacted := sandboxSafeOutputPreview(stdout.String(), s.root)
	stderrPreview, stderrRedacted := sandboxSafeOutputPreview(stderr.String(), s.root)
	return map[string]any{
		"tool":              productdata.ToolNameSandboxExecCommand,
		"scope":             "bounded_command",
		"operation":         "exec_command",
		"argv":              append([]string(nil), argv...),
		"cwd":               cwdRel,
		"exit_code":         exitCode,
		"timed_out":         timedOut,
		"stdout":            stdoutPreview,
		"stderr":            stderrPreview,
		"stdout_bytes":      stdout.total,
		"stderr_bytes":      stderr.total,
		"stdout_truncated":  stdout.Truncated(),
		"stderr_truncated":  stderr.Truncated(),
		"redaction_applied": stdoutRedacted || stderrRedacted,
	}, nil
}

func (s *SandboxProcessStore) start(ctx context.Context, scope workspaceScope, invocation ToolInvocation) (map[string]any, error) {
	if s == nil {
		return nil, errors.New("sandbox process store is unavailable")
	}
	if productdata.IsRunTerminal(invocation.RunStatus) {
		return nil, errors.New("sandbox process cannot start for a terminal run")
	}
	args := invocation.ArgumentsSummary
	for key := range args {
		if key != "argv" && key != "cwd" && key != "timeout_ms" && key != "max_output_bytes" && key != "stdin" {
			return nil, errors.New("sandbox process argument is not supported")
		}
	}
	argv, err := sandboxArgv(args["argv"])
	if err != nil {
		return nil, err
	}
	stdinEnabled := boolArg(args, "stdin", false)
	if !allowedSandboxProcessCommand(argv, stdinEnabled) {
		return nil, errors.New("sandbox process command is not allowed")
	}
	cwdArg := strings.TrimSpace(stringArg(args, "cwd", "."))
	if cwdArg == "" {
		cwdArg = "."
	}
	cwdPath, cwdRel, err := scope.resolveDir(cwdArg)
	if err != nil {
		return nil, err
	}
	timeoutMS := boundedInt(args, "timeout_ms", defaultSandboxProcessMS, maxSandboxProcessMS)
	outputLimit := boundedInt(args, "max_output_bytes", defaultSandboxOutputBytes, maxSandboxOutputBytes)
	runCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMS)*time.Millisecond)
	command := exec.CommandContext(runCtx, argv[0], argv[1:]...)
	command.Dir = cwdPath
	configureSandboxCommand(command)
	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		cancel()
		return nil, errors.New("sandbox process stdout could not be created")
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		cancel()
		return nil, errors.New("sandbox process stderr could not be created")
	}
	var stdinPipe io.WriteCloser
	if stdinEnabled {
		stdinPipe, err = command.StdinPipe()
		if err != nil {
			cancel()
			return nil, errors.New("sandbox process stdin could not be created")
		}
	}
	processID := sandboxProcessID(invocation.RunID, invocation.ToolCallID, argv)
	now := s.nowTime()
	process := &sandboxProcess{
		runID:     invocation.RunID,
		processID: processID,
		argv:      append([]string(nil), argv...),
		cwd:       cwdRel,
		status:    SandboxProcessStatusRunning,
		cancel:    cancel,
		command:   command,
		stdin:     stdinPipe,
		stdinOpen: stdinEnabled,
		stdout:    &boundedOutput{limit: outputLimit},
		stderr:    &boundedOutput{limit: outputLimit},
		done:      make(chan struct{}),
		exitCode:  -1,
		startedAt: now,
		updatedAt: now,
	}
	if err := command.Start(); err != nil {
		cancel()
		return nil, errors.New("sandbox process could not be started")
	}
	go func() { _, _ = io.Copy(process.stdout, stdoutPipe) }()
	go func() { _, _ = io.Copy(process.stderr, stderrPipe) }()
	go func() {
		err := command.Wait()
		process.mu.Lock()
		process.err = err
		process.timedOut = runCtx.Err() == context.DeadlineExceeded
		if command.ProcessState != nil {
			process.exitCode = command.ProcessState.ExitCode()
		}
		process.stdinOpen = false
		if process.terminated {
			process.status = SandboxProcessStatusTerminated
		} else if process.timedOut {
			process.status = SandboxProcessStatusExpired
			process.summary = "expired by sandbox process timeout"
		} else {
			process.status = SandboxProcessStatusExited
		}
		now := s.nowTime()
		process.updatedAt = now
		process.endedAt = &now
		cancel()
		close(process.done)
		process.mu.Unlock()
		_ = s.saveProcess(process)
	}()
	s.mu.Lock()
	if s.processes == nil {
		s.processes = map[string]*sandboxProcess{}
	}
	s.processes[processID] = process
	_ = s.saveProcessLocked(process)
	s.mu.Unlock()
	return process.result(productdata.ToolNameSandboxStartProcess, "start_process", scope.root, 0), nil
}

func (s *SandboxProcessStore) continueProcess(scope workspaceScope, invocation ToolInvocation) (map[string]any, error) {
	for key := range invocation.ArgumentsSummary {
		if key != "process_id" && key != "cursor" && key != "stdin_text" && key != "input_seq" && key != "close_stdin" {
			return nil, errors.New("sandbox process argument is not supported")
		}
	}
	if value, ok := invocation.ArgumentsSummary["stdin_text"]; ok {
		if _, ok := value.(string); !ok {
			return nil, errors.New("sandbox process stdin_text is invalid")
		}
	}
	if value, ok := invocation.ArgumentsSummary["close_stdin"]; ok {
		if _, ok := value.(bool); !ok {
			return nil, errors.New("sandbox process close_stdin is invalid")
		}
	}
	process, err := s.get(invocation.RunID, invocation.ArgumentsSummary)
	if err != nil {
		return nil, err
	}
	if productdata.IsRunTerminal(invocation.RunStatus) && sandboxContinueHasMutation(invocation.ArgumentsSummary) {
		return nil, errors.New("sandbox process cannot mutate for a terminal run")
	}
	s.reconcile(process)
	if process.isTerminal() && sandboxContinueHasMutation(invocation.ArgumentsSummary) {
		return nil, errors.New("sandbox process cannot mutate a terminal process")
	}
	if !process.isTerminal() {
		if err := process.applyInput(invocation.ArgumentsSummary); err != nil {
			return nil, err
		}
		process.touch(s.nowTime())
	}
	cursor, err := sandboxCursorArg(invocation.ArgumentsSummary, "cursor")
	if err != nil {
		return nil, err
	}
	result := process.result(productdata.ToolNameSandboxContinueProcess, "continue_process", scope.root, cursor)
	_ = s.saveProcess(process)
	return result, nil
}

func (s *SandboxProcessStore) terminateProcess(scope workspaceScope, invocation ToolInvocation) (map[string]any, error) {
	for key := range invocation.ArgumentsSummary {
		if key != "process_id" {
			return nil, errors.New("sandbox process argument is not supported")
		}
	}
	process, err := s.get(invocation.RunID, invocation.ArgumentsSummary)
	if err != nil {
		return nil, err
	}
	process.mu.Lock()
	process.terminated = true
	process.status = SandboxProcessStatusTerminated
	process.stdinOpen = false
	now := s.nowTime()
	process.updatedAt = now
	if process.endedAt == nil {
		process.endedAt = &now
	}
	cancel := process.cancel
	if process.command != nil && process.command.Process != nil {
		_ = killSandboxProcessGroup(process.command.Process, syscall.SIGTERM)
	}
	process.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	select {
	case <-process.done:
	case <-time.After(2 * time.Second):
		if process.command.Process != nil {
			_ = killSandboxProcessGroup(process.command.Process, syscall.SIGKILL)
		}
		<-process.done
	}
	result := process.result(productdata.ToolNameSandboxTerminateProcess, "terminate_process", scope.root, 0)
	_ = s.saveProcess(process)
	return result, nil
}

func (s *SandboxProcessStore) get(runID string, args map[string]any) (*sandboxProcess, error) {
	if s == nil {
		return nil, errors.New("sandbox process store is unavailable")
	}
	processID := strings.TrimSpace(stringArg(args, "process_id", ""))
	if processID == "" {
		return nil, errors.New("sandbox process id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	process, ok := s.processes[processID]
	if !ok || process.runID != runID {
		return nil, errors.New("sandbox process is unavailable")
	}
	return process, nil
}

func sandboxContinueHasMutation(args map[string]any) bool {
	if _, ok := args["stdin_text"]; ok {
		return true
	}
	return boolArg(args, "close_stdin", false)
}

func (p *sandboxProcess) touch(now time.Time) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.updatedAt = now
}

func (p *sandboxProcess) applyInput(args map[string]any) error {
	stdinText, hasStdinText := args["stdin_text"].(string)
	closeStdin := boolArg(args, "close_stdin", false)
	inputSeq := boundedInt(args, "input_seq", 0, 1_000_000)
	if hasStdinText && inputSeq <= 0 {
		return errors.New("sandbox process input_seq is required for stdin_text")
	}
	p.mu.Lock()
	stdin := p.stdin
	stdinOpen := p.stdinOpen
	if hasStdinText {
		if stdin == nil || !stdinOpen {
			p.mu.Unlock()
			return errors.New("sandbox process stdin is not open")
		}
		if inputSeq <= p.inputSeq {
			p.mu.Unlock()
			return errors.New("sandbox process input_seq must increase")
		}
		p.inputSeq = inputSeq
	}
	if closeStdin && stdinOpen {
		p.stdinOpen = false
	}
	p.mu.Unlock()
	if hasStdinText {
		if _, err := io.WriteString(stdin, stdinText); err != nil {
			return errors.New("sandbox process stdin could not be written")
		}
	}
	if closeStdin && stdin != nil {
		_ = stdin.Close()
	}
	return nil
}

func (p *sandboxProcess) isTerminal() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch p.status {
	case SandboxProcessStatusExited, SandboxProcessStatusTerminated, SandboxProcessStatusFailed, SandboxProcessStatusExpired:
		p.stdinOpen = false
		return true
	}
	if p.terminated {
		p.stdinOpen = false
		p.status = SandboxProcessStatusTerminated
		return true
	}
	select {
	case <-p.done:
		p.stdinOpen = false
		if p.status == "" || p.status == SandboxProcessStatusRunning {
			p.status = SandboxProcessStatusExited
		}
		return true
	default:
		return false
	}
}

func (p *sandboxProcess) result(toolName string, operation string, root string, cursor int) map[string]any {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := p.status
	if status == "" {
		status = SandboxProcessStatusRunning
	}
	if p.terminated {
		status = SandboxProcessStatusTerminated
	}
	select {
	case <-p.done:
		if status == SandboxProcessStatusRunning {
			status = SandboxProcessStatusExited
		}
	default:
	}
	stdout := p.stdout.StringFrom(cursor)
	nextCursor := p.stdout.Stored()
	stdoutPreview, stdoutRedacted := sandboxSafeOutputPreview(stdout, root)
	stderrPreview, stderrRedacted := sandboxSafeOutputPreview(p.stderr.String(), root)
	terminalSummary := ""
	if status != SandboxProcessStatusRunning {
		terminalSummary = string(status)
		if p.exitCode >= 0 {
			terminalSummary += " exit_code=" + strconv.Itoa(p.exitCode)
		}
		if p.timedOut {
			terminalSummary += " timed_out=true"
		}
		if p.summary != "" {
			terminalSummary += " " + p.summary
		}
	}
	exitCode := any(nil)
	if p.exitCode >= 0 {
		exitCode = p.exitCode
	}
	return map[string]any{
		"tool":              toolName,
		"scope":             "bounded_process",
		"operation":         operation,
		"process_id":        p.processID,
		"argv":              append([]string(nil), p.argv...),
		"argv_summary":      sandboxSafeArgvSummary(p.argv),
		"cwd":               p.cwd,
		"cwd_alias":         p.cwd,
		"status":            string(status),
		"exit_code":         exitCode,
		"timed_out":         p.timedOut,
		"stdout":            stdoutPreview,
		"stderr":            stderrPreview,
		"stdout_bytes":      p.stdout.Total(),
		"stderr_bytes":      p.stderr.Total(),
		"stdout_truncated":  p.stdout.Truncated(),
		"stderr_truncated":  p.stderr.Truncated(),
		"next_cursor":       nextCursor,
		"stdin_open":        p.stdinOpen,
		"input_seq":         p.inputSeq,
		"terminal_summary":  terminalSummary,
		"redaction_applied": stdoutRedacted || stderrRedacted,
		"started_at":        p.startedAt.UTC().Format(time.RFC3339Nano),
		"updated_at":        p.updatedAt.UTC().Format(time.RFC3339Nano),
		"ended_at":          sandboxTimeString(p.endedAt),
	}
}

func sandboxProcessID(runID string, toolCallID string, argv []string) string {
	sum := sha256.Sum256([]byte(runID + "\x00" + toolCallID + "\x00" + strings.Join(argv, "\x00")))
	return "sp_" + hex.EncodeToString(sum[:8])
}

func (s *SandboxProcessStore) reconcile(process *sandboxProcess) {
	if s == nil || process == nil {
		return
	}
	process.mu.Lock()
	now := s.nowTime()
	if process.status == SandboxProcessStatusRunning && s.processExpiredLocked(process, now) {
		process.status = SandboxProcessStatusExpired
		process.stdinOpen = false
		process.summary = "expired by sandbox process TTL"
		process.updatedAt = now
		process.endedAt = &now
		if process.cancel != nil {
			process.cancel()
		}
		if process.command != nil && process.command.Process != nil {
			_ = killSandboxProcessGroup(process.command.Process, syscall.SIGTERM)
		}
		process.mu.Unlock()
		_ = s.saveProcess(process)
		return
	}
	if process.status == SandboxProcessStatusRunning && process.restored && process.command == nil {
		process.status = SandboxProcessStatusFailed
		process.stdinOpen = false
		process.summary = "process missing after registry restore"
		process.updatedAt = now
		process.endedAt = &now
		process.mu.Unlock()
		_ = s.saveProcess(process)
		return
	}
	process.mu.Unlock()
}

func (s *SandboxProcessStore) processExpired(process *sandboxProcess, now time.Time) bool {
	process.mu.Lock()
	defer process.mu.Unlock()
	return s.processExpiredLocked(process, now)
}

func (s *SandboxProcessStore) processExpiredLocked(process *sandboxProcess, now time.Time) bool {
	if process == nil || process.status != SandboxProcessStatusRunning {
		return false
	}
	startedAt := process.startedAt
	if startedAt.IsZero() {
		startedAt = process.updatedAt
	}
	if !startedAt.IsZero() && s.maxLife > 0 && now.Sub(startedAt) > s.maxLife {
		return true
	}
	if !process.updatedAt.IsZero() && s.idle > 0 && now.Sub(process.updatedAt) > s.idle {
		return true
	}
	return false
}

func (s *SandboxProcessStore) saveProcess(process *sandboxProcess) error {
	if s == nil || process == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveProcessLocked(process)
}

func (s *SandboxProcessStore) saveProcessLocked(process *sandboxProcess) error {
	if s == nil || s.repo == nil || process == nil {
		return nil
	}
	record := process.record()
	return s.repo.SaveSandboxProcess(context.Background(), record)
}

func (p *sandboxProcess) record() SandboxProcessRecord {
	p.mu.Lock()
	defer p.mu.Unlock()
	status := p.status
	if status == "" {
		status = SandboxProcessStatusRunning
	}
	exitCode := (*int)(nil)
	if p.exitCode >= 0 {
		value := p.exitCode
		exitCode = &value
	}
	stdoutTail := p.stdout.String()
	stderrTail := p.stderr.String()
	return SandboxProcessRecord{
		RunID:        p.runID,
		ProcessID:    p.processID,
		Argv:         sandboxSafeArgvSummary(p.argv),
		CwdAlias:     p.cwd,
		Status:       status,
		Cursor:       p.stdout.Stored(),
		StartedAt:    p.startedAt,
		UpdatedAt:    p.updatedAt,
		EndedAt:      p.endedAt,
		ExitCode:     exitCode,
		StdoutTail:   stdoutTail,
		StdoutCursor: p.stdout.Stored(),
		StderrTail:   stderrTail,
		StderrCursor: p.stderr.Stored(),
		StdinOpen:    p.stdinOpen,
		InputSeq:     p.inputSeq,
		TimedOut:     p.timedOut,
		Summary:      p.summary,
		OutputLimit:  p.stdout.limit,
	}
}

func sandboxSafeArgvSummary(argv []string) []string {
	safe := make([]string, 0, len(argv))
	for _, arg := range argv {
		text := productdata.RedactEventText(strings.TrimSpace(arg))
		if filepath.IsAbs(text) || strings.Contains(text, "/Users/") {
			text = "[redacted-path]"
		}
		safe = append(safe, text)
	}
	return safe
}

func sandboxTimeString(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func configureSandboxCommand(command *exec.Cmd) {
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 2 * time.Second
	command.Cancel = func() error {
		return killSandboxProcessGroup(command.Process, syscall.SIGTERM)
	}
}

func killSandboxProcessGroup(process *os.Process, signal syscall.Signal) error {
	if process == nil {
		return nil
	}
	if err := syscall.Kill(-process.Pid, signal); err != nil {
		return process.Signal(signal)
	}
	return nil
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

func sandboxCursorArg(args map[string]any, key string) (int, error) {
	value, ok := args[key]
	if !ok || value == nil {
		return 0, nil
	}
	var parsed int
	switch typed := value.(type) {
	case int:
		parsed = typed
	case int64:
		parsed = int(typed)
	case float64:
		parsed = int(typed)
	default:
		return 0, errors.New("sandbox process cursor is invalid")
	}
	if parsed < 0 {
		return 0, errors.New("sandbox process cursor is invalid")
	}
	return parsed, nil
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
		return len(argv) == 1 || (len(argv) == 2 && sandboxPathArgAllowed(argv[1]))
	case "cat", "wc":
		return len(argv) >= 2 && sandboxPathArgsAllowed(argv[1:])
	case "head", "tail":
		return sandboxHeadTailArgsAllowed(argv[1:])
	case "sed":
		return sandboxSedArgsAllowed(argv[1:])
	case "rg":
		return sandboxRGArgsAllowed(argv[1:])
	case "git":
		return gitReadOnlyArgsAllowed(argv[1:])
	case "go":
		return sandboxGoArgsAllowed(argv[1:])
	case "bun":
		return sandboxJSValidationArgsAllowed(argv[1:])
	case "npm", "pnpm":
		return sandboxPackageValidationArgsAllowed(argv[1:])
	default:
		return false
	}
}

func allowedSandboxProcessCommand(argv []string, stdinEnabled bool) bool {
	if stdinEnabled && len(argv) == 1 && strings.ToLower(filepath.Base(strings.TrimSpace(argv[0]))) == "cat" {
		return true
	}
	return allowedReadOnlyCommand(argv)
}

func gitReadOnlyArgsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "status", "diff", "log", "show":
		return true
	default:
		return false
	}
}

func sandboxPathArgsAllowed(args []string) bool {
	for _, arg := range args {
		if !sandboxPathArgAllowed(arg) {
			return false
		}
	}
	return true
}

func sandboxPathArgAllowed(arg string) bool {
	arg = strings.TrimSpace(arg)
	if arg == "" {
		return false
	}
	if strings.HasPrefix(arg, "-") {
		return false
	}
	rel, err := cleanWorkspaceRelativePath(arg)
	if err != nil {
		return false
	}
	return rel == "." || !isSensitiveWorkspacePath(rel)
}

func sandboxHeadTailArgsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	paths := []string{}
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "-n" {
			i++
			if i >= len(args) || strings.TrimSpace(args[i]) == "" || strings.HasPrefix(strings.TrimSpace(args[i]), "-") {
				return false
			}
			continue
		}
		paths = append(paths, arg)
	}
	return len(paths) > 0 && sandboxPathArgsAllowed(paths)
}

func sandboxSedArgsAllowed(args []string) bool {
	if len(args) != 3 || args[0] != "-n" {
		return false
	}
	script := strings.TrimSpace(args[1])
	if script == "" || strings.ContainsAny(script, "we") {
		return false
	}
	if !strings.HasSuffix(script, "p") {
		return false
	}
	return sandboxPathArgAllowed(args[2])
}

func sandboxRGArgsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			return false
		}
		if arg == "-u" || arg == "-uu" || arg == "-uuu" || arg == "--hidden" || arg == "--no-ignore" {
			return false
		}
		if strings.HasPrefix(arg, "../") || filepath.IsAbs(arg) || isSensitiveWorkspacePath(filepath.ToSlash(filepath.Clean(arg))) {
			return false
		}
	}
	return true
}

func sandboxGoArgsAllowed(args []string) bool {
	if len(args) == 0 || args[0] != "test" {
		return false
	}
	for _, arg := range args[1:] {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			return false
		}
		if arg == "-o" || strings.HasPrefix(arg, "-o=") || arg == "-coverprofile" || strings.HasPrefix(arg, "-coverprofile=") {
			return false
		}
		if !strings.HasPrefix(arg, "-") && arg != "./..." && !sandboxPathArgAllowed(arg) {
			return false
		}
	}
	return true
}

func sandboxJSValidationArgsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "test" {
		return sandboxOptionalPathArgsAllowed(args[1:])
	}
	if args[0] != "run" {
		return false
	}
	rest := args[1:]
	if len(rest) >= 3 && rest[0] == "--cwd" {
		if !sandboxPathArgAllowed(rest[1]) {
			return false
		}
		rest = rest[2:]
	}
	if len(rest) == 0 {
		return false
	}
	if rest[0] != "build" && rest[0] != "test" {
		return false
	}
	return sandboxOptionalPathArgsAllowed(rest[1:])
}

func sandboxPackageValidationArgsAllowed(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "test" {
		return sandboxOptionalPathArgsAllowed(args[1:])
	}
	if len(args) >= 2 && args[0] == "run" && (args[1] == "build" || args[1] == "test") {
		return sandboxOptionalPathArgsAllowed(args[2:])
	}
	return false
}

func sandboxOptionalPathArgsAllowed(args []string) bool {
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			return false
		}
		if !strings.HasPrefix(arg, "-") && !sandboxPathArgAllowed(arg) {
			return false
		}
	}
	return true
}

func sandboxOutputPreview(content string, root string) string {
	preview, _ := sandboxSafeOutputPreview(content, root)
	return preview
}

func sandboxSafeOutputPreview(content string, root string) (string, bool) {
	original := content
	content = strings.ToValidUTF8(content, "")
	content = strings.ReplaceAll(content, root, ".")
	if strings.ContainsRune(content, 0) || !utf8.ValidString(content) {
		return "", true
	}
	content = productdata.RedactEventText(content)
	return content, content != original
}

func (w *boundedOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	written := len(p)
	w.total += len(p)
	if w.limit <= 0 {
		w.buf = nil
		w.start = w.total
		return written, nil
	}
	w.buf = append(w.buf, p...)
	if len(w.buf) > w.limit {
		over := len(w.buf) - w.limit
		w.buf = append([]byte(nil), w.buf[over:]...)
		w.start += over
	}
	return written, nil
}

func (w *boundedOutput) setTail(tail []byte, cursor int) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if cursor < 0 {
		cursor = 0
	}
	if w.limit > 0 && len(tail) > w.limit {
		tail = tail[len(tail)-w.limit:]
	}
	w.buf = append([]byte(nil), tail...)
	w.total = cursor
	w.start = cursor - len(w.buf)
	if w.start < 0 {
		w.start = 0
	}
}

func (w *boundedOutput) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return string(w.buf)
}

func (w *boundedOutput) StringFrom(offset int) string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if offset < w.start {
		offset = w.start
	}
	if offset >= w.total {
		return ""
	}
	bufferOffset := offset - w.start
	if bufferOffset < 0 {
		bufferOffset = 0
	}
	if bufferOffset >= len(w.buf) {
		return ""
	}
	return string(w.buf[bufferOffset:])
}

func (w *boundedOutput) Stored() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.total
}

func (w *boundedOutput) Truncated() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.start > 0
}

func (w *boundedOutput) Total() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.total
}

func (e SandboxToolExecutor) store() *SandboxProcessStore {
	if e.Store != nil {
		return e.Store
	}
	return defaultSandboxProcessStore
}

func (s *SandboxProcessStore) nowTime() time.Time {
	if s != nil && s.now != nil {
		return s.now()
	}
	return time.Now()
}
