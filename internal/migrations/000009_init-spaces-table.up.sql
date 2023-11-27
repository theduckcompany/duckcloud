CREATE TABLE IF NOT EXISTS spaces (
  "id" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "public" BOOLEAN NOT NULL,
  "owners" TEXT NOT NULL,
  "root_fs" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  "created_by" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_spaces_id ON spaces(id);
CREATE INDEX IF NOT EXISTS idx_spaces_root_fs ON spaces(root_fs);
