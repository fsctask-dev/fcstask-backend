-- +goose Up
INSERT INTO course_admin_permissions (role_id, permission)
SELECT role_id, 'is_super_admin'
FROM course_admin_permissions
WHERE permission = 'super_admin.create'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM course_admin_permissions
WHERE permission = 'is_super_admin';
