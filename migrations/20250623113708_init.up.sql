create table public.users
(
    id              uuid primary key,
    name            text        not null default '',
    email           text        not null,
    hashed_password text        not null,
    created_at      timestamptz not null,
    updated_at      timestamptz
);

create unique index idx_unique_users_email on public.users (email);