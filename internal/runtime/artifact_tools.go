package runtime

import (
	"context"
	"errors"
	"fmt"

	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

const defaultArtifactMaxBytes = 32 * 1024

type ArtifactToolExecutor struct {
	Artifacts productdata.ArtifactService
}

func ArtifactToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		{Name: productdata.ToolNameArtifactCreateText, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameArtifactRead, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
		{Name: productdata.ToolNameArtifactList, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyNoSideEffectInternal, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
	}
}

func (e ArtifactToolExecutor) Execute(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	if e.Artifacts == nil {
		return nil, errors.New("artifact service is unavailable")
	}
	switch invocation.ToolName {
	case productdata.ToolNameArtifactCreateText:
		return e.createText(ctx, invocation)
	case productdata.ToolNameArtifactRead:
		return e.read(ctx, invocation)
	case productdata.ToolNameArtifactList:
		return e.list(ctx, invocation)
	default:
		return nil, errors.New("artifact tool is not supported")
	}
}

func (e ArtifactToolExecutor) createText(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	title, _ := invocation.ArgumentsSummary["title"].(string)
	content, _ := invocation.ArgumentsSummary["content"].(string)
	artifact, err := e.Artifacts.CreateArtifact(ctx, identity.LocalDevIdentity(), productdata.CreateArtifactInput{
		ThreadID: invocation.ThreadID,
		RunID:    invocation.RunID,
		Title:    title,
		Content:  content,
		MaxBytes: boundedInt(invocation.ArgumentsSummary, "max_bytes", defaultArtifactMaxBytes, defaultArtifactMaxBytes),
	})
	if err != nil {
		return nil, err
	}
	return artifactSummary(productdata.ToolNameArtifactCreateText, "create_text", artifact), nil
}

func (e ArtifactToolExecutor) read(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	artifactID, _ := invocation.ArgumentsSummary["artifact_id"].(string)
	artifact, err := e.Artifacts.ReadArtifact(ctx, identity.LocalDevIdentity(), productdata.ReadArtifactInput{
		ThreadID:   invocation.ThreadID,
		ArtifactID: artifactID,
		MaxBytes:   boundedInt(invocation.ArgumentsSummary, "max_bytes", defaultArtifactMaxBytes, defaultArtifactMaxBytes),
	})
	if err != nil {
		return nil, err
	}
	return artifactSummary(productdata.ToolNameArtifactRead, "read", artifact), nil
}

func (e ArtifactToolExecutor) list(ctx context.Context, invocation ToolInvocation) (map[string]any, error) {
	artifacts, err := e.Artifacts.ListArtifacts(ctx, identity.LocalDevIdentity(), productdata.ListArtifactsInput{
		ThreadID: invocation.ThreadID,
		Limit:    boundedInt(invocation.ArgumentsSummary, "limit", 20, 50),
	})
	if err != nil {
		return nil, err
	}
	items := make([]map[string]any, 0, len(artifacts))
	for _, artifact := range artifacts {
		items = append(items, artifactSummary("", "list_item", artifact))
	}
	return map[string]any{
		"tool":              productdata.ToolNameArtifactList,
		"scope":             "artifact",
		"operation":         "list",
		"artifacts":         items,
		"count":             len(items),
		"redaction_applied": false,
	}, nil
}

func artifactSummary(tool string, operation string, artifact productdata.Artifact) map[string]any {
	summary := map[string]any{
		"scope":             "artifact",
		"operation":         operation,
		"artifact_id":       artifact.ID,
		"title":             artifact.Title,
		"artifact_type":     artifact.ArtifactType,
		"content_bytes":     artifact.ContentBytes,
		"text_excerpt":      artifact.TextExcerpt,
		"truncated":         artifact.Truncated,
		"redaction_applied": false,
	}
	if tool != "" {
		summary["tool"] = tool
	}
	if artifact.RunID != "" {
		summary["run_id"] = artifact.RunID
	}
	if artifact.ThreadID != "" {
		summary["thread_id"] = artifact.ThreadID
	}
	return summary
}

func artifactIDFromSummary(result map[string]any) (string, error) {
	id, _ := result["artifact_id"].(string)
	if id == "" {
		return "", fmt.Errorf("artifact id missing")
	}
	return id, nil
}
