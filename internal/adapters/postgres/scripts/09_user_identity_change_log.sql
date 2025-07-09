-- Table: user_identity_change_logs
CREATE TABLE IF NOT EXISTS user_identity_change_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    global_user_id UUID NOT NULL REFERENCES global_users(id) ON DELETE CASCADE,
    tenant VARCHAR(25) NOT NULL,
    identity_type VARCHAR(20) NOT NULL,     
    old_value VARCHAR(255),                 
    new_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_identity_change_user ON user_identity_change_logs (global_user_id);
CREATE INDEX IF NOT EXISTS idx_identity_change_tenant ON user_identity_change_logs (tenant);
