-- Step 1: Rename old table for backup
ALTER TABLE zalo_tokens RENAME TO zalo_tokens_old;

-- Step 2: Create new multi-tenant table
CREATE TABLE zalo_tokens (
    id SERIAL PRIMARY KEY,
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    app_id VARCHAR(255) NOT NULL,
    secret_key TEXT NOT NULL,
    access_token TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id)
);

CREATE INDEX idx_zalo_tokens_tenant_id ON zalo_tokens(tenant_id);
CREATE INDEX idx_zalo_tokens_expires_at ON zalo_tokens(expires_at);

-- Step 3: Create or get system/default tenant
DO $$
DECLARE
    system_tenant_id UUID;
BEGIN
    -- Try to find existing "system" tenant
    SELECT id INTO system_tenant_id FROM tenants WHERE name = 'System' LIMIT 1;

    -- If no system tenant exists, create one
    IF system_tenant_id IS NULL THEN
        INSERT INTO tenants (id, name, public_url, admin_url, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            'System',
            'https://system.default',
            'https://admin.system.default',
            NOW(),
            NOW()
        )
        RETURNING id INTO system_tenant_id;
    END IF;

    -- Step 4: Migrate existing token to system tenant
    -- Note: app_id and secret_key need to be populated manually
    INSERT INTO zalo_tokens (
        tenant_id,
        app_id,
        secret_key,
        access_token,
        refresh_token,
        expires_at,
        created_at,
        updated_at
    )
    SELECT
        system_tenant_id,
        'MIGRATION_REQUIRED',
        'MIGRATION_REQUIRED',
        access_token,
        refresh_token,
        expires_at,
        NOW(),
        updated_at
    FROM zalo_tokens_old
    ORDER BY updated_at DESC
    LIMIT 1;
END $$;

-- Keep old table for manual verification, can drop later
-- DROP TABLE zalo_tokens_old;
