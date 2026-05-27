package runtime

import (
	"testing"

	"github.com/sheridiany/loomi/internal/productdata"
)

func TestExecuteTodoWriteReturnsSafeTodoSnapshot(t *testing.T) {
	result, err := ExecuteTodoWrite(ToolInvocation{
		ToolName: productdata.ToolNameTodoWrite,
		ArgumentsSummary: map[string]any{"items": []any{
			map[string]any{"id": "todo-1", "title": "Review patch", "status": "running", "summary": "Check tests"},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["operation"] != "todo_write" || result["updated_by"] != "provider" {
		t.Fatalf("result = %+v", result)
	}
	items, ok := result["todo_items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %#v", result["todo_items"])
	}
	item := items[0].(map[string]any)
	if item["title"] != "Review patch" || item["status"] != "running" {
		t.Fatalf("item = %+v", item)
	}
}
