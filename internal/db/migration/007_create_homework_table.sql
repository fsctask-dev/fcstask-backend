-- +goose Up
CREATE TABLE IF NOT EXISTS homework (
    hw_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL,
    is_public BOOL,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_homework_course_id ON homework(course_id);
CREATE INDEX idx_homework_start_date ON homework(start_date);
CREATE INDEX idx_homework_end_date ON homework(end_date);

-- +goose Down
DROP TABLE IF EXISTS homework;