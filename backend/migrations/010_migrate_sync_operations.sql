-- Migration to transition from sync operations to sync jobs architecture
-- This migration is safe to run multiple times

-- Create default sync configurations for existing instances
INSERT INTO instance_sync_configs (id, cyberark_instance_id, sync_type, enabled, interval_minutes, page_size, retry_attempts, timeout_minutes, created_at, updated_at)
SELECT 
    'syncfg_' || gen_random_uuid()::text AS id,
    ci.id AS cyberark_instance_id,
    sync_type.type AS sync_type,
    COALESCE(ci.sync_enabled, true) AS enabled,
    60 AS interval_minutes,  -- Default 60 minutes
    COALESCE(ci.user_sync_page_size, 100) AS page_size,
    3 AS retry_attempts,
    30 AS timeout_minutes,
    NOW() AS created_at,
    NOW() AS updated_at
FROM cyberark_instances ci
CROSS JOIN (
    SELECT 'users' AS type
    UNION ALL SELECT 'safes'
    UNION ALL SELECT 'groups'
) AS sync_type
WHERE NOT EXISTS (
    SELECT 1 FROM instance_sync_configs isc 
    WHERE isc.cyberark_instance_id = ci.id 
    AND isc.sync_type = sync_type.type
    AND isc.deleted_at IS NULL
)
AND ci.deleted_at IS NULL;

-- Note: We keep the existing operations table as-is for user-initiated operations
-- Sync operations will continue to work during the transition period