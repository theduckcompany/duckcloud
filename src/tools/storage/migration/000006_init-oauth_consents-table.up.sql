CREATE TABLE IF NOT EXISTS oauth_consents (
  "id" TEXT PRIMARY KEY,
  "user_id" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "scopes" TEXT NOT NULL,
  "session_token" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);
