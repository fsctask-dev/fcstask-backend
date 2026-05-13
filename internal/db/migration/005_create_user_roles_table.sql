-- +goose Up
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    course_id UUID NOT NULL,
    role_id UUID NOT NULL
);

CREATE UNIQUE INDEX idx_user_roles_unique ON user_roles(user_id, course_id, role_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_course ON user_roles(course_id);
CREATE INDEX idx_user_roles_user_course ON user_roles(user_id, course_id);

-- +goose Down
DROP TABLE IF EXISTS user_roles;