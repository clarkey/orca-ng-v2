-- Create CyberArk instances table
CREATE TABLE IF NOT EXISTS cyberark_instances (
    id VARCHAR(30) PRIMARY KEY, -- ULID with prefix 'cai_'
    name VARCHAR(255) UNIQUE NOT NULL,
    base_url TEXT NOT NULL,
    username VARCHAR(255) NOT NULL,
    password_encrypted TEXT NOT NULL, -- Will be encrypted before storage
    is_active BOOLEAN DEFAULT true,
    last_test_at TIMESTAMP WITH TIME ZONE,
    last_test_success BOOLEAN,
    last_test_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(30) REFERENCES users(id),
    updated_by VARCHAR(30) REFERENCES users(id)
);

-- Create indexes
CREATE INDEX idx_cyberark_instances_name ON cyberark_instances(name);
CREATE INDEX idx_cyberark_instances_is_active ON cyberark_instances(is_active);

-- Add updated_at trigger
CREATE TRIGGER update_cyberark_instances_updated_at BEFORE UPDATE ON cyberark_instances
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add a comment to the table
COMMENT ON TABLE cyberark_instances IS 'Stores CyberArk PVWA instance configurations';
COMMENT ON COLUMN cyberark_instances.password_encrypted IS 'Password encrypted using AES-256-GCM';