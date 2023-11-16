CREATE TABLE IF NOT EXISTS config (
  "key" TEXT NOT NULL,
  "value" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_config_key ON config(key);
