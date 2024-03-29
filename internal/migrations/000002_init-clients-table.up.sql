CREATE TABLE IF NOT EXISTS oauth_clients (
  "id" TEXT  NOT NULL,
  "name" TEXT NOT NULL,
  "secret" TEXT NOT NULL,
  "redirect_uri" TEXT NOT NULL,
  "user_id" TEXT,
  "created_at" TEXT NOT NULL,
  "scopes" TEXT NOT NULL,
  "is_public" INTEGER NOT NULL,
  "skip_validation" INTEGER NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE RESTRICT
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_clients_id ON oauth_clients(id);
