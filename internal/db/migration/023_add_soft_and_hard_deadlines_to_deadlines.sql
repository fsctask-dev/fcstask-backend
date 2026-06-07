-- +goose Up
ALTER TABLE deadlines ADD COLUMN IF NOT EXISTS soft_deadline TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_deadlines_soft_deadline ON deadlines(soft_deadline);
ALTER TABLE deadlines ADD COLUMN IF NOT EXISTS hard_deadline TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_deadlines_hard_deadline ON deadlines(hard_deadline);

-- +goose Down
ALTER TABLE deadlines DROP COLUMN IF EXISTS soft_deadline;
ALTER TABLE deadlines DROP COLUMN IF EXISTS hard_deadline
