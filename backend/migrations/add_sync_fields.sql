-- Add sync configuration fields to cyberark_instances table
ALTER TABLE cyberark_instances 
ADD COLUMN IF NOT EXISTS sync_enabled BOOLEAN DEFAULT true NOT NULL,
ADD COLUMN IF NOT EXISTS user_sync_interval INTEGER DEFAULT 30,
ADD COLUMN IF NOT EXISTS group_sync_interval INTEGER DEFAULT 60,
ADD COLUMN IF NOT EXISTS safe_sync_interval INTEGER DEFAULT 120;

-- Update existing instances with default intervals if null
UPDATE cyberark_instances 
SET user_sync_interval = 30 
WHERE user_sync_interval IS NULL;

UPDATE cyberark_instances 
SET group_sync_interval = 60 
WHERE group_sync_interval IS NULL;

UPDATE cyberark_instances 
SET safe_sync_interval = 120 
WHERE safe_sync_interval IS NULL;