CREATE TABLE IF NOT EXISTS oauth_codes (
  "code" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  "expires_at" DATETIME NOT NULL,
  "user_id" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "redirect_uri" TEXT NOT NULL,
  "challenge" TEXT DEFAULT NULL,
  "challenge_method" TEXT DEFAULT NULL,
  "scope" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_codes_code ON oauth_codes(code);
