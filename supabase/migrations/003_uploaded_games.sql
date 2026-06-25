-- Community game uploads (metadata; .jar files stored client-side until object storage is added)
CREATE TABLE IF NOT EXISTS uploaded_games (
  id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name        text NOT NULL,
  category    text NOT NULL,
  description text NOT NULL DEFAULT '',
  file_name   text NOT NULL DEFAULT '',
  file_size   bigint NOT NULL DEFAULT 0,
  thumbnail   text NOT NULL DEFAULT '',
  created_at  timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_uploaded_games_user ON uploaded_games(user_id);
CREATE INDEX IF NOT EXISTS idx_uploaded_games_category ON uploaded_games(category);
CREATE INDEX IF NOT EXISTS idx_uploaded_games_created ON uploaded_games(created_at DESC);

GRANT ALL ON uploaded_games TO postgres;
GRANT ALL ON uploaded_games TO service_role;
