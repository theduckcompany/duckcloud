CREATE TABLE IF NOT EXISTS files (
  "id" TEXT NOT NULL,
  "size" NUMERIC NOT NULL,
  "mimetype" TEXT DEFAULT NULL,
  "checksum" TEXT NOT NULL,
  "key" BINARY NOT NULL,
  "uploaded_at" TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_files_id ON files(id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_files_checksum ON files(checksum);
