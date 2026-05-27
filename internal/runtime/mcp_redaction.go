package runtime

import "github.com/sheridiany/loomi/internal/productdata"

func RedactMCPSummary(summary map[string]any) map[string]any {
	return redactMCPSummaryMap(productdata.RedactEventMetadata(summary))
}

func redactMCPSummaryMap(summary map[string]any) map[string]any {
	redacted := make(map[string]any, len(summary))
	for key, value := range summary {
		redacted[key] = redactMCPSummaryValue(value)
	}
	return redacted
}

func redactMCPSummaryValue(value any) any {
	switch typed := value.(type) {
	case string:
		return RedactMCPText(typed)
	case []string:
		items := make([]string, len(typed))
		for i, item := range typed {
			items[i] = RedactMCPText(item)
		}
		return items
	case []any:
		items := make([]any, len(typed))
		for i, item := range typed {
			items[i] = redactMCPSummaryValue(item)
		}
		return items
	case map[string]any:
		return redactMCPSummaryMap(typed)
	default:
		return value
	}
}
