-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS course_admin_permissions (
    role_id UUID NOT NULL UNIQUE,
    permission VARCHAR(255) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_role_permission
    ON course_admin_permissions(role_id, permission);

-- +goose Down
DROP TABLE IF EXISTS course_admin_permissions;
