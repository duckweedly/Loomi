package runtime

import (
	"context"
	"fmt"
	"strings"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

const postRunMemoryContentLimit = 480

const (
	eventMemoryProviderCommitCompleted = "memory_provider_commit_completed"
	eventMemoryProviderCommitFailed    = "memory_provider_commit_failed"
)

func proposePostRunMemory(ctx context.Context, svc productdata.Service, ident identity.LocalIdentity, runID string) error {
	if svc == nil {
		return nil
	}
	run, err := svc.GetRun(ctx, ident, runID)
	if err != nil {
		return err
	}
	if run.Status != productdata.RunStatusCompleted {
		return nil
	}
	status, err := svc.GetMemoryProviderStatus(ctx, ident)
	if err != nil {
		return err
	}
	if !status.Enabled || !status.CommitAfterRun || !status.Configured {
		return nil
	}
	message, ok, err := postRunAssistantMessage(ctx, svc, ident, run)
	if err != nil || !ok {
		return err
	}
	content := postRunMemoryContent(message.Content)
	if content == "" {
		return nil
	}
	if status.Provider == productdata.MemoryProviderOpenViking || status.Provider == productdata.MemoryProviderNowledge {
		return commitPostRunExternalMemory(ctx, svc, ident, run, status, content)
	}
	_, err = svc.ProposeMemoryWrite(ctx, ident, productdata.ProposeMemoryWriteInput{
		ScopeType:      productdata.MemoryScopeThread,
		ScopeID:        run.ThreadID,
		Title:          "Run outcome",
		Content:        content,
		SourceThreadID: run.ThreadID,
		SourceRunID:    run.ID,
		IdempotencyKey: postRunMemoryIdempotencyKey(run.ID),
	})
	return err
}

func commitPostRunExternalMemory(ctx context.Context, svc productdata.Service, ident identity.LocalIdentity, run productdata.Run, status productdata.MemoryProviderStatus, content string) error {
	committed, err := svc.HasRunEventType(ctx, ident, run.ID, eventMemoryProviderCommitCompleted)
	if err != nil {
		return err
	}
	if committed {
		return nil
	}
	result, handled, err := MemoryToolExecutor{Service: svc, Ident: ident}.externalMemoryWrite(ctx, "Run outcome", content)
	if !handled {
		return nil
	}
	if err != nil {
		_, _ = svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: eventMemoryProviderCommitFailed, Summary: "Memory provider commit failed", Metadata: map[string]any{"provider": string(status.Provider), "error_code": "memory_provider_commit_failed"}})
		return nil
	}
	_, err = svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{Category: productdata.RunEventCategoryProgress, Type: eventMemoryProviderCommitCompleted, Summary: "Memory provider commit completed", Metadata: productdata.RedactEventMetadata(map[string]any{"provider": string(status.Provider), "operation": result["operation"], "entry_id": result["entry_id"], "status": result["status"], "redaction_applied": true})})
	return err
}

func postRunAssistantMessage(ctx context.Context, svc productdata.Service, ident identity.LocalIdentity, run productdata.Run) (productdata.Message, bool, error) {
	messages, err := svc.ListMessages(ctx, ident, run.ThreadID)
	if err != nil {
		return productdata.Message{}, false, err
	}
	for index := len(messages) - 1; index >= 0; index-- {
		message := messages[index]
		if message.Role != productdata.MessageRoleAssistant {
			continue
		}
		if metadataString(message.Metadata, "run_id") == run.ID {
			return message, true, nil
		}
	}
	return productdata.Message{}, false, nil
}

func postRunMemoryContent(content string) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	runes := []rune(content)
	if len(runes) > postRunMemoryContentLimit {
		content = string(runes[:postRunMemoryContentLimit])
	}
	return "Assistant outcome: " + content
}

func postRunMemoryIdempotencyKey(runID string) string {
	return fmt.Sprintf("post_run_memory:%s", strings.TrimSpace(runID))
}
