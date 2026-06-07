-- +goose Up
ALTER TABLE deadlines DROP COLUMN IF EXISTS due_date;

-- +goose Down
ALTER TABLE deadlines ADD COLUMN IF NOT EXISTS due_date TIMESTAMPTZ;
