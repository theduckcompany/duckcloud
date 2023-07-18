CREATE TABLE IF NOT EXISTS web_sessions (
  "token" TEXT PRIMARY KEY,
  "data" BLOB NOT NULL,
  "expiry" DATETIME NOT NULL
);

CREATE INDEX idx_web_sessions_expiry ON web_sessions(expiry);
