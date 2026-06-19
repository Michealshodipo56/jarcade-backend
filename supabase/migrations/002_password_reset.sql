-- Run in Supabase SQL Editor

create table if not exists public.password_reset_tokens (
  id         uuid primary key default gen_random_uuid(),
  user_id    uuid not null references public.users(id) on delete cascade,
  token_hash text not null,
  expires_at timestamptz not null,
  used_at    timestamptz,
  created_at timestamptz not null default now()
);

create index if not exists idx_password_reset_user on public.password_reset_tokens(user_id);
create index if not exists idx_password_reset_expires on public.password_reset_tokens(expires_at);

grant all on table public.password_reset_tokens to postgres, service_role;
