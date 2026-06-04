-- +goose Up
ALTER TABLE deadlines ADD COLUMN IF NOT EXISTS homework_id UUID;
ALTER TABLE deadlines ALTER COLUMN homework_id SET NOT NULL;
ALTER TABLE deadlines DROP CONSTRAINT IF EXISTS deadlines_homework_id_fkey;
ALTER TABLE deadlines ADD CONSTRAINT deadlines_homework_id_fkey 
    FOREIGN KEY (homework_id) REFERENCES homework(hw_id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_deadlines_homework_id ON deadlines(homework_id);

-- +goose Down
ALTER TABLE deadlines ALTER COLUMN homework_id DROP NOT NULL;
ALTER TABLE deadlines DROP CONSTRAINT IF EXISTS deadlines_homework_id_fkey;
ALTER TABLE deadlines ADD CONSTRAINT deadlines_homework_id_fkey 
    FOREIGN KEY (homework_id) REFERENCES homework(hw_id) ON DELETE SET NULL;