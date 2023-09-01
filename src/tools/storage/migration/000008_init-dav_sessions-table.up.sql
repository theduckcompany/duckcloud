CREATE TABLE IF NOT EXISTS dav_sessions (
  "id" TEXT PRIMARY KEY,
  "username" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "folders" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dav_sessions_username_password ON dav_sessions (username, password);
