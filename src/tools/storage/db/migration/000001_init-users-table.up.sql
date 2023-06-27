CREATE TABLE IF NOT EXISTS users (
  "id" TEXT NOT NULL PRIMARY KEY,
  "username" TEXT NOT NULL,
  "email" TEXT NOT NULL,
  "password" TEXT NOT NULL,
  "created_at" DATETIME NOT NULL
);
