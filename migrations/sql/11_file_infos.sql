-- Create the file_infos table
CREATE TABLE IF NOT EXISTS file_infos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for the file
    name VARCHAR(255) NOT NULL,                     -- File name
    share_count INT NOT NULL CHECK (share_count >= 0), -- Number of shares, must be >= 0
    owner_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Foreign key to the accounts table
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP  -- Last updated timestamp
);

-- Add an index for the owner_id column
CREATE INDEX IF NOT EXISTS idx_file_infos_owner_id
ON file_infos (owner_id);

-- Add an index for the created_at column
CREATE INDEX IF NOT EXISTS idx_file_infos_created_at
ON file_infos (created_at);

-- Add the updated_at trigger for the file_infos table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_file_infos_updated_at'
          AND tgrelid = 'file_infos'::regclass
    ) THEN
        DROP TRIGGER update_file_infos_updated_at ON file_infos;
    END IF;

    CREATE TRIGGER update_file_infos_updated_at
    BEFORE UPDATE ON file_infos
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
