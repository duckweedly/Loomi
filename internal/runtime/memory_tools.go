package runtime

import (
	"context"
	"errors"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

type MemoryToolExecutor struct {
	Service productdata.Service
	Ident   identity.LocalIdentity
}

func MemoryToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameMemorySearch, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryList, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryRead, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryWrite, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryEdit, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryForget, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryContext, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryTimeline, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryConnections, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryThreadSearch, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryThreadFetch, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameMemoryStatus, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameNotebookRead, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameNotebookWrite, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameNotebookEdit, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameNotebookForget, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e MemoryToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if e.Service == nil {
		return nil, errors.New("memory service is unavailable")
	}
	if _, err := productdata.ValidateToolCallRequestInput(productdata.RecordToolCallRequestInput{ToolCallID: firstNonEmpty(invocation.ToolCallID, "tc_memory_runtime"), ToolName: invocation.ToolName, ArgumentsSummary: invocation.ArgumentsSummary, ApprovalStatus: productdata.ToolCallApprovalRequired, ExecutionStatus: productdata.ToolCallExecutionBlocked}); err != nil {
		return nil, err
	}
	switch invocation.ToolName {
	case productdata.ToolNameMemorySearch:
		return e.search(ctx, invocation)
	case productdata.ToolNameMemoryList:
		return e.list(ctx, invocation)
	case productdata.ToolNameMemoryRead:
		return e.read(ctx, invocation)
	case productdata.ToolNameMemoryWrite:
		return e.write(ctx, invocation)
	case productdata.ToolNameMemoryEdit:
		return e.edit(ctx, invocation)
	case productdata.ToolNameMemoryForget:
		return e.forget(ctx, invocation)
	case productdata.ToolNameMemoryContext:
		return e.memoryContext(ctx, invocation)
	case productdata.ToolNameMemoryTimeline:
		return e.timeline(ctx, invocation)
	case productdata.ToolNameMemoryConnections:
		return e.connections(ctx, invocation)
	case productdata.ToolNameMemoryThreadSearch:
		return e.threadSearch(ctx, invocation)
	case productdata.ToolNameMemoryThreadFetch:
		return e.threadFetch(ctx, invocation)
	case productdata.ToolNameMemoryStatus:
		return e.status(ctx)
	case productdata.ToolNameNotebookRead:
		return e.notebookRead(ctx, invocation)
	case productdata.ToolNameNotebookWrite:
		return e.notebookWrite(ctx, invocation)
	case productdata.ToolNameNotebookEdit:
		return e.notebookEdit(ctx, invocation)
	case productdata.ToolNameNotebookForget:
		return e.notebookForget(ctx, invocation)
	default:
		return nil, errors.New("memory tool is not supported")
	}
}

func (e MemoryToolExecutor) ident() identity.LocalIdentity {
	if e.Ident.UserID == "" {
		return identity.LocalDevIdentity()
	}
	return e.Ident
}

