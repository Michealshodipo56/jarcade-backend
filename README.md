# JARCADE Backend API

Express + SQLite API for user authentication and favourites.

## Setup

```bash
cd jarcade-backend
npm install
cp .env.example .env   # if .env does not exist — set JWT_SECRET
npm run dev
```

API runs at **http://localhost:3001**

## Frontend

Serve the `jarcade/` folder with any static server (e.g. VS Code Live Server on port 5500).

The client auto-connects to `http://localhost:3001/api` when opened from localhost.

Override with:

```html
<script>window.JARCADE_API_URL = 'https://your-api.com/api';</script>
```

## Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| GET | `/api/health` | No | Health check |
| POST | `/api/auth/register` | No | Create account |
| POST | `/api/auth/login` | No | Sign in |
| GET | `/api/auth/me` | Bearer | Current user |
| POST | `/api/auth/logout` | No | Logout (client clears JWT) |
| GET | `/api/favourites` | Bearer | List favourites |
| POST | `/api/favourites` | Bearer | Add/update favourite |
| DELETE | `/api/favourites/:name` | Bearer | Remove favourite |

## Database

SQLite file at `./data/jarcade.db` (created automatically).

**Tables:**
- `users` — id, username, email, password_hash, created_at
- `favourites` — user_id, game_name, game_img, play_onclick

## Deploy on Render

1. Push this repo to GitHub and connect it on [Render](https://render.com).
2. Create a **Web Service** from the repo (or use `render.yaml` Blueprint).
3. Set environment variables:
   - `JWT_SECRET` — long random string (Render can auto-generate)
   - `CORS_ORIGIN` — your frontend URL(s), comma-separated  
     e.g. `https://michealshodipo56.github.io,http://localhost:5500`
4. **Persistent disk (recommended):** attach a disk at `/var/data` and set  
   `DATABASE_PATH=/var/data/jarcade.db`  
   Without a disk, SQLite data resets on redeploy.
5. After deploy, copy your service URL and set in the frontend `config.js`:
   ```js
   window.JARCADE_API_URL = 'https://YOUR-SERVICE.onrender.com/api';
   ```

## Environment

See `.env.example` for all variables.
