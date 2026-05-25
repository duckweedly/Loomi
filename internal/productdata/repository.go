package productdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sheridiany/loomi/internal/identity"
)

type PostgresRepository struct {
	Pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{Pool: pool}
}

func (r *PostgresRepository) CurrentIdentity(ctx context.Context, ident identity.LocalIdentity) (User, error) {
	return r.ensureUser(ctx, ident)
}

func (r *PostgresRepository) CreateThread(ctx context.Context, ident identity.LocalIdentity, input CreateThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	threadID := NewThreadID()
	personaID := strings.TrimSpace(input.PersonaID)
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Thread{}, err
	}
	defer tx.Rollback(ctx)
	if err := validatePersonaReferenceTx(ctx, tx, personaID); err != nil {
		return Thread{}, err
	}
	row := tx.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status, persona_id) values ($1, $2, $3, $4, $5, nullif($6, '')) returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, threadID, user.ID, title, input.Mode, ThreadLifecycleActive, personaID)
	thread, err := scanThread(row)
	if err != nil {
		return Thread{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func (r *PostgresRepository) UpsertSeedThread(ctx context.Context, ident identity.LocalIdentity, input SeedThreadInput) (Thread, error) {
	title, err := NormalizeThreadTitle(input.Title)
	if err != nil {
		return Thread{}, err
	}
	if err := ValidateThreadMode(input.Mode); err != nil {
		return Thread{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status) values ($1, $2, $3, $4, 'active') on conflict (id) do update set title=excluded.title, mode=excluded.mode, lifecycle_status='active', archived_at=null, updated_at=now() returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, input.ID, user.ID, title, input.Mode)
	return scanThread(row)
}

func (r *PostgresRepository) ListThreads(ctx context.Context, ident identity.LocalIdentity, includeArchived bool) ([]Thread, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	query := `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where user_id=$1`
	args := []any{user.ID}
	if !includeArchived {
		query += ` and lifecycle_status='active'`
	}
	query += ` order by updated_at desc, id desc`
	rows, err := r.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var threads []Thread
	for rows.Next() {
		thread, err := scanThread(rows)
		if err != nil {
			return nil, err
		}
		threads = append(threads, thread)
	}
	return threads, rows.Err()
}

func (r *PostgresRepository) GetThread(ctx context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at from threads where id=$1 and user_id=$2`, threadID, user.ID)
	thread, err := scanThread(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Thread{}, NewError(CodeThreadNotFound, "Thread not found.")
	}
	return thread, err
}

func (r *PostgresRepository) UpdateThread(ctx context.Context, ident identity.LocalIdentity, threadID string, input UpdateThreadInput) (Thread, error) {
	current, err := r.GetThread(ctx, ident, threadID)
	if err != nil {
		return Thread{}, err
	}
	title := current.Title
	mode := current.Mode
	personaID := current.PersonaID
	if input.Title != nil {
		normalized, err := NormalizeThreadTitle(*input.Title)
		if err != nil {
			return Thread{}, err
		}
		title = normalized
	}
	if input.Mode != nil {
		if err := ValidateThreadMode(*input.Mode); err != nil {
			return Thread{}, err
		}
		mode = *input.Mode
	}
	if input.PersonaID != nil {
		personaID = strings.TrimSpace(*input.PersonaID)
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Thread{}, err
	}
	defer tx.Rollback(ctx)
	if err := validatePersonaReferenceTx(ctx, tx, personaID); err != nil {
		return Thread{}, err
	}
	row := tx.QueryRow(ctx, `update threads set title=$1, mode=$2, persona_id=nullif($5, ''), updated_at=now() where id=$3 and user_id=$4 returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, title, mode, threadID, current.UserID, personaID)
	thread, err := scanThread(row)
	if err != nil {
		return Thread{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func (r *PostgresRepository) ArchiveThread(ctx context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	current, err := r.GetThread(ctx, ident, threadID)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `update threads set lifecycle_status='archived', archived_at=coalesce(archived_at, now()), updated_at=now() where id=$1 and user_id=$2 returning id, user_id, title, mode, lifecycle_status, coalesce(persona_id, ''), created_at, updated_at, archived_at`, threadID, current.UserID)
	return scanThread(row)
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, ident identity.LocalIdentity, threadID string, input CreateMessageInput) (Message, bool, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, false, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, false, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, false, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Message{}, false, err
	}
	defer tx.Rollback(ctx)
	var threadUserID string
	if err := tx.QueryRow(ctx, `select user_id from threads where id=$1 and user_id=$2`, threadID, user.ID).Scan(&threadUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Message{}, false, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, false, err
	}
	if clientMessageID != nil {
		message, err := scanMessage(tx.QueryRow(ctx, `select id, thread_id, user_id, role, content, metadata, client_message_id, created_at from messages where thread_id=$1 and user_id=$2 and client_message_id=$3`, threadID, user.ID, *clientMessageID))
		if err == nil {
			return message, false, tx.Commit(ctx)
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return Message{}, false, err
		}
	}
	messageID := NewMessageID()
	message, err := scanMessage(tx.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'user', $4, '{}'::jsonb, $5) returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, messageID, threadID, user.ID, content, clientMessageID))
	if err != nil {
		return Message{}, false, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Message{}, false, err
	}
	return message, true, tx.Commit(ctx)
}

func (r *PostgresRepository) AppendAssistantMessage(ctx context.Context, ident identity.LocalIdentity, threadID string, input AppendAssistantMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, err
	}
	metadata := RedactEventMetadata(input.Metadata)
	if runID, ok := metadata["run_id"].(string); ok && runID != "" {
		var existing string
		err := r.Pool.QueryRow(ctx, `select id from messages where thread_id=$1 and user_id=$2 and role='assistant' and metadata->>'run_id'=$3 limit 1`, threadID, user.ID, runID).Scan(&existing)
		if err == nil {
			return Message{}, NewError(CodeInvalidRequest, "Assistant message already exists for run.")
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return Message{}, err
		}
	}
	message, err := scanMessage(r.Pool.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'assistant', $4, $5, null) returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, NewMessageID(), threadID, user.ID, content, mustJSON(metadata)))
	if err != nil {
		if strings.Contains(err.Error(), "foreign key") {
			return Message{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, err
	}
	if _, err := r.Pool.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Message{}, err
	}
	return message, nil
}

