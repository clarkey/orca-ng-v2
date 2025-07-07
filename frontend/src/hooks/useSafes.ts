import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient, CreateSafeRequest, UpdateSafeRequest } from '@/api/client';

export const safeKeys = {
  all: ['safes'] as const,
  lists: () => [...safeKeys.all, 'list'] as const,
  list: (filters?: any) => [...safeKeys.lists(), filters] as const,
  details: () => [...safeKeys.all, 'detail'] as const,
  detail: (id: string) => [...safeKeys.details(), id] as const,
};

// Hook to fetch all safes
export function useSafes(filters?: any) {
  return useQuery({
    queryKey: safeKeys.list(filters),
    queryFn: () => apiClient.getSafes(filters),
  });
}

// Hook to fetch single safe
export function useSafe(id: string) {
  return useQuery({
    queryKey: safeKeys.detail(id),
    queryFn: () => apiClient.getSafe(id),
    enabled: !!id,
  });
}

// Hook to create safe
export function useCreateSafe() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateSafeRequest) => apiClient.createSafe(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: safeKeys.lists() });
    },
  });
}

// Hook to update safe
export function useUpdateSafe() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSafeRequest }) => 
      apiClient.updateSafe(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: safeKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: safeKeys.lists() });
    },
  });
}

// Hook to delete safe
export function useDeleteSafe() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => apiClient.deleteSafe(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: safeKeys.lists() });
    },
  });
}