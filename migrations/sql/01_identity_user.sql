SET TIMEZONE TO 'UTC';

-- Enable the uuid-ossp extension for generating UUIDs
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS identity_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    organization_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    full_name VARCHAR(255),
    user_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    password_hash TEXT,
    status BOOLEAN NOT NULL DEFAULT TRUE,
    lifeai_id VARCHAR(255),
    google_id VARCHAR(255),
    facebook_id VARCHAR(255),
    apple_id VARCHAR(255),
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create indexes if they do not exist
CREATE INDEX IF NOT EXISTS user_name_idx ON identity_users (name);
CREATE INDEX IF NOT EXISTS user_username_idx ON identity_users (user_name);
CREATE INDEX IF NOT EXISTS user_email_idx ON identity_users (email);
CREATE INDEX IF NOT EXISTS user_phone_idx ON identity_users (phone);

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
        WHERE tgname = 'update_identity_users_updated_at'
          AND tgrelid = 'identity_users'::regclass
    ) THEN
        DROP TRIGGER update_identity_users_updated_at ON identity_users;
    END IF;

    CREATE TRIGGER update_identity_users_updated_at
    BEFORE UPDATE ON identity_users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
END;
$$;
