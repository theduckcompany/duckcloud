CREATE TABLE IF NOT EXISTS fs_inodes (
  "id" TEXT PRIMARY KEY,
  "user_id" TEXT NOT NULL,
  "name" TEXT NOT NULL,
  "last_modified_at" DATETIME NOT NULL,
  "created_at" DATETIME NOT NULL
);