func (r *PostgresRepository) UpsertSeedMessage(ctx context.Context, ident identity.LocalIdentity, input SeedMessageInput) (Message, error) {
	content, err := NormalizeMessageContent(input.Content)
	if err != nil {
		return Message{}, err
	}
	clientMessageID, err := NormalizeClientMessageID(input.ClientMessageID)
	if err != nil {
		return Message{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Message{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Message{}, err
	}
	defer tx.Rollback(ctx)
	var threadUserID string
	if err := tx.QueryRow(ctx, `select user_id from threads where id=$1 and user_id=$2`, input.ThreadID, user.ID).Scan(&threadUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Message{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Message{}, err
	}
	message, err := scanMessage(tx.QueryRow(ctx, `insert into messages (id, thread_id, user_id, role, content, metadata, client_message_id) values ($1, $2, $3, 'user', $4, '{}'::jsonb, $5) on conflict (id) do update set content=excluded.content, client_message_id=excluded.client_message_id returning id, thread_id, user_id, role, content, metadata, client_message_id, created_at`, input.ID, input.ThreadID, user.ID, content, clientMessageID))
	if err != nil {
		return Message{}, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, input.ThreadID, user.ID); err != nil {
		return Message{}, err
	}
	return message, tx.Commit(ctx)
}

func (r *PostgresRepository) ListMessages(ctx context.Context, ident identity.LocalIdentity, threadID string) ([]Message, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from threads where id=$1 and user_id=$2)`, threadID, user.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewError(CodeThreadNotFound, "Thread not found.")
	}
	rows, err := r.Pool.Query(ctx, `select id, thread_id, user_id, role, content, metadata, client_message_id, created_at from messages where thread_id=$1 and user_id=$2 order by created_at asc, id asc`, threadID, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var messages []Message
	for rows.Next() {
		message, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	return messages, rows.Err()
}

func (r *PostgresRepository) StartRun(ctx context.Context, ident identity.LocalIdentity, threadID string, input StartRunInput) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return Run{}, err
	}
	defer tx.Rollback(ctx)
	var threadUserID, threadPersonaID string
	if err := tx.QueryRow(ctx, `select user_id, coalesce(persona_id, '') from threads where id=$1 and user_id=$2 and lifecycle_status='active'`, threadID, user.ID).Scan(&threadUserID, &threadPersonaID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Run{}, err
	}
	source, err := NormalizeRunSource(input.Source)
	if err != nil {
		return Run{}, err
	}
	snapshot, err := r.resolvePersonaSnapshotTx(ctx, tx, threadPersonaID, input.PersonaID)
	if err != nil {
		return Run{}, err
	}
	runID := NewRunID()
	run, err := scanRun(tx.QueryRow(ctx, `insert into runs (id, thread_id, user_id, status, source, title, persona_id) values ($1, $2, $3, 'queued', $4, $5, nullif($6, '')) returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, runID, threadID, user.ID, source, TitleForRunSource(source), snapshot.ID))
	if err != nil {
		if strings.Contains(err.Error(), "runs_one_active_per_thread_idx") {
			return Run{}, NewError(CodeActiveRunExists, "Thread already has an active run.")
		}
		return Run{}, err
	}
	run.PersonaID = snapshot.ID
	jobID := NewBackgroundJobID()
	metadata := map[string]any{"source": string(source), "job_id": jobID}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = runProviderID(input.ProviderID, snapshot)
		metadata["model"] = runModel(input.ProviderID, input.Model, snapshot)
	}
	if snapshot.ID != "" {
		metadata["persona_id"] = snapshot.ID
		metadata["persona_version"] = snapshot.Version
		metadata["persona_name"] = snapshot.Name
		metadata["persona_resolved_from"] = string(snapshot.ResolvedFrom)
		if err := insertPersonaSnapshot(ctx, tx, run.ID, snapshot); err != nil {
			return Run{}, err
		}
	}
	_, err = scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, 1, 'lifecycle', 'run_created', 'Run created', $5) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(metadata))))
	if err != nil {
		return Run{}, err
	}
	if _, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, 2, 'lifecycle', 'run_queued', 'Run queued', $5) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(map[string]any{"job_id": jobID})))); err != nil {
		return Run{}, err
	}
	if _, err := tx.Exec(ctx, `insert into background_jobs (id, run_id, thread_id, user_id, kind, status, max_attempts, metadata) values ($1, $2, $3, $4, 'run_execution', 'queued', 3, $5)`, jobID, run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(metadata))); err != nil {
		return Run{}, err
	}
	if _, err := tx.Exec(ctx, `update threads set updated_at=now() where id=$1 and user_id=$2`, threadID, user.ID); err != nil {
		return Run{}, err
	}
	return run, tx.Commit(ctx)
}

func (r *PostgresRepository) GetRun(ctx context.Context, ident identity.LocalIdentity, runID string) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return run, err
}

func (r *PostgresRepository) GetCurrentRun(ctx context.Context, ident identity.LocalIdentity, threadID string) (Run, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return Run{}, err
	}
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where thread_id=$1 and user_id=$2 order by case when status in ('pending','queued','running','recovering','blocked_on_tool_approval','stopping','retrying') then 0 else 1 end, updated_at desc, id desc limit 1`, threadID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, NewError(CodeRunNotFound, "Run not found.")
	}
	return run, err
}

func (r *PostgresRepository) ListRunEvents(ctx context.Context, ident identity.LocalIdentity, runID string, afterSequence int) ([]RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	var exists bool
	if err := r.Pool.QueryRow(ctx, `select exists(select 1 from runs where id=$1 and user_id=$2)`, runID, user.ID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, NewError(CodeRunNotFound, "Run not found.")
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where run_id=$1 and user_id=$2 and sequence>$3 order by sequence asc, id asc`, runID, user.ID, afterSequence)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []RunEvent
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (r *PostgresRepository) PrepareRunContext(ctx context.Context, ident identity.LocalIdentity, job BackgroundJob) (RunContext, error) {
	run, err := r.GetRun(ctx, ident, job.RunID)
	if err != nil {
		return RunContext{}, err
	}
	thread, err := r.GetThread(ctx, ident, run.ThreadID)
	if err != nil {
		return RunContext{}, err
	}
	if job.ID == "" || job.RunID != run.ID || job.ThreadID != thread.ID || job.UserID != run.UserID {
		return RunContext{}, NewError(CodeInvalidRequest, "Run context job boundary is invalid.")
	}
	messages, err := r.ListMessages(ctx, ident, thread.ID)
	if err != nil {
		return RunContext{}, err
	}
	events, err := r.ListRunEvents(ctx, ident, run.ID, 0)
	if err != nil {
		return RunContext{}, err
	}
	context, err := buildRunContext(run, thread, messages, job, events)
	if err != nil {
		return RunContext{}, err
	}
	snapshot, err := r.getPersonaSnapshot(ctx, run.ID)
	if err == nil {
		context.Persona = snapshot
		applyPersonaToRunContext(&context, events)
	}
	memories, err := r.SearchMemory(ctx, ident, MemorySearchInput{ScopeType: MemoryScopeThread, ScopeID: thread.ID, Limit: 5, Purpose: "run_context"})
	if err != nil {
		context.MemorySnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Limit: 5, LoadStatus: "unavailable"}
		return context, nil
	}
	status := "loaded"
	if len(memories.Items) == 0 {
		status = "empty"
	}
	context.MemorySnapshot = MemorySnapshot{RunID: run.ID, ThreadID: thread.ID, Entries: memories.Items, Limit: 5, TotalCandidates: len(memories.Items), LoadStatus: status, RedactionApplied: true}
	_, _ = r.AppendRunEvent(ctx, ident, run.ID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventMemorySnapshotLoaded, Summary: "Memory snapshot loaded", Metadata: memorySnapshotEventMetadata(context.MemorySnapshot)})
	return context, nil
}

