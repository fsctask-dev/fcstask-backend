-- +goose Up
CREATE TABLE IF NOT EXISTS course_late_policies (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id   UUID        NOT NULL UNIQUE,
    policy_type TEXT        NOT NULL CHECK (policy_type IN ('linear','step','coefficient')),
    soft_penalty NUMERIC(5,4) NOT NULL DEFAULT 0,
    hard_deadline_score NUMERIC(5,4) NOT NULL DEFAULT 0,
    step_percent NUMERIC(5,4),
    coefficient  NUMERIC(5,4),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
    );
-- +goose Down
DROP TABLE IF EXISTS course_late_policies;
