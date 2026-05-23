create table users (
    id text primary key,
    display_name text not null,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create table threads (
    id text primary key,
    user_id text not null references users(id) on delete cascade,
    title text not null check (char_length(trim(title)) between 1 and 120),
    mode text not null check (mode in ('chat', 'work')),
    lifecycle_status text not null check (lifecycle_status in ('active', 'archived')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    archived_at timestamptz
);

create index threads_user_active_updated_idx on threads (user_id, lifecycle_status, updated_at desc);

create table messages (
    id text primary key,
    thread_id text not null references threads(id) on delete cascade,
    user_id text not null references users(id) on delete cascade,
    role text not null check (role = 'user'),
    content text not null check (char_length(trim(content)) > 0),
    metadata jsonb not null default '{}'::jsonb,
    client_message_id text,
    created_at timestamptz not null default now()
);

create index messages_thread_created_idx on messages (thread_id, created_at asc, id asc);
create unique index messages_client_id_unique_idx on messages (thread_id, user_id, client_message_id) where client_message_id is not null;
