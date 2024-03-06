CREATE TABLE IF NOT EXISTS oauth_consents (
  "id" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "scopes" TEXT NOT NULL,
  "session_token" TEXT NOT NULL,
  "created_at" TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
  FOREIGN KEY(client_id) REFERENCES oauth_clients(id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_consents_id ON oauth_consents(id);
CREATE INDEX IF NOT EXISTS idx_oauth_consents_user_id ON oauth_consents(user_id);
