CREATE UNIQUE INDEX IF NOT EXISTS uniq_user_identities_global_user_id_type
ON user_identities (global_user_id, type);

DROP INDEX IF EXISTS idx_user_identities_type_value;

CREATE UNIQUE INDEX IF NOT EXISTS uniq_user_identities_type_value
ON user_identities (type, value);