-- +goose Up
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS is_public BOOL DEFAULT false;

-- +goose Down
ALTER TABLE tasks DROP COLUMN IF EXISTS is_public;