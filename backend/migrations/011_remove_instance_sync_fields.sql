-- Remove duplicate sync fields from cyberark_instances table
-- These fields are now managed in instance_sync_configs table

-- Drop the sync configuration columns if they exist
ALTER TABLE cyberark_instances 
  DROP COLUMN IF EXISTS sync_enabled,
  DROP COLUMN IF EXISTS user_sync_interval,
  DROP COLUMN IF EXISTS group_sync_interval,
  DROP COLUMN IF EXISTS safe_sync_interval,
  DROP COLUMN IF EXISTS user_sync_page_size,
  DROP COLUMN IF EXISTS last_user_sync_at,
  DROP COLUMN IF EXISTS last_user_sync_status,
  DROP COLUMN IF EXISTS last_user_sync_error,
  DROP COLUMN IF EXISTS last_group_sync_at,
  DROP COLUMN IF EXISTS last_group_sync_status,
  DROP COLUMN IF EXISTS last_group_sync_error,
  DROP COLUMN IF EXISTS last_safe_sync_at,
  DROP COLUMN IF EXISTS last_safe_sync_status,
  DROP COLUMN IF EXISTS last_safe_sync_error;

-- Note: The sync configuration is now stored in instance_sync_configs table
-- with per-sync-type settings (users, safes, groups)