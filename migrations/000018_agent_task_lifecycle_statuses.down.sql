update agent_tasks set status = 'completed' where status in ('in_progress', 'failed');
alter table agent_tasks drop constraint if exists agent_tasks_status_check;
alter table agent_tasks add constraint agent_tasks_status_check check (status in ('spawned', 'completed'));
