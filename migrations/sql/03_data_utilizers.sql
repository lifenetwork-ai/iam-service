CREATE TABLE IF NOT EXISTS data_utilizers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Use UUID as primary key
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Match UUID from accounts table
    organization_name VARCHAR(255) NOT NULL,
    industry VARCHAR(100),
    contact_name VARCHAR(100),
    phone_number VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the data_utilizers table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_data_utilizers_updated_at'
          AND tgrelid = 'data_utilizers'::regclass
    ) THEN
        DROP TRIGGER update_data_utilizers_updated_at ON data_utilizers;
    END IF;

    CREATE TRIGGER update_data_utilizers_updated_at
    BEFORE UPDATE ON data_utilizers
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
