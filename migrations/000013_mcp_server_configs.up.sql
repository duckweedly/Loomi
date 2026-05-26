create table mcp_server_configs (
    user_id text not null references users(id) on delete cascade,
    slug text not null,
    display_name text not null,
    enabled boolean not null default true,
    transport text not null,
    command text not null default '',
    args_json jsonb not null default '[]'::jsonb,
    env_json jsonb not null default '{}'::jsonb,
    timeout_ms integer not null default 5000,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (user_id, slug),
    constraint mcp_server_configs_transport_check check (transport in ('stdio')),
    constraint mcp_server_configs_timeout_check check (timeout_ms between 100 and 60000)
);

create index mcp_server_configs_user_idx on mcp_server_configs (user_id, slug);
