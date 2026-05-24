-- +goose Up
ALTER TABLE courses
    DROP CONSTRAINT IF EXISTS courses_invite_code_key;

DROP INDEX IF EXISTS idx_courses_invite_code;
DROP INDEX IF EXISTS idx_courses_invite_code_unique;

CREATE INDEX IF NOT EXISTS idx_courses_invite_code
    ON courses(invite_code);

-- +goose Down
DROP INDEX IF EXISTS idx_courses_invite_code;

ALTER TABLE courses
    ADD CONSTRAINT courses_invite_code_key UNIQUE (invite_code);
