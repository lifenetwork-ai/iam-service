CREATE TABLE IF NOT EXISTS validator_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Use UUID as primary key
    account_id UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE, -- Match UUID from accounts table
    validation_organization VARCHAR(255) NOT NULL,
    contact_person VARCHAR(100),
    phone_number VARCHAR(20),
    is_active BOOLEAN NOT NULL DEFAULT TRUE, -- Field to indicate if the validator is active
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Add the updated_at trigger for the validator_details table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_validator_details_updated_at'
          AND tgrelid = 'validator_details'::regclass
    ) THEN
        DROP TRIGGER update_validator_details_updated_at ON validator_details;
    END IF;

    CREATE TRIGGER update_validator_details_updated_at
    BEFORE UPDATE ON validator_details
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
