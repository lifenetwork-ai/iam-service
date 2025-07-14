-- Table: global_users
CREATE TABLE IF NOT EXISTS global_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Trigger for global_users
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger
        WHERE tgname = 'trigger_update_global_users_updated_at'
          AND tgrelid = 'global_users'::regclass
    ) THEN
        DROP TRIGGER trigger_update_global_users_updated_at ON global_users;
    END IF;

    CREATE TRIGGER trigger_update_global_users_updated_at
    BEFORE UPDATE ON global_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
