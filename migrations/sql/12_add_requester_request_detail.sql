-- Add validation status column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'data_access_request_requesters' 
        AND column_name = 'validation_status'
    ) THEN
        ALTER TABLE data_access_request_requesters
        ADD COLUMN validation_status VARCHAR(10) DEFAULT 'PENDING' 
        CHECK (validation_status IN ('VALID', 'INVALID', 'PENDING'));
    END IF;
END $$;

-- Add validation message column if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'data_access_request_requesters' 
        AND column_name = 'validation_message'
    ) THEN
        ALTER TABLE data_access_request_requesters
        ADD COLUMN validation_message TEXT DEFAULT NULL;
    END IF;
END $$;

-- Add index if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_indexes
        WHERE tablename = 'data_access_request_requesters'
        AND indexname = 'idx_request_requesters_validation_status'
    ) THEN
        CREATE INDEX idx_request_requesters_validation_status
        ON data_access_request_requesters (validation_status);
    END IF;
END $$;