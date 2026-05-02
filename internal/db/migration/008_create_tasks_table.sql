CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS tasks (
    task_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hw_id UUID NOT NULL,
    repo_url VARCHAR(500),
    task_url VARCHAR(255)
)

CREATE INDEX idx_tasks_hw_id ON tasks(hw_id);

DROP TABLE IF EXISTS tasks;