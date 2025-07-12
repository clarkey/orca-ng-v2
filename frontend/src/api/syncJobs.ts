import { apiClient } from './client';

// Types
export interface SyncJob {
  id: string;
  cyberark_instance_id: string;
  sync_type: 'users' | 'safes' | 'groups';
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  triggered_by: 'manual' | 'scheduled' | 'system';
  started_at?: string;
  completed_at?: string;
  next_run_at?: string;
  records_synced: number;
  records_created: number;
  records_updated: number;
  records_deleted: number;
  records_failed: number;
  error_message?: string;
  error_details?: string;
  duration_seconds?: number;
  created_at: string;
  updated_at: string;
  cyberark_instance?: {
    id: string;
    name: string;
  };
  created_by_user?: {
    id: string;
    username: string;
  };
}

export interface SyncJobsResponse {
  sync_jobs: SyncJob[];
  total: number;
  limit: number;
  offset: number;
}

export interface InstanceSyncConfig {
  id: string;
  cyberark_instance_id: string;
  sync_type: 'users' | 'safes' | 'groups';
  enabled: boolean;
  interval_minutes: number;
  page_size: number;
  retry_attempts: number;
  timeout_minutes: number;
  last_run_at?: string;
  last_run_status?: string;
  last_run_message?: string;
  next_run_at?: string;
  created_at: string;
  updated_at: string;
}

export interface InstanceSyncConfigsResponse {
  instance_id: string;
  configs: {
    users?: InstanceSyncConfig;
    safes?: InstanceSyncConfig;
    groups?: InstanceSyncConfig;
  };
}

export interface TriggerSyncRequest {
  instance_id: string;
  sync_type: 'users' | 'safes' | 'groups';
}

export interface UpdateSyncConfigRequest {
  enabled?: boolean;
  interval_minutes?: number;
  page_size?: number;
  retry_attempts?: number;
  timeout_minutes?: number;
}

// API functions

// Get sync jobs for a specific instance
export async function listSyncJobsForInstance(
  instanceId: string,
  params?: {
    sync_type?: string;
    status?: string;
    limit?: number;
    offset?: number;
  }
): Promise<SyncJobsResponse> {
  const response = await apiClient.get(`/instances/${instanceId}/sync-jobs`, { params });
  return response.data;
}

// Get a specific sync job by ID
export async function getSyncJob(id: string): Promise<SyncJob> {
  const response = await apiClient.get(`/sync-jobs/${id}`);
  return response.data;
}

// Trigger sync for a specific instance
export async function triggerSyncForInstance(
  instanceId: string,
  syncType: 'users' | 'safes' | 'groups'
): Promise<{ message: string; job_id: string; job: SyncJob }> {
  const response = await apiClient.post(`/instances/${instanceId}/sync-jobs/trigger`, { sync_type: syncType });
  return response.data;
}

// Legacy functions for backward compatibility
export async function listSyncJobs(params?: {
  instance_id?: string;
  sync_type?: string;
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<SyncJobsResponse> {
  // If instance_id is provided, use the new instance-specific endpoint
  if (params?.instance_id) {
    const { instance_id, ...restParams } = params;
    return listSyncJobsForInstance(instance_id, restParams);
  }
  // Otherwise, this would need a global endpoint which we don't have anymore
  throw new Error('Global sync jobs listing is no longer supported. Please provide an instance_id.');
}

export async function triggerSync(data: TriggerSyncRequest): Promise<{ message: string; job_id: string; job: SyncJob }> {
  return triggerSyncForInstance(data.instance_id, data.sync_type);
}

export async function getInstanceSyncConfigs(instanceId: string): Promise<InstanceSyncConfigsResponse> {
  const response = await apiClient.get(`/instances/${instanceId}/sync-configs`);
  return response.data;
}

export async function updateSyncConfig(
  instanceId: string, 
  syncType: 'users' | 'safes' | 'groups',
  data: UpdateSyncConfigRequest
): Promise<InstanceSyncConfig> {
  const response = await apiClient.patch(`/instances/${instanceId}/sync-configs/${syncType}`, data);
  return response.data;
}