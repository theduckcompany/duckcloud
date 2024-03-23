CREATE TABLE IF NOT EXISTS fs_inodes (
  "id" TEXT NOT NULL,
  "parent" TEXT DEFAULT NULL,
  "name" TEXT NOT NULL,
  "size" INTEGER NOT NULL,
  "space_id" TEXT NOT NULL,
  "file_id" TEXT DEFAULT NULL,
  "last_modified_at" TEXT NOT NULL,
  "created_at" TEXT NOT NULL,
  "created_by" TEXT NOT NULL,
  "deleted_at" TEXT DEFAULT NULL,
  FOREIGN KEY(space_id) REFERENCES spaces(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
  FOREIGN KEY(file_id) REFERENCES files(id) ON UPDATE RESTRICT ON DELETE RESTRICT
  FOREIGN KEY(parent) REFERENCES fs_inodes(id) ON UPDATE RESTRICT ON DELETE RESTRICT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_inodes_id ON fs_inodes(id);
CREATE INDEX IF NOT EXISTS idx_fs_inodes_parent_name ON fs_inodes(parent, name);
CREATE INDEX IF NOT EXISTS idx_fs_inodes_deleted ON fs_inodes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_fs_inodes_file_id ON fs_inodes(file_id);
