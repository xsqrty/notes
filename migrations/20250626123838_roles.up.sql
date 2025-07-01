create table public.roles
(
    id          uuid primary key,
    description text        not null default '',
    permissions text[] not null default '{}'::text[],
    label       text        not null default '',
    created_at  timestamptz not null,
    updated_at  timestamptz
);

create table public.roles_users
(
    id         uuid primary key,
    user_id    uuid references public.users (id) on delete cascade,
    role_id    uuid references public.roles (id) on delete cascade,
    created_at timestamptz not null,
    unique (user_id, role_id)
);

-- roles indexes
create index idx_roles_label on public.roles (label);

-- roles_users indexes
create index idx_roles_users_user_id on public.roles_users (user_id);
create index idx_roles_users_role_id on public.roles_users (role_id);

-- create default role
insert into public.roles
    (id, description, permissions, label, created_at)
values (gen_random_uuid(),
        'Default user role',
        '{notes.read,notes.create,notes.update,notes.delete}'::text[],
        'on_created',
        current_timestamp);

-- attach default role to existed users
insert into public.roles_users
    (id, user_id, role_id, created_at)
select gen_random_uuid(), users.id, roles.id, current_timestamp
from public.users
         join public.roles on label = 'on_created';
