CREATE TABLE IF NOT EXISTS fs_inodes (
  "id" TEXT PRIMARY KEY,
  "user_id" TEXT NOT NULL,
  "parent" TEXT DEFAULT NULL,
  "name" TEXT NOT NULL,
  "size" NUMERIC NOT NULL,
  "checksum" TEXT NOT NULL,
  "mode" NUMERIC NOT NULL,
  "last_modified_at" DATETIME NOT NULL,
  "created_at" DATETIME NOT NULL,
  "deleted_at" DATETIME DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_fs_inodes_owner_parent_name ON fs_inodes (user_id, parent, name);
