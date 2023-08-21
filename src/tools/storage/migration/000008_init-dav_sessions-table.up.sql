CREATE TABLE IF NOT EXISTS dav_sessions (
  "id" TEXT PRIMARY KEY,
  "username" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "fs_root" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_dav_sessions_username_password ON dav_sessions (username, password);
