package runtime

import (
	"errors"

	"github.com/sheridiany/loomi/internal/productdata"
)

func ExecuteTodoWrite(invocation ToolInvocation) (map[string]any, error) {
	if invocation.ToolName != productdata.ToolNameTodoWrite {
		return nil, errors.New("todo tool is not supported")
	}
	items, ok := invocation.ArgumentsSummary["items"].([]any)
	if !ok || len(items) == 0 {
		return nil, errors.New("todo items are required")
	}
	metadata := productdata.NormalizeWorkTodoMetadata(map[string]any{"todo_items": items, "updated_by": "provider"})
	normalized, _ := metadata["todo_items"].([]any)
	if len(normalized) == 0 {
		return nil, errors.New("todo items are required")
	}
	return map[string]any{
		"operation":         "todo_write",
		"todo_items":        normalized,
		"updated_by":        "provider",
		"redaction_applied": metadata["redaction_applied"],
	}, nil
}
