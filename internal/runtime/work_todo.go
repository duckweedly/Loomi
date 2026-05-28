package runtime

import (
	"context"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

func appendWorkTodoSnapshot(ctx context.Context, svc productdata.Service, run productdata.Run, updatedBy string) (productdata.RunEvent, bool) {
	if svc == nil || run.ThreadID == "" || run.ID == "" {
		return productdata.RunEvent{}, false
	}
	thread, err := svc.GetThread(ctx, identity.LocalDevIdentity(), run.ThreadID)
	if err != nil || thread.Mode != productdata.ThreadModeWork {
		return productdata.RunEvent{}, false
	}
	events, err := svc.ListRunEvents(ctx, identity.LocalDevIdentity(), run.ID, 0)
	if err != nil {
		return productdata.RunEvent{}, false
	}
	items := workTodoItemsFromEvents(events)
	if len(items) == 0 {
		return productdata.RunEvent{}, false
	}
	event, err := svc.AppendRunEvent(ctx, identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     productdata.EventWorkTodoUpdated,
		Summary:  "Work todo updated",
		Metadata: map[string]any{
			"todo_items":        items,
			"updated_by":        updatedBy,
			"redaction_applied": false,
		},
	})
	if err != nil {
		return productdata.RunEvent{}, false
	}
	return event, true
}

func appendProviderWorkTodoSnapshot(ctx context.Context, svc productdata.Service, run productdata.Run, result map[string]any) (productdata.RunEvent, bool) {
	if svc == nil || run.ThreadID == "" || run.ID == "" {
		return productdata.RunEvent{}, false
	}
	thread, err := svc.GetThread(ctx, identity.LocalDevIdentity(), run.ThreadID)
	if err != nil || thread.Mode != productdata.ThreadModeWork {
		return productdata.RunEvent{}, false
	}
	items, ok := result["todo_items"].([]any)
	if !ok || len(items) == 0 {
		return productdata.RunEvent{}, false
	}
	event, err := svc.AppendRunEvent(ctx, identity.LocalDevIdentity(), run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     productdata.EventWorkTodoUpdated,
		Summary:  "Work todo updated",
		Metadata: map[string]any{
			"todo_items":        items,
			"updated_by":        "provider",
			"redaction_applied": result["redaction_applied"],
		},
	})
	if err != nil {
		return productdata.RunEvent{}, false
	}
	return event, true
}

func workTodoItemsFromEvents(events []productdata.RunEvent) []any {
	type todoState struct {
		id       string
		toolName string
		status   string
		summary  string
	}
	order := []string{}
	byID := map[string]todoState{}
	for _, event := range events {
		if event.Type != productdata.EventToolCallRequested &&
			event.Type != productdata.EventToolCallApprovalRequired &&
			event.Type != productdata.EventToolCallApproved &&
			event.Type != productdata.EventToolCallExecuting &&
			event.Type != productdata.EventToolCallSucceeded &&
			event.Type != productdata.EventToolCallFailed &&
			event.Type != productdata.EventToolCallDenied {
			continue
		}
		id := metadataString(event.Metadata, "tool_call_id")
		if id == "" {
			continue
		}
		state, ok := byID[id]
		if !ok {
			order = append(order, id)
			state = todoState{id: "tool_" + id, toolName: metadataString(event.Metadata, "tool_name")}
		}
		if state.toolName == "" {
			state.toolName = metadataString(event.Metadata, "tool_name")
		}
		state.status = todoStatusForToolEvent(event.Type)
		state.summary = todoSummaryForToolEvent(event.Type, state.toolName)
		byID[id] = state
	}
	if len(order) > productdata.MaxWorkTodoItems {
		order = order[len(order)-productdata.MaxWorkTodoItems:]
	}
	items := make([]any, 0, len(order))
	for _, id := range order {
		state := byID[id]
		items = append(items, map[string]any{
			"id":      state.id,
			"title":   todoTitleForTool(state.toolName),
			"status":  state.status,
			"summary": state.summary,
		})
	}
	return items
}

func todoStatusForToolEvent(eventType string) string {
	switch eventType {
	case productdata.EventToolCallSucceeded:
		return "completed"
	case productdata.EventToolCallFailed, productdata.EventToolCallDenied:
		return "failed"
	case productdata.EventToolCallApprovalRequired:
		return "blocked"
	default:
		return "running"
	}
}

func todoSummaryForToolEvent(eventType string, toolName string) string {
	switch eventType {
	case productdata.EventToolCallApprovalRequired:
		return "Waiting for approval"
	case productdata.EventToolCallApproved:
		return "Approved"
	case productdata.EventToolCallExecuting:
		return "Executing"
	case productdata.EventToolCallSucceeded:
		return "Completed"
	case productdata.EventToolCallFailed:
		return "Failed"
	case productdata.EventToolCallDenied:
		return "Denied"
	default:
		if toolName == "" {
			return "Requested"
		}
		return "Requested " + toolName
	}
}

func todoTitleForTool(toolName string) string {
	switch {
	case toolName == productdata.ToolNameWorkspaceGlob || toolName == productdata.ToolNameWorkspaceGrep:
		return "Search project files"
	case toolName == productdata.ToolNameWorkspaceRead:
		return "Read project file"
	case toolName == productdata.ToolNameWorkspaceListDirectory:
		return "Read directory"
	case toolName == productdata.ToolNameWorkspaceTreeSummary:
		return "Summarize directory"
	case toolName == productdata.ToolNameWorkspaceWriteFile:
		return "Create workspace file"
	case toolName == productdata.ToolNameWorkspaceEdit:
		return "Edit workspace file"
	case toolName == productdata.ToolNameWorkspacePatchPreview:
		return "Preview workspace patch"
	case toolName == productdata.ToolNameWorkspacePatchApply:
		return "Apply workspace patch"
	case toolName == productdata.ToolNameSandboxExecCommand:
		return "Run validation command"
	case productdata.IsLSPToolName(toolName):
		return "Analyze code structure"
	case productdata.IsWebToolName(toolName):
		return "Check web information"
	case productdata.IsBrowserToolName(toolName):
		return "Inspect web page"
	case productdata.IsArtifactToolName(toolName):
		return "Handle artifact"
	case productdata.IsAgentToolName(toolName):
		return "Coordinate subtask"
	default:
		text := strings.TrimSpace(toolName)
		if text == "" {
			return "Use approved tool"
		}
		return "Use " + text
	}
}
