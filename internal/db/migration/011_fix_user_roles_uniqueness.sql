-- +goose Up
DROP INDEX IF EXISTS idx_user_roles_unique;
DROP INDEX IF EXISTS idx_user_roles_user_course;
DROP INDEX IF EXISTS idx_user_course_role;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_course_role
    ON user_roles(user_id, course_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_role_id
    ON user_roles(role_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_course_role;

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique
    ON user_roles(user_id, course_id, role_id);

CREATE INDEX IF NOT EXISTS idx_user_roles_user_course
    ON user_roles(user_id, course_id);
