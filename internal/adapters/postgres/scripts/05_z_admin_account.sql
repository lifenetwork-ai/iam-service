-- Description: Create admin accounts table with username field
-- Version: 1.1.0

BEGIN;

-- Create the table if it doesn't exist
CREATE TABLE IF NOT EXISTS admin_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'ADMIN',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Handle column rename idempotently
DO $$
BEGIN
    -- Check if email column exists and username doesn't
    IF EXISTS (SELECT 1 FROM information_schema.columns 
               WHERE table_name = 'admin_accounts' AND column_name = 'email')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns 
                       WHERE table_name = 'admin_accounts' AND column_name = 'username') THEN
        
        -- Rename email column to username
        ALTER TABLE admin_accounts RENAME COLUMN email TO username;
        
    END IF;
END $$;

-- Create indexes idempotently
CREATE UNIQUE INDEX IF NOT EXISTS idx_admin_accounts_username ON admin_accounts(username);
CREATE INDEX IF NOT EXISTS idx_admin_accounts_status ON admin_accounts(status);

COMMIT;