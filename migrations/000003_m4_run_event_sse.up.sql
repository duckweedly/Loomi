create table runs (
    id text primary key,
    thread_id text not null references threads(id) on delete cascade,
    user_id text not null references users(id) on delete cascade,
    status text not null,
    source text not null,
    title text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    completed_at timestamptz,
    error_code text,
    error_message text,
    constraint runs_status_check check (status in ('pending', 'running', 'completed', 'failed', 'stopped')),
    constraint runs_source_check check (source in ('local_simulated')),
    constraint runs_terminal_completed_at_check check ((status in ('completed', 'failed', 'stopped') and completed_at is not null) or (status in ('pending', 'running') and completed_at is null))
);

create unique index runs_one_active_per_thread_idx on runs (thread_id) where status in ('pending', 'running');
create index runs_thread_updated_idx on runs (thread_id, updated_at desc, id desc);
create index runs_user_updated_idx on runs (user_id, updated_at desc, id desc);

create table run_events (
    id text primary key,
    run_id text not null references runs(id) on delete cascade,
    thread_id text not null references threads(id) on delete cascade,
    user_id text not null references users(id) on delete cascade,
    sequence integer not null,
    category text not null,
    type text not null,
    summary text not null,
    content text,
    metadata jsonb not null default '{}'::jsonb,
    created_at timestamptz not null default now(),
    constraint run_events_sequence_check check (sequence >= 1),
    constraint run_events_category_check check (category in ('lifecycle', 'progress', 'message', 'error', 'final')),
    constraint run_events_run_sequence_unique unique (run_id, sequence)
);

create index run_events_run_sequence_idx on run_events (run_id, sequence asc, id asc);
create index run_events_thread_created_idx on run_events (thread_id, created_at asc, id asc);
