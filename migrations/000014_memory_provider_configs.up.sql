create table memory_provider_configs (
    user_id text primary key references users(id) on delete cascade,
    enabled boolean not null default true,
    provider text not null default 'local',
    commit_after_run boolean not null default false,
    semantic_endpoint text not null default '',
    diagnostic text not null default '',
    updated_at timestamptz not null default now(),
    constraint memory_provider_configs_provider_check check (provider <> '')
);