func (r *PostgresRepository) ListToolCatalog(ctx context.Context, ident identity.LocalIdentity) ([]ToolCatalogEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at from run_events where user_id=$1 and type in ('mcp_discovery_succeeded','mcp_discovery_failed','mcp_discovery_rejected') order by created_at asc, id asc`, user.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []RunEvent
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return SafeToolCatalogFromEvents(events), nil
}

func (r *PostgresRepository) CreateMemoryEntry(ctx context.Context, ident identity.LocalIdentity, input CreateMemoryEntryInput) (MemoryEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryEntry{}, err
	}
	scopeType, scopeID, err := normalizeMemoryScope(user.ID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryEntry{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryEntry{}, err
	}
	status := MemoryEntryApproved
	if safety == MemorySafetyBlocked {
		status = MemoryEntryDisabled
	}
	row := r.Pool.QueryRow(ctx, `insert into memory_entries (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, content_hash) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,nullif($10,''),nullif($11,''),nullif($12,''),$13) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`,
		NewMemoryEntryID(), user.ID, scopeType, scopeID, title, summary, content, status, safety, strings.TrimSpace(input.SourceThreadID), strings.TrimSpace(input.SourceRunID), strings.TrimSpace(input.SourceEventID), memoryContentHash(scopeType, scopeID, content))
	return scanMemoryEntry(row)
}

func (r *PostgresRepository) ListMemoryEntries(ctx context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	return r.SearchMemory(ctx, ident, input)
}

func (r *PostgresRepository) SearchMemory(ctx context.Context, ident identity.LocalIdentity, input MemorySearchInput) (MemorySearchOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemorySearchOutput{}, err
	}
	limit := memoryLimit(input.Limit)
	scopeType := input.ScopeType
	scopeID := strings.TrimSpace(input.ScopeID)
	if scopeType == "" {
		scopeType = MemoryScopeUser
	}
	query := strings.ToLower(strings.TrimSpace(input.Query))
	sql := `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where user_id=$1 and safety_state <> 'blocked'`
	args := []any{user.ID}
	if input.IncludeTombstoned {
		sql += ` and status in ('approved','tombstoned')`
	} else {
		sql += ` and status='approved'`
	}
	if scopeType == MemoryScopeThread {
		args = append(args, scopeID)
		sql += ` and ((scope_type='user' and scope_id=$1) or (scope_type='thread' and scope_id=$2))`
	} else {
		sql += ` and scope_type='user' and scope_id=$1`
	}
	if query != "" {
		args = append(args, "%"+query+"%")
		sql += ` and (lower(title) like $` + intPlaceholder(len(args)) + ` or lower(summary) like $` + intPlaceholder(len(args)) + ` or lower(content) like $` + intPlaceholder(len(args)) + `)`
	}
	if sourceRunID := strings.TrimSpace(input.SourceRunID); sourceRunID != "" {
		args = append(args, sourceRunID)
		sql += ` and source_run_id=$` + intPlaceholder(len(args))
	}
	if sourceThreadID := strings.TrimSpace(input.SourceThreadID); sourceThreadID != "" {
		args = append(args, sourceThreadID)
		sql += ` and source_thread_id=$` + intPlaceholder(len(args))
	}
	switch strings.TrimSpace(input.SourceType) {
	case "", "any":
	case "run":
		sql += ` and source_run_id is not null`
	case "thread":
		sql += ` and source_thread_id is not null`
	case "manual":
		sql += ` and source_run_id is null and source_thread_id is null`
	default:
		return MemorySearchOutput{}, NewError(CodeInvalidRequest, "Memory source type is invalid.")
	}
	args = append(args, limit)
	sql += ` order by updated_at desc, id desc limit $` + intPlaceholder(len(args))
	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return MemorySearchOutput{}, err
	}
	defer rows.Close()
	var items []MemorySearchResult
	for rows.Next() {
		entry, err := scanMemoryEntry(rows)
		if err != nil {
			return MemorySearchOutput{}, err
		}
		items = append(items, memorySearchResult(entry))
	}
	return MemorySearchOutput{Items: items}, rows.Err()
}

func (r *PostgresRepository) GetMemoryEntry(ctx context.Context, ident identity.LocalIdentity, entryID string, input MemoryEntryAccessInput) (MemoryEntry, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryEntry{}, err
	}
	entry, err := scanMemoryEntry(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1 and user_id=$2 and status in ('approved','tombstoned') and safety_state <> 'blocked'`, strings.TrimSpace(entryID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryEntry{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryEntry{}, err
	}
	if !memoryEntryReadableTo(entry, user.ID, input) {
		return MemoryEntry{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	entry.Content = ""
	return entry, err
}

func (r *PostgresRepository) ListMemoryAudit(ctx context.Context, ident identity.LocalIdentity, input MemoryAuditInput) (MemoryAuditOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryAuditOutput{}, err
	}
	limit := memoryLimit(input.Limit)
	sql := `select id, run_id, thread_id, user_id, 0, 'progress', type, summary, null, metadata, created_at from memory_audit_events where user_id=$1 and type = any($2)`
	args := []any{user.ID, []string{EventMemorySnapshotLoaded, EventMemoryWriteProposed, EventMemoryWriteApproved, EventMemoryWriteDenied, EventMemoryEntryDeleted}}
	if threadID := strings.TrimSpace(input.ThreadID); threadID != "" {
		args = append(args, threadID)
		sql += ` and thread_id=$` + intPlaceholder(len(args))
	}
	if runID := strings.TrimSpace(input.SourceRunID); runID != "" {
		args = append(args, runID)
		sql += ` and run_id=$` + intPlaceholder(len(args))
	}
	if eventType := strings.TrimSpace(input.EventType); eventType != "" {
		if eventType == "memory_deleted" {
			eventType = EventMemoryEntryDeleted
		}
		args = append(args, eventType)
		sql += ` and type=$` + intPlaceholder(len(args))
	}
	args = append(args, limit)
	sql += ` order by created_at desc, id desc limit $` + intPlaceholder(len(args))
	rows, err := r.Pool.Query(ctx, sql, args...)
	if err != nil {
		return MemoryAuditOutput{}, err
	}
	defer rows.Close()
	var items []MemoryAuditItem
	for rows.Next() {
		event, err := scanRunEvent(rows)
		if err != nil {
			return MemoryAuditOutput{}, err
		}
		items = append(items, memoryAuditItem(event))
	}
	return MemoryAuditOutput{Items: items}, rows.Err()
}

func (r *PostgresRepository) DeleteMemoryEntry(ctx context.Context, ident identity.LocalIdentity, entryID string, input DeleteMemoryEntryInput) (MemoryTombstone, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryTombstone{}, err
	}
	entry, err := scanMemoryEntry(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1 and user_id=$2`, strings.TrimSpace(entryID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryTombstone{}, err
	}
	if !memoryEntryReadableTo(entry, user.ID, MemoryEntryAccessInput{ScopeType: input.ScopeType, ScopeID: input.ScopeID, SourceThreadID: input.SourceThreadID, SourceRunID: input.SourceRunID}) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if entry.Status == MemoryEntryTombstoned && entry.DeletedAt != nil {
		return MemoryTombstone{EntryID: entry.ID, Status: string(MemoryEntryTombstoned), DeletedAt: *entry.DeletedAt}, nil
	}
	entry, err = scanMemoryEntry(r.Pool.QueryRow(ctx, `update memory_entries set status='tombstoned', content='', summary='[deleted]', deleted_at=now(), deleted_by_user_id=$3, delete_reason=$4, updated_at=now() where id=$1 and user_id=$2 and status <> 'tombstoned' returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`, entry.ID, user.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason))))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryTombstone{}, NewError(CodeMemoryNotFound, "Memory not found.")
	}
	if err != nil {
		return MemoryTombstone{}, err
	}
	r.appendMemoryAuditEvent(ctx, ident, entry.SourceRunID, EventMemoryEntryDeleted, "Memory entry deleted", memoryEntryAuditMetadata(entry, ""))
	return MemoryTombstone{EntryID: strings.TrimSpace(entryID), Status: string(MemoryEntryTombstoned), DeletedAt: *entry.DeletedAt}, nil
}

func (r *PostgresRepository) ProposeMemoryWrite(ctx context.Context, ident identity.LocalIdentity, input ProposeMemoryWriteInput) (MemoryWriteProposal, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	if key := strings.TrimSpace(input.IdempotencyKey); key != "" {
		proposal, err := scanMemoryProposal(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where user_id=$1 and idempotency_key=$2`, user.ID, key))
		if err == nil {
			return proposal, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return MemoryWriteProposal{}, err
		}
	}
	scopeType, scopeID, err := normalizeMemoryScope(user.ID, input.ScopeType, input.ScopeID)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	title, summary, content, safety, err := normalizeMemoryContent(input.Title, input.Content)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	status := MemoryWritePending
	if safety == MemorySafetyBlocked {
		status = MemoryWriteDenied
	}
	row := r.Pool.QueryRow(ctx, `insert into memory_write_proposals (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, idempotency_key) values ($1,$2,$3,$4,$5,$6,$7,$8,$9,nullif($10,''),nullif($11,''),nullif($12,''),$13) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`,
		NewMemoryProposalID(), user.ID, scopeType, scopeID, title, summary, content, status, safety, strings.TrimSpace(input.SourceThreadID), strings.TrimSpace(input.SourceRunID), strings.TrimSpace(input.SourceEventID), strings.TrimSpace(input.IdempotencyKey))
	proposal, err := scanMemoryProposal(row)
	if err != nil {
		return MemoryWriteProposal{}, err
	}
	r.appendMemoryAuditEvent(ctx, ident, proposal.SourceRunID, EventMemoryWriteProposed, "Memory write proposed", memoryProposalAuditMetadata(proposal, ""))
	return proposal, nil
}

