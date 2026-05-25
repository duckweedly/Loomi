package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/sheridiany/loomi/internal/config"
	"github.com/sheridiany/loomi/internal/db"
	"github.com/sheridiany/loomi/internal/diagnostics"
	"github.com/sheridiany/loomi/internal/identity"
	"github.com/sheridiany/loomi/internal/productdata"
)

const (
	seedThreadID        = "thr_local_demo"
	seedThreadTitle     = "M3 local demo thread"
	seedMessageID       = "msg_local_demo_001"
	seedMessageContent  = "This is a durable local user message seeded for M3."
	seedClientMessageID = "seed-m3-local-demo-message"

	m17SeedScenario        = "m17-work-artifact"
	m17SeedThreadID        = "thr_m17_work_artifact"
	m17SeedThreadTitle     = "M17 Work artifact evidence"
	m17SeedMessageID       = "msg_m17_work_artifact_001"
	m17SeedMessageContent  = "Close out M17 Work artifact evidence with real thread/message/run/event replay."
	m17SeedClientMessageID = "seed-m17-work-artifact-evidence"
	m17SeedEventType       = "work.plan.updated"
)

type seedResult struct {
	ThreadID  string
	MessageID string
	RunID     string
	EventID   string
}

func main() {
	logger := diagnostics.NewJSONLogger(os.Stdout, slog.LevelInfo)
	opID := diagnostics.NewOperationID("seed")
	cfg, err := config.Load()
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	level, err := diagnostics.ParseLevel(cfg.LogLevel)
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	logger = diagnostics.NewJSONLogger(os.Stdout, level)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.ReadinessTimeoutSeconds)*time.Second)
	defer cancel()
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	defer pool.Close()
	result, err := runSeed(ctx, productdata.NewPostgresRepository(pool), identity.LocalDevIdentity())
	if err != nil {
		logger.Error("m3 seed failed", "component", "seed", "operation_id", opID, "error", diagnostics.Redact(err.Error()))
		os.Exit(1)
	}
	logger.Info("seed complete", "component", "seed", "operation_id", opID, "thread_id", result.ThreadID, "message_id", result.MessageID, "run_id", result.RunID, "event_id", result.EventID)
}

func runSeed(ctx context.Context, svc productdata.SeedService, ident identity.LocalIdentity) (seedResult, error) {
	if os.Getenv("LOOMI_SEED_SCENARIO") == m17SeedScenario {
		return runM17WorkArtifactSeed(ctx, svc, ident)
	}
	if _, err := svc.CurrentIdentity(ctx, ident); err != nil {
		return seedResult{}, err
	}
	thread, err := svc.UpsertSeedThread(ctx, ident, productdata.SeedThreadInput{ID: seedThreadID, Title: seedThreadTitle, Mode: productdata.ThreadModeChat})
	if err != nil {
		return seedResult{}, err
	}
	message, err := svc.UpsertSeedMessage(ctx, ident, productdata.SeedMessageInput{ID: seedMessageID, ThreadID: thread.ID, Content: seedMessageContent, ClientMessageID: seedClientMessageID})
	if err != nil {
		return seedResult{}, err
	}
	return seedResult{ThreadID: thread.ID, MessageID: message.ID}, nil
}

func runM17WorkArtifactSeed(ctx context.Context, svc productdata.SeedService, ident identity.LocalIdentity) (seedResult, error) {
	if _, err := svc.CurrentIdentity(ctx, ident); err != nil {
		return seedResult{}, err
	}
	thread, err := svc.UpsertSeedThread(ctx, ident, productdata.SeedThreadInput{ID: m17SeedThreadID, Title: m17SeedThreadTitle, Mode: productdata.ThreadModeWork})
	if err != nil {
		return seedResult{}, err
	}
	message, err := svc.UpsertSeedMessage(ctx, ident, productdata.SeedMessageInput{ID: m17SeedMessageID, ThreadID: thread.ID, Content: m17SeedMessageContent, ClientMessageID: m17SeedClientMessageID})
	if err != nil {
		return seedResult{}, err
	}
	run, event, err := ensureM17WorkArtifactEvent(ctx, svc, ident, thread.ID)
	if err != nil {
		return seedResult{}, err
	}
	return seedResult{ThreadID: thread.ID, MessageID: message.ID, RunID: run.ID, EventID: event.ID}, nil
}

