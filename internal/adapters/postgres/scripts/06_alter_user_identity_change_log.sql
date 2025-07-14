DO $$
BEGIN
  -- 1. Add tenant_id column if not exists
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name='user_identity_change_logs' AND column_name='tenant_id'
  ) THEN
    ALTER TABLE user_identity_change_logs 
    ADD COLUMN tenant_id UUID;
  END IF;

  -- 2. Only update tenant_id if column 'tenant' exists and tenant_id has NULLs
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name='user_identity_change_logs' AND column_name='tenant'
  ) AND EXISTS (
    SELECT 1 FROM user_identity_change_logs WHERE tenant_id IS NULL
  ) THEN
    UPDATE user_identity_change_logs l
    SET tenant_id = t.id
    FROM tenants t
    WHERE t.name = l.tenant;
  END IF;

  -- 3. Make tenant_id NOT NULL (ignore error if already NOT NULL)
  BEGIN
    ALTER TABLE user_identity_change_logs
    ALTER COLUMN tenant_id SET NOT NULL;
  EXCEPTION WHEN others THEN
    NULL;
  END;

  -- 4. Add FK constraint if not already exists
  BEGIN
    ALTER TABLE user_identity_change_logs
    ADD CONSTRAINT fk_user_identity_change_logs_tenant 
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
  EXCEPTION WHEN others THEN
    NULL;
  END;

  -- 5. Drop old index on tenant (if it exists)
  IF EXISTS (
    SELECT 1 FROM pg_indexes WHERE tablename = 'user_identity_change_logs' AND indexname = 'idx_identity_change_tenant'
  ) THEN
    DROP INDEX IF EXISTS idx_identity_change_tenant;
  END IF;

  -- 6. Create new index on tenant_id (if not exists)
  CREATE INDEX IF NOT EXISTS idx_identity_change_tenant ON user_identity_change_logs (tenant_id);

  -- 7. Drop 'tenant' column if it exists
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name='user_identity_change_logs' AND column_name='tenant'
  ) THEN
    ALTER TABLE user_identity_change_logs
    DROP COLUMN tenant;
  END IF;

END $$;
