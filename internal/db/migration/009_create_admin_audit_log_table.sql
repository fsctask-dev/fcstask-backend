CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE admin_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    admin_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,

    -- Над чем совершено действие
    target_type VARCHAR(50),        -- course, task, user, subscription, deadline...
    target_id UUID,

    -- Тип действия
    action VARCHAR(50) NOT NULL,    -- create, update, delete, add_user, ...

    old_data JSONB,                 -- состояние до
    new_data JSONB,                 -- состояние после

    description TEXT,
    ip_address VARCHAR(45),

    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_admin_audit_admin ON admin_audit_log(admin_id);
CREATE INDEX idx_admin_audit_action ON admin_audit_log(action);
CREATE INDEX idx_admin_audit_target ON admin_audit_log(target_type, target_id);
CREATE INDEX idx_admin_audit_created ON admin_audit_log(created_at);

DROP TABLE IF EXISTS admin_audit_log;