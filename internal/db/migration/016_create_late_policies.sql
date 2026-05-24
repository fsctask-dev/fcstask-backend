-- +goose Up
CREATE TABLE IF NOT EXISTS late_policies (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hw_id       UUID NOT NULL REFERENCES homework(hw_id) ON DELETE CASCADE,
    soft_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    hard_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    soft_penalty  NUMERIC(4,2) NOT NULL DEFAULT 0.7,
    hard_penalty  NUMERIC(4,2) NOT NULL DEFAULT 0.0,
    UNIQUE(hw_id)
    );

-- +goose Down
DROP TABLE IF EXISTS late_policies;