func ensureM17WorkArtifactEvent(ctx context.Context, svc productdata.SeedService, ident identity.LocalIdentity, threadID string) (productdata.Run, productdata.RunEvent, error) {
	run, err := svc.GetCurrentRun(ctx, ident, threadID)
	if err == nil {
		if event, ok, err := findM17WorkArtifactEvent(ctx, svc, ident, run.ID); err != nil {
			return productdata.Run{}, productdata.RunEvent{}, err
		} else if ok {
			return run, event, nil
		}
		if productdata.IsRunTerminal(run.Status) {
			run, err = svc.StartRun(ctx, ident, threadID, productdata.StartRunInput{Source: productdata.RunSourceLocalSimulated, ScriptName: m17SeedScenario})
			if err != nil {
				return productdata.Run{}, productdata.RunEvent{}, err
			}
		}
	} else if productdata.ErrorCode(err) == productdata.CodeRunNotFound {
		run, err = svc.StartRun(ctx, ident, threadID, productdata.StartRunInput{Source: productdata.RunSourceLocalSimulated, ScriptName: m17SeedScenario})
		if err != nil {
			return productdata.Run{}, productdata.RunEvent{}, err
		}
	} else {
		return productdata.Run{}, productdata.RunEvent{}, err
	}

	event, err := svc.AppendRunEvent(ctx, ident, run.ID, productdata.AppendRunEventInput{
		Category: productdata.RunEventCategoryProgress,
		Type:     m17SeedEventType,
		Summary:  "M17 Work artifact evidence linked",
		Metadata: m17WorkArtifactMetadata(threadID, run.ID),
	})
	if err != nil {
		return productdata.Run{}, productdata.RunEvent{}, err
	}
	return run, event, nil
}

func findM17WorkArtifactEvent(ctx context.Context, svc productdata.SeedService, ident identity.LocalIdentity, runID string) (productdata.RunEvent, bool, error) {
	events, err := svc.ListRunEvents(ctx, ident, runID, 0)
	if err != nil {
		return productdata.RunEvent{}, false, err
	}
	for _, event := range events {
		if event.Type == m17SeedEventType && event.Metadata["m17_seed"] == m17SeedScenario {
			return event, true, nil
		}
	}
	return productdata.RunEvent{}, false, nil
}

func m17WorkArtifactMetadata(threadID string, runID string) map[string]any {
	return map[string]any{
		"m17_seed":  m17SeedScenario,
		"work_goal": "Close out M17 Work artifact evidence with repeatable real event replay",
		"work_steps": []any{
			map[string]any{"id": "m17-step-seed", "title": "Create local evidence seed", "status": "completed", "summary": "Reuse existing thread/message/run/event services."},
			map[string]any{"id": "m17-step-render", "title": "Render artifact evidence", "status": "running", "summary": "Project safe metadata in Work Plan View."},
			map[string]any{"id": "m17-step-smoke", "title": "Browser smoke Work and Chat modes", "status": "pending", "summary": "Verify Work evidence and Chat isolation."},
		},
		"work_artifacts": []any{
			map[string]any{
				"id":                "m17-artifact-evidence",
				"title":             "M17 Work artifact evidence",
				"type":              "markdown",
				"source_thread_id":  threadID,
				"source_run_id":     runID,
				"summary":           "Safe metadata-only artifact evidence from local seed.",
				"created_at":        "2026-05-25T00:00:00Z",
				"updated_at":        "2026-05-25T00:00:00Z",
				"redaction_applied": true,
				"command":           "open /Users/xuean/private/artifact.md",
				"private_path":      "/Users/xuean/private/artifact.md",
				"authorization":     "Bearer sk-m17-secret",
				"tool_output":       "provider trace token",
			},
		},
	}
}
