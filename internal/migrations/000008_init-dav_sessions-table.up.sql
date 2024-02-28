CREATE TABLE IF NOT EXISTS dav_sessions (
  "id" TEXT NOT NULL,
  "username" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "space_id" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id)
  FOREIGN KEY(space_id) REFERENCES spaces(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_dav_sessions_id ON dav_sessions(id);
CREATE INDEX IF NOT EXISTS idx_dav_sessions_user_id ON dav_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_dav_sessions_username_password ON dav_sessions(username, password);
