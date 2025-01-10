-- Create the iam_policies table
CREATE TABLE IF NOT EXISTS iam_policies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the iam_policies table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_iam_policies_updated_at'
          AND tgrelid = 'iam_policies'::regclass
    ) THEN
        DROP TRIGGER update_iam_policies_updated_at ON iam_policies;
    END IF;

    CREATE TRIGGER update_iam_policies_updated_at
    BEFORE UPDATE ON iam_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
