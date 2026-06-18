-- Run in Supabase SQL Editor (Dashboard → SQL → New query)

create extension if not exists "pgcrypto";

create table if not exists public.users (
  id            uuid primary key default gen_random_uuid(),
  email         text unique not null,
  password_hash text not null,
  created_at    timestamptz not null default now()
);

create index if not exists idx_users_email_lower on public.users (lower(email));

-- PostgREST (Supabase REST) access for service role only.
alter table public.users enable row level security;

-- No public policies: the Go API uses service role / direct Postgres only.
