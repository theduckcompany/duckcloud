CREATE TABLE IF NOT EXISTS tasks (
  "id" TEXT NOT NULL,
  "priority" INTEGER NOT NULL,
  "name" TEXT NOT NULL,
  "status" TEXT NOT NULL,
  "retries" INTEGER NOT NULL,
  "registered_at" TEXT NOT NULL,
  "args" BLOB NOT NULL
) STRICT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_id ON tasks(id);
CREATE INDEX IF NOT EXISTS idx_tasks_status_priority_registered ON tasks(status, priority, registered_at);
CREATE INDEX IF NOT EXISTS idx_tasks_name_registered ON tasks(name, registered_at);
