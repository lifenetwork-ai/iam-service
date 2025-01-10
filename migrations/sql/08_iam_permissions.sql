-- Create the iam_permissions table
CREATE TABLE IF NOT EXISTS iam_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Unique identifier for each permission
    policy_id UUID NOT NULL REFERENCES iam_policies(id) ON DELETE CASCADE, -- References the IAM policy
    resource VARCHAR(255) NOT NULL, -- The resource this permission applies to
    action VARCHAR(50) NOT NULL, -- The action this permission allows
    description TEXT, -- Optional description of the permission
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Creation timestamp
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Update timestamp
    UNIQUE(policy_id, resource, action) -- Ensure unique combinations of policy, resource, and action
);

-- Add composite index for resource, and action
CREATE INDEX IF NOT EXISTS idx_iam_permissions_resource_action ON iam_permissions (resource, action);

-- Add the updated_at trigger for the iam_permissions table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_iam_permissions_updated_at'
          AND tgrelid = 'iam_permissions'::regclass
    ) THEN
        DROP TRIGGER update_iam_permissions_updated_at ON iam_permissions;
    END IF;

    CREATE TRIGGER update_iam_permissions_updated_at
    BEFORE UPDATE ON iam_permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
