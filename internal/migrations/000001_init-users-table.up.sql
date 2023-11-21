CREATE TABLE IF NOT EXISTS users (
  "id" TEXT NOT NULL,
  "username" TEXT NOT NULL,
  "admin" BOOL NOT NULL,
  "password" TEXT NOT NULL,
  "space" TEXT NOT NULL,
  "status" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_id ON users(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_username ON users(username);
