-- Create operations table for the processing pipeline
CREATE TABLE IF NOT EXISTS operations (
    id VARCHAR(30) PRIMARY KEY, -- ULID with prefix 'op_'
    type VARCHAR(50) NOT NULL,
    priority VARCHAR(10) NOT NULL CHECK (priority IN ('low', 'normal', 'medium', 'high')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'cancelled')),
    payload JSONB NOT NULL,
    result JSONB,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    max_retries INT DEFAULT 3,
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_by VARCHAR(30) REFERENCES users(id),
    cyberark_instance_id VARCHAR(30), -- Reference to future cyberark_instances table
    correlation_id VARCHAR(30), -- For grouping related operations
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient queue processing
CREATE INDEX idx_operations_queue ON operations(status, priority, scheduled_at) 
    WHERE status IN ('pending', 'processing');
CREATE INDEX idx_operations_correlation ON operations(correlation_id) 
    WHERE correlation_id IS NOT NULL;
CREATE INDEX idx_operations_created_by ON operations(created_by);
CREATE INDEX idx_operations_type ON operations(type);
CREATE INDEX idx_operations_created_at ON operations(created_at);

-- Create pipeline configuration table
CREATE TABLE IF NOT EXISTS pipeline_config (
    id VARCHAR(30) PRIMARY KEY, -- ULID with prefix 'cfg_'
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert default pipeline configuration
INSERT INTO pipeline_config (id, key, value, description) VALUES 
(
    'cfg_01JQF8Y6XYZABC123456789DEF',
    'processing_capacity',
    '{"total": 20, "priority_allocation": {"high": 0.4, "medium": 0.3, "normal": 0.2, "low": 0.1}}',
    'Total processing capacity and priority lane allocation'
),
(
    'cfg_01JQF8Y6XYZABC123456789GHI',
    'retry_policy',
    '{"max_attempts": 3, "backoff_base_seconds": 1, "backoff_multiplier": 2, "backoff_jitter": true}',
    'Retry policy for failed operations'
),
(
    'cfg_01JQF8Y6XYZABC123456789JKL',
    'operation_timeouts',
    '{"default": 300, "safe_provision": 600, "bulk_sync": 1800}',
    'Operation timeout settings in seconds'
)
ON CONFLICT (key) DO NOTHING;

-- Add updated_at triggers
CREATE TRIGGER update_operations_updated_at BEFORE UPDATE ON operations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_pipeline_config_updated_at BEFORE UPDATE ON pipeline_config
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();