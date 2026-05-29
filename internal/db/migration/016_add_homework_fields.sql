-- +goose Up
ALTER TABLE homework ADD COLUMN title VARCHAR(255);
ALTER TABLE homework ADD COLUMN description TEXT;
ALTER TABLE homework ADD COLUMN position INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_homework_position
    ON homework(course_id, position);

-- +goose Down
DROP INDEX IF EXISTS idx_homework_position;
ALTER TABLE homework DROP COLUMN IF EXISTS position;
ALTER TABLE homework DROP COLUMN IF EXISTS description;
ALTER TABLE homework DROP COLUMN IF EXISTS title;
