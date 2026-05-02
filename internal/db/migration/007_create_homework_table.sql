CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS homework (
    hw_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id UUID NOT NULL,
    is_public BOOL,
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
)

CREATE INDEX idx_homework_course_id ON homework(course_id);
CREATE INDEX idx_homework_start_date ON homework(start_date);
CREATE INDEX idx_homework_end_date ON homework(end_date);
CREATE INDEX idx_homework_course_active ON homework(course_id, end_date) WHERE end_date > NOW();
CREATE INDEX idx_homework_public ON homework(course_id) WHERE is_public = true;

DROP TABLE IF EXISTS homework;