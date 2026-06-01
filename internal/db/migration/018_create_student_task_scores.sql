-- +goose Up
CREATE TABLE IF NOT EXISTS student_task_scores (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    task_id UUID NOT NULL REFERENCES tasks(task_id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    score INT NOT NULL DEFAULT 0,
    is_passed BOOLEAN NOT NULL DEFAULT false,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(student_id, task_id)
);

CREATE INDEX IF NOT EXISTS idx_sts_student ON student_task_scores(student_id);
CREATE INDEX IF NOT EXISTS idx_sts_course ON student_task_scores(course_id);
CREATE INDEX IF NOT EXISTS idx_sts_task ON student_task_scores(task_id);

-- +goose Down
DROP TABLE IF EXISTS student_task_scores;