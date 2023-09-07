CREATE TABLE IF NOT EXISTS fs_inodes (
  "id" TEXT NOT NULL,
  "parent" TEXT DEFAULT NULL,
  "name" TEXT NOT NULL,
  "size" NUMERIC NOT NULL,
  "checksum" TEXT NOT NULL,
  "mode" NUMERIC NOT NULL,
  "last_modified_at" DATETIME NOT NULL,
  "created_at" DATETIME NOT NULL,
  "deleted_at" DATETIME DEFAULT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_inodes_id ON fs_inodes(id);
CREATE INDEX IF NOT EXISTS idx_fs_inodes_parent_name ON fs_inodes(parent, name);
CREATE INDEX IF NOT EXISTS idx_fs_inodes_deleted ON fs_inodes(deleted_at);
