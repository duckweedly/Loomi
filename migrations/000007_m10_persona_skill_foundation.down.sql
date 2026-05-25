drop index if exists runs_persona_idx;
drop index if exists threads_persona_idx;
drop index if exists personas_active_default_idx;

drop table if exists run_persona_snapshots;

alter table runs drop column if exists persona_id;
alter table threads drop column if exists persona_id;

drop table if exists persona_versions;
drop table if exists personas;
