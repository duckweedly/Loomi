package runtime

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	defaultLSPResultLimit = 50
	maxLSPResultLimit     = 200
	maxLSPReadBytes       = 128 * 1024
	maxLSPLineBytes       = 1024 * 1024
)

var lspSymbolPattern = regexp.MustCompile(`^\s*(type|func|class|interface|const|let|var|function)\s+(?:\([^)]+\)\s*)?([A-Za-z_][A-Za-z0-9_]*)`)

type LSPToolExecutor struct {
	Root string
}

func LSPToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameLSPDiagnostics, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameLSPSymbols, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameLSPReferences, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameLSPDefinition, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameLSPHover, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e LSPToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	root := e.Root
	if root == "" {
		root = invocation.WorkspaceRoot
	}
	scope, err := newWorkspaceScope(root)
	if err != nil {
		return nil, err
	}
	switch invocation.ToolName {
	case productdata.ToolNameLSPSymbols:
		return scope.lspSymbols(invocation.ArgumentsSummary)
	case productdata.ToolNameLSPReferences:
		return scope.lspReferences(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameLSPDiagnostics:
		return scope.lspDiagnostics(invocation.ArgumentsSummary)
	case productdata.ToolNameLSPDefinition:
		return scope.lspDefinition(ctx, invocation.ArgumentsSummary)
	case productdata.ToolNameLSPHover:
		return scope.lspHover(ctx, invocation.ArgumentsSummary)
	default:
		return nil, errors.New("lsp tool is not supported")
	}
}

func (s workspaceScope) lspSymbols(args map[string]any) (map[string]any, error) {
	if err := validateLSPRuntimeArguments(productdata.ToolNameLSPSymbols, args); err != nil {
		return nil, err
	}
	path, rel, content, err := s.readLSPFile(args)
	if err != nil {
		return nil, err
	}
	_ = path
	query := strings.TrimSpace(stringArg(args, "query", ""))
	limit := boundedInt(args, "limit", defaultLSPResultLimit, maxLSPResultLimit)
	items := make([]map[string]any, 0)
	truncated := false
	for lineNumber, line := range strings.Split(content, "\n") {
		match := lspSymbolPattern.FindStringSubmatch(line)
		if len(match) != 3 {
			continue
		}
		name := match[2]
		if query != "" && !strings.HasPrefix(strings.ToLower(name), strings.ToLower(query)) {
			continue
		}
		if len(items) >= limit {
			truncated = true
			break
		}
		items = append(items, map[string]any{
			"name":    name,
			"kind":    match[1],
			"path":    rel,
			"line":    lineNumber + 1,
			"column":  strings.Index(line, name) + 1,
			"preview": safeLineExcerpt(line),
		})
	}
	return map[string]any{"tool": productdata.ToolNameLSPSymbols, "scope": "lsp", "operation": "symbols", "path": rel, "items": items, "count": len(items), "limit": limit, "truncated": truncated, "redaction_applied": false}, nil
}

