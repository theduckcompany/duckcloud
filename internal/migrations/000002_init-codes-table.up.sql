CREATE TABLE IF NOT EXISTS oauth_codes (
  "code" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  "expires_at" DATETIME NOT NULL,
  "user_id" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "redirect_uri" TEXT NOT NULL,
  "challenge" TEXT DEFAULT NULL,
  "challenge_method" TEXT DEFAULT NULL,
  "scope" TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id),
  FOREIGN KEY(client_id) REFERENCES oauth_clients(id) 
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_codes_code ON oauth_codes(code);
