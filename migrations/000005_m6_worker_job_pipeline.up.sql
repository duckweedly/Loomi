alter table runs drop constraint if exists runs_status_check;
alter table runs add constraint runs_status_check check (status in ('pending', 'queued', 'running', 'recovering', 'completed', 'failed', 'stopped'));

alter table runs drop constraint if exists runs_terminal_completed_at_check;
alter table runs add constraint runs_terminal_completed_at_check check ((status in ('completed', 'failed', 'stopped') and completed_at is not null) or (status in ('pending', 'queued', 'running', 'recovering') and completed_at is null));

alter table runs add column if not exists stop_requested_at timestamptz;

drop index if exists runs_one_active_per_thread_idx;
create unique index runs_one_active_per_thread_idx on runs (thread_id) where status in ('pending', 'queued', 'running', 'recovering');

create table background_jobs (
    id text primary key,
    run_id text not null references runs(id) on delete cascade,
    thread_id text not null references threads(id) on delete cascade,
    user_id text not null references users(id) on delete cascade,
    kind text not null,
    status text not null,
    priority integer not null default 100,
    attempt_count integer not null default 0,
    max_attempts integer not null default 3,
    scheduled_at timestamptz not null default now(),
    leased_by text,
    lease_expires_at timestamptz,
    ownership_version integer not null default 0,
    metadata jsonb not null default '{}'::jsonb,
    last_error_code text,
    last_error_message text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint background_jobs_kind_check check (kind in ('run_execution')),
    constraint background_jobs_status_check check (status in ('queued', 'leased', 'retrying', 'completed', 'failed', 'cancelled', 'dead')),
    constraint background_jobs_attempt_count_check check (attempt_count >= 0),
    constraint background_jobs_max_attempts_check check (max_attempts > 0),
    constraint background_jobs_lease_check check ((status in ('leased', 'retrying') and leased_by is not null and lease_expires_at is not null) or status not in ('leased', 'retrying'))
);

create unique index background_jobs_one_active_per_run_idx on background_jobs (run_id) where status in ('queued', 'leased', 'retrying');
create index background_jobs_claim_idx on background_jobs (status, scheduled_at asc, priority asc, created_at asc, id asc);
create index background_jobs_run_idx on background_jobs (run_id, created_at desc, id desc);
create index background_jobs_user_status_idx on background_jobs (user_id, status, updated_at desc);
