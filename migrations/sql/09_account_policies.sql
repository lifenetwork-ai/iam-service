-- Create the account_policies table
CREATE TABLE IF NOT EXISTS account_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for each account-policy assignment
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- References the accounts table
    policy_id UUID NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE, -- References the IAM policies table
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Update timestamp
    UNIQUE(account_id, policy_id) -- Ensure no duplicate assignments of the same policy to an account
);

-- Add the updated_at trigger for the account_policies table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_account_policies_updated_at'
          AND tgrelid = 'account_policies'::regclass
    ) THEN
        DROP TRIGGER update_account_policies_updated_at ON account_policies;
    END IF;

    CREATE TRIGGER update_account_policies_updated_at
    BEFORE UPDATE ON account_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
