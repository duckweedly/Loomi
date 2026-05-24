alter table messages drop constraint if exists messages_role_check;
alter table messages add constraint messages_role_check check (role in ('user', 'assistant'));

alter table runs drop constraint if exists runs_source_check;
alter table runs add constraint runs_source_check check (source in ('local_simulated', 'model_gateway'));
