package runtime

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultWorkspaceReadBytes  = 32 * 1024
	maxWorkspaceReadBytes      = 128 * 1024
	defaultWorkspaceListLimit  = 100
	maxWorkspaceListLimit      = 500
	maxWorkspaceLineBytes      = 1024 * 1024
	defaultWorkspaceWriteBytes = 32 * 1024
	maxWorkspaceWriteBytes     = 128 * 1024
)

type WorkspaceToolExecutor struct {
	Root string
}

type workspaceScope struct {
	root string
}

func WorkspaceToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameWorkspaceGlob, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameWorkspaceGrep, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameWorkspaceRead, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameWorkspaceWriteFile, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameWorkspaceEdit, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e WorkspaceToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	scope, err := newWorkspaceScope(e.Root)
	if err != nil {
		return nil, err
	}
	switch invocation.ToolName {
	case productdata.ToolNameWorkspaceGlob:
		return scope.glob(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameWorkspaceGrep:
		return scope.grep(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameWorkspaceRead:
		return scope.read(invocation.ArgumentsSummary)
	case productdata.ToolNameWorkspaceWriteFile:
		return scope.writeFile(invocation.ArgumentsSummary)
	case productdata.ToolNameWorkspaceEdit:
		return scope.edit(invocation.ArgumentsSummary)
	default:
		return nil, errors.New("workspace tool is not supported")
	}
}

func newWorkspaceScope(root string) (workspaceScope, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		root = strings.TrimSpace(os.Getenv("LOOMI_WORKSPACE_ROOT"))
	}
	if root == "" {
		guessed, err := defaultWorkspaceRoot()
		if err != nil {
			return workspaceScope{}, err
		}
		root = guessed
	}
	if !filepath.IsAbs(root) {
		abs, err := filepath.Abs(root)
		if err != nil {
			return workspaceScope{}, err
		}
		root = abs
	}
	real, err := filepath.EvalSymlinks(root)
	if err != nil {
		return workspaceScope{}, errors.New("workspace root is unavailable")
	}
	info, err := os.Stat(real)
	if err != nil || !info.IsDir() {
		return workspaceScope{}, errors.New("workspace root is unavailable")
	}
	return workspaceScope{root: filepath.Clean(real)}, nil
}

func defaultWorkspaceRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd, nil
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			return "", errors.New("workspace root is unavailable")
		}
		wd = parent
	}
}

func (s workspaceScope) glob(ctx context.Context, args map[string]any) (map[string]any, error) {
	pattern := strings.TrimSpace(stringArg(args, "pattern", ""))
	if pattern == "" {
		return nil, errors.New("workspace glob pattern is required")
	}
	if err := validateRelativePattern(pattern); err != nil {
		return nil, err
	}
	start, _, err := s.resolveDir(stringArg(args, "path", "."))
	if err != nil {
		return nil, err
	}
	limit := boundedInt(args, "limit", defaultWorkspaceListLimit, maxWorkspaceListLimit)
	matches := make([]map[string]any, 0)
	truncated := false
	err = filepath.WalkDir(start, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := s.relative(path)
		if err != nil {
			return nil
		}
		if rel != "." && isSensitiveWorkspacePath(rel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if rel == "." || entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if _, _, err := s.resolveFile(rel); err != nil {
				return nil
			}
		}
		ok, err := workspacePatternMatch(pattern, rel)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}
		if len(matches) >= limit {
			truncated = true
			return filepath.SkipAll
		}
		kind := "file"
		if entry.Type()&os.ModeSymlink != 0 {
			kind = "symlink"
		}
		matches = append(matches, map[string]any{"path": rel, "kind": kind})
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameWorkspaceGlob, "scope": "workspace", "matches": sortedStringMaps(matches), "match_count": len(matches), "limit": limit, "truncated": truncated}, nil
}

