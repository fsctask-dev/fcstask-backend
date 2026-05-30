-- +goose Up
ALTER TABLE deadlines ADD COLUMN IF NOT EXISTS homework_id UUID REFERENCES homework(hw_id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_deadlines_homework_id ON deadlines(homework_id);

-- +goose Down
ALTER TABLE deadlines DROP COLUMN IF EXISTS homework_id;
