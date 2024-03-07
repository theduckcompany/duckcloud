CREATE TABLE IF NOT EXISTS web_sessions (
  "token" TEXT NOT NULL,
  "user_id" TEXT NOT NULL,
  "ip" TEXT NOT NULL,
  "device" TEXT NOT NULL,
  "created_at" TEXT NOT NULL,
  FOREIGN KEY(user_id) REFERENCES users(id) ON UPDATE RESTRICT ON DELETE RESTRICT
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_web_sessions_token ON web_sessions(token);
CREATE INDEX IF NOT EXISTS idx_web_sessions_user_id ON web_sessions(user_id);
