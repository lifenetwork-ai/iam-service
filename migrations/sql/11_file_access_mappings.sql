-- Table to map user access to file IDs
CREATE TABLE IF NOT EXISTS file_access_mappings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique ID for the mapping
    file_id UUID NOT NULL, -- File ID
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- User account ID
    access_granted BOOLEAN NOT NULL DEFAULT TRUE, -- Whether access is granted
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Timestamp when the mapping was created
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP -- Timestamp when the mapping was last updated
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_file_access_file_id ON file_access_mappings (file_id);
CREATE INDEX IF NOT EXISTS idx_file_access_account_id ON file_access_mappings (account_id);

-- Add the updated_at trigger for the file_access_mappings table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_file_access_mappings_updated_at'
          AND tgrelid = 'file_access_mappings'::regclass
    ) THEN
        DROP TRIGGER update_file_access_mappings_updated_at ON file_access_mappings;
    END IF;

    CREATE TRIGGER update_file_access_mappings_updated_at
    BEFORE UPDATE ON file_access_mappings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
