CREATE TABLE IF NOT EXISTS fs_folders (
  "id" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "public" BOOLEAN NOT NULL,
  "owners" TEXT NOT NULL,
  "size" NUMERIC NOT NULL,
  "root_fs" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  "last_modified_at" DATETIME NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_folders_id ON fs_folders(id);
CREATE INDEX IF NOT EXISTS idx_fs_folders_root_fs ON fs_folders(root_fs);
