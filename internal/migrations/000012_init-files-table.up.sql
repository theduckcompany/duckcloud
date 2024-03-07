CREATE TABLE IF NOT EXISTS files (
  "id" TEXT NOT NULL,
  "size" REAL NOT NULL,
  "mimetype" TEXT DEFAULT NULL,
  "checksum" TEXT NOT NULL,
  "key" BLOB NOT NULL,
  "uploaded_at" TEXT NOT NULL
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_files_id ON files(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_files_checksum ON files(checksum);
