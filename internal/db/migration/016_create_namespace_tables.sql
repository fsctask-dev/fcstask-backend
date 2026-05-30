-- +goose Up
CREATE TABLE IF NOT EXISTS namespaces (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          VARCHAR(255) NOT NULL,
    slug          VARCHAR(255) NOT NULL UNIQUE,
    description   TEXT,
    gitlab_group_id VARCHAR(255),
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at    TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_namespaces_slug ON namespaces(slug);
CREATE INDEX IF NOT EXISTS idx_namespaces_deleted_at ON namespaces(deleted_at);

CREATE TABLE IF NOT EXISTS namespace_users (
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role         VARCHAR(50) NOT NULL DEFAULT 'student',
    UNIQUE (namespace_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_namespace_users_namespace ON namespace_users(namespace_id);
CREATE INDEX IF NOT EXISTS idx_namespace_users_user     ON namespace_users(user_id);

CREATE TABLE IF NOT EXISTS namespace_courses (
    namespace_id UUID NOT NULL REFERENCES namespaces(id) ON DELETE CASCADE,
    course_id    UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    PRIMARY KEY (namespace_id, course_id)
);

CREATE INDEX IF NOT EXISTS idx_namespace_courses_namespace ON namespace_courses(namespace_id);
CREATE INDEX IF NOT EXISTS idx_namespace_courses_course    ON namespace_courses(course_id);

-- +goose Down
DROP TABLE IF EXISTS namespace_courses;
DROP TABLE IF EXISTS namespace_users;
DROP TABLE IF EXISTS namespaces;
