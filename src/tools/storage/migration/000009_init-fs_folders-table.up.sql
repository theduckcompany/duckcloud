CREATE TABLE IF NOT EXISTS fs_folders (
  "id" TEXT PRIMARY KEY,
  "name" TEXT NOT NULL,
  "public" BOOLEAN NOT NULL,
  "owners" TEXT NOT NULL,
  "root_fs" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);
