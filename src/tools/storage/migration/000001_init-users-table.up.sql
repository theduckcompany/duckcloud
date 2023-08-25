CREATE TABLE IF NOT EXISTS users (
  "id" TEXT NOT NULL PRIMARY KEY,
  "username" TEXT NOT NULL,
  "admin" BOOL NOT NULL,
  "fs_root" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL,
  "deleted_at" DATETIME DEFAULT NULL
);
