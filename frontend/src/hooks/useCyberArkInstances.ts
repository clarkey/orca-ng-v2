import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { cyberarkApi, CreateCyberArkInstanceRequest as CreateInstanceRequest, UpdateCyberArkInstanceRequest as UpdateInstanceRequest } from '@/api/cyberark';

export const instanceKeys = {
  all: ['cyberark-instances'] as const,
  lists: () => [...instanceKeys.all, 'list'] as const,
  list: () => [...instanceKeys.lists()] as const,
  details: () => [...instanceKeys.all, 'detail'] as const,
  detail: (id: string) => [...instanceKeys.details(), id] as const,
};

// Hook to fetch all instances
export function useCyberArkInstances() {
  return useQuery({
    queryKey: instanceKeys.list(),
    queryFn: () => cyberarkApi.listInstances(),
  });
}

// Hook to fetch single instance
export function useCyberArkInstance(id: string) {
  return useQuery({
    queryKey: instanceKeys.detail(id),
    queryFn: () => cyberarkApi.getInstance(id),
    enabled: !!id,
  });
}

// Hook to create instance
export function useCreateCyberArkInstance() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (data: CreateInstanceRequest) => cyberarkApi.createInstance(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: instanceKeys.lists() });
    },
  });
}

// Hook to update instance
export function useUpdateCyberArkInstance() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateInstanceRequest }) => 
      cyberarkApi.updateInstance(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: instanceKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: instanceKeys.lists() });
    },
  });
}

// Hook to delete instance
export function useDeleteCyberArkInstance() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => cyberarkApi.deleteInstance(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: instanceKeys.lists() });
    },
  });
}

// Hook to test connection
export function useTestCyberArkConnection() {
  return useMutation({
    mutationFn: (id: string) => cyberarkApi.testInstanceConnection(id),
  });
}