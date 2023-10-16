CREATE TABLE IF NOT EXISTS tasks (
  "id" TEXT NOT NULL,
  "priority" NUMERIC NOT NULL,
  "name" TEXT NOT NULL,
  "status" TEXT NOT NULL,
  "retries" NUMERIC NOT NULL,
  "registered_at" DATETIME NOT NULL,
  "args" BLOB NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_id ON tasks(id);
CREATE INDEX IF NOT EXISTS idx_tasks_status_priority_registered ON tasks(status, priority, registered_at);
