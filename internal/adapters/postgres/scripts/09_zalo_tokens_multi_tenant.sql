-- *** START: Add Tenant Support to Zalo Tokens *** --

-- 1. ADD NEW COLUMNS (Idempotent)
-- Add tenant_id column if it doesn't exist
DO $$ 
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'zalo_tokens' 
        AND column_name = 'tenant_id'
        AND table_schema = 'public'
    ) THEN
        ALTER TABLE zalo_tokens 
        ADD COLUMN tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE;
        RAISE NOTICE 'Added tenant_id column to zalo_tokens';
    ELSE
        RAISE NOTICE 'Column tenant_id already exists';
    END IF;
END $$;

-- 2. MAKE TENANT_ID NOT NULL (After you manually set values)
-- Uncomment this block after you've set tenant_id for all existing records
/*
DO $$ 
BEGIN
    IF EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'zalo_tokens' 
        AND column_name = 'tenant_id'
        AND is_nullable = 'YES'
        AND table_schema = 'public'
    ) THEN
        ALTER TABLE zalo_tokens 
        ALTER COLUMN tenant_id SET NOT NULL;
        RAISE NOTICE 'Set tenant_id to NOT NULL';
    ELSE
        RAISE NOTICE 'tenant_id is already NOT NULL';
    END IF;
END $$;
*/


-- 4. CREATE INDEXES (Idempotent)
CREATE UNIQUE INDEX IF NOT EXISTS idx_zalo_tokens_tenant_id 
    ON zalo_tokens(tenant_id);
    
CREATE INDEX IF NOT EXISTS idx_zalo_tokens_expires_at 
    ON zalo_tokens(expires_at);
