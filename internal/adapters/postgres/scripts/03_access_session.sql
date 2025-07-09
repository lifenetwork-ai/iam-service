CREATE TABLE IF NOT EXISTS access_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    user_id UUID NOT NULL,
    access_token TEXT,
    refresh_token TEXT,
    device_id TEXT,
    firebase_token TEXT,
    access_expired_at TIMESTAMP,
    refresh_expired_at TIMESTAMP,
    last_revoked_at TIMESTAMP,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Drop existing trigger if it exists, then create it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_access_sessions_updated_at'
          AND tgrelid = 'access_sessions'::regclass
    ) THEN
        DROP TRIGGER update_access_sessions_updated_at ON access_sessions;
    END IF;

    CREATE TRIGGER update_access_sessions_updated_at
    BEFORE UPDATE ON access_sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
