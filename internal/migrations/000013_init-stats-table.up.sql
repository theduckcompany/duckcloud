CREATE TABLE IF NOT EXISTS stats (
  "key" TEXT NOT NULL,
  "value" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_stats_key ON stats(key);