func (r *PostgresRepository) ApproveMemoryWrite(ctx context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	defer tx.Rollback(ctx)
	proposal, err := scanMemoryProposal(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where id=$1 and user_id=$2 for update`, strings.TrimSpace(proposalID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	if proposal.Status == MemoryWriteApproved && proposal.CreatedEntryID != "" {
		entry, err := scanMemoryEntry(tx.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'') from memory_entries where id=$1`, proposal.CreatedEntryID))
		if err != nil {
			return MemoryWriteDecision{}, err
		}
		entry.Content = ""
		if err := tx.Commit(ctx); err != nil {
			return MemoryWriteDecision{}, err
		}
		return MemoryWriteDecision{Proposal: proposal, Entry: entry}, nil
	}
	if proposal.Status != MemoryWritePending || proposal.SafetyState == MemorySafetyBlocked {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Memory write cannot be approved.")
	}
	entry, err := scanMemoryEntry(tx.QueryRow(ctx, `insert into memory_entries (id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, source_thread_id, source_run_id, source_event_id, content_hash) values ($1,$2,$3,$4,$5,$6,$7,'approved',$8,nullif($9,''),nullif($10,''),nullif($11,''),$12) returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), content_hash, created_at, updated_at, deleted_at, coalesce(deleted_by_user_id,''), coalesce(delete_reason,'')`,
		NewMemoryEntryID(), user.ID, proposal.ScopeType, proposal.ScopeID, proposal.Title, proposal.Summary, proposal.Content, proposal.SafetyState, proposal.SourceThreadID, proposal.SourceRunID, proposal.SourceEventID, memoryContentHash(proposal.ScopeType, proposal.ScopeID, proposal.Content)))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	proposal, err = scanMemoryProposal(tx.QueryRow(ctx, `update memory_write_proposals set status='approved', created_entry_id=$1, decided_at=now(), decided_by_user_id=$2, decision_reason=$3 where id=$4 returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`, entry.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason)), proposal.ID))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	entry.Content = ""
	if err := tx.Commit(ctx); err != nil {
		return MemoryWriteDecision{}, err
	}
	r.appendMemoryAuditEvent(ctx, ident, proposal.SourceRunID, EventMemoryWriteApproved, "Memory write approved", memoryProposalAuditMetadata(proposal, entry.ID))
	return MemoryWriteDecision{Proposal: proposal, Entry: entry}, nil
}

func (r *PostgresRepository) DenyMemoryWrite(ctx context.Context, ident identity.LocalIdentity, proposalID string, input MemoryWriteDecisionInput) (MemoryWriteDecision, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	proposal, err := scanMemoryProposal(r.Pool.QueryRow(ctx, `select id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'') from memory_write_proposals where id=$1 and user_id=$2`, strings.TrimSpace(proposalID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return MemoryWriteDecision{}, NewError(CodeMemoryNotFound, "Memory proposal not found.")
	}
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	if proposal.Status == MemoryWriteDenied {
		return MemoryWriteDecision{Proposal: proposal}, nil
	}
	if proposal.Status == MemoryWriteApproved {
		return MemoryWriteDecision{}, NewError(CodeInvalidRequest, "Approved memory write cannot be denied.")
	}
	proposal, err = scanMemoryProposal(r.Pool.QueryRow(ctx, `update memory_write_proposals set status='denied', decided_at=now(), decided_by_user_id=$3, decision_reason=$4 where id=$1 and user_id=$2 and status='pending' returning id, user_id, scope_type, scope_id, title, summary, content, status, safety_state, coalesce(source_thread_id,''), coalesce(source_run_id,''), coalesce(source_event_id,''), idempotency_key, coalesce(created_entry_id,''), created_at, decided_at, coalesce(decided_by_user_id,''), coalesce(decision_reason,'')`, proposal.ID, user.ID, user.ID, RedactEventText(strings.TrimSpace(input.Reason))))
	if err != nil {
		return MemoryWriteDecision{}, err
	}
	r.appendMemoryAuditEvent(ctx, ident, proposal.SourceRunID, EventMemoryWriteDenied, "Memory write denied", memoryProposalAuditMetadata(proposal, ""))
	return MemoryWriteDecision{Proposal: proposal}, nil
}

func (r *PostgresRepository) appendMemoryAuditEvent(ctx context.Context, ident identity.LocalIdentity, runID string, eventType string, summary string, metadata map[string]any) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return
	}
	runID = strings.TrimSpace(runID)
	threadID := ""
	var run Run
	if runID != "" {
		run, err = scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2`, runID, user.ID))
		if err == nil {
			threadID = run.ThreadID
		}
	}
	_, _ = r.Pool.Exec(ctx, `insert into memory_audit_events (id, user_id, thread_id, run_id, type, summary, metadata) values ($1,$2,$3,$4,$5,$6,$7)`, NewRunEventID(), user.ID, threadID, runID, eventType, RedactEventText(summary), mustJSON(RedactEventMetadata(metadata)))
	if run.ID != "" {
		_, _ = insertRunEventIgnoringTerminal(ctx, r.Pool, run, RunEventCategoryProgress, eventType, summary, nil, metadata)
	}
}

func (r *PostgresRepository) AppendRunEvent(ctx context.Context, ident identity.LocalIdentity, runID string, input AppendRunEventInput) (RunEvent, error) {
	input, err := NormalizeRunEventInput(input)
	if err != nil {
		return RunEvent{}, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return RunEvent{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return RunEvent{}, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return RunEvent{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return RunEvent{}, err
	}
	if IsRunTerminal(run.Status) {
		return RunEvent{}, NewError(CodeInvalidRequest, "Terminal run cannot accept new events.")
	}
	var nextSequence int
	if err := tx.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return RunEvent{}, err
	}
	event, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, user.ID, nextSequence, input.Category, input.Type, input.Summary, input.Content, mustJSON(input.Metadata)))
	if err != nil {
		return RunEvent{}, err
	}
	if isMemoryAuditEvent(event.Type) {
		if _, err := tx.Exec(ctx, `insert into memory_audit_events (id, user_id, thread_id, run_id, type, summary, metadata, created_at) values ($1,$2,$3,$4,$5,$6,$7,$8)`, NewRunEventID(), user.ID, run.ThreadID, run.ID, event.Type, event.Summary, mustJSON(event.Metadata), event.CreatedAt); err != nil {
			return RunEvent{}, err
		}
	}
	status := run.Status
	completedAtSQL := `completed_at`
	errorCode := run.ErrorCode
	errorMessage := run.ErrorMessage
	if input.Category == RunEventCategoryFinal {
		status = statusFromFinalType(input.Type)
		completedAtSQL = `now()`
		if input.ErrorCode != "" {
			errorCode = &input.ErrorCode
		}
		if input.ErrorMessage != "" {
			errorMessage = &input.ErrorMessage
		}
	}
	if _, err := tx.Exec(ctx, `update runs set status=$1, updated_at=now(), completed_at=`+completedAtSQL+`, error_code=$4, error_message=$5 where id=$2 and user_id=$3`, status, run.ID, user.ID, errorCode, errorMessage); err != nil {
		return RunEvent{}, err
	}
	return event, tx.Commit(ctx)
}

func (r *PostgresRepository) GetToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, err
	}
	call, err := scanToolCall(r.Pool.QueryRow(ctx, `select tc.id, tc.thread_id, tc.run_id, tc.tool_call_id, tc.tool_name, tc.candidate_schema_hash, tc.arguments_summary, tc.approval_status, tc.execution_status, tc.result_summary, tc.error_code, tc.error_message, tc.requested_at, tc.updated_at from tool_calls tc join runs r on r.id=tc.run_id where tc.thread_id=$1 and tc.run_id=$2 and tc.tool_call_id=$3 and r.user_id=$4`, threadID, runID, strings.TrimSpace(toolCallID), user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return ToolCall{}, err
	}
	return call, nil
}

func (r *PostgresRepository) RecordToolCallRequest(ctx context.Context, ident identity.LocalIdentity, runID string, input RecordToolCallRequestInput) (ToolCall, []RunEvent, error) {
	input, err := ValidateToolCallRequestInput(input)
	if err != nil {
		return ToolCall{}, nil, err
	}
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ToolCall{}, nil, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return ToolCall{}, nil, err
	}
	if IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Terminal runs cannot request tools.")
	}
	var existingToolCallID string
	err = tx.QueryRow(ctx, `select tool_call_id from tool_calls where run_id=$1 limit 1`, run.ID).Scan(&existingToolCallID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ToolCall{}, nil, err
	}
	if existingToolCallID != "" && existingToolCallID != input.ToolCallID {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Only one tool call is supported per run.")
	}
	arguments := RedactEventMetadata(input.ArgumentsSummary)
	call, err := scanToolCall(tx.QueryRow(ctx, `insert into tool_calls (id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, arguments_hash, approval_status, execution_status) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) on conflict (run_id, tool_call_id) do nothing returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, NewToolCallID(), run.ThreadID, run.ID, input.ToolCallID, input.ToolName, input.CandidateSchemaHash, mustJSON(arguments), input.ArgumentsHash, input.ApprovalStatus, input.ExecutionStatus))
	if errors.Is(err, pgx.ErrNoRows) {
		existing, err := scanToolCall(tx.QueryRow(ctx, `select id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at from tool_calls where run_id=$1 and tool_call_id=$2`, run.ID, input.ToolCallID))
		if err != nil {
			return ToolCall{}, nil, err
		}
		return existing, nil, tx.Commit(ctx)
	}
	if err != nil {
		return ToolCall{}, nil, err
	}
	run.Status = RunStatusBlockedOnToolApproval
	if _, err := tx.Exec(ctx, `update runs set status=$1, updated_at=now() where id=$2 and user_id=$3`, run.Status, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	metadata := toolCallEventMetadata(call)
	requested, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallRequested, "Tool call requested", nil, metadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	required, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApprovalRequired, "Tool approval required", nil, metadata)
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{requested, required}, tx.Commit(ctx)
}

