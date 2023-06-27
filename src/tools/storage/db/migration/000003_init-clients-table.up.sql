CREATE TABLE IF NOT EXISTS oauth_clients (
  "id" TEXT  NOT NULL PRIMARY KEY,
  "secret" TEXT NOT NULL,
  "redirect_uri" TEXT NOT NULL,
  "user_id" TEXT,
  "created_at" DATETIME NOT NULL,
  "scopes" TEXT NOT NULL,
  "is_public" BOOLEAN NOT NULL,
  "skip_validation" BOOLEAN NOT NULL
);
