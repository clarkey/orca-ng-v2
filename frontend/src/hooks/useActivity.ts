import { useQuery } from '@tanstack/react-query';
import * as activityApi from '@/api/activity';

// Query keys
export const activityKeys = {
  all: ['activity'] as const,
  lists: () => [...activityKeys.all, 'list'] as const,
  list: (params?: any) => [...activityKeys.lists(), params] as const,
};

// Hook
export function useActivity(params?: {
  instance_id?: string;
  type?: 'operation' | 'sync';
  status?: string;
  limit?: number;
  offset?: number;
}) {
  return useQuery({
    queryKey: activityKeys.list(params),
    queryFn: () => activityApi.listActivity(params),
    refetchInterval: 5000, // Refetch every 5 seconds for real-time updates
  });
}