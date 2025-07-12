import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import * as syncJobsApi from '@/api/syncJobs';

// Query keys
export const syncJobsKeys = {
  all: ['syncJobs'] as const,
  lists: () => [...syncJobsKeys.all, 'list'] as const,
  list: (params?: any) => [...syncJobsKeys.lists(), params] as const,
  details: () => [...syncJobsKeys.all, 'detail'] as const,
  detail: (id: string) => [...syncJobsKeys.details(), id] as const,
  configs: () => ['syncConfigs'] as const,
  config: (instanceId: string) => [...syncJobsKeys.configs(), instanceId] as const,
};

// Hooks
export function useSyncJobs(params?: {
  instance_id?: string;
  sync_type?: string;
  status?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: syncJobsKeys.list(params),
    queryFn: () => syncJobsApi.listSyncJobs(params),
  });
}

export function useSyncJob(id: string) {
  return useQuery({
    queryKey: syncJobsKeys.detail(id),
    queryFn: () => syncJobsApi.getSyncJob(id),
    enabled: !!id,
  });
}

export function useTriggerSync() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: syncJobsApi.triggerSync,
    onSuccess: () => {
      // Invalidate sync jobs list
      queryClient.invalidateQueries({ queryKey: syncJobsKeys.lists() });
      // Also invalidate activity list
      queryClient.invalidateQueries({ queryKey: ['activity'] });
    },
  });
}

export function useInstanceSyncConfigs(instanceId: string) {
  return useQuery({
    queryKey: syncJobsKeys.config(instanceId),
    queryFn: () => syncJobsApi.getInstanceSyncConfigs(instanceId),
    enabled: !!instanceId,
  });
}

export function useUpdateSyncConfig() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ 
      instanceId, 
      syncType, 
      data 
    }: { 
      instanceId: string; 
      syncType: 'users' | 'safes' | 'groups';
      data: syncJobsApi.UpdateSyncConfigRequest;
    }) => syncJobsApi.updateSyncConfig(instanceId, syncType, data),
    onSuccess: (_, variables) => {
      // Invalidate the specific instance config
      queryClient.invalidateQueries({ 
        queryKey: syncJobsKeys.config(variables.instanceId) 
      });
    },
  });
}