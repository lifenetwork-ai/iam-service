CREATE TABLE IF NOT EXISTS partner_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Use UUID as primary key
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Match UUID from accounts table
    company_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(100),
    phone_number VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the partner_details table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_partner_details_updated_at'
          AND tgrelid = 'partner_details'::regclass
    ) THEN
        DROP TRIGGER update_partner_details_updated_at ON partner_details;
    END IF;

    CREATE TRIGGER update_partner_details_updated_at
    BEFORE UPDATE ON partner_details
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
