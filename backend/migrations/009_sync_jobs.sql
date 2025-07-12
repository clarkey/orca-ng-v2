-- Create sync_jobs table
CREATE TABLE IF NOT EXISTS sync_jobs (
    id VARCHAR(30) PRIMARY KEY,
    cyberark_instance_id VARCHAR(30) NOT NULL,
    sync_type VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    triggered_by VARCHAR(50) NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    records_synced INTEGER DEFAULT 0,
    records_created INTEGER DEFAULT 0,
    records_updated INTEGER DEFAULT 0,
    records_deleted INTEGER DEFAULT 0,
    records_failed INTEGER DEFAULT 0,
    error_message TEXT,
    error_details TEXT,
    duration_seconds DOUBLE PRECISION,
    created_by VARCHAR(30),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT fk_sync_jobs_instance
        FOREIGN KEY (cyberark_instance_id)
        REFERENCES cyberark_instances(id)
        ON DELETE CASCADE,
        
    CONSTRAINT fk_sync_jobs_user
        FOREIGN KEY (created_by)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- Create indexes
CREATE INDEX idx_sync_jobs_instance_id ON sync_jobs(cyberark_instance_id);
CREATE INDEX idx_sync_jobs_sync_type ON sync_jobs(sync_type);
CREATE INDEX idx_sync_jobs_status ON sync_jobs(status);
CREATE INDEX idx_sync_jobs_created_at ON sync_jobs(created_at);
CREATE INDEX idx_sync_jobs_deleted_at ON sync_jobs(deleted_at);

-- Create instance_sync_configs table
CREATE TABLE IF NOT EXISTS instance_sync_configs (
    id VARCHAR(30) PRIMARY KEY,
    cyberark_instance_id VARCHAR(30) NOT NULL,
    sync_type VARCHAR(50) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    interval_minutes INTEGER NOT NULL DEFAULT 60,
    page_size INTEGER DEFAULT 100,
    retry_attempts INTEGER DEFAULT 3,
    timeout_minutes INTEGER DEFAULT 30,
    last_run_at TIMESTAMPTZ,
    last_run_status VARCHAR(20),
    last_run_message TEXT,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT fk_instance_sync_configs_instance
        FOREIGN KEY (cyberark_instance_id)
        REFERENCES cyberark_instances(id)
        ON DELETE CASCADE,
        
    CONSTRAINT uq_instance_sync_type
        UNIQUE (cyberark_instance_id, sync_type, deleted_at)
);

-- Create indexes
CREATE INDEX idx_instance_sync_configs_instance_id ON instance_sync_configs(cyberark_instance_id);
CREATE INDEX idx_instance_sync_configs_next_run_at ON instance_sync_configs(next_run_at);
CREATE INDEX idx_instance_sync_configs_deleted_at ON instance_sync_configs(deleted_at);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_sync_jobs_updated_at BEFORE UPDATE
    ON sync_jobs FOR EACH ROW EXECUTE FUNCTION
    update_updated_at_column();
    
CREATE TRIGGER update_instance_sync_configs_updated_at BEFORE UPDATE
    ON instance_sync_configs FOR EACH ROW EXECUTE FUNCTION
    update_updated_at_column();