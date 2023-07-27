CREATE TABLE IF NOT EXISTS web_sessions (
  "token" TEXT PRIMARY KEY,
  "user_id" TEXT NOT NULL,
  "ip" TEXT NOT NULL,
  "client_id" TEXT NOT NULL,
  "device" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);
