DO $$
BEGIN
  -- 1. Add is_primary column if it does not exist
  IF NOT EXISTS (
    SELECT 1 FROM information_schema.columns 
    WHERE table_name = 'user_identities' AND column_name = 'is_primary'
  ) THEN
    ALTER TABLE user_identities 
    ADD COLUMN is_primary BOOLEAN NOT NULL DEFAULT TRUE;
  END IF;
END $$;
