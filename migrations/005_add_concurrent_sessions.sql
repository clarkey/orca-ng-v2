-- Add concurrent_sessions column to cyberark_instances table
ALTER TABLE cyberark_instances 
ADD COLUMN concurrent_sessions BOOLEAN NOT NULL DEFAULT true;

-- Add comment for the new column
COMMENT ON COLUMN cyberark_instances.concurrent_sessions IS 'Whether to allow concurrent sessions when connecting to this CyberArk instance';