package productdata

import (
	"context"
	"encoding/json"
	"errors"
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
	row := r.Pool.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status) values ($1, $2, $3, $4, $5) returning id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at`, NewThreadID(), user.ID, title, input.Mode, ThreadLifecycleActive)
	return scanThread(row)
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
	row := r.Pool.QueryRow(ctx, `insert into threads (id, user_id, title, mode, lifecycle_status) values ($1, $2, $3, $4, 'active') on conflict (id) do update set title=excluded.title, mode=excluded.mode, lifecycle_status='active', archived_at=null, updated_at=now() returning id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at`, input.ID, user.ID, title, input.Mode)
	return scanThread(row)
}

func (r *PostgresRepository) ListThreads(ctx context.Context, ident identity.LocalIdentity, includeArchived bool) ([]Thread, error) {
	user, err := r.ensureUser(ctx, ident)
	if err != nil {
		return nil, err
	}
	query := `select id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at from threads where user_id=$1`
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
	row := r.Pool.QueryRow(ctx, `select id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at from threads where id=$1 and user_id=$2`, threadID, user.ID)
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
	row := r.Pool.QueryRow(ctx, `update threads set title=$1, mode=$2, updated_at=now() where id=$3 and user_id=$4 returning id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at`, title, mode, threadID, current.UserID)
	return scanThread(row)
}

func (r *PostgresRepository) ArchiveThread(ctx context.Context, ident identity.LocalIdentity, threadID string) (Thread, error) {
	current, err := r.GetThread(ctx, ident, threadID)
	if err != nil {
		return Thread{}, err
	}
	row := r.Pool.QueryRow(ctx, `update threads set lifecycle_status='archived', archived_at=coalesce(archived_at, now()), updated_at=now() where id=$1 and user_id=$2 returning id, user_id, title, mode, lifecycle_status, created_at, updated_at, archived_at`, threadID, current.UserID)
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
	var threadUserID string
	if err := tx.QueryRow(ctx, `select user_id from threads where id=$1 and user_id=$2 and lifecycle_status='active'`, threadID, user.ID).Scan(&threadUserID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Run{}, NewError(CodeThreadNotFound, "Thread not found.")
		}
		return Run{}, err
	}
	source, err := NormalizeRunSource(input.Source)
	if err != nil {
		return Run{}, err
	}
	runID := NewRunID()
	run, err := scanRun(tx.QueryRow(ctx, `insert into runs (id, thread_id, user_id, status, source, title) values ($1, $2, $3, 'running', $4, $5) returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message`, runID, threadID, user.ID, source, TitleForRunSource(source)))
	if err != nil {
		if strings.Contains(err.Error(), "runs_one_active_per_thread_idx") {
			return Run{}, NewError(CodeActiveRunExists, "Thread already has an active run.")
		}
		return Run{}, err
	}
	metadata := map[string]any{"source": string(source)}
	if source == RunSourceLocalSimulated {
		metadata["script_name"] = NormalizeScriptName(input.ScriptName)
	} else {
		metadata["message_id"] = input.MessageID
		metadata["provider_id"] = input.ProviderID
		metadata["model"] = input.Model
	}
	_, err = scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, 1, 'lifecycle', 'run_created', 'Run created', $5) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, threadID, user.ID, mustJSON(RedactEventMetadata(metadata))))
	if err != nil {
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
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message from runs where id=$1 and user_id=$2`, runID, user.ID))
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
	run, err := scanRun(r.Pool.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message from runs where thread_id=$1 and user_id=$2 order by case when status in ('pending','running') then 0 else 1 end, updated_at desc, id desc limit 1`, threadID, user.ID))
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
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
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
	run, err := scanRun(tx.QueryRow(ctx, `select id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message from runs where id=$1 and user_id=$2 for update`, runID, user.ID))
	if errors.Is(err, pgx.ErrNoRows) {
		return StopRunOutput{}, NewError(CodeRunNotFound, "Run not found.")
	}
	if err != nil {
		return StopRunOutput{}, err
	}
	if IsRunTerminal(run.Status) {
		return StopRunOutput{Run: run, Result: StopRunResultAlreadyTerminal}, nil
	}
	var nextSequence int
	if err := tx.QueryRow(ctx, `select coalesce(max(sequence), 0) + 1 from run_events where run_id=$1`, run.ID).Scan(&nextSequence); err != nil {
		return StopRunOutput{}, err
	}
	lifecycle, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, $5, 'lifecycle', 'run_stopped', 'Run stopped', '{}'::jsonb) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, user.ID, nextSequence))
	if err != nil {
		return StopRunOutput{}, err
	}
	final, err := scanRunEvent(tx.QueryRow(ctx, `insert into run_events (id, run_id, thread_id, user_id, sequence, category, type, summary, metadata) values ($1, $2, $3, $4, $5, 'final', 'run_stopped', 'Run stopped', '{}'::jsonb) returning id, run_id, thread_id, user_id, sequence, category, type, summary, content, metadata, created_at`, NewRunEventID(), run.ID, run.ThreadID, user.ID, nextSequence+1))
	if err != nil {
		return StopRunOutput{}, err
	}
	stopped, err := scanRun(tx.QueryRow(ctx, `update runs set status='stopped', updated_at=now(), completed_at=now() where id=$1 and user_id=$2 returning id, thread_id, user_id, status, source, title, created_at, updated_at, completed_at, error_code, error_message`, run.ID, user.ID))
	if err != nil {
		return StopRunOutput{}, err
	}
	return StopRunOutput{Run: stopped, Result: StopRunResultStopped, Events: []RunEvent{lifecycle, final}}, tx.Commit(ctx)
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
	if err := row.Scan(&thread.ID, &thread.UserID, &thread.Title, &thread.Mode, &thread.LifecycleStatus, &thread.CreatedAt, &thread.UpdatedAt, &thread.ArchivedAt); err != nil {
		return Thread{}, err
	}
	return thread, nil
}

func scanRun(row scanner) (Run, error) {
	var run Run
	if err := row.Scan(&run.ID, &run.ThreadID, &run.UserID, &run.Status, &run.Source, &run.Title, &run.CreatedAt, &run.UpdatedAt, &run.CompletedAt, &run.ErrorCode, &run.ErrorMessage); err != nil {
		return Run{}, err
	}
	return run, nil
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
