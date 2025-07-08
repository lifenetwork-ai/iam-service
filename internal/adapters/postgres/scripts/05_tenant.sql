-- Description: Create tenant table
-- Version: 1.0.0

BEGIN;

CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    public_url VARCHAR(255) NOT NULL,
    admin_url VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at);

COMMIT;

-- Pre-create two tenant records if they don't exist
-- This should be run after the tenants table has been created

INSERT INTO tenants (
    id,
    name,
    public_url,
    admin_url
) 
SELECT 
    gen_random_uuid(),
    'life_ai',
    'https://auth.develop.lifenetwork.ai',
    'https://human-network-kratos-admin-develop-802449703053.asia-southeast1.run.app'
WHERE NOT EXISTS (
    SELECT 1 FROM tenants WHERE name = 'life_ai'
);

INSERT INTO tenants (
    id,
    name,
    public_url,
    admin_url
) 
SELECT 
    gen_random_uuid(),
    'genetica',
    'https://auth.develop.lifenetwork.ai',
    'https://human-network-kratos-admin-develop-802449703053.asia-southeast1.run.app'
WHERE NOT EXISTS (
    SELECT 1 FROM tenants WHERE name = 'genetica'
);

-- Verify the records exist:
SELECT id, name, public_url, admin_url, created_at FROM tenants WHERE name IN ('life_ai', 'genetica');