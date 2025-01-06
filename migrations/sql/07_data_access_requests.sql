-- Create the data_access_requests table
CREATE TABLE IF NOT EXISTS data_access_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Use UUID as the primary key
    request_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Account whose data is being requested
    requester_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Account requesting access to data
    requester_role VARCHAR(20) NOT NULL, -- Role of the requester (e.g., CUSTOMER, PARTNER)
    reason_for_request TEXT NOT NULL, -- Reason why the requester is asking for access
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- Status: PENDING, APPROVED, REJECTED
    reason_for_rejection TEXT, -- Optional reason for rejecting the request
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add composite index for request_account_id and status
CREATE INDEX IF NOT EXISTS idx_request_account_status
ON data_access_requests (request_account_id, status);

-- Add index for requester_account_id
CREATE INDEX IF NOT EXISTS idx_requester_account
ON data_access_requests (requester_account_id);

-- Add the updated_at trigger for the data_access_requests table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_data_access_requests_updated_at'
          AND tgrelid = 'data_access_requests'::regclass
    ) THEN
        DROP TRIGGER update_data_access_requests_updated_at ON data_access_requests;
    END IF;

    CREATE TRIGGER update_data_access_requests_updated_at
    BEFORE UPDATE ON data_access_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
