import { useEffect, useRef, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { apiClient } from '@/api/client';

export type SSEEventHandler = (event: MessageEvent) => void;

interface UseSSEOptions {
  endpoint: string;
  onMessage?: SSEEventHandler;
  onError?: (error: Event) => void;
  onOpen?: () => void;
  enabled?: boolean;
}

export function useSSE({
  endpoint,
  onMessage,
  onError,
  onOpen,
  enabled = true,
}: UseSSEOptions) {
  const eventSourceRef = useRef<EventSource | null>(null);
  const queryClient = useQueryClient();

  const connect = useCallback(() => {
    if (!enabled || eventSourceRef.current) return;

    try {
      // Get the base URL from the API client
      const baseURL = apiClient.defaults.baseURL || '';
      const url = `${baseURL}${endpoint}`;
      
      // Get auth token
      const token = localStorage.getItem('token');
      if (!token) {
        console.warn('No auth token available for SSE connection');
        return;
      }

      // Create EventSource with auth token
      const eventSource = new EventSource(`${url}?token=${encodeURIComponent(token)}`);
      eventSourceRef.current = eventSource;

      eventSource.onopen = () => {
        console.log('SSE connection opened');
        onOpen?.();
      };

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          console.log('SSE message received:', data);
          onMessage?.(event);
        } catch (error) {
          console.error('Failed to parse SSE message:', error);
        }
      };

      eventSource.onerror = (error) => {
        console.error('SSE error:', error);
        onError?.(error);
        
        // Reconnect after error
        eventSource.close();
        eventSourceRef.current = null;
        
        // Attempt to reconnect after 5 seconds
        setTimeout(() => {
          if (enabled) {
            connect();
          }
        }, 5000);
      };

      // Listen for specific event types
      eventSource.addEventListener('connected', (event) => {
        console.log('SSE connected:', event.data);
      });

      eventSource.addEventListener('heartbeat', (event) => {
        console.log('SSE heartbeat:', event.data);
      });

    } catch (error) {
      console.error('Failed to create SSE connection:', error);
    }
  }, [endpoint, onMessage, onError, onOpen, enabled]);

  const disconnect = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
      console.log('SSE connection closed');
    }
  }, []);

  useEffect(() => {
    if (enabled) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [connect, disconnect, enabled]);

  return {
    isConnected: !!eventSourceRef.current,
    reconnect: connect,
    disconnect,
  };
}

// Hook for activity stream updates
export function useActivityStream() {
  const queryClient = useQueryClient();

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      
      // Invalidate relevant queries based on event type
      switch (data.type) {
        case 'created':
        case 'updated':
        case 'completed':
        case 'failed':
        case 'cancelled':
          // Invalidate operations queries
          queryClient.invalidateQueries({ queryKey: ['operations'] });
          queryClient.invalidateQueries({ queryKey: ['activity'] });
          break;
          
        case 'sync_created':
        case 'sync_updated':
          // Invalidate sync job queries
          queryClient.invalidateQueries({ queryKey: ['syncJobs'] });
          queryClient.invalidateQueries({ queryKey: ['activity'] });
          break;
      }
    } catch (error) {
      console.error('Failed to handle activity stream message:', error);
    }
  }, [queryClient]);

  return useSSE({
    endpoint: '/activity/stream',
    onMessage: handleMessage,
  });
}

// Hook for operations stream updates
export function useOperationsStream() {
  const queryClient = useQueryClient();

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      
      // Invalidate operations queries
      queryClient.invalidateQueries({ queryKey: ['operations'] });
      
      // If we have a specific operation ID, invalidate that too
      if (data.operation?.id) {
        queryClient.invalidateQueries({ queryKey: ['operations', data.operation.id] });
      }
    } catch (error) {
      console.error('Failed to handle operations stream message:', error);
    }
  }, [queryClient]);

  return useSSE({
    endpoint: '/operations/stream',
    onMessage: handleMessage,
  });
}

// Hook for sync jobs stream updates
export function useSyncJobsStream() {
  const queryClient = useQueryClient();

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const data = JSON.parse(event.data);
      
      // Invalidate sync jobs queries
      queryClient.invalidateQueries({ queryKey: ['syncJobs'] });
      
      // If we have a specific sync job ID, invalidate that too
      if (data.sync_job?.id) {
        queryClient.invalidateQueries({ queryKey: ['syncJobs', 'detail', data.sync_job.id] });
      }
      
      // Also invalidate sync configs if job completed
      if (data.sync_job?.cyberark_instance_id && 
          (data.type === 'sync_updated' && data.sync_job.status === 'completed')) {
        queryClient.invalidateQueries({ 
          queryKey: ['syncConfigs', data.sync_job.cyberark_instance_id] 
        });
      }
    } catch (error) {
      console.error('Failed to handle sync jobs stream message:', error);
    }
  }, [queryClient]);

  return useSSE({
    endpoint: '/sync-jobs/stream',
    onMessage: handleMessage,
  });
}