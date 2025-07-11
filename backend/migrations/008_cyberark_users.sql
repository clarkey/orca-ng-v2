-- Create table for storing CyberArk users
CREATE TABLE IF NOT EXISTS cyberark_users (
    id VARCHAR(30) PRIMARY KEY,
    cyberark_instance_id VARCHAR(30) NOT NULL REFERENCES cyberark_instances(id) ON DELETE CASCADE,
    username VARCHAR(255) NOT NULL,
    user_id VARCHAR(255) NOT NULL,  -- CyberArk's internal user ID
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    email VARCHAR(255),
    user_type VARCHAR(50) NOT NULL,
    location VARCHAR(255),
    suspended BOOLEAN NOT NULL DEFAULT false,
    enable_user BOOLEAN NOT NULL DEFAULT true,
    change_pass_on_next_logon BOOLEAN NOT NULL DEFAULT false,
    expiry_date TIMESTAMPTZ,
    last_successful_login_at TIMESTAMPTZ,
    
    -- Sync metadata
    last_synced_at TIMESTAMPTZ NOT NULL,
    cyberark_last_modified TIMESTAMPTZ,
    is_deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes and constraints
    CONSTRAINT uk_cyberark_users_instance_username UNIQUE (cyberark_instance_id, username)
);

CREATE INDEX idx_cyberark_users_instance_id ON cyberark_users(cyberark_instance_id);
CREATE INDEX idx_cyberark_users_is_deleted ON cyberark_users(is_deleted);
CREATE INDEX idx_cyberark_users_user_type ON cyberark_users(user_type);

-- Add pagination and sync status fields to instances
ALTER TABLE cyberark_instances 
ADD COLUMN IF NOT EXISTS user_sync_page_size INTEGER DEFAULT 100,
ADD COLUMN IF NOT EXISTS last_user_sync_at TIMESTAMPTZ,
ADD COLUMN IF NOT EXISTS last_user_sync_status VARCHAR(20),
ADD COLUMN IF NOT EXISTS last_user_sync_error TEXT;