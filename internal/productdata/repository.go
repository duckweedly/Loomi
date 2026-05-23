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
