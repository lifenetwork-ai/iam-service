-- *** START: Add Tenant Support to Zalo Tokens *** --

-- 1. ADD NEW COLUMNS (with IF NOT EXISTS checks)
ALTER TABLE zalo_tokens 
ADD COLUMN IF NOT EXISTS app_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS secret_key VARCHAR(255),
ADD COLUMN IF NOT EXISTS otp_template_id VARCHAR(255),
ADD COLUMN IF NOT EXISTS tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- 1.5 DELETE ANY ROWS WITH tenant_id IS NULL (before enforcing NOT NULL)
DELETE FROM zalo_tokens
WHERE tenant_id IS NULL;

-- 2. MAKE COLUMNS NOT NULL (Uncomment after confirming cleanup)
-- /*
-- ALTER TABLE zalo_tokens 
-- ALTER COLUMN app_id SET NOT NULL,
-- ALTER COLUMN secret_key SET NOT NULL,
-- ALTER COLUMN otp_template_id SET NOT NULL,
-- ALTER COLUMN tenant_id SET NOT NULL;
-- */

-- 3. CREATE INDEXES
CREATE UNIQUE INDEX IF NOT EXISTS idx_zalo_tokens_tenant_id 
    ON zalo_tokens(tenant_id);

CREATE INDEX IF NOT EXISTS idx_zalo_tokens_expires_at 
    ON zalo_tokens(expires_at);
