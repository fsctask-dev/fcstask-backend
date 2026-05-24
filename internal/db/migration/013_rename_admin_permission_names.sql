-- +goose Up
UPDATE course_admin_permissions
SET permission = CASE permission
    WHEN 'admin.roles.assign' THEN 'course.roles.assign'
    WHEN 'admin.roles.revoke' THEN 'course.roles.revoke'
    WHEN 'admin.roles.list' THEN 'course.roles.list'
    WHEN 'admin.permissions.add' THEN 'course.permissions.add'
    WHEN 'admin.permissions.remove' THEN 'course.permissions.remove'
    WHEN 'admin.permissions.list' THEN 'course.permissions.list'
    WHEN 'admin.super_admins.create' THEN 'super_admin.create'
    ELSE permission
END
WHERE permission IN (
    'admin.roles.assign',
    'admin.roles.revoke',
    'admin.roles.list',
    'admin.permissions.add',
    'admin.permissions.remove',
    'admin.permissions.list',
    'admin.super_admins.create'
);

INSERT INTO course_admin_permissions (role_id, permission)
SELECT role_id, 'course.create'
FROM course_admin_permissions
WHERE permission = 'super_admin.create'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM course_admin_permissions
WHERE permission = 'course.create';

UPDATE course_admin_permissions
SET permission = CASE permission
    WHEN 'course.roles.assign' THEN 'admin.roles.assign'
    WHEN 'course.roles.revoke' THEN 'admin.roles.revoke'
    WHEN 'course.roles.list' THEN 'admin.roles.list'
    WHEN 'course.permissions.add' THEN 'admin.permissions.add'
    WHEN 'course.permissions.remove' THEN 'admin.permissions.remove'
    WHEN 'course.permissions.list' THEN 'admin.permissions.list'
    WHEN 'super_admin.create' THEN 'admin.super_admins.create'
    ELSE permission
END
WHERE permission IN (
    'course.roles.assign',
    'course.roles.revoke',
    'course.roles.list',
    'course.permissions.add',
    'course.permissions.remove',
    'course.permissions.list',
    'super_admin.create'
);
