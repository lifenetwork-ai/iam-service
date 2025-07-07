-- Table: user_identities
CREATE TABLE IF NOT EXISTS user_identities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    global_user_id UUID NOT NULL REFERENCES global_users(id) ON DELETE CASCADE,
    type VARCHAR(20) NOT NULL, -- email, phone, google, wallet, etc.
    value VARCHAR(255) NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    primary BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (type, value)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_user_identities_global_user ON user_identities (global_user_id);
CREATE INDEX IF NOT EXISTS idx_user_identities_type_value ON user_identities (type, value);

-- Trigger for user_identities
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
        WHERE tgname = 'trigger_update_user_identities_updated_at'
          AND tgrelid = 'user_identities'::regclass
    ) THEN
        DROP TRIGGER trigger_update_user_identities_updated_at ON user_identities;
    END IF;

    CREATE TRIGGER trigger_update_user_identities_updated_at
    BEFORE UPDATE ON user_identities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
