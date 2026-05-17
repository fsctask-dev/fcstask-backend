-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS scores (ls
    id SERIAL PRIMARY KEY,
    student VARCHAR(255) NOT NULL,
    score INTEGER NOT NULL DEFAULT 0,
    submitted_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    course_id UUID NOT NULL REFERENCES courses(id)
);

CREATE INDEX IF NOT EXISTS idx_scores_course_id ON scores(course_id);
CREATE INDEX IF NOT EXISTS idx_scores_student ON scores(student);

-- +goose Down
DROP TABLE IF EXISTS scores;