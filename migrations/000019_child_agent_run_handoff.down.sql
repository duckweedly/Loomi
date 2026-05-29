drop index if exists agent_tasks_user_child_run_idx;

alter table agent_tasks drop column if exists delegated_at;
alter table agent_tasks drop column if exists parent_tool_call_id;
alter table agent_tasks drop column if exists child_run_id;
alter table agent_tasks drop column if exists child_thread_id;
