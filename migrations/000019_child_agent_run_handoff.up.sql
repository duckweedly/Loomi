alter table agent_tasks add column if not exists child_thread_id text references threads(id) on delete set null;
alter table agent_tasks add column if not exists child_run_id text references runs(id) on delete set null;
alter table agent_tasks add column if not exists parent_tool_call_id text;
alter table agent_tasks add column if not exists delegated_at timestamptz;

create index if not exists agent_tasks_user_child_run_idx on agent_tasks (user_id, child_run_id) where child_run_id is not null;