func (e MemoryToolExecutor) list(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	output, err := e.Service.ListMemoryEntries(ctx, e.ident(), productdata.MemorySearchInput{
		ScopeType:      memoryScopeArg(invocation),
		ScopeID:        memoryScopeIDArg(invocation),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
		SourceType:     stringArg(invocation.ArgumentsSummary, "source_type", ""),
		Limit:          boundedInt(invocation.ArgumentsSummary, "limit", 5, 20),
		Purpose:        "tool",
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, memorySearchItemSummary(item))
	}
	return map[string]any{"tool": productdata.ToolNameMemoryList, "scope": "memory", "operation": "list", "items": items, "count": len(items), "excluded_count": output.ExcludedCount, "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) search(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	query := stringArg(invocation.ArgumentsSummary, "query", "")
	limit := boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)
	if hits, handled, err := e.externalMemorySearch(ctx, query, limit); handled {
		if err != nil {
			return nil, err
		}
		items := externalSearchItems(hits)
		return map[string]any{"tool": productdata.ToolNameMemorySearch, "scope": "memory", "operation": "search", "items": items, "count": len(items), "provider": string(hitsProvider(hits)), "redaction_applied": true}, nil
	}
	output, err := e.Service.SearchMemory(ctx, e.ident(), productdata.MemorySearchInput{
		Query:          query,
		ScopeType:      memoryScopeArg(invocation),
		ScopeID:        memoryScopeIDArg(invocation),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
		SourceType:     stringArg(invocation.ArgumentsSummary, "source_type", ""),
		Limit:          limit,
		Purpose:        "tool",
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, memorySearchItemSummary(item))
	}
	return map[string]any{"tool": productdata.ToolNameMemorySearch, "scope": "memory", "operation": "search", "items": items, "count": len(items), "excluded_count": output.ExcludedCount, "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) read(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	entryID := stringArg(invocation.ArgumentsSummary, "entry_id", "")
	if result, handled, err := e.externalMemoryRead(ctx, entryID); handled {
		return result, err
	}
	entry, err := e.Service.GetMemoryEntry(ctx, e.ident(), entryID, productdata.MemoryEntryAccessInput{
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	return memoryEntrySummary(productdata.ToolNameMemoryRead, "read", entry), nil
}

func (e MemoryToolExecutor) write(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalMemoryWrite(ctx, stringArg(invocation.ArgumentsSummary, "title", ""), stringArg(invocation.ArgumentsSummary, "content", "")); handled {
		return result, err
	}
	proposal, err := e.Service.ProposeMemoryWrite(ctx, e.ident(), productdata.ProposeMemoryWriteInput{
		ScopeType:      memoryScopeArg(invocation),
		ScopeID:        memoryScopeIDArg(invocation),
		Title:          stringArg(invocation.ArgumentsSummary, "title", ""),
		Content:        stringArg(invocation.ArgumentsSummary, "content", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_run_id", ""), invocation.RunID),
		SourceEventID:  stringArg(invocation.ArgumentsSummary, "source_event_id", ""),
		IdempotencyKey: stringArg(invocation.ArgumentsSummary, "idempotency_key", ""),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryWrite, "scope": "memory", "operation": "write_proposal", "proposal_id": proposal.ID, "status": string(proposal.Status), "title": proposal.Title, "summary": proposal.Summary, "safety_state": string(proposal.SafetyState), "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) edit(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalMemoryEdit(ctx, stringArg(invocation.ArgumentsSummary, "entry_id", ""), stringArg(invocation.ArgumentsSummary, "title", ""), stringArg(invocation.ArgumentsSummary, "content", "")); handled {
		return result, err
	}
	proposalID := stringArg(invocation.ArgumentsSummary, "proposal_id", "")
	if proposalID != "" {
		proposal, err := e.Service.UpdateMemoryWriteProposal(ctx, e.ident(), proposalID, productdata.MemoryWriteProposalUpdateInput{Title: stringArg(invocation.ArgumentsSummary, "title", ""), Summary: stringArg(invocation.ArgumentsSummary, "content", "")})
		if err != nil {
			return nil, err
		}
		return map[string]any{"tool": productdata.ToolNameMemoryEdit, "scope": "memory", "operation": "edit_proposal", "proposal_id": proposal.ID, "status": string(proposal.Status), "title": proposal.Title, "summary": proposal.Summary, "safety_state": string(proposal.SafetyState), "redaction_applied": true}, nil
	}
	proposal, err := e.Service.ProposeMemoryWrite(ctx, e.ident(), productdata.ProposeMemoryWriteInput{
		ScopeType:      memoryScopeArg(invocation),
		ScopeID:        memoryScopeIDArg(invocation),
		Title:          stringArg(invocation.ArgumentsSummary, "title", ""),
		Content:        stringArg(invocation.ArgumentsSummary, "content", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_run_id", ""), invocation.RunID),
		SourceEventID:  stringArg(invocation.ArgumentsSummary, "source_event_id", ""),
		IdempotencyKey: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "idempotency_key", ""), "memory_edit:"+stringArg(invocation.ArgumentsSummary, "entry_id", "")),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryEdit, "scope": "memory", "operation": "edit_replacement_proposal", "entry_id": stringArg(invocation.ArgumentsSummary, "entry_id", ""), "proposal_id": proposal.ID, "status": string(proposal.Status), "title": proposal.Title, "summary": proposal.Summary, "safety_state": string(proposal.SafetyState), "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) forget(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalMemoryForget(ctx, stringArg(invocation.ArgumentsSummary, "entry_id", "")); handled {
		return result, err
	}
	tombstone, err := e.Service.DeleteMemoryEntry(ctx, e.ident(), stringArg(invocation.ArgumentsSummary, "entry_id", ""), productdata.DeleteMemoryEntryInput{
		Reason:         stringArg(invocation.ArgumentsSummary, "reason", ""),
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryForget, "scope": "memory", "operation": "forget", "entry_id": tombstone.EntryID, "status": tombstone.Status, "deleted_at": tombstone.DeletedAt, "audit_event_id": tombstone.AuditEventID, "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) memoryContext(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	status, err := e.status(ctx)
	if err != nil {
		return nil, err
	}
	list, err := e.search(ctx, ToolInvocation{ThreadID: invocation.ThreadID, RunID: invocation.RunID, ToolCallID: invocation.ToolCallID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{
		"query":            firstNonEmpty(stringArg(invocation.ArgumentsSummary, "query", ""), "memory"),
		"limit":            boundedInt(invocation.ArgumentsSummary, "limit", 5, 20),
		"scope_type":       stringArg(invocation.ArgumentsSummary, "scope_type", ""),
		"scope_id":         stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		"source_thread_id": stringArg(invocation.ArgumentsSummary, "source_thread_id", ""),
		"source_run_id":    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
		"source_type":      stringArg(invocation.ArgumentsSummary, "source_type", ""),
	}})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryContext, "scope": "memory", "operation": "context", "provider": status, "items": list["items"], "count": list["count"], "redaction_applied": true}, nil
}

func hitsProvider(hits []externalMemoryHit) productdata.MemoryProviderID {
	if len(hits) == 0 {
		return ""
	}
	return hits[0].Provider
}

func (e MemoryToolExecutor) timeline(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalNowledgeTimeline(ctx, boundedInt(invocation.ArgumentsSummary, "limit", 10, 50)); handled {
		return result, err
	}
	output, err := e.Service.ListMemoryAudit(ctx, e.ident(), productdata.MemoryAuditInput{
		ThreadID:    firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID: stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
		EventType:   stringArg(invocation.ArgumentsSummary, "source_type", ""),
		Limit:       boundedInt(invocation.ArgumentsSummary, "limit", 10, 50),
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, map[string]any{"audit_id": item.ID, "event_type": item.EventType, "summary": item.Summary, "memory_entry_id": item.MemoryEntryID, "memory_proposal_id": item.MemoryProposalID, "status": item.Status, "scope_type": item.ScopeType, "thread_id": item.ThreadID, "run_id": item.RunID, "source_type": item.SourceType, "redaction_applied": item.RedactionApplied, "occurred_at": item.OccurredAt})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryTimeline, "scope": "memory", "operation": "timeline", "items": items, "count": len(items), "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) connections(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalOpenVikingConnections(ctx, stringArg(invocation.ArgumentsSummary, "entry_id", ""), boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)); handled {
		return result, err
	}
	if result, handled, err := e.externalNowledgeConnections(ctx, stringArg(invocation.ArgumentsSummary, "entry_id", ""), boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)); handled {
		return result, err
	}
	query := stringArg(invocation.ArgumentsSummary, "query", "")
	if query == "" && stringArg(invocation.ArgumentsSummary, "entry_id", "") != "" {
		entry, err := e.Service.GetMemoryEntry(ctx, e.ident(), stringArg(invocation.ArgumentsSummary, "entry_id", ""), productdata.MemoryEntryAccessInput{ScopeType: productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")), ScopeID: stringArg(invocation.ArgumentsSummary, "scope_id", ""), SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID), SourceRunID: stringArg(invocation.ArgumentsSummary, "source_run_id", "")})
		if err != nil {
			return nil, err
		}
		query = entry.Title + " " + entry.Summary
	}
	result, err := e.search(ctx, ToolInvocation{ThreadID: invocation.ThreadID, RunID: invocation.RunID, ToolCallID: invocation.ToolCallID, ToolName: productdata.ToolNameMemorySearch, ArgumentsSummary: map[string]any{"query": query, "limit": boundedInt(invocation.ArgumentsSummary, "limit", 5, 20), "scope_type": stringArg(invocation.ArgumentsSummary, "scope_type", ""), "scope_id": stringArg(invocation.ArgumentsSummary, "scope_id", ""), "source_thread_id": stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), "source_run_id": stringArg(invocation.ArgumentsSummary, "source_run_id", "")}})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryConnections, "scope": "memory", "operation": "connections", "target_entry_id": stringArg(invocation.ArgumentsSummary, "entry_id", ""), "items": result["items"], "count": result["count"], "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) threadSearch(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalNowledgeThreadSearch(ctx, stringArg(invocation.ArgumentsSummary, "query", ""), boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)); handled {
		return result, err
	}
	threads, err := e.Service.ListThreads(ctx, e.ident(), false)
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(stringArg(invocation.ArgumentsSummary, "query", ""))
	limit := boundedInt(invocation.ArgumentsSummary, "limit", 5, 20)
	items := []map[string]any{}
	for _, thread := range threads {
		if len(items) >= limit {
			break
		}
		messages, _ := e.Service.ListMessages(ctx, e.ident(), thread.ID)
		matched := strings.Contains(strings.ToLower(thread.Title), query)
		excerpt := ""
		for _, message := range messages {
			if strings.Contains(strings.ToLower(message.Content), query) {
				matched = true
				excerpt = safeMemoryExcerpt(message.Content)
				break
			}
		}
		if matched {
			items = append(items, map[string]any{"thread_id": thread.ID, "title": thread.Title, "mode": string(thread.Mode), "excerpt": excerpt, "updated_at": thread.UpdatedAt, "redaction_applied": true})
		}
	}
	return map[string]any{"tool": productdata.ToolNameMemoryThreadSearch, "scope": "memory", "operation": "thread_search", "items": items, "count": len(items), "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) threadFetch(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if result, handled, err := e.externalNowledgeThreadFetch(ctx, stringArg(invocation.ArgumentsSummary, "thread_id", ""), boundedInt(invocation.ArgumentsSummary, "limit", 10, 50)); handled {
		return result, err
	}
	thread, err := e.Service.GetThread(ctx, e.ident(), stringArg(invocation.ArgumentsSummary, "thread_id", ""))
	if err != nil {
		return nil, err
	}
	messages, err := e.Service.ListMessages(ctx, e.ident(), thread.ID)
	if err != nil {
		return nil, err
	}
	limit := boundedInt(invocation.ArgumentsSummary, "limit", 10, 50)
	if len(messages) > limit {
		messages = messages[len(messages)-limit:]
	}
	items := make([]map[string]any, 0, len(messages))
	for _, message := range messages {
		items = append(items, map[string]any{"message_id": message.ID, "role": string(message.Role), "excerpt": safeMemoryExcerpt(message.Content), "created_at": message.CreatedAt, "redaction_applied": true})
	}
	return map[string]any{"tool": productdata.ToolNameMemoryThreadFetch, "scope": "memory", "operation": "thread_fetch", "thread_id": thread.ID, "title": thread.Title, "items": items, "count": len(items), "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) status(ctx context.Context) (map[string]any, error) {
	status, err := e.Service.GetMemoryProviderStatus(ctx, e.ident())
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameMemoryStatus, "scope": "memory", "operation": "status", "enabled": status.Enabled, "provider": string(status.Provider), "label": status.Label, "state": string(status.State), "configured": status.Configured, "commit_after_run": status.CommitAfterRun, "diagnostic_code": status.Diagnostic.Code, "diagnostic_message": status.Diagnostic.Message, "redaction_applied": true}, nil
}

func (e MemoryToolExecutor) notebookRead(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	entry, err := e.Service.GetMemoryEntry(ctx, e.ident(), stringArg(invocation.ArgumentsSummary, "entry_id", ""), productdata.MemoryEntryAccessInput{
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	if entry.SourceEventID != "notebook" {
		return nil, errors.New("notebook entry is not available")
	}
	if entry.Status != productdata.MemoryEntryApproved {
		return nil, errors.New("notebook entry is not available")
	}
	return memoryEntrySummary(productdata.ToolNameNotebookRead, "notebook_read", entry), nil
}

func (e MemoryToolExecutor) notebookWrite(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	entry, err := e.Service.CreateMemoryEntry(ctx, e.ident(), productdata.CreateMemoryEntryInput{
		ScopeType:      memoryScopeArg(invocation),
		ScopeID:        memoryScopeIDArg(invocation),
		Title:          stringArg(invocation.ArgumentsSummary, "title", ""),
		Content:        stringArg(invocation.ArgumentsSummary, "content", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_run_id", ""), invocation.RunID),
		SourceEventID:  "notebook",
	})
	if err != nil {
		return nil, err
	}
	return memoryEntrySummary(productdata.ToolNameNotebookWrite, "notebook_write", entry), nil
}

func (e MemoryToolExecutor) notebookEdit(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	entryID := stringArg(invocation.ArgumentsSummary, "entry_id", "")
	oldEntry, err := e.Service.GetMemoryEntry(ctx, e.ident(), entryID, productdata.MemoryEntryAccessInput{
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	if oldEntry.SourceEventID != "notebook" {
		return nil, errors.New("notebook entry is not available")
	}
	tombstone, err := e.Service.DeleteMemoryEntry(ctx, e.ident(), entryID, productdata.DeleteMemoryEntryInput{
		Reason:         "replaced by notebook edit",
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	newEntry, err := e.Service.CreateMemoryEntry(ctx, e.ident(), productdata.CreateMemoryEntryInput{
		ScopeType:      oldEntry.ScopeType,
		ScopeID:        oldEntry.ScopeID,
		Title:          stringArg(invocation.ArgumentsSummary, "title", ""),
		Content:        stringArg(invocation.ArgumentsSummary, "content", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_run_id", ""), invocation.RunID),
		SourceEventID:  "notebook",
	})
	if err != nil {
		return nil, err
	}
	result := memoryEntrySummary(productdata.ToolNameNotebookEdit, "notebook_edit", newEntry)
	result["replaced_entry_id"] = tombstone.EntryID
	return result, nil
}

func (e MemoryToolExecutor) notebookForget(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	entry, err := e.Service.GetMemoryEntry(ctx, e.ident(), stringArg(invocation.ArgumentsSummary, "entry_id", ""), productdata.MemoryEntryAccessInput{
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	if entry.SourceEventID != "notebook" {
		return nil, errors.New("notebook entry is not available")
	}
	tombstone, err := e.Service.DeleteMemoryEntry(ctx, e.ident(), entry.ID, productdata.DeleteMemoryEntryInput{
		Reason:         stringArg(invocation.ArgumentsSummary, "reason", ""),
		ScopeType:      productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")),
		ScopeID:        stringArg(invocation.ArgumentsSummary, "scope_id", ""),
		SourceThreadID: firstNonEmpty(stringArg(invocation.ArgumentsSummary, "source_thread_id", ""), invocation.ThreadID),
		SourceRunID:    stringArg(invocation.ArgumentsSummary, "source_run_id", ""),
	})
	if err != nil {
		return nil, err
	}
	return map[string]any{"tool": productdata.ToolNameNotebookForget, "scope": "memory", "operation": "notebook_forget", "entry_id": tombstone.EntryID, "status": tombstone.Status, "deleted_at": tombstone.DeletedAt, "audit_event_id": tombstone.AuditEventID, "redaction_applied": true}, nil
}

func memorySearchItemSummary(item productdata.MemorySearchResult) map[string]any {
	return map[string]any{"entry_id": item.ID, "title": item.Title, "summary": item.Summary, "scope_type": string(item.ScopeType), "scope_id": item.ScopeID, "status": item.Status, "safety_state": item.SafetyState, "source_thread_id": item.SourceThreadID, "source_run_id": item.SourceRunID, "source_type": item.SourceType, "rank_reason": item.RankReason, "redaction_applied": true}
}

func memoryEntrySummary(tool string, operation string, entry productdata.MemoryEntry) map[string]any {
	return map[string]any{"tool": tool, "scope": "memory", "operation": operation, "entry_id": entry.ID, "title": entry.Title, "summary": entry.Summary, "scope_type": string(entry.ScopeType), "scope_id": entry.ScopeID, "status": string(entry.Status), "safety_state": string(entry.SafetyState), "source_thread_id": entry.SourceThreadID, "source_run_id": entry.SourceRunID, "redaction_applied": true}
}

func safeMemoryExcerpt(content string) string {
	excerpt := productdata.RedactEventText(strings.TrimSpace(content))
	runes := []rune(excerpt)
	if len(runes) > 240 {
		excerpt = string(runes[:240])
	}
	return excerpt
}

func memoryScopeArg(invocation ToolInvocation) productdata.MemoryScopeType {
	if scopeType := productdata.MemoryScopeType(stringArg(invocation.ArgumentsSummary, "scope_type", "")); scopeType != "" {
		return scopeType
	}
	return productdata.MemoryScopeThread
}

func memoryScopeIDArg(invocation ToolInvocation) string {
	if scopeID := stringArg(invocation.ArgumentsSummary, "scope_id", ""); scopeID != "" {
		return scopeID
	}
	return invocation.ThreadID
}
