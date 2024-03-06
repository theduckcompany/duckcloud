CREATE TABLE IF NOT EXISTS spaces (
  "id" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "owners" TEXT NOT NULL,
  "created_at" TEXT NOT NULL,
  "created_by" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spaces_id ON spaces(id);
