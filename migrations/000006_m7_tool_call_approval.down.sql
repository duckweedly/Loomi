drop table if exists tool_calls;

drop index if exists runs_one_active_per_thread_idx;
create unique index runs_one_active_per_thread_idx on runs (thread_id) where status in ('pending', 'queued', 'running', 'recovering');

alter table runs drop constraint if exists runs_terminal_completed_at_check;
alter table runs add constraint runs_terminal_completed_at_check check ((status in ('completed', 'failed', 'stopped') and completed_at is not null) or (status in ('pending', 'queued', 'running', 'recovering') and completed_at is null));

alter table runs drop constraint if exists runs_status_check;
alter table runs add constraint runs_status_check check (status in ('pending', 'queued', 'running', 'recovering', 'completed', 'failed', 'stopped'));