func (s workspaceScope) lspReferences(ctx context.Context, args map[string]any) (map[string]any, error) {
	if err := validateLSPRuntimeArguments(productdata.ToolNameLSPReferences, args); err != nil {
		return nil, err
	}
	_, rel, content, err := s.readLSPFile(args)
	if err != nil {
		return nil, err
	}
	line := boundedInt(args, "line", 0, 1<<30)
	column := boundedInt(args, "column", 0, 1<<30)
	if line <= 0 || column <= 0 {
		return nil, errors.New("lsp reference position is required")
	}
	token := lspTokenAt(content, line, column)
	if token == "" {
		return nil, errors.New("lsp reference position has no symbol")
	}
	limit := boundedInt(args, "limit", defaultLSPResultLimit, maxLSPResultLimit)
	items := make([]map[string]any, 0)
	truncated := false
	tokenRe := regexp.MustCompile(`\b` + regexp.QuoteMeta(token) + `\b`)
	err = filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		candidateRel, err := s.relative(path)
		if err != nil {
			return nil
		}
		if candidateRel != "." && isSensitiveWorkspacePath(candidateRel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		if !isLSPTextPath(candidateRel) {
			return nil
		}
		filePath, resolvedRel, err := s.resolveFile(candidateRel)
		if err != nil {
			return nil
		}
		fileContent, err := readLSPTextFile(filePath)
		if err != nil {
			return nil
		}
		for lineNumber, candidateLine := range strings.Split(fileContent, "\n") {
			for _, loc := range tokenRe.FindAllStringIndex(candidateLine, -1) {
				if len(items) >= limit {
					truncated = true
					return filepath.SkipAll
				}
				items = append(items, map[string]any{"name": token, "path": resolvedRel, "line": lineNumber + 1, "column": loc[0] + 1, "preview": safeLineExcerpt(candidateLine)})
			}
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameLSPReferences, "scope": "lsp", "operation": "references", "path": rel, "name": token, "items": items, "count": len(items), "limit": limit, "truncated": truncated, "redaction_applied": false}, nil
}

func (s workspaceScope) lspDefinition(ctx context.Context, args map[string]any) (map[string]any, error) {
	if err := validateLSPRuntimeArguments(productdata.ToolNameLSPDefinition, args); err != nil {
		return nil, err
	}
	_, rel, content, err := s.readLSPFile(args)
	if err != nil {
		return nil, err
	}
	line := boundedInt(args, "line", 0, 1<<30)
	column := boundedInt(args, "column", 0, 1<<30)
	token := lspTokenAt(content, line, column)
	if token == "" {
		return nil, errors.New("lsp definition position has no symbol")
	}
	limit := boundedInt(args, "limit", defaultLSPResultLimit, maxLSPResultLimit)
	items, truncated, err := s.findLSPDefinitions(ctx, token, limit)
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameLSPDefinition, "scope": "lsp", "operation": "definition", "path": rel, "name": token, "items": items, "count": len(items), "limit": limit, "truncated": truncated, "redaction_applied": false}, nil
}

func (s workspaceScope) lspHover(ctx context.Context, args map[string]any) (map[string]any, error) {
	if err := validateLSPRuntimeArguments(productdata.ToolNameLSPHover, args); err != nil {
		return nil, err
	}
	_, rel, content, err := s.readLSPFile(args)
	if err != nil {
		return nil, err
	}
	line := boundedInt(args, "line", 0, 1<<30)
	column := boundedInt(args, "column", 0, 1<<30)
	token := lspTokenAt(content, line, column)
	if token == "" {
		return nil, errors.New("lsp hover position has no symbol")
	}
	definitions, _, err := s.findLSPDefinitions(ctx, token, 1)
	if err != nil {
		return nil, err
	}
	hover := map[string]any{"name": token, "path": rel, "line": line, "column": column}
	if len(definitions) > 0 {
		for key, value := range definitions[0] {
			hover[key] = value
		}
	}
	return map[string]any{"tool": productdata.ToolNameLSPHover, "scope": "lsp", "operation": "hover", "path": rel, "name": token, "hover": hover, "redaction_applied": false}, nil
}

func (s workspaceScope) lspDiagnostics(args map[string]any) (map[string]any, error) {
	if err := validateLSPRuntimeArguments(productdata.ToolNameLSPDiagnostics, args); err != nil {
		return nil, err
	}
	_, rel, content, err := s.readLSPFile(args)
	if err != nil {
		return nil, err
	}
	limit := boundedInt(args, "limit", defaultLSPResultLimit, maxLSPResultLimit)
	items := make([]map[string]any, 0)
	if strings.Count(content, "(") != strings.Count(content, ")") && limit > 0 {
		line, preview := firstLineContaining(content, "(")
		items = append(items, map[string]any{"path": rel, "line": line, "column": 1, "severity": "error", "source": "loomi-lsp-lite", "message": "unbalanced parentheses", "preview": preview})
	}
	return map[string]any{"tool": productdata.ToolNameLSPDiagnostics, "scope": "lsp", "operation": "diagnostics", "path": rel, "items": items, "count": len(items), "limit": limit, "truncated": false, "redaction_applied": false}, nil
}

