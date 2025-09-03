ALTER TABLE user_identities
ADD COLUMN IF NOT EXISTS tenant_id UUID;

ALTER TABLE user_identities
ALTER COLUMN tenant_id SET NOT NULL;

DROP INDEX IF EXISTS idx_user_identities_type_value;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_global_user_type
ON user_identities (tenant_id, global_user_id, type);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_tenant_type_value
ON user_identities (tenant_id, type, value);
