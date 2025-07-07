import { useQuery } from '@tanstack/react-query';
import { operationsApi } from '@/api/operations';

// Query keys
export const pipelineKeys = {
  all: ['pipeline'] as const,
  metrics: () => [...pipelineKeys.all, 'metrics'] as const,
  config: () => [...pipelineKeys.all, 'config'] as const,
};

// Hook to fetch pipeline metrics with auto-refresh
export function usePipelineMetrics(refetchInterval = 5000) {
  return useQuery({
    queryKey: pipelineKeys.metrics(),
    queryFn: () => operationsApi.getMetrics(),
    refetchInterval,
    refetchIntervalInBackground: true,
  });
}

// Hook to fetch pipeline config
export function usePipelineConfig() {
  return useQuery({
    queryKey: pipelineKeys.config(),
    queryFn: () => operationsApi.getConfig(),
  });
}

// Combined hook for both metrics and config
export function usePipelineData(refetchInterval = 5000) {
  const metrics = usePipelineMetrics(refetchInterval);
  const config = usePipelineConfig();
  
  return {
    metrics: metrics.data,
    config: config.data,
    isLoading: metrics.isLoading || config.isLoading,
    isError: metrics.isError || config.isError,
    error: metrics.error || config.error,
    refetch: () => {
      metrics.refetch();
      config.refetch();
    },
  };
}