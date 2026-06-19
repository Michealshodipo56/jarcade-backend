-- Run in Supabase SQL Editor (Dashboard → SQL → New query)

create extension if not exists "pgcrypto";

create table if not exists public.users (
  id            uuid primary key default gen_random_uuid(),
  email         text unique not null,
  password_hash text not null,
  created_at    timestamptz not null default now()
);

create index if not exists idx_users_email_lower on public.users (lower(email));

alter table public.users enable row level security;

-- Required for Supabase REST API (service_role key used by the Go backend)
grant usage on schema public to postgres, anon, authenticated, service_role;
grant all on table public.users to postgres, service_role;

-- No public RLS policies: only the backend (service_role / direct Postgres) may access rows.
