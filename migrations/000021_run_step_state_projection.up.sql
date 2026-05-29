create table run_step_state_projections (
    run_id text primary key references runs(id) on delete cascade,
    thread_id text not null references threads(id) on delete cascade,
    user_id text not null references users(id) on delete cascade,
    last_sequence integer not null default 0,
    state jsonb not null default '{}'::jsonb,
    updated_at timestamptz not null default now(),
    constraint run_step_state_last_sequence_check check (last_sequence >= 0)
);

create index run_step_state_projections_user_updated_idx on run_step_state_projections (user_id, updated_at desc);
