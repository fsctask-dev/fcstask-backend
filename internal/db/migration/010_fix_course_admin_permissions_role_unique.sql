-- +goose Up
ALTER TABLE course_admin_permissions
    DROP CONSTRAINT IF EXISTS course_admin_permissions_role_id_key;

DROP INDEX IF EXISTS idx_perm_role_permission;

CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_role_permission
    ON course_admin_permissions(role_id, permission);

-- +goose Down
DROP INDEX IF EXISTS idx_perm_role_permission;

ALTER TABLE course_admin_permissions
    ADD CONSTRAINT course_admin_permissions_role_id_key UNIQUE (role_id);

CREATE UNIQUE INDEX IF NOT EXISTS idx_perm_role_permission
    ON course_admin_permissions(role_id, permission);
