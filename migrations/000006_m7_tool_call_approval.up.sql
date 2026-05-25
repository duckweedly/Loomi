alter table runs drop constraint if exists runs_status_check;
alter table runs add constraint runs_status_check check (status in ('pending', 'queued', 'running', 'recovering', 'blocked_on_tool_approval', 'completed', 'failed', 'stopped'));

alter table runs drop constraint if exists runs_terminal_completed_at_check;
alter table runs add constraint runs_terminal_completed_at_check check ((status in ('completed', 'failed', 'stopped') and completed_at is not null) or (status in ('pending', 'queued', 'running', 'recovering', 'blocked_on_tool_approval') and completed_at is null));

drop index if exists runs_one_active_per_thread_idx;
create unique index runs_one_active_per_thread_idx on runs (thread_id) where status in ('pending', 'queued', 'running', 'recovering', 'blocked_on_tool_approval');

create table tool_calls (
    id text primary key,
    thread_id text not null references threads(id) on delete cascade,
    run_id text not null references runs(id) on delete cascade,
    tool_call_id text not null,
    tool_name text not null,
    arguments_summary jsonb not null default '{}'::jsonb,
    arguments_hash text not null,
    approval_status text not null,
    execution_status text not null,
    result_summary jsonb,
    error_code text,
    error_message text,
    requested_at timestamptz not null default now(),
    decided_at timestamptz,
    started_at timestamptz,
    completed_at timestamptz,
    updated_at timestamptz not null default now(),
    constraint tool_calls_unique_call_per_run unique (run_id, tool_call_id),
    constraint tool_calls_approval_status_check check (approval_status in ('not_required', 'required', 'approved', 'denied', 'cancelled')),
    constraint tool_calls_execution_status_check check (execution_status in ('not_started', 'blocked', 'executing', 'succeeded', 'failed', 'cancelled'))
);

create index tool_calls_run_idx on tool_calls (run_id, requested_at asc, id asc);
create index tool_calls_thread_run_idx on tool_calls (thread_id, run_id, tool_call_id);
