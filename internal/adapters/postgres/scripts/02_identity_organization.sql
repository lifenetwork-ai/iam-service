CREATE TABLE IF NOT EXISTS identity_organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    parent_id UUID REFERENCES identity_organizations(id),
    parent_path TEXT,
    self_authenticate BOOLEAN DEFAULT FALSE,
    authenticate_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes if they do not exist
CREATE INDEX IF NOT EXISTS organization_name_idx ON identity_organizations (name);
CREATE INDEX IF NOT EXISTS organization_code_idx ON identity_organizations (code);
CREATE INDEX IF NOT EXISTS organization_parent_path_idx ON identity_organizations (parent_path);

-- Drop existing trigger if it exists, then create it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_identity_organizations_updated_at'
          AND tgrelid = 'identity_organizations'::regclass
    ) THEN
        DROP TRIGGER update_identity_organizations_updated_at ON identity_organizations;
    END IF;

    CREATE TRIGGER update_identity_organizations_updated_at
    BEFORE UPDATE ON identity_organizations
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
