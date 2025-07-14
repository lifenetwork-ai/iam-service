DO $$
BEGIN
  -- 1. Add tenant_id column if not exists
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'user_identifier_mapping' AND column_name = 'tenant_id'
  ) THEN
    ALTER TABLE user_identifier_mapping 
    ADD COLUMN tenant_id UUID;
  END IF;

  -- 2. Update tenant_id from tenant column if both exist
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'user_identifier_mapping' AND column_name = 'tenant'
  ) THEN
    UPDATE user_identifier_mapping m
    SET tenant_id = t.id
    FROM tenants t
    WHERE t.name = m.tenant AND m.tenant_id IS NULL;
  END IF;

  -- 3. Drop old unique constraint if exists
  IF EXISTS (
    SELECT 1 FROM information_schema.table_constraints 
    WHERE constraint_name = 'user_identifier_mapping_tenant_tenant_user_id_key'
  ) THEN
    ALTER TABLE user_identifier_mapping
    DROP CONSTRAINT user_identifier_mapping_tenant_tenant_user_id_key;
  END IF;

  -- 4. Set tenant_id NOT NULL
  BEGIN
    ALTER TABLE user_identifier_mapping
    ALTER COLUMN tenant_id SET NOT NULL;
  EXCEPTION WHEN others THEN
    NULL;
  END;

  -- 5. Add FK constraint (skip if exists)
  BEGIN
    ALTER TABLE user_identifier_mapping
    ADD CONSTRAINT fk_user_identifier_mapping_tenant
    FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE CASCADE;
  EXCEPTION WHEN others THEN
    NULL;
  END;

  -- 6. Add new unique constraint
  BEGIN
    ALTER TABLE user_identifier_mapping
    ADD CONSTRAINT user_identifier_mapping_tenant_id_tenant_user_id_key 
    UNIQUE (tenant_id, tenant_user_id);
  EXCEPTION WHEN others THEN
    NULL;
  END;

  -- 7. Drop tenant column
  IF EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'user_identifier_mapping' AND column_name = 'tenant'
  ) THEN
    ALTER TABLE user_identifier_mapping
    DROP COLUMN tenant;
  END IF;

END $$;