func (r *PostgresRepository) ApproveToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalApproved {
		if call.ExecutionStatus == ToolCallExecutionNotStarted || call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed {
			return call, nil, tx.Commit(ctx)
		}
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	if call.ApprovalStatus != ToolCallApprovalRequired || call.ExecutionStatus != ToolCallExecutionBlocked || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be approved.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set approval_status='approved', execution_status='not_started', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	run.Status = RunStatusQueued
	if _, err := tx.Exec(ctx, `update runs set status='queued', updated_at=now() where id=$1 and user_id=$2`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	jobID := NewBackgroundJobID()
	metadata := RedactEventMetadata(map[string]any{"source": string(run.Source), "job_id": jobID, "tool_call_id": call.ToolCallID, "resume_reason": "tool_call_approved"})
	if _, err := tx.Exec(ctx, `insert into background_jobs (id, run_id, thread_id, user_id, kind, status, priority, max_attempts, scheduled_at, metadata) values ($1, $2, $3, $4, $5, 'queued', 50, 3, now(), $6)`, jobID, run.ID, run.ThreadID, user.ID, BackgroundJobKindRunExecution, mustJSON(metadata)); err != nil {
		return ToolCall{}, nil, err
	}
	event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallApproved, "Tool call approved", nil, toolCallEventMetadata(call))
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{event}, tx.Commit(ctx)
}

func (r *PostgresRepository) DenyToolCall(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ApprovalStatus == ToolCallApprovalDenied {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed || call.ExecutionStatus == ToolCallExecutionCancelled || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot be denied.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set approval_status='denied', execution_status='cancelled', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return ToolCall{}, nil, err
	}
	stopped, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', completed_at=now(), updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	denied, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryProgress, EventToolCallDenied, "Tool call denied by user", nil, toolCallEventMetadata(call))
	if err != nil {
		return ToolCall{}, nil, err
	}
	final, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryFinal, EventRunStopped, "Run stopped after tool denial", nil, map[string]any{"tool_call_id": call.ToolCallID, "reason": "tool_call_denied"})
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{denied, final}, tx.Commit(ctx)
}

func (r *PostgresRepository) StartToolCallExecution(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionExecuting || call.ExecutionStatus == ToolCallExecutionSucceeded || call.ExecutionStatus == ToolCallExecutionFailed || call.ExecutionStatus == ToolCallExecutionCancelled {
		return call, nil, tx.Commit(ctx)
	}
	if call.ApprovalStatus != ToolCallApprovalApproved || call.ExecutionStatus != ToolCallExecutionNotStarted || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot execute.")
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='executing', updated_at=now() where id=$1 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	event, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventToolCallExecuting, "Tool call executing", nil, toolCallEventMetadata(call))
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{event}, tx.Commit(ctx)
}

func (r *PostgresRepository) CompleteToolCallSuccess(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, resultSummary map[string]any) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionSucceeded {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot succeed.")
	}
	result := RedactEventMetadata(resultSummary)
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='succeeded', result_summary=$1, updated_at=now() where id=$2 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, mustJSON(result), call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	running, err := scanRun(tx.QueryRow(ctx, `update runs set status='running', completed_at=null, updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	succeeded, err := insertRunEvent(ctx, tx, running, RunEventCategoryProgress, EventToolCallSucceeded, "Tool call succeeded", nil, toolCallEventMetadata(call))
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{succeeded}, tx.Commit(ctx)
}

func (r *PostgresRepository) FailToolCallExecution(ctx context.Context, ident identity.LocalIdentity, threadID string, runID string, toolCallID string, errorCode string, errorMessage string) (ToolCall, []RunEvent, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return ToolCall{}, nil, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return ToolCall{}, nil, err
	}
	defer tx.Rollback(ctx)
	run, call, err := scopedPostgresToolCall(ctx, tx, user.ID, threadID, runID, toolCallID)
	if err != nil {
		return ToolCall{}, nil, err
	}
	if call.ExecutionStatus == ToolCallExecutionFailed {
		return call, nil, tx.Commit(ctx)
	}
	if call.ExecutionStatus != ToolCallExecutionExecuting || IsRunTerminal(run.Status) {
		return ToolCall{}, nil, NewError(CodeInvalidRequest, "Tool call cannot fail.")
	}
	code := strings.TrimSpace(errorCode)
	if code == "" {
		code = "tool_execution_failed"
	}
	message := RedactEventText(strings.TrimSpace(errorMessage))
	if message == "" {
		message = "Tool execution failed."
	}
	call, err = scanToolCall(tx.QueryRow(ctx, `update tool_calls set execution_status='failed', error_code=$1, error_message=$2, updated_at=now() where id=$3 returning id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at`, code, message, call.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	failedRun, err := scanRun(tx.QueryRow(ctx, `update runs set status='failed', completed_at=now(), updated_at=now(), error_code=$1, error_message=$2 where id=$3 and user_id=$4 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, code, message, run.ID, user.ID))
	if err != nil {
		return ToolCall{}, nil, err
	}
	failed, err := insertRunEvent(ctx, tx, failedRun, RunEventCategoryError, EventToolCallFailed, message, nil, toolCallEventMetadata(call))
	if err != nil {
		return ToolCall{}, nil, err
	}
	final, err := insertRunEvent(ctx, tx, failedRun, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"tool_call_id": call.ToolCallID, "error_code": code})
	if err != nil {
		return ToolCall{}, nil, err
	}
	return call, []RunEvent{failed, final}, tx.Commit(ctx)
}

