# JARCADE Backend API

Production Go API for user authentication, backed by **Supabase PostgreSQL**.

## Stack

- **Go** + [chi](https://github.com/go-chi/chi) router
- **Supabase PostgreSQL** (`users` table)
- **bcrypt** password hashing
- **JWT** session tokens (HS256)

## Setup

### 1. Supabase

1. Create a project at [supabase.com](https://supabase.com).
2. Open **SQL Editor** and run `supabase/migrations/001_users.sql`.
3. Copy **Transaction pooler** connection string → `DATABASE_URL` (see Render note below).

### 2. Local API

```bash
cd jarcade-backend
cp .env.example .env   # fill in secrets
go run ./cmd/server
```

API runs at **http://localhost:8080**

### 3. Frontend

Serve the `jarcade/` folder (e.g. Live Server on port 5500).  
The client auto-connects to `http://localhost:8080/api` on localhost.

After deploying, set in `jarcade/config.js`:

```js
window.JARCADE_API_URL = 'https://YOUR-SERVICE.onrender.com/api';
```

Set `CORS_ORIGIN` on Render to your frontend URL(s).

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Health check |
| POST | `/api/auth/signup` | No | Create account |
| POST | `/api/auth/login` | No | Sign in (rate limited) |
| GET | `/api/auth/me` | Bearer | Current user |
| POST | `/api/auth/logout` | No | Logout (client clears JWT) |

## Database

`users` table (UUID primary key):

```sql
id            uuid
email         text unique
password_hash text
created_at    timestamptz
```

Plain passwords are never stored.

## Deploy on Render

1. Push this repo to GitHub and connect on [Render](https://render.com).
2. Use `render.yaml` or create a **Web Service** with runtime **Go**.
3. Set environment variables:
   - `DATABASE_URL` — **Transaction pooler** URI (port **6543**), **not** `db.*.supabase.co:5432`
   - `JWT_SECRET` — long random string (≥ 32 chars)
   - `CORS_ORIGIN` — comma-separated frontend origins

**Important:** Render cannot connect to Supabase's direct database host (`db.xxxx.supabase.co:5432`).  
In Supabase → **Settings → Database → Connection string**, choose **URI** and **Transaction pooler**:

```text
postgresql://postgres.[project-ref]:[password]@aws-0-[region].pooler.supabase.com:6543/postgres
```

4. Deploy and point the frontend `config.js` at your service URL.

## Environment

See `.env.example` for all variables.
