-- +goose Up
CREATE TABLE IF NOT EXISTS namespace_users (
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'student',
    UNIQUE(namespace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_namespace_users_namespace ON namespace_users(namespace_id);
CREATE INDEX IF NOT EXISTS idx_namespace_users_user ON namespace_users(user_id);

-- +goose Down
DROP TABLE IF EXISTS namespace_users;
