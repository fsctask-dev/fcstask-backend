-- +goose Up
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS title VARCHAR(255) NOT NULL DEFAULT 'Untitled';

-- +goose Down
ALTER TABLE tasks DROP COLUMN IF EXISTS title;