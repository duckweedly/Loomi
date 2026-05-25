create table personas (
    id text primary key,
    slug text not null,
    name text not null,
    description text not null default '',
    source text not null,
    is_default boolean not null default false,
    is_active boolean not null default true,
    active_version text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    constraint personas_source_check check (source in ('built_in')),
    constraint personas_unique_slug_source unique (slug, source)
);

create table persona_versions (
    persona_id text not null references personas(id) on delete cascade,
    version text not null,
    system_prompt text not null,
    model_route jsonb not null default '{}'::jsonb,
    allowed_tool_names text[] not null default '{}',
    reasoning_mode text not null,
    budget_summary text not null,
    created_at timestamptz not null default now(),
    primary key (persona_id, version)
);

alter table threads add column persona_id text references personas(id);
alter table runs add column persona_id text references personas(id);

create table run_persona_snapshots (
    run_id text primary key references runs(id) on delete cascade,
    persona_id text not null references personas(id),
    persona_slug text not null,
    version text not null,
    name text not null,
    description text not null,
    system_prompt text not null,
    model_route jsonb not null default '{}'::jsonb,
    allowed_tool_names text[] not null default '{}',
    reasoning_mode text not null,
    budget_summary text not null,
    resolved_from text not null,
    created_at timestamptz not null default now(),
    constraint run_persona_snapshots_resolved_from_check check (resolved_from in ('run', 'thread', 'default'))
);

create index personas_active_default_idx on personas (source, is_default) where is_active=true;
create index threads_persona_idx on threads (persona_id);
create index runs_persona_idx on runs (persona_id);
