import React from 'react';
import { useQuery, useMutation, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import { operationsApi, ListOperationsParams as OperationsListParams } from '@/api/operations';
import { apiClient } from '@/api/client';

// Query key factory
export const operationKeys = {
  all: ['operations'] as const,
  lists: () => [...operationKeys.all, 'list'] as const,
  list: (params: OperationsListParams) => [...operationKeys.lists(), params] as const,
  details: () => [...operationKeys.all, 'detail'] as const,
  detail: (id: string) => [...operationKeys.details(), id] as const,
};

// Hook to fetch operations list with pagination
export function useOperations(params: OperationsListParams, options?: { prefetchNext?: boolean }) {
  const queryClient = useQueryClient();
  const query = useQuery({
    queryKey: operationKeys.list(params),
    queryFn: () => operationsApi.list(params),
    placeholderData: (previousData) => previousData,
    staleTime: 0, // Always refetch when params change
  });

  // Prefetch next page
  React.useEffect(() => {
    if (options?.prefetchNext && query.data?.pagination?.has_next && params.page) {
      const nextPageParams = { ...params, page: params.page + 1 };
      queryClient.prefetchQuery({
        queryKey: operationKeys.list(nextPageParams),
        queryFn: () => operationsApi.list(nextPageParams),
        staleTime: 10 * 1000, // 10 seconds
      });
    }
  }, [queryClient, params, query.data, options?.prefetchNext]);

  return query;
}

// Hook for infinite scrolling operations
export function useInfiniteOperations(params: Omit<OperationsListParams, 'page'>) {
  return useInfiniteQuery({
    queryKey: [...operationKeys.lists(), 'infinite', params],
    queryFn: ({ pageParam = 1 }) => 
      operationsApi.list({ ...params, page: pageParam }),
    getNextPageParam: (lastPage) => {
      const { pagination } = lastPage;
      return pagination.has_next ? pagination.page + 1 : undefined;
    },
    getPreviousPageParam: (firstPage) => {
      const { pagination } = firstPage;
      return pagination.has_prev ? pagination.page - 1 : undefined;
    },
    initialPageParam: 1,
  });
}

// Hook to fetch single operation
export function useOperation(id: string) {
  return useQuery({
    queryKey: operationKeys.detail(id),
    queryFn: () => operationsApi.get(id),
    enabled: !!id,
  });
}

// Hook to cancel an operation
export function useCancelOperation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => operationsApi.cancel(id),
    onSuccess: (data, id) => {
      // Update the specific operation in cache
      queryClient.setQueryData(operationKeys.detail(id), data);
      
      // Invalidate lists to refetch
      queryClient.invalidateQueries({ queryKey: operationKeys.lists() });
    },
  });
}

// Hook to retry an operation
export function useRetryOperation() {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: (id: string) => apiClient.post(`/operations/${id}/retry`),
    onSuccess: (data, id) => {
      // Update the specific operation in cache
      queryClient.setQueryData(operationKeys.detail(id), data);
      
      // Invalidate lists to refetch
      queryClient.invalidateQueries({ queryKey: operationKeys.lists() });
    },
  });
}