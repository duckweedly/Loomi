alter table tool_calls add column if not exists candidate_schema_hash text not null default '';
