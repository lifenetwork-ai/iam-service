ALTER TABLE identity_users
ADD COLUMN IF NOT EXISTS self_authenticate_id VARCHAR(255);