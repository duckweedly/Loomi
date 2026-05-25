create table memory_audit_events (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    thread_id text not null default '',
    run_id text not null default '',
    type text not null,
    summary text not null,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now()
);

create index memory_audit_events_user_created_idx on memory_audit_events (user_id, created_at desc, id desc);
create index memory_audit_events_run_idx on memory_audit_events (user_id, run_id, created_at desc) where run_id <> '';
create index memory_audit_events_thread_idx on memory_audit_events (user_id, thread_id, created_at desc) where thread_id <> '';
