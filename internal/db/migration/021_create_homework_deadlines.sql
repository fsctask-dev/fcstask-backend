-- +goose Up
CREATE TABLE IF NOT EXISTS homework_deadlines (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hw_id         UUID NOT NULL UNIQUE,
    soft_deadline TIMESTAMPTZ NOT NULL,
    hard_deadline TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS homework_deadlines;
