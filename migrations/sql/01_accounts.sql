SET TIMEZONE TO 'UTC';

-- Enable the uuid-ossp extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Check if the enum type exists, and create it if it doesn't
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'account_role') THEN
        CREATE TYPE account_role AS ENUM ('USER', 'PARTNER', 'CUSTOMER', 'VALIDATOR');
    END IF;
END;
$$;

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Use UUID as primary key
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash TEXT, -- NULL if OAuth or API key is used
    api_key VARCHAR(255) UNIQUE, -- Only for API-based roles (partner, validator)
    role account_role NOT NULL, -- Enum: USER, PARTNER, CUSTOMER, VALIDATOR
    oauth_provider VARCHAR(50), -- Google, Facebook, etc. (nullable)
    oauth_id VARCHAR(255), -- ID from OAuth provider (nullable)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes if they do not exist
CREATE INDEX IF NOT EXISTS accounts_email_idx ON accounts (email);
CREATE INDEX IF NOT EXISTS accounts_username_idx ON accounts (username);
CREATE INDEX IF NOT EXISTS accounts_api_key_idx ON accounts (api_key);

-- Create a trigger function to update 'updated_at' column on update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists, then create it
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM pg_trigger
        WHERE tgname = 'update_accounts_updated_at'
          AND tgrelid = 'accounts'::regclass
    ) THEN
        DROP TRIGGER update_accounts_updated_at ON accounts;
    END IF;

    CREATE TRIGGER update_accounts_updated_at
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
