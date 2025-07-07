-- Table: user_identifier_mapping
CREATE TABLE IF NOT EXISTS user_identifier_mapping (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    global_user_id UUID NOT NULL REFERENCES global_users(id) ON DELETE CASCADE,
    tenant VARCHAR(25) NOT NULL,
    tenant_user_id UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (tenant, tenant_user_id)
);

-- Index
CREATE INDEX IF NOT EXISTS idx_user_identifier_mapping_global_user ON user_identifier_mapping (global_user_id);

-- Trigger for user_identifier_mapping
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
        WHERE tgname = 'trigger_update_user_identifier_mapping_updated_at'
          AND tgrelid = 'user_identifier_mapping'::regclass
    ) THEN
        DROP TRIGGER trigger_update_user_identifier_mapping_updated_at ON user_identifier_mapping;
    END IF;

    CREATE TRIGGER trigger_update_user_identifier_mapping_updated_at
    BEFORE UPDATE ON user_identifier_mapping
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