func (s workspaceScope) grep(ctx context.Context, args map[string]any) (map[string]any, error) {
	query := strings.TrimSpace(stringArg(args, "query", ""))
	if query == "" {
		query = strings.TrimSpace(stringArg(args, "pattern", ""))
	}
	if query == "" {
		return nil, errors.New("workspace grep query is required")
	}
	if !boolArg(args, "case_sensitive", true) {
		query = "(?i)" + query
	}
	re, err := regexp.Compile(query)
	if err != nil {
		return nil, errors.New("workspace grep query is invalid")
	}
	include := strings.TrimSpace(stringArg(args, "include", ""))
	if include != "" {
		if err := validateRelativePattern(include); err != nil {
			return nil, err
		}
	}
	start, _, err := s.resolvePathOrDir(stringArg(args, "path", "."))
	if err != nil {
		return nil, err
	}
	limit := boundedInt(args, "limit", defaultWorkspaceListLimit, maxWorkspaceListLimit)
	matches := make([]map[string]any, 0)
	truncated := false
	visit := func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		rel, err := s.relative(path)
		if err != nil {
			return nil
		}
		if rel != "." && isSensitiveWorkspacePath(rel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if include != "" {
			ok, err := workspacePatternMatch(include, rel)
			if err != nil {
				return err
			}
			baseOK, _ := filepath.Match(include, filepath.Base(rel))
			if !ok && !baseOK {
				return nil
			}
		}
		filePath, _, err := s.resolveFile(rel)
		if err != nil {
			return nil
		}
		fileMatches, err := grepFile(filePath, rel, re, limit-len(matches))
		if err != nil {
			return nil
		}
		matches = append(matches, fileMatches...)
		if len(matches) >= limit {
			truncated = true
			return filepath.SkipAll
		}
		return nil
	}
	info, err := os.Stat(start)
	if err != nil {
		return nil, errors.New("workspace path is unavailable")
	}
	if info.IsDir() {
		err = filepath.WalkDir(start, visit)
	} else {
		rel, relErr := s.relative(start)
		if relErr != nil {
			return nil, relErr
		}
		err = visit(start, fileInfoDirEntry{info: info}, nil)
		if err == nil && rel == "." {
			err = errors.New("workspace path is invalid")
		}
	}
	if err != nil && err != filepath.SkipAll {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameWorkspaceGrep, "scope": "workspace", "matches": sortedStringMaps(matches), "match_count": len(matches), "limit": limit, "truncated": truncated}, nil
}

func (s workspaceScope) read(args map[string]any) (map[string]any, error) {
	relArg := strings.TrimSpace(stringArg(args, "path", ""))
	if relArg == "" {
		return nil, errors.New("workspace read path is required")
	}
	path, rel, err := s.resolveFile(relArg)
	if err != nil {
		return nil, err
	}
	offset := boundedInt(args, "offset", 0, 1<<30)
	if offset < 0 {
		offset = 0
	}
	limit := boundedInt(args, "limit", defaultWorkspaceReadBytes, maxWorkspaceReadBytes)
	maxBytes := boundedInt(args, "max_bytes", limit, maxWorkspaceReadBytes)
	if maxBytes < limit {
		limit = maxBytes
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.New("workspace file is unavailable")
	}
	defer file.Close()
	if offset > 0 {
		if _, err := file.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, errors.New("workspace read offset is invalid")
		}
	}
	buf := make([]byte, limit+1)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return nil, errors.New("workspace file could not be read")
	}
	raw := buf[:n]
	truncated := n > limit
	if truncated {
		raw = raw[:limit]
	}
	if bytes.Contains(raw, []byte{0}) {
		return map[string]any{"tool": productdata.ToolNameWorkspaceRead, "scope": "workspace", "path": rel, "content": "", "bytes_read": 0, "offset": offset, "limit": limit, "truncated": false, "utf8_valid": false, "summary": "unsupported binary content"}, nil
	}
	content := string(raw)
	valid := utf8.ValidString(content)
	if !valid {
		content = strings.ToValidUTF8(content, "")
	}
	return map[string]any{"tool": productdata.ToolNameWorkspaceRead, "scope": "workspace", "path": rel, "content": content, "bytes_read": len([]byte(content)), "offset": offset, "limit": limit, "truncated": truncated, "utf8_valid": valid}, nil
}

