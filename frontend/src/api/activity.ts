import { apiClient } from './client';
import { Operation } from './operations';
import { SyncJob } from './syncJobs';

export interface ActivityItem {
  id: string;
  type: 'operation' | 'sync';
  status: string;
  title: string;
  subtitle: string;
  instance?: {
    id: string;
    name: string;
  };
  created_by?: {
    id: string;
    username: string;
  };
  created_at: string;
  started_at?: string;
  completed_at?: string;
  duration_seconds?: number;
  error?: string;
  operation?: Operation;
  sync_job?: SyncJob;
}

export interface ActivityResponse {
  activities: ActivityItem[];
  total: number;
  limit: number;
  offset: number;
}

export async function listActivity(params?: {
  instance_id?: string;
  type?: 'operation' | 'sync';
  status?: string;
  limit?: number;
  offset?: number;
}): Promise<ActivityResponse> {
  const response = await apiClient.get('/activity', { params });
  return response.data;
}