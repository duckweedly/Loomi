create table artifacts (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    thread_id text not null references threads(id) on delete cascade,
    run_id text not null references runs(id) on delete cascade,
    title text not null,
    artifact_type text not null,
    content text not null,
    content_bytes integer not null,
    text_excerpt text not null,
    truncated boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index artifacts_user_thread_created_idx on artifacts (user_id, thread_id, created_at asc, id asc);
create index artifacts_user_thread_run_idx on artifacts (user_id, thread_id, run_id, created_at asc);

create table agent_tasks (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    thread_id text not null references threads(id) on delete cascade,
    run_id text not null references runs(id) on delete cascade,
    role text not null,
    goal text not null,
    status text not null,
    result_summary text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint agent_tasks_status_check check (status in ('spawned', 'completed'))
);

create index agent_tasks_user_thread_created_idx on agent_tasks (user_id, thread_id, created_at asc, id asc);
create index agent_tasks_user_thread_run_idx on agent_tasks (user_id, thread_id, run_id, created_at asc);
