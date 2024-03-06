CREATE TABLE IF NOT EXISTS oauth_sessions (
  "access_token" TEXT NOT NULL,
  "access_created_at" TEXT NOT NULL,
  "access_expires_at" TEXT NOT NULL,
  "refresh_token" TEXT NOT NULL,
  "refresh_created_at" TEXT NOT NULL,
  "refresh_expires_at" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "scope" TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
  FOREIGN KEY(client_id) REFERENCES oauth_clients(id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_sessions_access_token ON oauth_sessions(access_token);
CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_sessions_refresh_token ON oauth_sessions(refresh_token);
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_user_id ON oauth_sessions(user_id);
