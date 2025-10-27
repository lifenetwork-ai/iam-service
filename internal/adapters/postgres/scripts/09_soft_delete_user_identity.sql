-- Soft delete support and partial unique indexes for user_identities

-- 1) Add deleted_at column if not exists
ALTER TABLE IF EXISTS user_identities
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;

-- 2) Drop existing unique indexes that don't account for deleted rows
DROP INDEX IF EXISTS uniq_tenant_type_value;
DROP INDEX IF EXISTS uniq_tenant_global_user_type;

-- 3) Recreate unique indexes as partial uniques for non-deleted rows
CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_type_value_active
ON user_identities (tenant_id, type, value)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_global_user_type_active
ON user_identities (tenant_id, global_user_id, type)
WHERE deleted_at IS NULL;



