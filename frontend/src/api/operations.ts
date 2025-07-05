import { apiClient } from './client';

export type Priority = 'high' | 'medium' | 'normal' | 'low';
export type Status = 'pending' | 'processing' | 'completed' | 'failed' | 'cancelled';
export type OperationType = 
  | 'safe_provision' 
  | 'safe_modify' 
  | 'safe_delete'
  | 'access_grant'
  | 'access_revoke'
  | 'user_sync'
  | 'safe_sync'
  | 'group_sync';

export interface Operation {
  id: string;
  type: OperationType;
  priority: Priority;
  status: Status;
  result?: any;
  error_message?: string;
  created_at: string;
  completed_at?: string;
}

export interface CreateOperationRequest {
  type: OperationType;
  priority?: Priority;
  payload: any;
  wait?: boolean;
  wait_timeout_seconds?: number;
  scheduled_at?: string;
  correlation_id?: string;
}

export interface OperationsListResponse {
  operations: Operation[];
  count: number;
}

export interface PipelineMetrics {
  queue_depth: Record<Priority, number>;
  processing_count: Record<Priority, number>;
  completed_count: Record<OperationType, number>;
  failed_count: Record<OperationType, number>;
  avg_processing_time: Record<OperationType, number>;
  worker_utilization: number;
}

export interface PipelineConfig {
  total_capacity: number;
  priority_allocation: Record<Priority, number>;
  retry_policy: {
    max_attempts: number;
    backoff_base_seconds: number;
    backoff_multiplier: number;
    backoff_jitter: boolean;
  };
  operation_timeouts: Record<OperationType, number>;
  default_timeout: number;
}

export const operationsApi = {
  // Create a new operation
  create: async (request: CreateOperationRequest): Promise<Operation> => {
    return await apiClient.post<Operation>('/operations', request);
  },

  // Get operation by ID
  get: async (id: string): Promise<Operation> => {
    return await apiClient.get<Operation>(`/operations/${id}`);
  },

  // List operations with optional filters
  list: async (params?: {
    status?: Status;
    type?: OperationType;
    priority?: Priority;
    correlation_id?: string;
  }): Promise<OperationsListResponse> => {
    return await apiClient.get<OperationsListResponse>('/operations', { params });
  },

  // Cancel an operation
  cancel: async (id: string): Promise<void> => {
    await apiClient.post<void>(`/operations/${id}/cancel`);
  },

  // Get pipeline metrics
  getMetrics: async (): Promise<PipelineMetrics> => {
    return await apiClient.get<PipelineMetrics>('/pipeline/metrics');
  },

  // Get pipeline configuration
  getConfig: async (): Promise<PipelineConfig> => {
    return await apiClient.get<PipelineConfig>('/pipeline/config');
  },

  // Update pipeline configuration (admin only)
  updateConfig: async (updates: Partial<PipelineConfig>): Promise<void> => {
    await apiClient.put<void>('/admin/pipeline/config', updates);
  },
};

// Helper function to get human-readable operation type
export function getOperationTypeLabel(type: OperationType): string {
  const labels: Record<OperationType, string> = {
    safe_provision: 'Safe Provision',
    safe_modify: 'Safe Modify',
    safe_delete: 'Safe Delete',
    access_grant: 'Grant Access',
    access_revoke: 'Revoke Access',
    user_sync: 'User Sync',
    safe_sync: 'Safe Sync',
    group_sync: 'Group Sync',
  };
  return labels[type] || type;
}

// Helper function to get priority color
export function getPriorityColor(priority: Priority): string {
  const colors: Record<Priority, string> = {
    high: 'text-red-600 bg-red-100',
    medium: 'text-yellow-600 bg-yellow-100',
    normal: 'text-blue-600 bg-blue-100',
    low: 'text-gray-600 bg-gray-100',
  };
  return colors[priority] || 'text-gray-600 bg-gray-100';
}

// Helper function to get status color
export function getStatusColor(status: Status): string {
  const colors: Record<Status, string> = {
    pending: 'text-gray-600 bg-gray-100',
    processing: 'text-blue-600 bg-blue-100',
    completed: 'text-green-600 bg-green-100',
    failed: 'text-red-600 bg-red-100',
    cancelled: 'text-orange-600 bg-orange-100',
  };
  return colors[status] || 'text-gray-600 bg-gray-100';
}