func (r *PostgresRepository) ClaimBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input ClaimBackgroundJobInput) (BackgroundJob, Run, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	workerID := strings.TrimSpace(input.WorkerID)
	if workerID == "" {
		return BackgroundJob{}, Run{}, false, NewError(CodeInvalidRequest, "Worker id is required.")
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	defer tx.Rollback(ctx)
	job, err := scanBackgroundJob(tx.QueryRow(ctx, `select id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at from background_jobs where user_id=$1 and status='queued' and scheduled_at<=now() order by priority asc, created_at asc, id asc for update skip locked limit 1`, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, Run{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, job.RunID, user.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	if IsRunTerminal(run.Status) || run.StopRequestedAt != nil {
		cancelled, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='cancelled', updated_at=now() where id=$1 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, job.ID))
		if err != nil {
			return BackgroundJob{}, Run{}, false, err
		}
		return cancelled, run, false, tx.Commit(ctx)
	}
	leased, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='leased', leased_by=$1, lease_expires_at=now() + ($2::int * interval '1 second'), attempt_count=attempt_count+1, ownership_version=ownership_version+1, updated_at=now() where id=$3 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, workerID, leaseSeconds, job.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	run, err = scanRun(tx.QueryRow(ctx, `update runs set status='running', updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	if _, err := insertRunEvent(ctx, tx, run, RunEventCategoryProgress, EventJobClaimed, "Job claimed", nil, map[string]any{"job_id": leased.ID, "worker_id": workerID, "attempt": leased.AttemptCount}); err != nil {
		return BackgroundJob{}, Run{}, false, err
	}
	return leased, run, true, tx.Commit(ctx)
}

func (r *PostgresRepository) RenewBackgroundJobLease(ctx context.Context, ident identity.LocalIdentity, input RenewBackgroundJobLeaseInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	leaseSeconds := input.LeaseSeconds
	if leaseSeconds <= 0 {
		leaseSeconds = 30
	}
	job, err := scanBackgroundJob(r.Pool.QueryRow(ctx, `update background_jobs set lease_expires_at=now() + ($1::int * interval '1 second'), updated_at=now() where id=$2 and user_id=$3 and leased_by=$4 and ownership_version=$5 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, leaseSeconds, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, false, err
	}
	run, err := r.GetRun(ctx, ident, job.RunID)
	if err == nil && !IsRunTerminal(run.Status) {
		_, _ = r.AppendRunEvent(ctx, ident, job.RunID, AppendRunEventInput{Category: RunEventCategoryProgress, Type: EventLeaseRenewed, Summary: "Lease renewed", Metadata: map[string]any{"job_id": job.ID, "worker_id": strings.TrimSpace(input.WorkerID), "ownership_version": input.OwnershipVersion}})
	}
	return job, true, nil
}

func (r *PostgresRepository) RecoverBackgroundJobs(ctx context.Context, ident identity.LocalIdentity, input RecoverBackgroundJobsInput) ([]BackgroundJobRecovery, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 10
	}
	code := strings.TrimSpace(input.ErrorCode)
	if code == "" {
		code = "worker_lease_expired"
	}
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	if message == "" {
		message = "Worker lease expired."
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `select id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at from background_jobs where user_id=$1 and status='leased' and lease_expires_at < now() order by lease_expires_at asc, id asc for update skip locked limit $2`, user.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	jobs := []BackgroundJob{}
	for rows.Next() {
		job, err := scanBackgroundJob(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	recoveries := []BackgroundJobRecovery{}
	for _, job := range jobs {
		run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, job.RunID, user.ID))
		if err != nil || IsRunTerminal(run.Status) {
			continue
		}
		previousWorkerID := ""
		if job.LeasedBy != nil {
			previousWorkerID = *job.LeasedBy
		}
		if job.AttemptCount >= job.MaxAttempts {
			dead, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='dead', leased_by=null, lease_expires_at=null, last_error_code=$1, last_error_message=$2, updated_at=now() where id=$3 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, code, message, job.ID))
			if err != nil {
				return nil, err
			}
			failed, err := scanRun(tx.QueryRow(ctx, `update runs set status='failed', completed_at=now(), updated_at=now(), error_code=$1, error_message=$2 where id=$3 and user_id=$4 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, code, message, run.ID, user.ID))
			if err != nil {
				return nil, err
			}
			exhausted, err := insertRunEvent(ctx, tx, failed, RunEventCategoryError, EventJobRetryExhausted, message, nil, map[string]any{"job_id": dead.ID, "attempt_count": dead.AttemptCount, "error_code": code})
			if err != nil {
				return nil, err
			}
			final, err := insertRunEvent(ctx, tx, failed, RunEventCategoryFinal, EventRunFailed, message, nil, map[string]any{"job_id": dead.ID, "error_code": code})
			if err != nil {
				return nil, err
			}
			recoveries = append(recoveries, BackgroundJobRecovery{Job: dead, Run: failed, Events: []RunEvent{exhausted, final}, Exhausted: true})
			continue
		}
		backoffSeconds := int(retryBackoffDuration(job.AttemptCount).Seconds())
		queued, err := scanBackgroundJob(tx.QueryRow(ctx, `update background_jobs set status='queued', leased_by=null, lease_expires_at=null, scheduled_at=now() + ($1::int * interval '1 second'), last_error_code=$2, last_error_message=$3, updated_at=now() where id=$4 returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, backoffSeconds, code, message, job.ID))
		if err != nil {
			return nil, err
		}
		recoveringRun, err := scanRun(tx.QueryRow(ctx, `update runs set status='recovering', updated_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
		if err != nil {
			return nil, err
		}
		recovering, err := insertRunEvent(ctx, tx, recoveringRun, RunEventCategoryProgress, EventJobRecovering, "Job recovering", nil, map[string]any{"job_id": queued.ID, "previous_worker_id": previousWorkerID, "attempt": queued.AttemptCount})
		if err != nil {
			return nil, err
		}
		retry, err := insertRunEvent(ctx, tx, recoveringRun, RunEventCategoryProgress, EventJobRetryScheduled, "Job retry scheduled", nil, map[string]any{"job_id": queued.ID, "next_attempt": queued.AttemptCount + 1, "scheduled_at": queued.ScheduledAt})
		if err != nil {
			return nil, err
		}
		recoveries = append(recoveries, BackgroundJobRecovery{Job: queued, Run: recoveringRun, Events: []RunEvent{recovering, retry}})
	}
	return recoveries, tx.Commit(ctx)
}

func (r *PostgresRepository) CompleteBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input CompleteBackgroundJobInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	job, err := scanBackgroundJob(r.Pool.QueryRow(ctx, `update background_jobs set status='completed', updated_at=now() where id=$1 and user_id=$2 and leased_by=$3 and ownership_version=$4 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	return job, true, err
}

func (r *PostgresRepository) FailBackgroundJob(ctx context.Context, ident identity.LocalIdentity, input FailBackgroundJobInput) (BackgroundJob, bool, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return BackgroundJob{}, false, err
	}
	code := strings.TrimSpace(input.ErrorCode)
	message := RedactEventText(strings.TrimSpace(input.ErrorMessage))
	job, err := scanBackgroundJob(r.Pool.QueryRow(ctx, `update background_jobs set status='failed', last_error_code=$1, last_error_message=$2, updated_at=now() where id=$3 and user_id=$4 and leased_by=$5 and ownership_version=$6 and status='leased' returning id, run_id, thread_id, user_id, kind, status, priority, attempt_count, max_attempts, scheduled_at, leased_by, lease_expires_at, ownership_version, metadata, last_error_code, last_error_message, created_at, updated_at`, code, message, input.JobID, user.ID, strings.TrimSpace(input.WorkerID), input.OwnershipVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return BackgroundJob{}, false, nil
	}
	if err != nil {
		return BackgroundJob{}, false, err
	}
	_, _ = r.AppendRunEvent(ctx, ident, job.RunID, AppendRunEventInput{Category: RunEventCategoryError, Type: EventJobAttemptFailed, Summary: message, Metadata: map[string]any{"job_id": job.ID, "attempt": job.AttemptCount, "error_code": code}, ErrorCode: code, ErrorMessage: message})
	_, _ = r.AppendRunEvent(ctx, ident, job.RunID, AppendRunEventInput{Category: RunEventCategoryFinal, Type: EventRunFailed, Summary: message, Metadata: map[string]any{"job_id": job.ID, "error_code": code}, ErrorCode: code, ErrorMessage: message})
	return job, true, nil
}

func (r *PostgresRepository) WorkerQueueDiagnostics(ctx context.Context, ident identity.LocalIdentity) (WorkerQueueDiagnostics, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	diagnostics := WorkerQueueDiagnostics{QueueStatus: WorkerQueueStatusReady, WorkerStatus: WorkerStatusReady}
	row := r.Pool.QueryRow(ctx, `select
		count(*) filter (where status='queued'),
		count(*) filter (where status='leased'),
		count(*) filter (where status='leased' and lease_expires_at < now()),
		count(*) filter (where status='retrying'),
		count(*) filter (where status='dead'),
		now()
		from background_jobs where user_id=$1`, user.ID)
	if err := row.Scan(&diagnostics.QueuedCount, &diagnostics.LeasedCount, &diagnostics.StaleCount, &diagnostics.RetryingCount, &diagnostics.DeadCount, &diagnostics.UpdatedAt); err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	toolRow := r.Pool.QueryRow(ctx, `select
			count(*) filter (where tc.approval_status='required' and tc.execution_status='blocked'),
			count(*) filter (where tc.approval_status='approved' and tc.execution_status='not_started')
			from tool_calls tc join runs r on r.id=tc.run_id where r.user_id=$1`, user.ID)
	if err := toolRow.Scan(&diagnostics.BlockedToolApprovalCount, &diagnostics.ResumableToolCallCount); err != nil {
		return WorkerQueueDiagnostics{}, err
	}
	if diagnostics.StaleCount > 0 || diagnostics.RetryingCount > 0 || diagnostics.DeadCount > 0 {
		diagnostics.QueueStatus = WorkerQueueStatusDegraded
		diagnostics.WorkerStatus = WorkerStatusDegraded
	}
	return diagnostics, nil
}

func insertRunEvent(ctx context.Context, tx pgx.Tx, run Run, category RunEventCategory, eventType string, summary string, content *string, metadata map[string]any) (RunEvent, error) {
	var nextSequence int
	if err := tx.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return RunEvent{}, err
	}
	return scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, run.UserID, nextSequence, category, eventType, RedactEventText(summary), content, mustJSON(RedactEventMetadata(metadata))))
}

func insertRunEventIgnoringTerminal(ctx context.Context, pool *pgxpool.Pool, run Run, category RunEventCategory, eventType string, summary string, content *string, metadata map[string]any) (RunEvent, error) {
	var nextSequence int
	if err := pool.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return RunEvent{}, err
	}
	return scanRunEvent(pool.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, run.UserID, nextSequence, category, eventType, RedactEventText(summary), content, mustJSON(RedactEventMetadata(metadata))))
}

