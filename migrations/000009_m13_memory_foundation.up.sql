create table memory_entries (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    scope_type text not null check (scope_type in ('user', 'thread')),
    scope_id text not null,
    title text not null,
    summary text not null,
    content text not null default '',
    status text not null check (status in ('approved', 'tombstoned', 'disabled')),
    safety_state text not null check (safety_state in ('safe', 'redacted', 'blocked')),
    source_thread_id text null,
    source_run_id text null,
    source_event_id text null,
    content_hash text not null,
    deleted_at timestamptz null,
    deleted_by_user_id text null,
    delete_reason text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index memory_entries_visible_idx on memory_entries (user_id, scope_type, scope_id, status, updated_at desc);
create index memory_entries_content_hash_idx on memory_entries (user_id, content_hash);

create table memory_write_proposals (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    scope_type text not null check (scope_type in ('user', 'thread')),
    scope_id text not null,
    title text not null,
    summary text not null,
    content text not null default '',
    status text not null check (status in ('pending', 'approved', 'denied')),
    safety_state text not null check (safety_state in ('safe', 'redacted', 'blocked')),
    source_thread_id text null,
    source_run_id text null,
    source_event_id text null,
    idempotency_key text not null default '',
    created_entry_id text null references memory_entries(id),
    decided_at timestamptz null,
    decided_by_user_id text null,
    decision_reason text not null default '',
    created_at timestamptz not null default now()
);

create unique index memory_write_proposals_idempotency_idx on memory_write_proposals (user_id, idempotency_key) where idempotency_key <> '';
