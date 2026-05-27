create table workspace_root_configs (
    user_id text primary key references users(id) on delete cascade,
    root_path text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);
