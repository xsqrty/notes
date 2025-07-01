create table public.notes
(
    id         uuid primary key,
    name       text        not null default '',
    text       text        not null default '',
    user_id    uuid        not null references public.users (id) on delete cascade,
    created_at timestamptz not null,
    updated_at timestamptz
);

create index idx_notes_user_id on public.notes (user_id);