package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		{Name: productdata.ToolNameArtifactCreateVisual, ApprovalPolicy: ToolApprovalAlwaysRequired, SafetyClass: ToolSafetyWorkspaceMutation, Source: ToolSourceInternal, ExecutionState: ToolExecutionAllowlisted},
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
		return e.create(ctx, invocation, "text", "create_text")
	case productdata.ToolNameArtifactCreateVisual:
		return e.create(ctx, invocation, "visual", "create_visual")
	case productdata.ToolNameArtifactRead:
		return e.read(ctx, invocation)
	case productdata.ToolNameArtifactList:
		return e.list(ctx, invocation)
	default:
		return nil, errors.New("artifact tool is not supported")
	}
}

func (e ArtifactToolExecutor) create(ctx context.Context, invocation ToolInvocation, artifactType string, operation string) (map[string]any, error) {
	title, _ := invocation.ArgumentsSummary["title"].(string)
	content, _ := invocation.ArgumentsSummary["content"].(string)
	artifact, err := e.Artifacts.CreateArtifact(ctx, identity.LocalDevIdentity(), productdata.CreateArtifactInput{
		ThreadID:     invocation.ThreadID,
		RunID:        invocation.RunID,
		Title:        title,
		ArtifactType: artifactType,
		Content:      content,
		MaxBytes:     boundedInt(invocation.ArgumentsSummary, "max_bytes", defaultArtifactMaxBytes, defaultArtifactMaxBytes),
	})
	if err != nil {
		return nil, err
	}
	summary := artifactSummary(invocation.ToolName, operation, artifact)
	summary["artifacts"] = []map[string]any{artifactRefSummary(artifact, artifactRefMetadataFromArguments(invocation.ArgumentsSummary))}
	return summary, nil
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
	summary := artifactSummary(productdata.ToolNameArtifactRead, "read", artifact)
	summary["artifacts"] = []map[string]any{artifactRefSummary(artifact, artifactRefMetadata{})}
	return summary, nil
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
		items = append(items, artifactRefSummary(artifact, artifactRefMetadata{}))
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

type artifactRefMetadata struct {
	Filename string
	MIMEType string
	Display  string
}

func artifactRefMetadataFromArguments(args map[string]any) artifactRefMetadata {
	filename, _ := args["filename"].(string)
	mimeType, _ := args["mime_type"].(string)
	display, _ := args["display"].(string)
	return artifactRefMetadata{
		Filename: strings.TrimSpace(filename),
		MIMEType: strings.TrimSpace(mimeType),
		Display:  strings.TrimSpace(display),
	}
}

func artifactRefSummary(artifact productdata.Artifact, metadata artifactRefMetadata) map[string]any {
	display := metadata.Display
	if display == "" {
		display = "panel"
	}
	mimeType := metadata.MIMEType
	if mimeType == "" {
		mimeType = defaultArtifactMIMEType(metadata.Filename, artifact.ArtifactType)
	}
	item := map[string]any{
		"key":           artifact.ID,
		"artifact_id":   artifact.ID,
		"title":         artifact.Title,
		"mime_type":     mimeType,
		"display":       display,
		"size":          artifact.ContentBytes,
		"content_bytes": artifact.ContentBytes,
		"text_excerpt":  artifact.TextExcerpt,
	}
	if metadata.Filename != "" {
		item["filename"] = metadata.Filename
	}
	if artifact.ArtifactType == "visual" {
		item["content"] = artifact.Content
	}
	return item
}

func defaultArtifactMIMEType(filename string, artifactType string) string {
	lower := strings.ToLower(strings.TrimSpace(filename))
	if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".markdown") {
		return "text/markdown"
	}
	if strings.HasSuffix(lower, ".svg") {
		return "image/svg+xml"
	}
	if strings.HasSuffix(lower, ".html") || strings.HasSuffix(lower, ".htm") {
		return "text/html"
	}
	if artifactType == "visual" {
		return "image/svg+xml"
	}
	return "text/plain"
}

func artifactIDFromSummary(result map[string]any) (string, error) {
	id, _ := result["artifact_id"].(string)
	if id == "" {
		return "", fmt.Errorf("artifact id missing")
	}
	return id, nil
}
