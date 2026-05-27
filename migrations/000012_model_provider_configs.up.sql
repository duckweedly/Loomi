create table model_provider_configs (
    id text not null,
    user_id text not null references users(id) on delete cascade,
    family text not null,
    base_url text not null default '',
    api_key text not null,
    model text not null,
    enabled boolean not null default true,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (user_id, id),
    constraint model_provider_configs_family_check check (family in ('anthropic', 'openai', 'gemini', 'openai_compatible'))
);

create index model_provider_configs_user_idx on model_provider_configs (user_id, id);

create table web_search_configs (
    user_id text primary key references users(id) on delete cascade,
    tavily_api_key text not null default '',
    brave_api_key text not null default '',
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
