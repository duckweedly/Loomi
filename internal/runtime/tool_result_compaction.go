package runtime

import (
	"encoding/json"
	"strings"
)

func compactToolResultPayload(input map[string]any, maxBytes int) map[string]any {
	result := make(map[string]any, len(input))
	for key, value := range input {
		result[key] = compactToolResultValue(value, maxBytes)
	}
	if maxBytes <= 0 {
		return result
	}
	encoded, err := json.Marshal(result)
	if err != nil || len(encoded) <= maxBytes {
		return result
	}
	return compactAggregateToolResultPayload(string(encoded), len(encoded), maxBytes)
}

func compactAggregateToolResultPayload(encoded string, encodedBytes int, maxBytes int) map[string]any {
	result := map[string]any{
		"summary":       "tool result compacted before provider continuation",
		"truncated":     true,
		"omitted_bytes": encodedBytes - maxBytes,
		"excerpt":       compactToolResultText(encoded, maxBytes/2),
	}
	for {
		next, err := json.Marshal(result)
		if err != nil || len(next) <= maxBytes {
			return result
		}
		excerpt, ok := result["excerpt"].(string)
		if !ok || excerpt == "" {
			return map[string]any{
				"summary":       "tool result compacted before provider continuation",
				"truncated":     true,
				"omitted_bytes": encodedBytes - maxBytes,
			}
		}
		if len(excerpt) <= 32 {
			delete(result, "excerpt")
			continue
		}
		result["excerpt"] = compactToolResultText(excerpt, len(excerpt)/2)
	}
}

func compactToolResultValue(value any, maxBytes int) any {
	switch typed := value.(type) {
	case string:
		return compactToolResultText(typed, maxBytes)
	case map[string]any:
		return compactToolResultPayload(typed, maxBytes)
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, compactToolResultValue(item, maxBytes))
		}
		return result
	default:
		return value
	}
}

func compactToolResultText(input string, maxBytes int) string {
	input = redactToolResultText(input)
	if maxBytes <= 0 || len(input) <= maxBytes {
		return input
	}
	const marker = "\n[tool output compacted]"
	budget := maxBytes - len(marker)
	if budget <= 0 {
		return input[:maxBytes]
	}
	lines := strings.Split(input, "\n")
	kept := make([]string, 0, 24)
	seen := map[string]bool{}
	addLine := func(line string) {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || seen[trimmed] {
			return
		}
		if len(strings.Join(append(kept, trimmed), "\n")) > budget {
			return
		}
		seen[trimmed] = true
		kept = append(kept, trimmed)
	}

	for i, line := range lines {
		if i >= 8 {
			break
		}
		addLine(line)
	}
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, keyword := range []string{"error", "failed", "panic", "status", "path", "thread_id", "run_id"} {
			if strings.Contains(lower, keyword) {
				addLine(line)
				break
			}
		}
	}
	start := len(lines) - 8
	if start < 0 {
		start = 0
	}
	for _, line := range lines[start:] {
		addLine(line)
	}

	if len(kept) == 0 {
		return input[:budget] + marker
	}
	return strings.Join(kept, "\n") + marker
}

func redactToolResultText(input string) string {
	if strings.TrimSpace(input) == "" {
		return input
	}
	lines := strings.Split(input, "\n")
	for index, line := range lines {
		lines[index] = redactToolResultLine(line)
	}
	return strings.Join(lines, "\n")
}

func redactToolResultLine(line string) string {
	lower := strings.ToLower(line)
	for _, marker := range []string{"postgres://", "postgresql://", "password=", "api_key", " key=", "_key=", "bearer ", "secret", "token", "credential", "authorization", "sk-", ".ssh", "id_ed25519", "id_rsa", ".env", "env="} {
		if strings.Contains(lower, marker) {
			return "[redacted]"
		}
	}
	if strings.Contains(line, "/Users/") || strings.Contains(line, "/home/") || strings.Contains(line, "\\Users\\") || strings.Contains(line, ":\\") {
		return "[redacted]"
	}
	return line
}
