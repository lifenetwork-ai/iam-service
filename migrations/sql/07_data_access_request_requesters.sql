-- Create a join table for requesters
CREATE TABLE IF NOT EXISTS data_access_request_requesters (
    id SERIAL PRIMARY KEY, -- Auto-incrementing primary key
    request_id UUID NOT NULL REFERENCES data_access_requests(id) ON DELETE CASCADE, -- Data access request ID
    requester_account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Account requesting access to data
    UNIQUE (request_id, requester_account_id) -- Ensure unique pairs of request and requester
);

-- Add an index for the request_id column in the join table
CREATE INDEX IF NOT EXISTS idx_request_requesters_request_id
ON data_access_request_requesters (request_id);

-- Add an index for the requester_account_id column in the join table
CREATE INDEX IF NOT EXISTS idx_request_requesters_account_id
ON data_access_request_requesters (requester_account_id);