func (s workspaceScope) writeFile(args map[string]any) (map[string]any, error) {
	relArg := strings.TrimSpace(stringArg(args, "path", ""))
	if relArg == "" {
		return nil, errors.New("workspace write path is required")
	}
	content := stringArg(args, "content", "")
	if !utf8.ValidString(content) || strings.ContainsRune(content, 0) {
		return nil, errors.New("workspace write content must be UTF-8 text")
	}
	maxBytes := boundedInt(args, "max_bytes", defaultWorkspaceWriteBytes, maxWorkspaceWriteBytes)
	raw := []byte(content)
	if len(raw) > maxBytes {
		return nil, errors.New("workspace write content is too large")
	}
	path, rel, err := s.resolveNewFile(relArg)
	if err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if os.IsExist(err) {
			return nil, errors.New("workspace write target already exists")
		}
		return nil, errors.New("workspace file could not be written")
	}
	if _, err := file.Write(raw); err != nil {
		_ = file.Close()
		return nil, errors.New("workspace file could not be written")
	}
	if err := file.Close(); err != nil {
		return nil, errors.New("workspace file could not be written")
	}
	return map[string]any{
		"tool":              productdata.ToolNameWorkspaceWriteFile,
		"scope":             "workspace",
		"operation":         "write_file",
		"path":              rel,
		"changed":           true,
		"bytes_written":     len(raw),
		"line_count_after":  countTextLines(content),
		"truncated":         false,
		"redaction_applied": false,
	}, nil
}

func (s workspaceScope) edit(args map[string]any) (map[string]any, error) {
	relArg := strings.TrimSpace(stringArg(args, "path", ""))
	if relArg == "" {
		return nil, errors.New("workspace edit path is required")
	}
	oldText := stringArg(args, "old_text", "")
	if oldText == "" {
		return nil, errors.New("workspace edit old text is required")
	}
	newText := stringArg(args, "new_text", "")
	if !utf8.ValidString(oldText) || strings.ContainsRune(oldText, 0) || !utf8.ValidString(newText) || strings.ContainsRune(newText, 0) {
		return nil, errors.New("workspace edit text must be UTF-8 text")
	}
	maxBytes := boundedInt(args, "max_bytes", defaultWorkspaceWriteBytes, maxWorkspaceWriteBytes)
	path, rel, err := s.resolveFile(relArg)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.New("workspace file is unavailable")
	}
	if info.Size() > int64(maxWorkspaceWriteBytes) {
		return nil, errors.New("workspace edit file is too large")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("workspace file could not be read")
	}
	if bytes.Contains(raw, []byte{0}) || !utf8.Valid(raw) {
		return nil, errors.New("workspace edit target must be UTF-8 text")
	}
	content := string(raw)
	count := strings.Count(content, oldText)
	if count == 0 {
		return nil, errors.New("workspace edit old text was not found")
	}
	if count > 1 {
		return nil, errors.New("workspace edit old text is not unique")
	}
	updated := strings.Replace(content, oldText, newText, 1)
	if len([]byte(updated)) > maxBytes {
		return nil, errors.New("workspace edit result is too large")
	}
	if err := os.WriteFile(path, []byte(updated), info.Mode().Perm()); err != nil {
		return nil, errors.New("workspace file could not be written")
	}
	return map[string]any{
		"tool":              productdata.ToolNameWorkspaceEdit,
		"scope":             "workspace",
		"operation":         "edit",
		"path":              rel,
		"changed":           true,
		"bytes_before":      len(raw),
		"bytes_after":       len([]byte(updated)),
		"line_count_before": countTextLines(content),
		"line_count_after":  countTextLines(updated),
		"truncated":         false,
		"redaction_applied": false,
	}, nil
}

func (s workspaceScope) resolveDir(relArg string) (string, string, error) {
	path, rel, err := s.resolveWorkspacePath(relArg)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return "", "", errors.New("workspace path is not a directory")
	}
	return path, rel, nil
}

