-- +goose Up
INSERT INTO course_admin_permissions (role_id, permission)
SELECT DISTINCT ur.role_id, 'course.invite.regenerate'
FROM user_roles ur
INNER JOIN course_admin_permissions cap ON cap.role_id = ur.role_id
WHERE cap.permission = 'homework.create'
  AND NOT EXISTS (
    SELECT 1 FROM course_admin_permissions cap2
    WHERE cap2.role_id = ur.role_id
      AND cap2.permission = 'course.invite.regenerate'
  );

-- +goose Down
DELETE FROM course_admin_permissions
WHERE permission = 'course.invite.regenerate';