CREATE TABLE IF NOT EXISTS fs_uploads (
  "id" TEXT NOT NULL,
  "folder_id" TEXT NOT NULL,
  "file_id" TEXT NOT NULL,
  "directory" TEXT NOT NULL,
  "file_name" TEXT NOT NULL,
  "uploaded_at" DATETIME NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_fs_uploads_id ON fs_uploads(id);
CREATE INDEX IF NOT EXISTS idx_fs_uploads_uploaded_at ON fs_uploads(uploaded_at);
