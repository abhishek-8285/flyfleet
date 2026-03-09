-- OIDC Providers table
CREATE TABLE IF NOT EXISTS oidc_providers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT UNIQUE NOT NULL, -- e.g., 'google', 'azure'
    issuer TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret_hash TEXT NOT NULL,
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User OIDC links
CREATE TABLE IF NOT EXISTS user_oidc_links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    provider_id UUID REFERENCES oidc_providers(id) ON DELETE CASCADE,
    provider_user_id TEXT NOT NULL, -- 'sub' from OIDC
    email TEXT,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(provider_id, provider_user_id)
);

-- Default providers (optional, can be added via API)
INSERT INTO oidc_providers (name, issuer, client_id, client_secret_hash)
VALUES ('google', 'https://accounts.google.com', 'google-client-id', 'google-mock-secret')
ON CONFLICT (name) DO NOTHING;
