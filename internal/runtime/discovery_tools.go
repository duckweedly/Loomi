package runtime

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/sheridiany/loomi/internal/productdata"
)

type DiscoveryToolExecutor struct {
	SkillInput SkillDiscoveryInput
}

func DiscoveryToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameLoadTools, ApprovalPolicy: ToolApprovalNotRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameLoadSkill, ApprovalPolicy: ToolApprovalNotRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e DiscoveryToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	switch invocation.ToolName {
	case productdata.ToolNameLoadTools:
		return e.loadTools(invocation), nil
	case productdata.ToolNameLoadSkill:
		return e.loadSkill(ctx, invocation)
	default:
		return nil, errors.New("discovery tool is not supported")
	}
}

func (e DiscoveryToolExecutor) loadTools(invocation ToolInvocation) map[string]any {
	limit := boundedInt(invocation.ArgumentsSummary, "limit", 12, 30)
	names := normalizeDiscoveryToolNames(stringListArg(invocation.ArgumentsSummary["names"]))
	queries := stringListArg(invocation.ArgumentsSummary["queries"])
	enabled := map[string]bool{}
	for _, tool := range invocation.EnabledTools {
		if tool.ExecutionState == string(productdata.ToolExecutionStateExecutable) {
			enabled[tool.Name] = true
		}
	}
	matches := []map[string]any{}
	for _, entry := range invocation.Catalog {
		if !enabled[entry.Name] || !discoveryToolMatches(entry, names, queries) {
			continue
		}
		matches = append(matches, map[string]any{
			"name":            entry.Name,
			"display_name":    entry.DisplayName,
			"description":     entry.Description,
			"group":           entry.Group,
			"risk_level":      entry.RiskLevel,
			"approval_policy": entry.ApprovalPolicy,
			"safe_metadata":   entry.SafeMetadata,
		})
	}
	sort.Slice(matches, func(i, j int) bool {
		return stringValue(matches[i], "name") < stringValue(matches[j], "name")
	})
	truncated := false
	if len(matches) > limit {
		matches = matches[:limit]
		truncated = true
	}
	return map[string]any{
		"tool":                  productdata.ToolNameLoadTools,
		"scope":                 "runtime_catalog",
		"operation":             "load_tools",
		"tools":                 matches,
		"count":                 len(matches),
		"truncated":             truncated,
		"dynamic_schema_loader": false,
	}
}

func (e DiscoveryToolExecutor) loadSkill(_ context.Context, invocation ToolInvocation) (map[string]any, error) {
	name := strings.ToLower(strings.TrimSpace(stringArg(invocation.ArgumentsSummary, "name", "")))
	if name == "" {
		return nil, errors.New("skill name is required")
	}
	limit := boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)
	input := e.SkillInput
	if input.HomeDir == "" && input.WorkspaceDir == "" && len(input.ExtraRoots) == 0 {
		input = DefaultSkillDiscoveryInput()
	}
	skills, err := DiscoverInstalledSkills(input)
	if err != nil {
		return nil, errors.New("skill discovery failed")
	}
	matches := []map[string]any{}
	for _, skill := range skills {
		if !skillMatches(skill, name) {
			continue
		}
		matches = append(matches, map[string]any{
			"id":                  skill.ID,
			"name":                skill.Name,
			"description":         skill.Description,
			"source":              skill.Source,
			"source_label":        skill.SourceLabel,
			"package":             skill.Package,
			"installed":           skill.Installed,
			"instruction_loaded":  false,
			"instruction_excerpt": "",
		})
	}
	truncated := false
	if len(matches) > limit {
		matches = matches[:limit]
		truncated = true
	}
	return map[string]any{
		"tool":                 productdata.ToolNameLoadSkill,
		"scope":                "skill_manifest",
		"operation":            "load_skill",
		"skills":               matches,
		"count":                len(matches),
		"truncated":            truncated,
		"returns_skill_body":   false,
		"executes_skill_code":  false,
		"requires_host_access": false,
	}, nil
}

func discoveryToolMatches(entry productdata.ToolCatalogEntry, names []string, queries []string) bool {
	if len(names) == 0 && len(queries) == 0 {
		return true
	}
	lowerName := strings.ToLower(entry.Name)
	for _, name := range names {
		if lowerName == strings.ToLower(strings.TrimSpace(name)) {
			return true
		}
	}
	text := strings.ToLower(entry.Name + " " + entry.DisplayName + " " + entry.Description + " " + string(entry.Group))
	for _, query := range queries {
		query = strings.ToLower(strings.TrimSpace(query))
		if query != "" && strings.Contains(text, query) {
			return true
		}
		for _, token := range discoveryQueryTokens(query) {
			if strings.Contains(text, token) {
				return true
			}
		}
	}
	return false
}

func discoveryQueryTokens(query string) []string {
	parts := strings.FieldsFunc(strings.ToLower(query), func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '.')
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) < 3 || discoveryQueryStopWord(part) {
			continue
		}
		out = append(out, part)
	}
	return out
}

func discoveryQueryStopWord(token string) bool {
	switch token {
	case "and", "the", "for", "with", "from", "into", "under", "root", "tool", "tools":
		return true
	default:
		return false
	}
}

func skillMatches(skill InstalledSkill, query string) bool {
	text := strings.ToLower(skill.Name + " " + skill.Description + " " + skill.Package + " " + string(skill.Source))
	return strings.Contains(text, query)
}

func stringListArg(value any) []string {
	var items []any
	switch typed := value.(type) {
	case string:
		items = []any{typed}
	case []string:
		items = make([]any, 0, len(typed))
		for _, item := range typed {
			items = append(items, item)
		}
	case []any:
		items = typed
	default:
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if ok && strings.TrimSpace(text) != "" {
			out = append(out, strings.TrimSpace(text))
		}
	}
	return out
}

func normalizeDiscoveryToolNames(names []string) []string {
	if len(names) == 0 {
		return nil
	}
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		normalized = append(normalized, internalProviderToolName(name))
	}
	return normalized
}
