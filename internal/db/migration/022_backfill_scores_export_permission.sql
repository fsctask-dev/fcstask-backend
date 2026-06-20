-- +goose Up
INSERT INTO course_admin_permissions (role_id, permission)
SELECT DISTINCT cap.role_id, 'scores.export'
FROM course_admin_permissions cap
WHERE cap.permission = 'task.score.update'
  AND NOT EXISTS (
    SELECT 1 FROM course_admin_permissions x
    WHERE x.role_id = cap.role_id AND x.permission = 'scores.export'
  );

-- +goose Down
DELETE FROM course_admin_permissions WHERE permission = 'scores.export';
