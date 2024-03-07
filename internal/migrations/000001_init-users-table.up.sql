CREATE TABLE IF NOT EXISTS users (
  "id" TEXT NOT NULL,
  "username" TEXT NOT NULL,
  "admin" INTEGER NOT NULL,
  "password" TEXT NOT NULL,
  "status" TEXT NOT NULL,
  "password_changed_at" TEXT NOT NULL,
  "created_at" TEXT NOT NULL,
  "created_by" TEXT NOT NULL
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_id ON users(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
