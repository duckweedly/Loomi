create table if not exists context_sources (
  id text primary key,
  user_id text not null references users(id) on delete cascade,
  thread_id text not null references threads(id) on delete cascade,
  kind text not null,
  title text not null,
  locator text not null default '',
  summary text not null default '',
  status text not null,
  metadata jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint context_sources_kind_check check (kind in ('url', 'github_repo', 'workspace_path', 'note')),
  constraint context_sources_status_check check (status in ('registered'))
);

create index if not exists context_sources_user_thread_created_idx on context_sources(user_id, thread_id, created_at asc, id asc);
