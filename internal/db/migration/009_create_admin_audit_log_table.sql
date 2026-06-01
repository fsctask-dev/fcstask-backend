-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID REFERENCES users(id) ON DELETE SET NULL,
    target_type VARCHAR(50),
    target_id UUID,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    description TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_admin_audit_admin
    ON admin_audit_log(admin_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_action
    ON admin_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_target
    ON admin_audit_log(target_type, target_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_created
    ON admin_audit_log(created_at);

-- +goose Down
DROP TABLE IF EXISTS admin_audit_log;