func scopedPostgresToolCall(ctx context.Context, tx pgx.Tx, userID string, threadID string, runID string, toolCallID string) (Run, ToolCall, error) {
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and thread_id=$2 and user_id=$3 for update`, runID, threadID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return Run{}, ToolCall{}, err
	}
	call, err := scanToolCall(tx.QueryRow(ctx, `select id, thread_id, run_id, tool_call_id, tool_name, candidate_schema_hash, arguments_summary, approval_status, execution_status, result_summary, error_code, error_message, requested_at, updated_at from tool_calls where thread_id=$1 and run_id=$2 and tool_call_id=$3 for update`, threadID, runID, strings.TrimSpace(toolCallID)))
	if errors.Is(err, pgx.ErrNoRows) {
		return Run{}, ToolCall{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return Run{}, ToolCall{}, err
	}
	return run, call, nil
}

func (r *PostgresRepository) StopRun(ctx context.Context, ident identity.LocalIdentity, runID string) (StopRunOutput, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return StopRunOutput{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return StopRunOutput{}, err
	}
	defer tx.Rollback(ctx)
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return StopRunOutput{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return StopRunOutput{}, err
	}
	if IsRunTerminal(run.Status) {
		return StopRunOutput{Run: run, Result: StopRunResultAlreadyTerminal}, nil
	}
	if _, err := tx.Exec(ctx, `update background_jobs set status='cancelled', updated_at=now() where run_id=$1 and user_id=$2 and status in ('queued', 'leased', 'retrying')`, run.ID, user.ID); err != nil {
		return StopRunOutput{}, err
	}
	stopped, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', stop_requested_at=coalesce(stop_requested_at, now()), updated_at=now(), completed_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, stop_requested_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return StopRunOutput{}, err
	}
	lifecycle, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryProgress, EventStopRequested, "Stop requested", nil, map[string]any{})
	if err != nil {
		return StopRunOutput{}, err
	}
	final, err := insertRunEvent(ctx, tx, stopped, RunEventCategoryFinal, EventRunStopped, "Run stopped", nil, map[string]any{})
	if err != nil {
		return StopRunOutput{}, err
	}
	return StopRunOutput{Run: stopped, Result: StopRunResultStopped, Events: []RunEvent{lifecycle, final}}, tx.Commit(ctx)
}

func (r *PostgresRepository) SyncBuiltInPersonas(ctx context.Context, ident identity.LocalIdentity, configs []BuiltInPersonaConfig) (PersonaSyncResult, error) {
	if _, err := r.ensureUser(ctx, ident); err != nil {
		return PersonaSyncResult{}, err
	}
	if err := validateBuiltInPersonaConfigs(configs); err != nil {
		return PersonaSyncResult{}, err
	}
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return PersonaSyncResult{}, err
	}
	defer tx.Rollback(ctx)
	result := PersonaSyncResult{Synced: len(configs)}
	for _, config := range configs {
		slug := strings.TrimSpace(config.Slug)
		personaID := ""
		err := tx.QueryRow(ctx, `select id from personas where slug=$1 and source='built_in'`, slug).Scan(&personaID)
		if errors.Is(err, pgx.ErrNoRows) {
			personaID = NewPersonaID()
			result.CreatedPersonas++
			if _, err := tx.Exec(ctx, `insert into personas (id, slug, name, description, source, is_default, is_active, active_version) values ($1, $2, $3, $4, 'built_in', $5, true, $6)`, personaID, slug, strings.TrimSpace(config.Name), strings.TrimSpace(config.Description), config.IsDefault, strings.TrimSpace(config.Version)); err != nil {
				return PersonaSyncResult{}, err
			}
		} else if err != nil {
			return PersonaSyncResult{}, err
		} else if _, err := tx.Exec(ctx, `update personas set name=$1, description=$2, is_default=$3, is_active=true, active_version=$4, updated_at=now() where id=$5`, strings.TrimSpace(config.Name), strings.TrimSpace(config.Description), config.IsDefault, strings.TrimSpace(config.Version), personaID); err != nil {
			return PersonaSyncResult{}, err
		}
		if config.IsDefault {
			result.DefaultPersonaSlug = slug
			if _, err := tx.Exec(ctx, `update personas set is_default=false, updated_at=now() where source='built_in' and id<>$1`, personaID); err != nil {
				return PersonaSyncResult{}, err
			}
		}
		tag, err := tx.Exec(ctx, `insert into persona_versions (persona_id, version, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary) values ($1, $2, $3, $4, $5, $6, $7) on conflict (persona_id, version) do nothing`, personaID, strings.TrimSpace(config.Version), strings.TrimSpace(config.SystemPrompt), mustJSON(config.ModelRoute), config.AllowedToolNames, strings.TrimSpace(config.ReasoningMode), strings.TrimSpace(config.BudgetSummary))
		if err != nil {
			return PersonaSyncResult{}, err
		}
		if tag.RowsAffected() > 0 {
			result.CreatedVersions++
		}
		result.ActivatedVersions++
	}
	if result.DefaultPersonaSlug == "" {
		_ = tx.QueryRow(ctx, `select slug from personas where source='built_in' and is_default=true and is_active=true limit 1`).Scan(&result.DefaultPersonaSlug)
	}
	return result, tx.Commit(ctx)
}

func (r *PostgresRepository) ListPersonas(ctx context.Context, ident identity.LocalIdentity) ([]Persona, error) {
	if _, err := r.ensureUser(ctx, ident); err != nil {
		return nil, err
	}
	rows, err := r.Pool.Query(ctx, `select id, slug, name, description, source, is_default, is_active, active_version, created_at, updated_at from personas where is_active=true order by is_default desc, name asc`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var personas []Persona
	for rows.Next() {
		var persona Persona
		if err := rows.Scan(&persona.ID, &persona.Slug, &persona.Name, &persona.Description, &persona.Source, &persona.IsDefault, &persona.IsActive, &persona.ActiveVersion, &persona.CreatedAt, &persona.UpdatedAt); err != nil {
			return nil, err
		}
		personas = append(personas, persona)
	}
	return personas, rows.Err()
}

func (r *PostgresRepository) resolvePersonaSnapshotTx(ctx context.Context, tx pgx.Tx, threadPersonaID string, runPersonaID string) (PersonaSnapshot, error) {
	if personaID := strings.TrimSpace(runPersonaID); personaID != "" {
		return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromRun)
	}
	if personaID := strings.TrimSpace(threadPersonaID); personaID != "" {
		return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromThread)
	}
	var personaID string
	err := tx.QueryRow(ctx, `select id from personas where source='built_in' and is_default=true and is_active=true limit 1`).Scan(&personaID)
	if errors.Is(err, pgx.ErrNoRows) {
		return PersonaSnapshot{}, nil
	}
	if err != nil {
		return PersonaSnapshot{}, err
	}
	return selectPersonaSnapshot(ctx, tx, personaID, PersonaResolvedFromDefault)
}

func selectPersonaSnapshot(ctx context.Context, tx pgx.Tx, personaID string, resolvedFrom PersonaResolvedFrom) (PersonaSnapshot, error) {
	var snapshot PersonaSnapshot
	var rawRoute []byte
	err := tx.QueryRow(ctx, `select p.id, p.slug, p.active_version, p.name, p.description, pv.system_prompt, pv.model_route, pv.allowed_tool_names, pv.reasoning_mode, pv.budget_summary from personas p join persona_versions pv on pv.persona_id=p.id and pv.version=p.active_version where p.id=$1 and p.is_active=true`, personaID).Scan(&snapshot.ID, &snapshot.Slug, &snapshot.Version, &snapshot.Name, &snapshot.Description, &snapshot.SystemPrompt, &rawRoute, &snapshot.AllowedToolNames, &snapshot.ReasoningMode, &snapshot.BudgetSummary)
	if errors.Is(err, pgx.ErrNoRows) {
		return PersonaSnapshot{}, NewError(CodeInvalidRequest, "Persona could not be resolved for this run.")
	}
	if err != nil {
		return PersonaSnapshot{}, err
	}
	if len(rawRoute) > 0 {
		_ = json.Unmarshal(rawRoute, &snapshot.ModelRoute)
	}
	snapshot.ResolvedFrom = resolvedFrom
	return snapshot, nil
}

func validatePersonaReferenceTx(ctx context.Context, tx pgx.Tx, personaID string) error {
	personaID = strings.TrimSpace(personaID)
	if personaID == "" {
		return nil
	}
	var exists int
	err := tx.QueryRow(ctx, `select 1 from personas p join persona_versions pv on pv.persona_id=p.id and pv.version=p.active_version where p.id=$1 and p.is_active=true`, personaID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return NewError(CodeInvalidRequest, "Persona could not be resolved for this thread.")
	}
	return err
}

func insertPersonaSnapshot(ctx context.Context, tx pgx.Tx, runID string, snapshot PersonaSnapshot) error {
	if snapshot.ID == "" {
		return nil
	}
	_, err := tx.Exec(ctx, `insert into run_persona_snapshots (run_id, persona_id, persona_slug, version, name, description, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary, resolved_from) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`, runID, snapshot.ID, snapshot.Slug, snapshot.Version, snapshot.Name, snapshot.Description, snapshot.SystemPrompt, mustJSON(snapshot.ModelRoute), snapshot.AllowedToolNames, snapshot.ReasoningMode, snapshot.BudgetSummary, string(snapshot.ResolvedFrom))
	return err
}

func (r *PostgresRepository) getPersonaSnapshot(ctx context.Context, runID string) (PersonaSnapshot, error) {
	var snapshot PersonaSnapshot
	var rawRoute []byte
	err := r.Pool.QueryRow(ctx, `select persona_id, persona_slug, version, name, description, system_prompt, model_route, allowed_tool_names, reasoning_mode, budget_summary, resolved_from from run_persona_snapshots where run_id=$1`, runID).Scan(&snapshot.ID, &snapshot.Slug, &snapshot.Version, &snapshot.Name, &snapshot.Description, &snapshot.SystemPrompt, &rawRoute, &snapshot.AllowedToolNames, &snapshot.ReasoningMode, &snapshot.BudgetSummary, &snapshot.ResolvedFrom)
	if err != nil {
		return PersonaSnapshot{}, err
	}
	if len(rawRoute) > 0 {
		_ = json.Unmarshal(rawRoute, &snapshot.ModelRoute)
	}
	return snapshot, nil
}

func (r *PostgresRepository) ensureUser(ctx context.Context, ident identity.LocalIdentity) (User, error) {
	row := r.Pool.QueryRow(ctx, `insert into users (id, display_name) values ($1, $2) on conflict (id) do update set display_name=excluded.display_name, updated_at=users.updated_at returning id, display_name, created_at, updated_at`, ident.UserID, ident.DisplayName)
	var user User
	if err := row.Scan(&user.ID, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt); err != nil {
		return User{}, err
	}
	return user, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanThread(row scanner) (Thread, error) {
	var thread Thread
	if err := row.Scan(&thread.ID, &thread.UserID, &thread.Title, &thread.Mode, &thread.LifecycleStatus, &thread.PersonaID, &thread.CreatedAt, &thread.UpdatedAt, &thread.ArchivedAt); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func scanRun(row scanner) (Run, error) {
	var run Run
	if err := row.Scan(&run.ID, &run.ThreadID, &run.UserID, &run.Status, &run.Source, &run.Title, &run.CreatedAt, &run.UpdatedAt, &run.CompletedAt, &run.StopRequestedAt, &run.ErrorCode, &run.ErrorMessage); err != nil {
		return Run{}, err
	}
	return run, nil
}

func scanBackgroundJob(row scanner) (BackgroundJob, error) {
	var job BackgroundJob
	var rawMetadata []byte
	if err := row.Scan(&job.ID, &job.RunID, &job.ThreadID, &job.UserID, &job.Kind, &job.Status, &job.Priority, &job.AttemptCount, &job.MaxAttempts, &job.ScheduledAt, &job.LeasedBy, &job.LeaseExpiresAt, &job.OwnershipVersion, &rawMetadata, &job.LastErrorCode, &job.LastError, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return BackgroundJob{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &job.Metadata)
	}
	if job.Metadata == nil {
		job.Metadata = map[string]any{}
	}
	return job, nil
}

func scanToolCall(row scanner) (ToolCall, error) {
	var call ToolCall
	var rawArguments []byte
	var rawResult []byte
	if err := row.Scan(&call.ID, &call.ThreadID, &call.RunID, &call.ToolCallID, &call.ToolName, &call.CandidateSchemaHash, &rawArguments, &call.ApprovalStatus, &call.ExecutionStatus, &rawResult, &call.ErrorCode, &call.ErrorMessage, &call.RequestedAt, &call.UpdatedAt); err != nil {
		return ToolCall{}, err
	}
	if len(rawArguments) > 0 {
		_ = json.Unmarshal(rawArguments, &call.ArgumentsSummary)
	}
	if len(rawResult) > 0 {
		_ = json.Unmarshal(rawResult, &call.ResultSummary)
	}
	if call.ArgumentsSummary == nil {
		call.ArgumentsSummary = map[string]any{}
	}
	return call, nil
}

func scanMemoryEntry(row scanner) (MemoryEntry, error) {
	var entry MemoryEntry
	if err := row.Scan(&entry.ID, &entry.UserID, &entry.ScopeType, &entry.ScopeID, &entry.Title, &entry.Summary, &entry.Content, &entry.Status, &entry.SafetyState, &entry.SourceThreadID, &entry.SourceRunID, &entry.SourceEventID, &entry.ContentHash, &entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt, &entry.DeletedBy, &entry.DeleteReason); err != nil {
		return MemoryEntry{}, err
	}
	return entry, nil
}

func scanMemoryProposal(row scanner) (MemoryWriteProposal, error) {
	var proposal MemoryWriteProposal
	if err := row.Scan(&proposal.ID, &proposal.UserID, &proposal.ScopeType, &proposal.ScopeID, &proposal.Title, &proposal.Summary, &proposal.Content, &proposal.Status, &proposal.SafetyState, &proposal.SourceThreadID, &proposal.SourceRunID, &proposal.SourceEventID, &proposal.IdempotencyKey, &proposal.CreatedEntryID, &proposal.CreatedAt, &proposal.DecidedAt, &proposal.DecidedBy, &proposal.DecisionReason); err != nil {
		return MemoryWriteProposal{}, err
	}
	return proposal, nil
}

func intPlaceholder(index int) string {
	return fmt.Sprintf("%d", index)
}

func scanRunEvent(row scanner) (RunEvent, error) {
	var event RunEvent
	var rawMetadata []byte
	if err := row.Scan(&event.ID, &event.RunID, &event.ThreadID, &event.UserID, &event.Sequence, &event.Category, &event.Type, &event.Summary, &event.Content, &rawMetadata, &event.CreatedAt); err != nil {
		return RunEvent{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &event.Metadata)
	}
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}
	return event, nil
}

func mustJSON(value any) []byte {
	raw, err := json.Marshal(value)
	if err != nil {
		return []byte(`{}`)
	}
	return raw
}

func scanMessage(row scanner) (Message, error) {
	var message Message
	var rawMetadata []byte
	if err := row.Scan(&message.ID, &message.ThreadID, &message.UserID, &message.Role, &message.Content, &rawMetadata, &message.ClientMessageID, &message.CreatedAt); err != nil {
		return Message{}, err
	}
	if len(rawMetadata) > 0 {
		_ = json.Unmarshal(rawMetadata, &message.Metadata)
	}
	if message.Metadata == nil {
		message.Metadata = map[string]any{}
	}
	if message.ClientMessageID != nil {
		trimmed := strings.TrimSpace(*message.ClientMessageID)
		message.ClientMessageID = &trimmed
	}
	return message, nil
}