func (s workspaceScope) resolvePathOrDir(relArg string) (string, string, error) {
	return s.resolveWorkspacePath(relArg)
}

func (s workspaceScope) resolveFile(relArg string) (string, string, error) {
	path, rel, err := s.resolveWorkspacePath(relArg)
	if err != nil {
		return "", "", err
	}
	info, err := os.Stat(path)
	if err != nil {
		return "", "", errors.New("workspace file is unavailable")
	}
	if info.IsDir() {
		return "", "", errors.New("workspace path is a directory")
	}
	return path, rel, nil
}

func (s workspaceScope) resolveNewFile(relArg string) (string, string, error) {
	rel, err := cleanWorkspaceRelativePath(relArg)
	if err != nil {
		return "", "", err
	}
	if rel == "." || isSensitiveWorkspacePath(rel) {
		return "", "", errors.New("workspace path is sensitive")
	}
	parentRel := filepath.ToSlash(filepath.Dir(rel))
	if parentRel == "." {
		parentRel = "."
	}
	parent, _, err := s.resolveDir(parentRel)
	if err != nil {
		return "", "", err
	}
	target := filepath.Join(parent, filepath.Base(filepath.FromSlash(rel)))
	if !s.contains(target) {
		return "", "", errors.New("workspace path is outside the allowed scope")
	}
	if _, err := os.Lstat(target); err == nil {
		return "", "", errors.New("workspace write target already exists")
	} else if !os.IsNotExist(err) {
		return "", "", errors.New("workspace path is unavailable")
	}
	return target, rel, nil
}

func (s workspaceScope) resolveWorkspacePath(relArg string) (string, string, error) {
	rel, err := cleanWorkspaceRelativePath(relArg)
	if err != nil {
		return "", "", err
	}
	if rel != "." && isSensitiveWorkspacePath(rel) {
		return "", "", errors.New("workspace path is sensitive")
	}
	candidate := filepath.Join(s.root, filepath.FromSlash(rel))
	real, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", "", errors.New("workspace path is unavailable")
	}
	if !s.contains(real) {
		return "", "", errors.New("workspace path is outside the allowed scope")
	}
	resolvedRel, err := s.relative(real)
	if err != nil {
		return "", "", err
	}
	if resolvedRel != "." && isSensitiveWorkspacePath(resolvedRel) {
		return "", "", errors.New("workspace path is sensitive")
	}
	return real, resolvedRel, nil
}

func (s workspaceScope) relative(path string) (string, error) {
	rel, err := filepath.Rel(s.root, path)
	if err != nil {
		return "", errors.New("workspace path is outside the allowed scope")
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	if rel == "." {
		return ".", nil
	}
	if strings.HasPrefix(rel, "../") || rel == ".." {
		return "", errors.New("workspace path is outside the allowed scope")
	}
	return rel, nil
}

func (s workspaceScope) contains(path string) bool {
	rel, err := filepath.Rel(s.root, filepath.Clean(path))
	if err != nil {
		return false
	}
	return rel == "." || (!strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != "..")
}

func cleanWorkspaceRelativePath(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "."
	}
	if strings.HasPrefix(value, "~") || filepath.IsAbs(value) {
		return "", errors.New("workspace path must be relative")
	}
	value = filepath.ToSlash(filepath.Clean(value))
	if value == ".." || strings.HasPrefix(value, "../") {
		return "", errors.New("workspace path is outside the allowed scope")
	}
	return value, nil
}

func validateRelativePattern(pattern string) error {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return errors.New("workspace pattern is required")
	}
	if strings.HasPrefix(pattern, "~") || filepath.IsAbs(pattern) || pattern == ".." || strings.HasPrefix(filepath.ToSlash(pattern), "../") || strings.Contains(filepath.ToSlash(pattern), "/../") {
		return errors.New("workspace pattern is outside the allowed scope")
	}
	return nil
}

