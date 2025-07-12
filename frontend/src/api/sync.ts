import { apiClient } from './client';

export interface EntitySchedule {
  entityType: "users" | "groups" | "safes";
  enabled: boolean;
  interval: number; // minutes
  lastSyncAt?: string;
  lastStatus?: "success" | "failed" | "running";
  nextSyncAt: string;
  recordCount?: number;
  lastError?: string;
}

export interface SyncSchedule {
  instanceId: string;
  instanceName: string;
  enabled: boolean;
  schedules: EntitySchedule[];
  userSyncPageSize?: number;
}

export interface UpdateScheduleRequest {
  enabled?: boolean;
  schedules?: Partial<EntitySchedule>[];
}

export const syncApi = {
  // Get all sync schedules
  getSchedules: async (): Promise<SyncSchedule[]> => {
    const schedules = await apiClient.get<SyncSchedule[]>('/sync/schedules');
    return schedules || [];
  },

  // Update sync schedule for an instance
  updateSchedule: async (instanceId: string, data: UpdateScheduleRequest): Promise<void> => {
    await apiClient.put(`/instances/${instanceId}/sync-schedules`, data);
  },

  // Update individual entity schedule
  updateEntitySchedule: async (
    instanceId: string, 
    entityType: string, 
    data: Partial<EntitySchedule>
  ): Promise<void> => {
    await apiClient.put(`/instances/${instanceId}/sync-schedules/${entityType}`, data);
  },

  // Trigger immediate sync
  triggerSync: async (instanceId: string, entityType: string): Promise<void> => {
    await apiClient.post(`/instances/${instanceId}/sync-schedules/${entityType}/trigger`);
  },

  // Pause/resume instance sync
  pauseInstance: async (instanceId: string): Promise<void> => {
    await apiClient.put(`/instances/${instanceId}/sync-schedules/pause`);
  },

  resumeInstance: async (instanceId: string): Promise<void> => {
    await apiClient.put(`/instances/${instanceId}/sync-schedules/resume`);
  },

  // Global pause/resume
  pauseAll: async (): Promise<void> => {
    await apiClient.post('/sync/schedules/pause-all');
  },

  resumeAll: async (): Promise<void> => {
    await apiClient.post('/sync/schedules/resume-all');
  }
};