func validateLSPRuntimeArguments(toolName string, args map[string]any) error {
	allowed := map[string]struct{}{"path": {}, "query": {}, "line": {}, "column": {}, "include_declaration": {}, "language": {}, "limit": {}}
	for key := range args {
		if _, ok := allowed[key]; !ok {
			return errors.New("lsp argument is not supported")
		}
	}
	if strings.TrimSpace(stringArg(args, "path", "")) == "" {
		return errors.New("lsp path is required")
	}
	if (toolName == productdata.ToolNameLSPReferences || toolName == productdata.ToolNameLSPDefinition || toolName == productdata.ToolNameLSPHover) && (boundedInt(args, "line", 0, 1<<30) <= 0 || boundedInt(args, "column", 0, 1<<30) <= 0) {
		return errors.New("lsp position is required")
	}
	return nil
}

func (s workspaceScope) findLSPDefinitions(ctx context.Context, token string, limit int) ([]map[string]any, bool, error) {
	items := make([]map[string]any, 0)
	truncated := false
	err := filepath.WalkDir(s.root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		candidateRel, err := s.relative(path)
		if err != nil {
			return nil
		}
		if candidateRel != "." && isSensitiveWorkspacePath(candidateRel) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !isLSPTextPath(candidateRel) {
			return nil
		}
		filePath, resolvedRel, err := s.resolveFile(candidateRel)
		if err != nil {
			return nil
		}
		fileContent, err := readLSPTextFile(filePath)
		if err != nil {
			return nil
		}
		for lineNumber, candidateLine := range strings.Split(fileContent, "\n") {
			match := lspSymbolPattern.FindStringSubmatch(candidateLine)
			if len(match) != 3 || match[2] != token {
				continue
			}
			if len(items) >= limit {
				truncated = true
				return filepath.SkipAll
			}
			items = append(items, map[string]any{"name": token, "kind": match[1], "path": resolvedRel, "line": lineNumber + 1, "column": strings.Index(candidateLine, token) + 1, "preview": safeLineExcerpt(candidateLine)})
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return nil, false, err
	}
	return items, truncated, nil
}

func (s workspaceScope) readLSPFile(args map[string]any) (string, string, string, error) {
	relArg := strings.TrimSpace(stringArg(args, "path", ""))
	if relArg == "" {
		return "", "", "", errors.New("lsp path is required")
	}
	path, rel, err := s.resolveFile(relArg)
	if err != nil {
		return "", "", "", err
	}
	content, err := readLSPTextFile(path)
	if err != nil {
		return "", "", "", err
	}
	return path, rel, content, nil
}

func readLSPTextFile(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", errors.New("lsp file is unavailable")
	}
	if info.Size() > maxLSPReadBytes {
		return "", errors.New("lsp file is too large")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("lsp file could not be read")
	}
	if bytes.Contains(raw, []byte{0}) {
		return "", errors.New("lsp file must be text")
	}
	content := string(raw)
	if !utf8.ValidString(content) {
		content = strings.ToValidUTF8(content, "")
	}
	return content, nil
}

func lspTokenAt(content string, lineNumber int, column int) string {
	lines := strings.Split(content, "\n")
	if lineNumber < 1 || lineNumber > len(lines) {
		return ""
	}
	line := lines[lineNumber-1]
	runes := []rune(line)
	index := column - 1
	if index < 0 || index >= len(runes) || !isIdentifierRune(runes[index]) {
		return ""
	}
	start := index
	for start > 0 && isIdentifierRune(runes[start-1]) {
		start--
	}
	end := index + 1
	for end < len(runes) && isIdentifierRune(runes[end]) {
		end++
	}
	return string(runes[start:end])
}

func isIdentifierRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isLSPTextPath(rel string) bool {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".md", ".txt", ".json", ".yaml", ".yml":
		return true
	default:
		return false
	}
}

func firstLineContaining(content string, needle string) (int, string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Buffer(make([]byte, 64*1024), maxLSPLineBytes)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if strings.Contains(line, needle) {
			return lineNumber, safeLineExcerpt(line)
		}
	}
	return 1, ""
}