func workspacePatternMatch(pattern string, rel string) (bool, error) {
	pattern = filepath.ToSlash(filepath.Clean(pattern))
	rel = filepath.ToSlash(filepath.Clean(rel))
	if pattern == "." {
		return rel == ".", nil
	}
	expr := regexp.QuoteMeta(pattern)
	expr = strings.ReplaceAll(expr, `\*\*`, `.*`)
	expr = strings.ReplaceAll(expr, `\*`, `[^/]*`)
	expr = strings.ReplaceAll(expr, `\?`, `[^/]`)
	re, err := regexp.Compile("^" + expr + "$")
	if err != nil {
		return false, errors.New("workspace glob pattern is invalid")
	}
	if re.MatchString(rel) {
		return true, nil
	}
	if !strings.Contains(pattern, "/") {
		return filepath.Match(pattern, filepath.Base(rel))
	}
	return false, nil
}

func isSensitiveWorkspacePath(rel string) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, part := range parts {
		lower := strings.ToLower(strings.TrimSpace(part))
		if lower == "" || lower == "." {
			continue
		}
		if lower == ".git" || lower == ".ssh" || lower == ".aws" || lower == ".gnupg" || lower == "secrets" || lower == "credentials" {
			return true
		}
		if strings.HasPrefix(lower, ".env") || strings.HasPrefix(lower, "id_rsa") || strings.HasPrefix(lower, "id_ed25519") || strings.HasSuffix(lower, ".pem") {
			return true
		}
	}
	return false
}

func grepFile(path string, rel string, re *regexp.Regexp, remaining int) ([]map[string]any, error) {
	if remaining <= 0 {
		return nil, nil
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), maxWorkspaceLineBytes)
	var matches []map[string]any
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.ContainsRune(line, 0) {
			return nil, nil
		}
		if !re.MatchString(line) {
			continue
		}
		matches = append(matches, map[string]any{"path": rel, "line": lineNumber, "text": safeLineExcerpt(line)})
		if len(matches) >= remaining {
			return matches, nil
		}
	}
	return matches, nil
}

func safeLineExcerpt(line string) string {
	line = strings.ToValidUTF8(line, "")
	if len([]rune(line)) <= 240 {
		return line
	}
	runes := []rune(line)
	return string(runes[:240])
}

func countTextLines(content string) int {
	if content == "" {
		return 0
	}
	lines := strings.Count(content, "\n")
	if !strings.HasSuffix(content, "\n") {
		lines++
	}
	return lines
}

func stringArg(args map[string]any, key string, fallback string) string {
	value, ok := args[key]
	if !ok || value == nil {
		return fallback
	}
	text, ok := value.(string)
	if !ok {
		return fallback
	}
	return text
}

func boolArg(args map[string]any, key string, fallback bool) bool {
	value, ok := args[key]
	if !ok || value == nil {
		return fallback
	}
	boolean, ok := value.(bool)
	if !ok {
		return fallback
	}
	return boolean
}

func boundedInt(args map[string]any, key string, fallback int, max int) int {
	value, ok := args[key]
	if !ok || value == nil {
		return fallback
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
		return fallback
	}
	if parsed <= 0 {
		return fallback
	}
	if parsed > max {
		return max
	}
	return parsed
}

type fileInfoDirEntry struct {
	info os.FileInfo
}

func (e fileInfoDirEntry) Name() string               { return e.info.Name() }
func (e fileInfoDirEntry) IsDir() bool                { return e.info.IsDir() }
func (e fileInfoDirEntry) Type() fs.FileMode          { return e.info.Mode().Type() }
func (e fileInfoDirEntry) Info() (os.FileInfo, error) { return e.info, nil }

func sortedStringMaps(items []map[string]any) []map[string]any {
	sort.SliceStable(items, func(i, j int) bool {
		left, _ := items[i]["path"].(string)
		right, _ := items[j]["path"].(string)
		if left == right {
			li, _ := items[i]["line"].(int)
			ri, _ := items[j]["line"].(int)
			return li < ri
		}
		return left < right
	})
	return items
}
