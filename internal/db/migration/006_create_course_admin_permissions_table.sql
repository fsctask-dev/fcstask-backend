-- +goose Up
CREATE TABLE IF NOT EXISTS course_admin_permissions (
    role_id UUID NOT NULL,
    permission VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX idx_perm_role_permission ON course_admin_permissions(role_id, permission);
CREATE INDEX idx_perm_role_id ON course_admin_permissions(role_id);

-- +goose Down
DROP TABLE IF EXISTS course_admin_permissions;