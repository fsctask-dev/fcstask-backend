-- +goose Up
CREATE TABLE IF NOT EXISTS oauth_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    email_registration_id UUID,
    provider VARCHAR(32) NOT NULL,
    provider_uid VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    raw_profile JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_oauth_provider_uid ON oauth_identities(provider, provider_uid);
CREATE INDEX IF NOT EXISTS idx_oauth_identities_user_id ON oauth_identities(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_identities_email_registration_id ON oauth_identities(email_registration_id);

-- +goose Down
DROP TABLE IF EXISTS oauth_identities;
