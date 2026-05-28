create table sandbox_process_records (
    process_id text primary key,
    run_id text not null references runs(id) on delete cascade,
    argv_summary jsonb not null default '[]'::jsonb,
    cwd_alias text not null default '.',
    status text not null,
    cursor integer not null default 0,
    stdout_tail text not null default '',
    stdout_cursor integer not null default 0,
    stderr_tail text not null default '',
    stderr_cursor integer not null default 0,
    stdout_bytes integer not null default 0,
    stderr_bytes integer not null default 0,
    stdin_open boolean not null default false,
    input_seq integer not null default 0,
    timed_out boolean not null default false,
    started_at timestamptz not null,
    updated_at timestamptz not null,
    ended_at timestamptz,
    exit_code integer,
    terminal_summary text not null default '',
    output_limit integer not null default 0,
    constraint sandbox_process_records_status_check check (status in ('running', 'exited', 'terminated', 'failed', 'expired', 'lost'))
);

create index sandbox_process_records_run_idx on sandbox_process_records (run_id, process_id);
create index sandbox_process_records_updated_idx on sandbox_process_records (updated_at asc, process_id asc);
